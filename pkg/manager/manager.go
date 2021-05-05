package manager

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/harvester/support-bundle-utils/pkg/manager/client"
	"github.com/harvester/support-bundle-utils/pkg/utils"
)

type SupportBundleManager struct {
	HarvesterNamespace string
	HarvesterVersion   string
	BundleName         string
	BundleFileName     string
	bundleFileSize     int64
	OutputDir          string
	WaitTimeout        time.Duration
	LonghornAPI        string
	ManagerPodIP       string
	Standalone         bool
	ImageName          string
	ImagePullPolicy    string

	context context.Context

	restConfig *rest.Config
	k8s        *client.KubernetesClient
	k8sMetrics *client.MetricsClient
	harvester  *client.HarvesterClient

	state StateStoreInterface

	ch            chan struct{}
	done          bool
	lock          sync.Mutex
	expectedNodes map[string]string
}

func (m *SupportBundleManager) check() error {
	if m.HarvesterNamespace == "" {
		return errors.New("namespace is not specified")
	}
	if m.BundleName == "" {
		return errors.New("support bundle name is not specified")
	}
	if m.ManagerPodIP == "" {
		return errors.New("manager pod IP is not specified")
	}
	if m.ImageName == "" {
		return errors.New("image name is not specified")
	}
	if m.ImagePullPolicy == "" {
		return errors.New("image pull policy is not specified")
	}
	if m.OutputDir == "" {
		m.OutputDir = filepath.Join(os.TempDir(), "harvester-support-bundle")
	}
	if err := os.MkdirAll(m.getWorkingDir(), os.FileMode(0755)); err != nil {
		return err
	}
	return nil
}

func (m *SupportBundleManager) getWorkingDir() string {
	return filepath.Join(m.OutputDir, "bundle")
}

func (m *SupportBundleManager) getBundlefile() string {
	return filepath.Join(m.OutputDir, m.BundleFileName)
}

func (m *SupportBundleManager) getBundlefilesize() (int64, error) {
	finfo, err := os.Stat(m.getBundlefile())
	if err != nil {
		return 0, err
	}
	return finfo.Size(), nil
}

func (m *SupportBundleManager) Run() error {
	if err := m.check(); err != nil {
		return err
	}

	m.context = context.Background()
	err := m.initClients()
	if err != nil {
		return err
	}

	m.initStateStore()

	state, err := m.state.GetState(m.HarvesterNamespace, m.BundleName)
	if err != nil {
		return err
	}
	if state != StateGenerating {
		return fmt.Errorf("invalid start state %s", state)
	}

	cluster := NewCluster(m.context, m)
	bundleName, err := cluster.GenerateClusterBundle(m.getWorkingDir())
	if err != nil {
		wErr := errors.Wrap(err, "fail to generate cluster bundle")
		if e := m.state.SetError(m.HarvesterNamespace, m.BundleName, wErr); e != nil {
			return e
		}
		return wErr
	}
	m.BundleFileName = bundleName

	err = m.waitNodeBundles()
	if err != nil {
		// Ignore error here, since in some failure cases we might not receive all node bundles.
		// A support bundle with partital data is also useful.
		logrus.Error(err)
	}

	err = m.compressBundle()
	if err != nil {
		if e := m.state.SetError(m.HarvesterNamespace, m.BundleName, err); e != nil {
			return e
		}
		return err
	}

	err = m.state.Done(m.HarvesterNamespace, m.BundleName, m.BundleFileName, m.bundleFileSize)
	if err != nil {
		return err
	}

	logrus.Infof("support bundle %s ready for downloading", m.getBundlefile())
	select {}
}

func (m *SupportBundleManager) initClients() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	m.restConfig = config
	hvst, err := client.NewHarvesterClient(m.context, m.restConfig)
	if err != nil {
		return err
	}
	m.harvester = hvst

	k8s, err := client.NewKubernetesClient(m.context, m.restConfig)
	if err != nil {
		return err
	}
	m.k8s = k8s

	k8sMetrics, err := client.NewMetricsClient(m.context, m.restConfig)
	if err != nil {
		return err
	}
	m.k8sMetrics = k8sMetrics
	return nil
}

func (m *SupportBundleManager) initStateStore() {
	if m.Standalone {
		m.state = NewLocalStore(m.HarvesterNamespace, m.BundleName)
		return
	}
	m.state = NewK8sStore(m.harvester)
}

func (m *SupportBundleManager) waitNodeBundles() error {
	m.ch = make(chan struct{})

	err := m.refreshHarvesterNodes()
	if err != nil {
		return err
	}
	logrus.Debugf("expected bundles from nodes: %+v", m.expectedNodes)

	// create a http server to receive node bundles
	s := HttpServer{context: m.context}
	go s.Run(m)

	// create a daemonset to collect node bundles and push back
	agents := &AgentDaemonSet{sbm: m}
	err = agents.Create(m.ImageName, fmt.Sprintf("http://%s:8080", m.ManagerPodIP))
	if err != nil {
		return err
	}

	logrus.Infof("wating node bundles, (timeout: %s)", m.WaitTimeout)
	select {
	case <-m.ch:
		logrus.Info("all node bundles are received.")

		// Clean up when everything is fine, leave ds there for debugging.
		// The ds will be garbage-collected when manager pod is gone
		err := agents.Cleanup()
		if err != nil {
			return errors.Wrap(err, "fail to cleanup agent daemonset")
		}
		return nil
	case <-time.After(m.WaitTimeout):
		return fmt.Errorf("fail to wait node bundles. missing: %+v", m.expectedNodes)
	}
}

func (m *SupportBundleManager) getBundle(w http.ResponseWriter, req *http.Request) {
	bundleFile := m.getBundlefile()
	f, err := os.Open(bundleFile)
	if err != nil {
		e := errors.Wrap(err, "fail to open bundle file")
		logrus.Error(e)
		utils.HttpResponseError(w, http.StatusNotFound, e)
		return
	}
	defer f.Close()

	fstat, err := f.Stat()
	if err != nil {
		e := errors.Wrap(err, "fail to stat bundle file")
		logrus.Error(e)
		utils.HttpResponseError(w, http.StatusNotFound, e)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Length", strconv.FormatInt(fstat.Size(), 10))
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(bundleFile))
	if _, err := io.Copy(w, f); err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, err)
		return
	}
}

func (m *SupportBundleManager) createNodeBundle(w http.ResponseWriter, req *http.Request) {
	node := mux.Vars(req)["nodeName"]
	if node == "" {
		utils.HttpResponseError(w, http.StatusBadRequest, errors.New("empty node name"))
		return
	}

	logrus.Debugf("handle create node bundle for %s", node)
	nodesDir := filepath.Join(m.getWorkingDir(), "nodes")
	err := os.MkdirAll(nodesDir, os.FileMode(0775))
	if err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, fmt.Errorf("fail to create directory %s: %s", nodesDir, err))
		return
	}

	nodeBundle := filepath.Join(nodesDir, node+".zip")
	f, err := os.Create(nodeBundle)
	if err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, fmt.Errorf("fail to create file %s: %s", nodeBundle, err))
		return
	}
	defer f.Close()
	_, err = io.Copy(f, req.Body)
	if err != nil {
		utils.HttpResponseError(w, http.StatusInternalServerError, err)
		return
	}

	err = m.verifyNodeBundle(nodeBundle)
	if err != nil {
		logrus.Errorf("fail to verify file %s: %s", nodeBundle, err)
		utils.HttpResponseError(w, http.StatusBadRequest, err)
		return
	}
	m.completeNode(node)
	utils.HttpResponseStatus(w, http.StatusCreated)
}

func (m *SupportBundleManager) verifyNodeBundle(file string) error {
	_, err := zip.OpenReader(file)
	return err
}

func (m *SupportBundleManager) completeNode(node string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.expectedNodes[node]
	if ok {
		logrus.Debugf("complete node %s", node)
		delete(m.expectedNodes, node)
	} else {
		logrus.Warnf("complete an unknown node %s", node)
	}

	if len(m.expectedNodes) == 0 {
		if !m.done {
			logrus.Debugf("all nodes are completed.")
			m.ch <- struct{}{}
			m.done = true
		}
	}
}

func (m *SupportBundleManager) compressBundle() error {
	bundleDir := strings.TrimSuffix(m.BundleFileName, filepath.Ext(m.getBundlefile()))
	bundleDirPath := filepath.Join(m.OutputDir, bundleDir)
	err := os.Rename(m.getWorkingDir(), bundleDirPath)
	if err != nil {
		return errors.Wrap(err, "fail to compress bundle")
	}
	cmd := exec.Command("zip", "-r", m.getBundlefile(), bundleDir)
	cmd.Dir = m.OutputDir
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "fail to compress bundle")
	}

	size, err := m.getBundlefilesize()
	if err != nil {
		return errors.Wrap(err, "fail to get bundle file size")
	}
	m.bundleFileSize = size
	return nil
}

func (m *SupportBundleManager) refreshHarvesterNodes() error {
	nodes, err := m.k8s.GetNodesListByLabels(fmt.Sprintf("%s=%s", HarvesterNodeLabelKey, HarvesterNodeLabelValue))
	if err != nil {
		return err
	}

	if len(nodes.Items) == 0 {
		return errors.New("no Harvester nodes are found")
	}

	m.expectedNodes = make(map[string]string)
	for _, node := range nodes.Items {
		m.expectedNodes[node.Name] = ""
	}
	return nil
}

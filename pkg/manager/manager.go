package manager

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/harvester/harvester/pkg/controller/master/supportbundle/types"
	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/harvester/support-bundle-utils/pkg/manager/client"
)

type SupportBundleManager struct {
	HarvesterNamespace string
	BundleName         string
	bundleFileName     string
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

	state  StateStoreInterface
	status ManagerStatus

	ch            chan struct{}
	done          bool
	nodesLock     sync.Mutex
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
	return filepath.Join(m.OutputDir, m.bundleFileName)
}

func (m *SupportBundleManager) getBundlefilesize() (int64, error) {
	finfo, err := os.Stat(m.getBundlefile())
	if err != nil {
		return 0, err
	}
	return finfo.Size(), nil
}

func (m *SupportBundleManager) Run() error {
	phases := []struct {
		Name string
		Run  func() error
	}{
		{
			PhaseInit,
			m.phaseInit,
		},
		{
			PhaseClusterBundle,
			m.phaseCollectClusterBundle,
		},
		{
			PhaseNodeBundle,
			m.phaseCollectNodeBundles,
		},
		{
			PhasePackaging,
			m.phasePackaging,
		},
		{
			PhaseDone,
			m.phaseDone,
		},
	}

	for i, phase := range phases {
		logrus.Infof("running phase %s", phase.Name)
		m.status.SetPhase(phase.Name)
		if err := phase.Run(); err != nil {
			m.status.SetError(err.Error())
			logrus.Errorf("fail to run phase %s: %s", phase.Name, err.Error())
			break
		}

		progress := 100 * (i + 1) / len(phases)
		m.status.SetProgress(progress)
		logrus.Infof("succeed to run phase %s. Progress (%d).", phase.Name, progress)
	}

	<-m.context.Done()
	return nil
}

func (m *SupportBundleManager) phaseInit() error {
	if err := m.check(); err != nil {
		return err
	}

	m.context = signals.SetupSignalHandler(context.Background())
	err := m.initClients()
	if err != nil {
		return err
	}
	m.initStateStore()

	state, err := m.state.GetState(m.HarvesterNamespace, m.BundleName)
	if err != nil {
		return err
	}
	if state != types.StateGenerating {
		return fmt.Errorf("invalid start state %s", state)
	}

	// create a http server to
	// (1) provide status to controller
	// (2) accept node bundles from agent daemonset
	s := HttpServer{
		context: m.context,
		manager: m,
	}

	go s.Run(m)

	return nil
}

func (m *SupportBundleManager) phaseCollectClusterBundle() error {
	cluster := NewCluster(m.context, m)
	bundleName, err := cluster.GenerateClusterBundle(m.getWorkingDir())
	if err != nil {
		return errors.Wrap(err, "fail to generate cluster bundle")
	}
	m.bundleFileName = bundleName
	return nil
}

func (m *SupportBundleManager) phaseCollectNodeBundles() error {
	err := m.collectNodeBundles()
	if err != nil {
		// Ignore error here, since in some failure cases we might not receive all node bundles.
		// A support bundle with partital data is also useful.
		logrus.Error(err)
	}
	return nil
}

func (m *SupportBundleManager) phasePackaging() error {
	return m.compressBundle()
}

func (m *SupportBundleManager) phaseDone() error {
	logrus.Infof("support bundle %s ready to download", m.getBundlefile())
	return nil
}

func (m *SupportBundleManager) initClients() error {
	var err error
	m.restConfig, err = rest.InClusterConfig()
	if err != nil {
		return err
	}

	m.harvester, err = client.NewHarvesterClient(m.context, m.restConfig)
	if err != nil {
		return err
	}

	m.k8s, err = client.NewKubernetesClient(m.context, m.restConfig)
	if err != nil {
		return err
	}

	m.k8sMetrics, err = client.NewMetricsClient(m.context, m.restConfig)
	if err != nil {
		return err
	}
	return nil
}

func (m *SupportBundleManager) initStateStore() {
	if m.Standalone {
		m.state = NewLocalStore(m.HarvesterNamespace, m.BundleName)
		return
	}
	m.state = NewK8sStore(m.harvester)
}

// collectNodeBundles spawns a daemonset on each node and waits for agents on
// each node to push node bundles
func (m *SupportBundleManager) collectNodeBundles() error {
	m.ch = make(chan struct{})

	err := m.refreshHarvesterNodes()
	if err != nil {
		return err
	}
	logrus.Debugf("expected bundles from nodes: %+v", m.expectedNodes)

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

		// Clean up when everything is fine. If something went wrong, keep ds for debugging.
		// The ds will be garbage-collected when manager pod is gone.
		err := agents.Cleanup()
		if err != nil {
			return errors.Wrap(err, "fail to cleanup agent daemonset")
		}
		return nil
	case <-time.After(m.WaitTimeout):
		return fmt.Errorf("fail to wait node bundles. missing: %+v", m.expectedNodes)
	}
}

func (m *SupportBundleManager) verifyNodeBundle(file string) error {
	_, err := zip.OpenReader(file)
	return err
}

func (m *SupportBundleManager) completeNode(node string) {
	m.nodesLock.Lock()
	defer m.nodesLock.Unlock()

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
	bundleDir := strings.TrimSuffix(m.bundleFileName, filepath.Ext(m.getBundlefile()))
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
	m.status.SetFileinfo(m.bundleFileName, size)
	return nil
}

func (m *SupportBundleManager) refreshHarvesterNodes() error {
	nodes, err := m.k8s.GetNodesListByLabels(fmt.Sprintf("%s=%s", types.HarvesterNodeLabelKey, types.HarvesterNodeLabelValue))
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

package manager

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	"github.com/rancher/support-bundle-kit/pkg/manager/client"
	"github.com/rancher/support-bundle-kit/pkg/types"
	"github.com/rancher/support-bundle-kit/pkg/utils"
)

type SupportBundleManager struct {
	Namespaces      []string
	BundleName      string
	bundleFileName  string
	OutputDir       string
	WaitTimeout     time.Duration
	ManagerPodIP    string
	Standalone      bool
	ImageName       string
	ImagePullPolicy string
	KubeConfig      string
	PodNamespace    string
	NodeSelector    string
	TaintToleration string
	RegistrySecret  string
	IssueURL        string
	Description     string
	NodeTimeout     time.Duration

	ExcludeResources    []schema.GroupResource
	ExcludeResourceList []string
	BundleCollectors    []string
	SpecifyCollector    string

	context context.Context

	restConfig *rest.Config
	k8s        *client.KubernetesClient
	k8sMetrics *client.MetricsClient
	discovery  *client.DiscoveryClient

	state  StateStoreInterface
	status ManagerStatus

	ch            chan struct{}
	done          bool
	nodesLock     sync.Mutex
	expectedNodes map[string]string
}

func (m *SupportBundleManager) check() error {
	if len(m.Namespaces) == 0 || len(m.Namespaces[0]) == 0 {
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
		m.OutputDir = filepath.Join(os.TempDir(), "support-bundle-kit")
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
		Name types.ManagerPhase
		Run  func() error
	}{
		{
			types.ManagerPhaseInit,
			m.phaseInit,
		},
		{
			types.ManagerPhaseClusterBundle,
			m.phaseCollectClusterBundle,
		},
		{
			types.ManagerPhasePrometheusBundle,
			m.phaseCollectPrometheusBundle,
		},
		{
			types.ManagerPhaseNodeBundle,
			m.phaseCollectNodeBundles,
		},
		{
			types.ManagerPhasePackaging,
			m.phasePackaging,
		},
		{
			types.ManagerPhaseDone,
			m.phaseDone,
		},
	}

	for i, phase := range phases {
		logrus.Infof("Running phase %s", phase.Name)
		m.status.SetPhase(phase.Name)
		if err := phase.Run(); err != nil {
			m.status.SetError(err.Error())
			logrus.Errorf("Failed to run phase %s: %s", phase.Name, err.Error())
			break
		}

		progress := 100 * (i + 1) / len(phases)
		m.status.SetProgress(progress)
		logrus.Infof("Succeed to run phase %s. Progress (%d).", phase.Name, progress)
	}

	<-m.context.Done()
	return nil
}

func (m *SupportBundleManager) phaseInit() error {
	// Init default collector
	m.BundleCollectors = append(m.BundleCollectors, "cluster", "default")
	m.ExcludeResources = []schema.GroupResource{
		// Default exclusion
		{Group: v1.GroupName, Resource: "secrets"},
	}
	for _, res := range m.ExcludeResourceList {
		gr := schema.ParseGroupResource(res)
		if !gr.Empty() {
			m.ExcludeResources = append(m.ExcludeResources, gr)
		}
	}
	if err := m.check(); err != nil {
		return err
	}

	m.context = signals.SetupSignalContext()
	err := m.initClients()
	if err != nil {
		return err
	}

	m.PodNamespace = utils.PodNamespace()

	m.initStateStore()

	state, err := m.state.GetState(m.PodNamespace, m.BundleName)
	if err != nil {
		return err
	}
	if state != types.SupportBundleStateGenerating {
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

func (m *SupportBundleManager) phaseCollectPrometheusBundle() error {
	pods, err := m.k8s.GetPodsListByLabels("cattle-monitoring-system", "app.kubernetes.io/name=prometheus")
	if err != nil {
		if apierrors.IsNotFound(err) {
			logrus.Info("prometheus pods not found")
			return nil
		}

		return errors.Wrap(err, "failed to get prometheus pods")
	}

	if len(pods.Items) == 0 {
		logrus.Info("prometheus pods not found")
		return nil
	}

	if len(pods.Items) > 1 {
		return fmt.Errorf("multiple %d prometheus pods found", len(pods.Items))
	}

	targetPod := pods.Items[0]
	p, err := utils.NewPrometheus(targetPod.Status.PodIP)
	if err != nil {
		logrus.Debugf("host: %s, port: %d", targetPod.Status.PodIP, utils.PrometheusPort)
		return errors.Wrap(err, "failed to new prometheus")
	}

	alerts, err := p.GetAlerts(m.context)
	if err != nil {
		return errors.Wrap(err, "failed to get prometheus alert")
	}

	b, err := json.MarshalIndent(alerts, "", "\t")
	if err != nil {
		return errors.Wrap(err, "failed to marshal prometheus alert")
	}

	if err := os.WriteFile(fmt.Sprintf("%s/prometheus-alerts.json", m.getWorkingDir()), b, 0644); err != nil {
		return errors.Wrap(err, "failed to write prometheus alert")
	}

	return nil
}

func (m *SupportBundleManager) phaseCollectNodeBundles() error {
	err := m.collectNodeBundles()
	if err != nil {
		// Ignore error here, since in some failure cases we might not receive all node bundles.
		// A support bundle with partital data is also useful.
		logrus.WithError(err).Error("Failed to collect node bundles")
	}
	return nil
}

func (m *SupportBundleManager) phasePackaging() error {
	return m.compressBundle()
}

func (m *SupportBundleManager) phaseDone() error {
	logrus.Infof("Support bundle %s ready to download", m.getBundlefile())
	return nil
}

func (m *SupportBundleManager) initClients() error {
	var err error
	m.restConfig, err = rest.InClusterConfig()
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

	m.discovery, err = client.NewDiscoveryClient(m.context, m.restConfig)
	if err != nil {
		return err
	}
	return nil
}

func (m *SupportBundleManager) initStateStore() {
	m.state = NewLocalStore(m.PodNamespace, m.BundleName)
}

// collectNodeBundles spawns a daemonset on each node and waits for agents on
// each node to push node bundles
func (m *SupportBundleManager) collectNodeBundles() error {
	m.ch = make(chan struct{})

	// create a daemonset to collect node bundles and push back
	agents := &AgentDaemonSet{sbm: m}
	agentDaemonSet, err := agents.Create(m.ImageName, fmt.Sprintf("http://%s:8080", m.ManagerPodIP))
	if err != nil {
		return err
	}

	err = m.refreshNodes(agentDaemonSet)
	if err != nil {
		return err
	}

	m.waitNodesCompleted()

	// Clean up when everything is fine. If something went wrong, keep ds for debugging.
	// The ds will be garbage-collected when manager pod is gone.
	err = agents.Cleanup()
	if err != nil {
		return errors.Wrap(err, "fail to cleanup agent daemonset")
	}
	return nil
}

func (m *SupportBundleManager) verifyNodeBundle(file string) error {
	f, err := zip.OpenReader(file)
	if err == nil {
		_ = f.Close()
	}
	return err
}

func (m *SupportBundleManager) printTimeoutNodes() {
	for node := range m.expectedNodes {
		logrus.Warnf("Collection timed out for node: %s", node)
	}
}

func (m *SupportBundleManager) waitNodesCompleted() {
	select {
	case <-m.ch:
		logrus.Info("All node bundles are received.")
	case <-m.timeout():
		logrus.Info("Some nodes are timeout, not all node bundles are received.")
		m.printTimeoutNodes()
	}
}

func (m *SupportBundleManager) timeout() <-chan time.Time {
	if m.NodeTimeout == 0 {
		return time.After(30 * time.Minute) // default time out
	}

	return time.After(m.NodeTimeout)
}

func (m *SupportBundleManager) completeNode(node string) {
	m.nodesLock.Lock()
	defer m.nodesLock.Unlock()

	_, ok := m.expectedNodes[node]
	if ok {
		logrus.Debugf("Complete node %s", node)
		delete(m.expectedNodes, node)
	} else {
		logrus.Warnf("Complete an unknown node %s", node)
	}

	if len(m.expectedNodes) == 0 {
		if !m.done {
			logrus.Debugf("All nodes are completed")
			close(m.ch)
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

func (m *SupportBundleManager) getAgentPodsCreatedBy(daemonSet *appsv1.DaemonSet) (*v1.PodList, error) {
	startTime := time.Now()
	ticker := time.NewTicker(types.PodCreationWaitInterval)
	defer ticker.Stop()

	for range ticker.C {
		logrus.Debug("Waiting for the creation of agent DaemonSet Pods for scheduled node names collection")

		pods, err := m.k8s.GetPodsListByLabels(m.PodNamespace, fmt.Sprintf("app=%s", types.SupportBundleAgent))
		if err != nil {
			return nil, err
		}

		// Filter out pods not created by the current agent DaemonSet or without assigned node names
		filteredPods := make([]v1.Pod, 0, len(pods.Items))
		for _, pod := range pods.Items {
			if len(pod.OwnerReferences) != 1 {
				return nil, fmt.Errorf("unexpected OwnerReferences in %v: %+v", pod.Name, pod.OwnerReferences)
			}

			if pod.OwnerReferences[0].Name == daemonSet.Name && pod.Spec.NodeName != "" {
				filteredPods = append(filteredPods, pod)
			}
		}

		// Get the latest agent DaemonSet status
		daemonSet, err = m.k8s.GetDaemonSetBy(daemonSet.Namespace, daemonSet.Name)
		if err != nil {
			return nil, err
		}

		// Check if all desired Pods have been scheduled
		if len(filteredPods) != 0 && len(filteredPods) == int(daemonSet.Status.DesiredNumberScheduled) {
			return &v1.PodList{Items: filteredPods}, nil
		}

		if time.Since(startTime) > types.PodCreationTimeout {
			return nil, fmt.Errorf("timed out (%d) waiting for the agent DaemonSet Pods to be scheduled", types.PodCreationTimeout)
		}
	}

	return nil, fmt.Errorf("unexpected error: stopped waiting for creating DaemonSet Pod or timing out")
}

func (m *SupportBundleManager) getAgentNodesIn(podList *v1.PodList) ([]*v1.Node, error) {
	var nodes []*v1.Node
	for _, pod := range podList.Items {
		node, err := m.k8s.GetNodeBy(pod.Spec.NodeName)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (m *SupportBundleManager) refreshNodes(agentDaemonSet *appsv1.DaemonSet) error {
	m.nodesLock.Lock()
	defer m.nodesLock.Unlock()

	podList, err := m.getAgentPodsCreatedBy(agentDaemonSet)
	if err != nil {
		return err
	}

	nodes, err := m.getAgentNodesIn(podList)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("no nodes are found")
	}

	m.expectedNodes = make(map[string]string)
	defer logrus.Debugf("Expecting bundles from nodes: %+v", m.expectedNodes)

NODE_LOOP:
	for _, node := range nodes {
		for _, cond := range node.Status.Conditions {
			switch cond.Type {
			case v1.NodeReady:
				if cond.Status != v1.ConditionTrue {
					continue NODE_LOOP
				}
			case v1.NodeNetworkUnavailable:
				if cond.Status == v1.ConditionTrue {
					continue NODE_LOOP
				}
			}
		}
		m.expectedNodes[node.Name] = ""
	}

	return nil
}

func (m *SupportBundleManager) getNodeSelector() map[string]string {
	nodeSelector := map[string]string{}
	if m.NodeSelector != "" {
		// parse key1=value1,key2=value2,...
		for _, s := range strings.Split(m.NodeSelector, ",") {
			kv := strings.Split(s, "=")
			if len(kv) != 2 {
				logrus.Warnf("Unable to parse %s", s)
				continue
			}
			nodeSelector[kv[0]] = kv[1]
		}
	}
	return nodeSelector
}

func (m *SupportBundleManager) getTaintToleration() []v1.Toleration {
	taintToleration := []v1.Toleration{}

	m.TaintToleration = strings.ReplaceAll(m.TaintToleration, " ", "")
	if m.TaintToleration == "" {
		return taintToleration
	}

	tolerationList := strings.Split(m.TaintToleration, ",")
	for _, toleration := range tolerationList {
		toleration, err := parseToleration(toleration)
		if err != nil {
			logrus.WithError(err).Warnf("Invalid toleration: %s", toleration)
			continue
		}
		taintToleration = append(taintToleration, *toleration)
	}
	return taintToleration
}

func parseToleration(taintToleration string) (*v1.Toleration, error) {
	// The schema should be `key=value:effect` or `key:effect`
	parts := strings.Split(taintToleration, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("missing key/value and effect pair")
	}

	// parse `key=value` or `key`
	key, value, operator := "", "", v1.TolerationOperator("")
	pair := strings.Split(parts[0], "=")
	switch len(pair) {
	case 1:
		key, value, operator = parts[0], "", v1.TolerationOpExists
	case 2:
		key, value, operator = pair[0], pair[1], v1.TolerationOpEqual
	}

	effect := v1.TaintEffect(parts[1])
	switch effect {
	case "", v1.TaintEffectNoExecute, v1.TaintEffectNoSchedule, v1.TaintEffectPreferNoSchedule:
	default:
		return nil, fmt.Errorf("invalid effect: %v", parts[1])
	}

	return &v1.Toleration{
		Key:      key,
		Value:    value,
		Operator: operator,
		Effect:   effect,
	}, nil
}

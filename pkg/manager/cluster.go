package manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/harvester/support-bundle-utils/pkg/manager/external"
	"github.com/harvester/support-bundle-utils/pkg/utils"
)

type Cluster struct {
	sbm *SupportBundleManager
}

func NewCluster(ctx context.Context, sbm *SupportBundleManager) *Cluster {
	return &Cluster{
		sbm: sbm,
	}
}

func (c *Cluster) GenerateClusterBundle(bundleDir string) (string, error) {
	logrus.Debug("generating cluster bundle...")
	namespace, err := c.sbm.k8s.GetNamespace(c.sbm.HarvesterNamespace)
	if err != nil {
		return "", errors.Wrap(err, "cannot get harvester namespace")
	}
	kubeVersion, err := c.sbm.k8s.GetKubernetesVersion()
	if err != nil {
		return "", errors.Wrap(err, "cannot get kubernetes version")
	}
	sb, err := c.sbm.harvester.GetSupportBundle(c.sbm.HarvesterNamespace, c.sbm.BundleName)
	if err != nil {
		return "", errors.Wrap(err, "cannot get support bundle")
	}

	bundleMeta := &BundleMeta{
		ProjectName:          "Harvester",
		ProjectVersion:       c.sbm.HarvesterVersion,
		BundleVersion:        BundleVersion,
		KubernetesVersion:    kubeVersion.GitVersion,
		ProjectNamespaceUUID: string(namespace.UID),
		BundleCreatedAt:      utils.Now(),
		IssueURL:             sb.Spec.IssueURL,
		IssueDescription:     sb.Spec.Description,
	}

	bundleName := fmt.Sprintf("harvester-supportbundle_%s_%s.zip",
		bundleMeta.ProjectNamespaceUUID,
		strings.Replace(bundleMeta.BundleCreatedAt, ":", "-", -1))

	errLog, err := os.Create(filepath.Join(bundleDir, "bundleGenerationError.log"))
	if err != nil {
		logrus.Errorf("Failed to create bundle generation log")
		return "", err
	}
	defer errLog.Close()

	metaFile := filepath.Join(bundleDir, "metadata.yaml")
	encodeToYAMLFile(bundleMeta, metaFile, errLog)

	yamlsDir := filepath.Join(bundleDir, "yamls")
	c.generateSupportBundleYAMLs(yamlsDir, errLog)

	logsDir := filepath.Join(bundleDir, "logs")
	c.generateSupportBundleLogs(logsDir, errLog)

	externalDir := filepath.Join(bundleDir, "external")
	c.getExternalSupportBundles(bundleMeta, externalDir, errLog)

	return bundleName, nil
}

func (c *Cluster) generateSupportBundleYAMLs(yamlsDir string, errLog io.Writer) {
	// Cluster scope
	globalDir := filepath.Join(yamlsDir, "cluster")
	c.generateKubernetesClusterYAMLs(globalDir, errLog)
	c.generateHarvesterClusterYAMLs(globalDir, errLog)

	// Namespaced scope: k8s resources
	namespaces := []string{"default", "kube-system", "cattle-system", c.sbm.HarvesterNamespace}
	for _, namespace := range namespaces {
		namespacedDir := filepath.Join(yamlsDir, "namespaced", namespace)
		c.generateKubernetesNamespacedYAMLs(namespace, namespacedDir, errLog)
	}

	// Namespaced scope: harvester cr
	namespaces = []string{"default", c.sbm.HarvesterNamespace}
	for _, namespace := range namespaces {
		namespacedDir := filepath.Join(yamlsDir, "namespaced", namespace)
		c.generateHarvesterNamespacedYAMLs(namespace, namespacedDir, errLog)
	}
}

type NamespacedGetter func(string) (runtime.Object, error)

func wrap(ns string, getter NamespacedGetter) GetRuntimeObjectListFunc {
	wrapped := func() (runtime.Object, error) {
		return getter(ns)
	}
	return wrapped
}

func (c *Cluster) generateKubernetesClusterYAMLs(dir string, errLog io.Writer) {
	toDir := filepath.Join(dir, "kubernetes")
	getListAndEncodeToYAML("nodes", c.sbm.k8s.GetAllNodesList, toDir, errLog)
	getListAndEncodeToYAML("volumeattachments", c.sbm.k8s.GetAllVolumeAttachments, toDir, errLog)
	getListAndEncodeToYAML("nodemetrics", c.sbm.k8sMetrics.GetAllNodeMetrics, toDir, errLog)
}

func (c *Cluster) generateKubernetesNamespacedYAMLs(namespace string, dir string, errLog io.Writer) {
	toDir := filepath.Join(dir, "kubernetes")
	getListAndEncodeToYAML("events", wrap(namespace, c.sbm.k8s.GetAllEventsList), toDir, errLog)
	getListAndEncodeToYAML("pods", wrap(namespace, c.sbm.k8s.GetAllPodsList), toDir, errLog)
	getListAndEncodeToYAML("services", wrap(namespace, c.sbm.k8s.GetAllServicesList), toDir, errLog)
	getListAndEncodeToYAML("deployments", wrap(namespace, c.sbm.k8s.GetAllDeploymentsList), toDir, errLog)
	getListAndEncodeToYAML("daemonsets", wrap(namespace, c.sbm.k8s.GetAllDaemonSetsList), toDir, errLog)
	getListAndEncodeToYAML("statefulsets", wrap(namespace, c.sbm.k8s.GetAllStatefulSetsList), toDir, errLog)
	getListAndEncodeToYAML("jobs", wrap(namespace, c.sbm.k8s.GetAllJobsList), toDir, errLog)
	getListAndEncodeToYAML("cronjobs", wrap(namespace, c.sbm.k8s.GetAllCronJobsList), toDir, errLog)
	getListAndEncodeToYAML("configmaps", wrap(namespace, c.sbm.k8s.GetAllConfigMaps), toDir, errLog)
	getListAndEncodeToYAML("podmetrics", wrap(namespace, c.sbm.k8sMetrics.GetAllPodMetrics), toDir, errLog)
}

func (c *Cluster) generateHarvesterClusterYAMLs(dir string, errLog io.Writer) {
	toDir := filepath.Join(dir, "harvester")
	getListAndEncodeToYAML("settings", c.sbm.harvester.GetAllSettings, toDir, errLog)
	getListAndEncodeToYAML("users", c.sbm.harvester.GetAllUsers, toDir, errLog)
}

func (c *Cluster) generateHarvesterNamespacedYAMLs(namespace string, dir string, errLog io.Writer) {
	// Harvester
	toDir := filepath.Join(dir, "harvester")
	getListAndEncodeToYAML("keypairs", wrap(namespace, c.sbm.harvester.GetAllKeypairs), toDir, errLog)
	getListAndEncodeToYAML("preferences", wrap(namespace, c.sbm.harvester.GetAllPreferences), toDir, errLog)
	getListAndEncodeToYAML("upgrades", wrap(namespace, c.sbm.harvester.GetAllUpgrades), toDir, errLog)
	getListAndEncodeToYAML("virtualmachinebackups", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineBackups), toDir, errLog)
	getListAndEncodeToYAML("virtualmachinebackupcontents", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineBackupContents), toDir, errLog)
	getListAndEncodeToYAML("virtualmachineimages", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineImages), toDir, errLog)
	getListAndEncodeToYAML("virtualmachinerestores", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineRestores), toDir, errLog)
	getListAndEncodeToYAML("virtualmachinetemplates", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineTemplates), toDir, errLog)
	getListAndEncodeToYAML("virtualmachinetemplateversions", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineTemplateVersions), toDir, errLog)

	// KubeVirt
	toDir = filepath.Join(dir, "kubevirt")
	getListAndEncodeToYAML("virtualmachines", wrap(namespace, c.sbm.harvester.GetAllVirtualMachines), toDir, errLog)
	getListAndEncodeToYAML("virtualmachineinstances", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineInstances), toDir, errLog)
	getListAndEncodeToYAML("virtualmachineinstancemigrations", wrap(namespace, c.sbm.harvester.GetAllVirtualMachineInstanceMigrations), toDir, errLog)

	// CDI
	toDir = filepath.Join(dir, "cdi")
	getListAndEncodeToYAML("datavolumes", wrap(namespace, c.sbm.harvester.GetAllDataVolumes), toDir, errLog)
}

func encodeToYAMLFile(obj interface{}, path string, errLog io.Writer) {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintf(errLog, "Support Bundle: failed to generate %v: %v\n", path, err)
		}
	}()
	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	switch v := obj.(type) {
	case runtime.Object:
		serializer := k8sjson.NewSerializerWithOptions(k8sjson.DefaultMetaFactory, nil, nil, k8sjson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		})
		if err = serializer.Encode(v, f); err != nil {
			return
		}
	default:
		encoder := yaml.NewEncoder(f)
		if err = encoder.Encode(obj); err != nil {
			return
		}
		if err = encoder.Close(); err != nil {
			return
		}
	}
}

type GetRuntimeObjectListFunc func() (runtime.Object, error)

func getListAndEncodeToYAML(name string, getListFunc GetRuntimeObjectListFunc, yamlsDir string, errLog io.Writer) {
	obj, err := getListFunc()
	if err != nil {
		fmt.Fprintf(errLog, "Support Bundle: failed to get %v: %v\n", name, err)
	}
	encodeToYAMLFile(obj, filepath.Join(yamlsDir, name+".yaml"), errLog)
}

func (c *Cluster) generateSupportBundleLogs(logsDir string, errLog io.Writer) {
	namespaces := []string{c.sbm.HarvesterNamespace, "default", "kube-system", "cattle-system"}

	for _, ns := range namespaces {
		list, err := c.sbm.k8s.GetAllPodsList(ns)
		if err != nil {
			fmt.Fprintf(errLog, "Support bundle: cannot get pod list: %v\n", err)
			return
		}
		podList, ok := list.(*corev1.PodList)
		if !ok {
			fmt.Fprintf(errLog, "BUG: Support bundle: didn't get pod list\n")
			return
		}
		for _, pod := range podList.Items {
			podName := pod.Name
			podDir := filepath.Join(logsDir, ns, podName)
			for _, container := range pod.Spec.Containers {
				req := c.sbm.k8s.GetPodContainerLogRequest(ns, podName, container.Name)
				logFileName := filepath.Join(podDir, container.Name+".log")
				stream, err := req.Stream(c.sbm.context)
				if err != nil {
					fmt.Fprintf(errLog, "BUG: Support bundle: cannot get log for pod %v container %v: %v\n",
						podName, container.Name, err)
					continue
				}
				streamLogToFile(stream, logFileName, errLog)
				stream.Close()
			}
		}
	}
}

func streamLogToFile(logStream io.ReadCloser, path string, errLog io.Writer) {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintf(errLog, "Support Bundle: failed to generate %v: %v\n", path, err)
		}
	}()
	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(f, logStream)
	if err != nil {
		return
	}
}

func (c *Cluster) getExternalSupportBundles(bundleMeta *BundleMeta, toDir string, errLog io.Writer) {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintf(errLog, "Support Bundle: failed to get external bundle: %v\n", err)
		}
	}()
	err = os.Mkdir(toDir, os.FileMode(0755))
	if err != nil {
		return
	}

	lh := external.NewLonghornSupportBundleManager(c.sbm.context, c.sbm.LonghornAPI)
	err = lh.GetLonghornSupportBundle(bundleMeta.IssueURL, bundleMeta.IssueDescription, toDir)
	if err != nil {
		return
	}
}

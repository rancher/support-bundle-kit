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

	"github.com/harvester/support-bundle-utils/pkg/manager/client"
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
	namespace, err := c.sbm.k8s.GetNamespace()
	if err != nil {
		return "", errors.Wrap(err, "cannot get harvester namespace")
	}
	kubeVersion, err := c.sbm.k8s.GetKubernetesVersion()
	if err != nil {
		return "", errors.Wrap(err, "cannot get kubernetes version")
	}
	sb, err := c.sbm.harvester.GetSupportBundle(c.sbm.BundleName)
	if err != nil {
		return "", errors.Wrap(err, "cannot get support bundle")
	}

	bundleMeta := &BundleMeta{
		ProjectName:          "Harvester",
		ProjectVersion:       c.sbm.HarvesterVersion,
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
	kubernetesDir := filepath.Join(yamlsDir, "kubernetes")
	c.generateSupportBundleYAMLsForKubernetes(kubernetesDir, errLog)
	harvesterDir := filepath.Join(yamlsDir, "harvester")
	c.generateSupportBundleYAMLsForHarvester(harvesterDir, errLog)
}

func (c *Cluster) generateSupportBundleYAMLsForKubernetes(dir string, errLog io.Writer) {
	getListAndEncodeToYAML("events", c.sbm.k8s.GetAllEventsList, dir, errLog)
	getListAndEncodeToYAML("pods", c.sbm.k8s.GetAllPodsList, dir, errLog)
	getListAndEncodeToYAML("services", c.sbm.k8s.GetAllServicesList, dir, errLog)
	getListAndEncodeToYAML("deployments", c.sbm.k8s.GetAllDeploymentsList, dir, errLog)
	getListAndEncodeToYAML("daemonsets", c.sbm.k8s.GetAllDaemonSetsList, dir, errLog)
	getListAndEncodeToYAML("statefulsets", c.sbm.k8s.GetAllStatefulSetsList, dir, errLog)
	getListAndEncodeToYAML("jobs", c.sbm.k8s.GetAllJobsList, dir, errLog)
	getListAndEncodeToYAML("cronjobs", c.sbm.k8s.GetAllCronJobsList, dir, errLog)
	getListAndEncodeToYAML("nodes", c.sbm.k8s.GetAllNodesList, dir, errLog)
	getListAndEncodeToYAML("configmaps", c.sbm.k8s.GetAllConfigMaps, dir, errLog)
	getListAndEncodeToYAML("volumeattachments", c.sbm.k8s.GetAllVolumeAttachments, dir, errLog)
}

func (c *Cluster) generateSupportBundleYAMLsForHarvester(dir string, errLog io.Writer) {

	// Harvester
	for _, ns := range []string{c.sbm.HarvesterNamespace, "default"} {
		harvester, err := client.NewHarvesterStore(c.sbm.context, ns, c.sbm.restConfig)
		if err != nil {
			fmt.Fprint(errLog, err)
			continue
		}
		toDir := filepath.Join(dir, "harvester", ns)
		getListAndEncodeToYAML("keypairs", harvester.GetAllKeypairs, toDir, errLog)
		getListAndEncodeToYAML("preferences", harvester.GetAllPreferences, toDir, errLog)
		getListAndEncodeToYAML("settings", harvester.GetAllSettings, toDir, errLog)
		getListAndEncodeToYAML("upgrades", harvester.GetAllUpgrades, toDir, errLog)
		getListAndEncodeToYAML("users", harvester.GetAllUsers, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinebackups", harvester.GetAllVirtualMachineBackups, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinebackupcontents", harvester.GetAllVirtualMachineBackupContents, toDir, errLog)
		getListAndEncodeToYAML("virtualmachineimages", harvester.GetAllVirtualMachineImages, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinerestores", harvester.GetAllVirtualMachineRestores, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinetemplates", harvester.GetAllVirtualMachineTemplates, toDir, errLog)
		getListAndEncodeToYAML("virtualmachinetemplateversions", harvester.GetAllVirtualMachineTemplateVersions, toDir, errLog)
	}

	// KubeVirt & CDI
	ns := "default"
	harvester, err := client.NewHarvesterStore(c.sbm.context, ns, c.sbm.restConfig)
	if err != nil {
		fmt.Fprint(errLog, err)
		return
	}
	toDir := filepath.Join(dir, "kubevirt", ns)
	getListAndEncodeToYAML("virtualmachines", harvester.GetAllVirtualMachines, toDir, errLog)
	getListAndEncodeToYAML("virtualmachineinstances", harvester.GetAllVirtualMachineInstances, toDir, errLog)
	getListAndEncodeToYAML("virtualmachineinstancemigrations", harvester.GetAllVirtualMachineInstanceMigrations, toDir, errLog)

	toDir = filepath.Join(dir, "cdi", ns)
	getListAndEncodeToYAML("datavolumes", harvester.GetAllDataVolumes, toDir, errLog)
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
		k8s, err := client.NewKubernetesStore(c.sbm.context, ns, c.sbm.restConfig)
		if err != nil {
			fmt.Fprint(errLog, err)
			continue
		}
		list, err := k8s.GetAllPodsList()
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
				req := k8s.GetPodContainerLogRequest(podName, container.Name)
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

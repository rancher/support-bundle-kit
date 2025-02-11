package objects

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	bundlekit "github.com/rancher/support-bundle-kit/pkg/simulator/apis/supportbundlekit.io/v1"
)

const (
	DefaultNodeDir      = "nodes"
	DefaultPodNamespace = "support-bundle-node-info"
)

var NodeInfoNS = v1.Namespace{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Namespace",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: DefaultPodNamespace,
	},
}

var NodeInfoSA = v1.ServiceAccount{
	TypeMeta: metav1.TypeMeta{
		Kind:       "ServiceAccount",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "default",
		Namespace: DefaultPodNamespace,
	},
	Secrets: []v1.ObjectReference{
		{
			Name:       "default",
			Kind:       "Secret",
			APIVersion: "v1",
			Namespace:  DefaultPodNamespace,
		},
	},
}

var NodeInfoSASecret = v1.Secret{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "default",
		Namespace: DefaultPodNamespace,
	},
	StringData: make(map[string]string),
}

// CreateNodeZipObjects will create associated objects in the cluster
func (o *ObjectManager) CreateNodeZipObjects() error {
	noStatusObjs, withStatusObjs, err := o.ProcessNodeZipObjects()
	if err != nil {
		return err
	}

	progressMgr := NewProgressManager("Step 3/4: NodeZipObjects (no status)")
	err = o.ApplyObjects(noStatusObjs, false, nil, progressMgr.progress)
	if err != nil {
		return err
	}

	progressMgr = NewProgressManager("Step 3/4: NodeZipObjects (status)")
	return o.ApplyObjects(withStatusObjs, true, nil, progressMgr.progress)
}

// ProcessNodeZipObjects will read the contents of the zip file and generate associated runtime.Objects
func (o *ObjectManager) ProcessNodeZipObjects() (noStatusObjs []runtime.Object, withStatusObjs []runtime.Object, err error) {

	// unzip files
	// drop them into the yamls folder
	// clean up later

	bundleAbsPath, err := filepath.Abs(o.path)

	if err != nil {
		return noStatusObjs, withStatusObjs, fmt.Errorf("error evaulating absolute path to node dirs: %v", err)
	}

	nodeZipList, err := generateNodeZipList(bundleAbsPath)

	if err != nil {
		return noStatusObjs, withStatusObjs, err
	}

	podList, nodeConfig, err := generateObjects(nodeZipList)
	if err != nil {
		return noStatusObjs, withStatusObjs, err
	}

	noStatusObjs = []runtime.Object{
		&NodeInfoNS, &NodeInfoSASecret, &NodeInfoSA,
	}

	for _, v := range podList {
		withStatusObjs = append(withStatusObjs, v)
	}

	for _, v := range nodeConfig {
		noStatusObjs = append(noStatusObjs, v)
	}
	return noStatusObjs, withStatusObjs, nil
}

func generateNodeZipList(bundleAbsPath string) ([]string, error) {

	var nodeZipList []string
	err := filepath.Walk(filepath.Join(bundleAbsPath, DefaultNodeDir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.Contains(path, ".zip") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			nodeZipList = append(nodeZipList, absPath)
		}

		return nil
	})

	if err != nil {
		return nodeZipList, fmt.Errorf("error during dir walk: %v", err)
	}

	return nodeZipList, nil
}

func generateObjects(nodeZipList []string) ([]*v1.Pod, []*bundlekit.NodeConfig, error) {
	var podList []*v1.Pod
	var configList []*bundlekit.NodeConfig

	for _, zipFile := range nodeZipList {
		pod, nodeconfig, err := walkZipFiles(zipFile)
		if err != nil {
			return nil, nil, err
		}

		// ignore node.zip files where no containers were found
		if pod != nil {
			podList = append(podList, pod)
		}
		configList = append(configList, nodeconfig)
	}

	return podList, configList, nil
}

func walkZipFiles(zipFile string) (*v1.Pod, *bundlekit.NodeConfig, error) {
	nodeName := strings.Split(zipFile, ".zip")
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      filepath.Base(nodeName[0]),
			Namespace: DefaultPodNamespace,
		},
	}
	nConfig := &bundlekit.NodeConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NodeConfig",
			APIVersion: "supportbundlekit.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      filepath.Base(nodeName[0]),
			Namespace: DefaultPodNamespace,
		},
	}

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return pod, nConfig, fmt.Errorf("error opening zip file %v:", err)
	}
	defer r.Close()

	var nodeConfigSpec []bundlekit.NodeConfigSpec
	var containers []v1.Container
	// iterate files into pod and nodeConfig object
	for _, f := range r.File {

		// config files
		if strings.Contains(f.Name, "configs") && !f.FileInfo().IsDir() {
			contentBytes, err := ReadContent(f)
			if err != nil {
				return pod, nConfig, err
			}
			ncSpec := bundlekit.NodeConfigSpec{
				FileName: f.Name,
				Content:  string(contentBytes),
			}
			nodeConfigSpec = append(nodeConfigSpec, ncSpec)
		}

		// generate pod object, skip parent directory
		if strings.Contains(f.Name, "logs") && !f.FileInfo().IsDir() {
			containerName := strings.Split(f.Name, ".log")
			c := v1.Container{
				Name:  filepath.Base(containerName[0]),
				Image: "noimage",
			}
			containers = append(containers, c)
		}
	}
	nConfig.Spec = nodeConfigSpec
	if len(containers) == 0 {
		logrus.Warnf("No pod being created for node %s as zip file has no log files associated with this node", filepath.Base(nodeName[0]))
		return nil, nConfig, nil
	}
	pod.Spec.Containers = containers
	pod.Spec.NodeName = filepath.Base(nodeName[0])
	pod.Status = *generatePodStatus(pod)
	return pod, nConfig, nil
}

func ReadContent(f *zip.File) ([]byte, error) {
	zipReader, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", f.Name, err)
	}
	defer zipReader.Close()
	contentBytes, err := io.ReadAll(zipReader)
	return contentBytes, err
}

func generatePodStatus(pod *v1.Pod) *v1.PodStatus {
	podStatus := pod.Status.DeepCopy()
	podStatus.Phase = v1.PodRunning
	podStatus.StartTime = &metav1.Time{Time: time.Now()}
	podStatus.Conditions = []v1.PodCondition{
		v1.PodCondition{
			Type:   v1.PodReady,
			Status: v1.ConditionTrue,
		},
		v1.PodCondition{
			Type:   v1.PodScheduled,
			Status: v1.ConditionTrue,
		},
		v1.PodCondition{
			Type:   v1.PodInitialized,
			Status: v1.ConditionTrue,
		},
		v1.PodCondition{
			Type:   v1.ContainersReady,
			Status: v1.ConditionTrue,
		},
	}

	var cStats []v1.ContainerStatus
	for _, c := range pod.Spec.Containers {
		cStats = append(cStats, v1.ContainerStatus{
			Name:         c.Name,
			Ready:        true,
			RestartCount: 0,
			State: v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: metav1.Time{Time: time.Now()},
				},
			},
		})
	}

	podStatus.ContainerStatuses = cStats
	return podStatus
}

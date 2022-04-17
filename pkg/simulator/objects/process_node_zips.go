package objects

import (
	"archive/zip"
	"fmt"
	bundlekit "github.com/rancher/support-bundle-kit/pkg/simulator/apis/supportbundlekit.io/v1"
	"github.com/rancher/support-bundle-kit/pkg/simulator/crd"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultNodeDir      = "nodes"
	DefaultPodNamespace = "support-bundle-node-info"
)

var NodeInfoNS = v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: DefaultPodNamespace,
	},
}

var NodeInfoSA = v1.ServiceAccount{
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
	ObjectMeta: metav1.ObjectMeta{
		Name:      "default",
		Namespace: DefaultPodNamespace,
	},
	StringData: make(map[string]string),
}

// CreateNodeZipObjects will create associated objects in the cluster
func (o *ObjectManager) CreateNodeZipObjects() error {
	objs, err := o.ProcessNodeZipObjects()
	if err != nil {
		return err
	}

	return o.ApplyObjects(objs, false, nil)
}

// ProcessNodeZipObjects will read the contents of the zip file and generate associated runtime.Objects
func (o *ObjectManager) ProcessNodeZipObjects() ([]runtime.Object, error) {

	// unzip files
	// drop them into the yamls folder
	// clean up later

	var objs, nodeObjs []runtime.Object
	bundleAbsPath, err := filepath.Abs(o.path)

	if err != nil {
		return nodeObjs, fmt.Errorf("error evaulating absolute path to node dirs: %v", err)
	}

	crdObjects, err := crd.Objects(false)
	if err != nil {
		return nodeObjs, fmt.Errorf("error generating CRD objects: %v", err)
	}

	nodeZipList, err := generateNodeZipList(bundleAbsPath)

	if err != nil {
		return nodeObjs, err
	}

	podList, nodeConfig, err := generateObjects(nodeZipList)
	if err != nil {
		return nodeObjs, err
	}

	nodeObjs = []runtime.Object{
		&NodeInfoNS, &NodeInfoSASecret, &NodeInfoSA, podList, nodeConfig,
	}
	
	objs = append(objs, crdObjects...)
	objs = append(objs, nodeObjs...)

	return objs, nil
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

func generateObjects(nodeZipList []string) (*v1.PodList, *bundlekit.NodeConfigList, error) {
	podList := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
	}

	configList := &bundlekit.NodeConfigList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
	}

	for _, zipFile := range nodeZipList {
		pod, nodeconfig, err := walkZipFiles(zipFile)
		if err != nil {
			return nil, nil, err
		}
		podList.Items = append(podList.Items, pod)
		configList.Items = append(configList.Items, nodeconfig)
	}

	return podList, configList, nil
}

func walkZipFiles(zipFile string) (v1.Pod, bundlekit.NodeConfig, error) {
	nodeName := strings.Split(zipFile, ".zip")
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      filepath.Base(nodeName[0]),
			Namespace: DefaultPodNamespace,
		},
	}
	nConfig := bundlekit.NodeConfig{
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

	pod.Spec.Containers = containers
	nConfig.Spec = nodeConfigSpec
	return pod, nConfig, nil
}

func ReadContent(f *zip.File) ([]byte, error) {
	zipReader, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", f.Name, err)
	}
	defer zipReader.Close()
	contentBytes, err := ioutil.ReadAll(zipReader)
	return contentBytes, err
}

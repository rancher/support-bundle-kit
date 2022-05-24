package objects

import (
	"github.com/rancher/support-bundle-kit/pkg/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateNodeZipList(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "zipfiles-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}

	list, err := generateNodeZipList(filepath.Join(tmpDir, supportBundleDir))
	if err != nil {
		t.Fatalf("error reading zip file")
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 got %d", len(list))
	}

	if filepath.Base(list[0]) != "node1.zip" {
		t.Fatalf("expected to find node1.zip but found %v", list[0])
	}
}

func TestGenerateNodeZipObjects(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "zipfiles-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}

	list, err := generateNodeZipList(filepath.Join(tmpDir, supportBundleDir))
	if err != nil {
		t.Fatalf("error reading zip file")
	}

	podList, nodeConfig, err := generateObjects(list)
	if err != nil {
		t.Fatalf("error generating objects %v", err)
	}

	// verify pod
	if len(podList) != 1 {
		t.Fatalf("expected to find 1 pod created to match the node, but found %d", len(podList))
	}

	// parse containers in the pod
	node1Pod := podList[0]
	if node1Pod.Name != "node1" {
		t.Fatalf("expected pod name to be node1 but got %s", node1Pod.Name)
	}

	if len(node1Pod.Spec.Containers) != 8 {
		t.Fatalf("expected 8 containers but found %d", len(node1Pod.Spec.Containers))
	}

	if len(nodeConfig) != 1 {
		t.Fatalf("expected to find 1 node but found %d", len(nodeConfig))
	}

	// parse content
	if len(nodeConfig[0].Spec) != 18 {
		t.Fatalf("expected to find 18 files, but found %d", len(nodeConfig[0].Spec))
	}
}

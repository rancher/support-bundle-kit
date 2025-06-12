package objects

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rancher/support-bundle-kit/pkg/utils"
)

func TestGenerateNodeZipList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "zipfiles-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}

	list, err := generateNodeZipList(filepath.Join(tmpDir, supportBundleDir))
	if err != nil {
		t.Fatalf("error reading zip file")
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 got %d", len(list))
	}

	if filepath.Base(list[0]) != "harvester-node-0.zip" {
		t.Fatalf("expected to find harvester-node-0.zip but found %v", list[0])
	}
}

func TestGenerateNodeZipObjects(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "zipfiles-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

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
	if len(podList) != 2 {
		t.Fatalf("expected to find 1 pod created to match the node, but found %d", len(podList))
	}

	// parse containers in the pod
	node1Pod := podList[0]
	if node1Pod.Name != "harvester-node-0" {
		t.Fatalf("expected pod name to be node1 but got %s", node1Pod.Name)
	}

	if len(node1Pod.Spec.Containers) != 11 {
		t.Fatalf("expected 11 containers but found %d", len(node1Pod.Spec.Containers))
	}

	if len(nodeConfig) != 2 {
		t.Fatalf("expected to find 1 node but found %d", len(nodeConfig))
	}

	// parse content
	if len(nodeConfig[0].Spec) != 30 {
		t.Fatalf("expected to find 30 files, but found %d", len(nodeConfig[0].Spec))
	}
}

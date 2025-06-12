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
	const (
		expectedPodCount       = 2
		expectedPodName        = "harvester-node-0"
		expectedContainerCount = 11
		expectedConfigLength   = 2
		expectedFileCount      = 30
	)

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
	if len(podList) != expectedPodCount {
		t.Fatalf("expected to find %d pod created to match the node, but found %d", expectedPodCount, len(podList))
	}

	// parse containers in the pod
	node1Pod := podList[0]
	if node1Pod.Name != expectedPodName {
		t.Fatalf("expected pod name to be %s but got %s", expectedPodName, node1Pod.Name)
	}

	if len(node1Pod.Spec.Containers) != expectedContainerCount {
		t.Fatalf("expected %d containers but found %d", expectedContainerCount, len(node1Pod.Spec.Containers))
	}

	if len(nodeConfig) != expectedConfigLength {
		t.Fatalf("expected to find %d node but found %d", expectedConfigLength, len(nodeConfig))
	}

	// parse content
	if len(nodeConfig[0].Spec) != expectedFileCount {
		t.Fatalf("expected to find %d files, but found %d", expectedFileCount, len(nodeConfig[0].Spec))
	}
}

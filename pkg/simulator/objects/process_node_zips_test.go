package objects

import (
	"path/filepath"
	"testing"
)

func TestGenerateNodeZipList(t *testing.T) {
	list, err := generateNodeZipList("../../../tests/integration/sampleSupportBundle")
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
	list, err := generateNodeZipList("../../../tests/integration/sampleSupportBundle")
	if err != nil {
		t.Fatalf("error reading zip file")
	}

	podList, nodeConfig, err := generateObjects(list)
	if err != nil {
		t.Fatalf("error generating objects %v", err)
	}

	// verify pod
	if len(podList.Items) != 1 {
		t.Fatalf("expected to find 1 pod created to match the node, but found %d", len(podList.Items))
	}

	// parse containers in the pod
	node1Pod := podList.Items[0]
	if node1Pod.Name != "node1" {
		t.Fatalf("expected pod name to be node1 but got %s", node1Pod.Name)
	}

	if len(node1Pod.Spec.Containers) != 8 {
		t.Fatalf("expected 8 containers but found %d", len(node1Pod.Spec.Containers))
	}

	if len(nodeConfig.Items) != 1 {
		t.Fatalf("expected to find 1 node but found %d", len(nodeConfig.Items))
	}

	// parse content
	if len(nodeConfig.Items[0].Spec) != 18 {
		t.Fatalf("expected to find 18 files, but found %d", len(nodeConfig.Items[0].Spec))
	}
}

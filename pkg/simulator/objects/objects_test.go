package objects

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rancher/support-bundle-kit/pkg/utils"
)

// TestGenerateClusterScopedRuntimeObjects will test cluster scoped object generation from a sample support bundle
func TestGenerateClusterScopedRuntimeObjects(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "objects-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}

	crds, clusterObjs, err := GenerateClusterScopedRuntimeObjects(filepath.Join(tmpDir, supportBundleDir))
	if err != nil {
		t.Fatalf("error processing crds and cluster scoped objects from support bundle %v", err)
	}

	t.Logf("found %d crds in support bundle", len(crds))
	t.Logf("found %d cluster scoped objects in support bundle", len(clusterObjs))
}

// TestGenerateNamepsacedRuntimeObjects will test namespaced cluster objects.
func TestGenerateNamespacedRuntimeObjects(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "objects-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	nonpodObjs, podObjs, err := GenerateNamespacedRuntimeObjects(filepath.Join(tmpDir, supportBundleDir))
	if err != nil {
		t.Fatalf("error processing non namespaced objects and pods from support bundle %v", err)
	}

	t.Logf("found %d namespaced non-pod objects", len(nonpodObjs))
	t.Logf("found %d namespaced pod objects", len(podObjs))
}

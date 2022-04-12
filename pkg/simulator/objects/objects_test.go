package objects

import (
	"testing"
)

// TestGenerateClusterScopedRuntimeObjects will test cluster scoped object generation from a sample support bundle
func TestGenerateClusterScopedRuntimeObjects(t *testing.T) {
	crds, clusterObjs, err := GenerateClusterScopedRuntimeObjects("../../../tests/integration/sampleSupportBundle")
	if err != nil {
		t.Fatalf("error processing crds and cluster scoped objects from support bundle %v", err)
	}

	t.Logf("found %d crds in support bundle", len(crds))
	t.Logf("found %d cluster scoped objects in support bundle", len(clusterObjs))
}

// TestGenerateNamepsacedRuntimeObjects will test namespaced cluster objects.
func TestGenerateNamespacedRuntimeObjects(t *testing.T) {
	nonpodObjs, podObjs, err := GenerateNamespacedRuntimeObjects("../../../tests/integration/sampleSupportBundle")
	if err != nil {
		t.Fatalf("error processing non namespaced objects and pods from support bundle %v", err)
	}

	t.Logf("found %d namespaced non-pod objects", len(nonpodObjs))
	t.Logf("found %d namespaced pod objects", len(podObjs))
}

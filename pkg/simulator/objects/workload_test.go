package objects

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"

	"github.com/rancher/support-bundle-kit/pkg/utils"
)

const (
	testDSFilePath      = "yamls/namespaced/cattle-monitoring-system/apps/v1/daemonsets.yaml"
	testSTSFilePath     = "yamls/namespaced/cattle-monitoring-system/apps/v1/statefulsets.yaml"
	testDeployFilePath  = "yamls/namespaced/cattle-monitoring-system/apps/v1/deployments.yaml"
	testRSFilePath      = "yamls/namespaced/cattle-monitoring-system/apps/v1/replicasets.yaml"
	testIngressFilePath = "yamls/namespaced/cattle-system/networking.k8s.io/v1/ingresses.yaml"
	testSettingsPath    = "yamls/cluster/management.cattle.io/v3/settings.yaml"
	testJobPath         = "yamls/namespaced/harvester-system/batch/v1/jobs.yaml"
	nodePath            = "yamls/cluster/v1/nodes.yaml"
)

// TestParseDaemonSet will verify a sample Daemonset
func TestParseDaemonSetObject(t *testing.T) {
	// read our sample node-exporter daemonset in the samples directory
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	verifyTestWorkloads(t, filepath.Join(tmpDir, supportBundleDir, testDSFilePath))
}

// TestParseReplicaSetObject will verify a sample Replicaset
func TestParseReplicaSetObject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	verifyTestWorkloads(t, filepath.Join(tmpDir, supportBundleDir, testRSFilePath))
}

// TestParseStatefulSetObject will verify a sample StatefulSet
func TestParseStatefulSetObject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	verifyTestWorkloads(t, filepath.Join(tmpDir, supportBundleDir, testSTSFilePath))
}

// TestParseStatefulSetObject will verify a sample StatefulSet
func TestParseDeploymentObject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	verifyTestWorkloads(t, filepath.Join(tmpDir, supportBundleDir, testDeployFilePath))
}

func TestIngressObject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	verifyTestWorkloads(t, filepath.Join(tmpDir, supportBundleDir, testIngressFilePath))
}

func TestSettings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	verifyTestWorkloads(t, filepath.Join(tmpDir, supportBundleDir, testSettingsPath))
}

func verifyTestWorkloads(t *testing.T, path string) {
	objs, err := GenerateObjects(path)
	if err != nil {
		t.Fatalf("error reading sample daemonset file %s %v", testDSFilePath, err)
	}

	for _, obj := range objs {
		unstructObj, err := wranglerunstructured.ToUnstructured(obj)
		if err != nil {
			t.Fatalf("error converting runtime object to unstructured object %v", err)
		}

		err = cleanupObjects(unstructObj.Object)
		if err != nil {
			t.Fatalf("error cleaning up objects %v", err)
		}

		if err != nil {
			t.Fatalf("error performing housekeeping on unstructured daemonset")
		}

		// verify objects
		ok, err := verifyObject(unstructObj, nil, "metadata", "resourceVersion")
		if err != nil {
			t.Fatalf("error verifying resource version object %v", err)
		}
		if ok {
			t.Fatalf("expected to find no resourceVersion but found one at metadata.resourceVersion")
		}

		ok, err = verifyObject(unstructObj, nil, "spec", "template", "metadata", "creationTimestamp")
		if err != nil {
			t.Fatalf("error verifying creationTimeStamp %v", err)
		}
		if ok {
			t.Fatalf("expected to find no creationTimeStamp but found one at spec.template.metadata.creationTimeStamp")
		}

		ok, err = verifyObject(unstructObj, func(v interface{}) (bool, error) {
			var volOk, innerOK bool
			var err error
			for _, volumes := range v.([]interface{}) {
				_, innerOK, err = unstructured.NestedFieldNoCopy(volumes.(map[string]interface{}), "hostPath", "type")
				if err != nil {
					return false, err
				}
				volOk = volOk || innerOK

			}
			return ok, err
		}, "spec", "template", "spec", "volumes")

		if err != nil {
			t.Fatalf("error verifying volumes in daemonset %v", err)
		}

		if ok {
			t.Fatalf("expected to find no type for hostPath volume definitions")
		}
	}
}

func TestVerifyJob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	objs, err := GenerateObjects(filepath.Join(tmpDir, supportBundleDir, testJobPath))
	if err != nil {
		t.Fatalf("error reading sample daemonset file %s %v", testDSFilePath, err)
	}
	for _, obj := range objs {
		unstructObj, err := wranglerunstructured.ToUnstructured(obj)
		if err != nil {
			t.Fatal(err)
		}

		err = cleanupObjects(unstructObj.Object)
		if err != nil {
			t.Fatal(err)
		}

		err = objectHousekeeping(unstructObj)
		if err != nil {
			t.Fatal(err)
		}
		labels, ok, err := unstructured.NestedStringMap(unstructObj.Object, "spec", "template", "metadata", "labels")
		if err != nil {
			t.Fatal(err)
		}

		if ok {
			t.Fatalf("expect to find no labels but found labels %v", labels)
		}
	}
}

func TestVerifyNode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "workload-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}

	objs, err := GenerateObjects(filepath.Join(tmpDir, supportBundleDir, nodePath))
	if err != nil {
		t.Fatalf("error reading sample daemonset file %s %v", testDSFilePath, err)
	}
	for _, obj := range objs {
		unstructObj, err := wranglerunstructured.ToUnstructured(obj)
		if err != nil {
			t.Log(unstructObj.Object["status"])
			t.Fatal(err)
		}

		err = cleanupObjects(unstructObj.Object)
		if err != nil {
			t.Fatal(err)
		}

		err = objectHousekeeping(unstructObj)
		if err != nil {
			t.Fatal(err)
		}

		addresses, ok, err := unstructured.NestedFieldCopy(unstructObj.Object, "status", "addresses")
		if err != nil {
			t.Fatalf("error looking up addresses %v", err)
		}

		if !ok {
			t.Fatalf("found no addresses in node")
		}

		addressList, ok := addresses.([]interface{})
		if !ok {
			t.Fatal("unable to assert addresses into []interface{}")
		}

		for _, addressInterface := range addressList {
			address, ok := addressInterface.(map[string]interface{})
			if !ok {
				t.Fatal("unable to assert address into map[string]interface")
			}
			value, ok := address["address"]
			if !ok || value.(string) == "localhost" {
				t.Fatalf("expected address to be %v but found localhost", address)
			}
		}

	}
}

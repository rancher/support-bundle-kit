package objects

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rancher/support-bundle-kit/pkg/utils"

	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"
)

const (
	testCoreEventsFile = "yamls/namespaced/default/v1/events.yaml"
	testEventsFile     = "yamls/namespaced/default/events.k8s.io/v1/events.yaml"
	bundleZipPath      = "../../../tests/integration/sampleSupportBundle.zip"
	supportBundleDir   = "sampleSupportBundle"
)

func TestCoreEvents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "events-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}

	testCoreEvents := filepath.Join(tmpDir, supportBundleDir, testCoreEventsFile)

	objs, err := GenerateObjects(testCoreEvents)
	if err != nil {
		t.Fatalf("error parsing core events %v", err)
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

		err = objectHousekeeping(unstructObj)
		t.Log(unstructObj)
		if err != nil {
			t.Fatalf("error during object housekeeping: %v", err)
		}

		if _, ok := unstructObj.Object["eventTime"]; !ok {
			t.Fatalf("expected to find eventTime but did not find one")
		}

		if _, ok := unstructObj.Object["reportingController"]; !ok {
			t.Fatalf("expected to find reportingController but did not find one")
		}

		if _, ok := unstructObj.Object["reportingInstance"]; !ok {
			t.Fatalf("expected to find reportingInstance but did not find one")
		}

		if _, ok := unstructObj.Object["series"]; ok {
			t.Fatalf("expected to not find a series but found one")
		}

	}
}

func TestEvents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "events-")
	if err != nil {
		t.Fatalf("Error creating tmp directory %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}

	testEvents := filepath.Join(tmpDir, supportBundleDir, testEventsFile)

	objs, err := GenerateObjects(testEvents)
	if err != nil {
		t.Fatalf("error parsing core events %v", err)
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

		err = objectHousekeeping(unstructObj)

		if err != nil {
			t.Fatalf("error during object housekeeping: %v", err)
		}

		t.Log(unstructObj)
		if _, ok := unstructObj.Object["eventTime"]; !ok {
			t.Fatalf("expected to find eventTime but did not find one")
		}

		if _, ok := unstructObj.Object["reportingController"]; !ok {
			t.Fatalf("expected to find reportingController but did not find one")
		}

		if _, ok := unstructObj.Object["reportingInstance"]; !ok {
			t.Fatalf("expected to find reportingInstance but did not find one")
		}

		if _, ok := unstructObj.Object["series"]; ok {
			t.Fatalf("expected to not find a series but found one")
		}
	}
}

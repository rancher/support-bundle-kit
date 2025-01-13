package objects

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"
)

const extensionsIngressSample = `
apiVersion: v1
items:
- apiVersion: extensions/v1beta1
  kind: Ingress
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"annotations":{},"name":"rancher-expose","namespace":"cattle-system"},"spec":{"rules":[{"http":{"paths":[{"backend":{"service":{"name":"rancher","port":{"number":80}}},"path":"/","pathType":"Prefix"}]}}]}}
    creationTimestamp: "2022-04-11T08:17:02Z"
    generation: 1
    managedFields:
    - apiVersion: networking.k8s.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:kubectl.kubernetes.io/last-applied-configuration: {}
        f:spec:
          f:rules: {}
      manager: kubectl-client-side-apply
      operation: Update
      time: "2022-04-11T08:17:02Z"
    - apiVersion: networking.k8s.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:status:
          f:loadBalancer:
            f:ingress: {}
      manager: nginx-ingress-controller
      operation: Update
      time: "2022-04-11T08:18:00Z"
    name: rancher-expose
    namespace: cattle-system
    resourceVersion: "58705"
    uid: 32dcceb4-1f5a-47f3-b75c-b33fe826ea4d
  spec:
    rules:
    - http:
        paths:
        - backend:
            serviceName: rancher
            servicePort: 80
          path: /
          pathType: Prefix
  status:
    loadBalancer:
      ingress:
      - ip: 10.10.0.11
      - ip: 10.10.0.12
      - ip: 10.10.0.13
      - ip: 10.10.0.14
kind: List
metadata:
  resourceVersion: "65340"

`

func Test_cleanupIngress(t *testing.T) {
	assert := require.New(t)

	tmpFile, err := os.CreateTemp("/tmp", "ingress")
	assert.NoError(err, "expected no error during creation of tmp ingress file")
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(extensionsIngressSample))
	assert.NoError(err, "expected no error during writing to tmp ingress file")
	assert.NoError(tmpFile.Close(), "expect no error during file close")

	objs, err := GenerateObjects(tmpFile.Name())
	assert.NoError(err, "expect no error during ingress object generation")

	for _, o := range objs {
		unstructObj, err := wranglerunstructured.ToUnstructured(o)
		assert.NoError(err, "expected no error during conversion to unstructured object")
		err = cleanupIngress(unstructObj)
		assert.NoError(err, "expected no error during cleanup on ingress objects")

		// check if new fields now exist
		assert.Equal(unstructObj.GetAPIVersion(), "networking.k8s.io/v1", "expected update to new apiversion")
		rules, _, err := unstructured.NestedSlice(unstructObj.Object, "spec", "rules")
		for _, r := range rules {
			rMap, ok := r.(map[string]interface{})
			assert.True(ok, "expected successful assertion to map")
			paths, _, err := unstructured.NestedSlice(rMap, "http", "paths")
			assert.NoError(err, "expected no error during lookup of paths")
			for _, p := range paths {
				pMap, ok := p.(map[string]interface{})
				assert.True(ok, "expected successful assertion to map")
				_, ok, err := unstructured.NestedFieldCopy(pMap, "backend", "service", "name")
				assert.NoError(err, "expected no error looking up backend.service.name")
				assert.True(ok, "expected to find backend.service.name")
				_, ok, err = unstructured.NestedFieldCopy(pMap, "backend", "service", "port", "number")
				assert.NoError(err, "expected no error looking up backend.service.port.number")
				assert.True(ok, "expected to find backend.service.port.number")
			}
		}
		assert.NoError(err, "expected no error during query of rules")
	}

}

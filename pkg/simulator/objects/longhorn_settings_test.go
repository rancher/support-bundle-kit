package objects

import (
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

var (
	nullSetting = `apiVersion: longhorn.io/v1beta2
kind: Setting
metadata:
  creationTimestamp: "2022-09-08T18:48:13Z"
  generation: 1
  managedFields:
  - apiVersion: longhorn.io/v1beta1
    fieldsType: FieldsV1
    fieldsV1:
      f:value: {}
    manager: longhorn-manager
    operation: Update
    time: "2022-09-08T18:48:13Z"
  name: default-replica-count
  namespace: longhorn-system
  resourceVersion: "4346"
  uid: c2ba018f-f523-48a5-b4ca-45667642fab0
value: "null"`

	normalSetting = `apiVersion: longhorn.io/v1beta2
kind: Setting
metadata:
  creationTimestamp: "2022-09-08T18:48:13Z"
  generation: 1
  managedFields:
  - apiVersion: longhorn.io/v1beta1
    fieldsType: FieldsV1
    fieldsV1:
      f:value: {}
    manager: longhorn-manager
    operation: Update
    time: "2022-09-08T18:48:13Z"
  name: default-replica-count
  namespace: longhorn-system
  resourceVersion: "4346"
  uid: c2ba018f-f523-48a5-b4ca-45667642fab0
value: "3"`

	missingSetting = `apiVersion: longhorn.io/v1beta2
kind: Setting
metadata:
  creationTimestamp: "2022-09-08T18:48:13Z"
  generation: 1
  managedFields:
  - apiVersion: longhorn.io/v1beta1
    fieldsType: FieldsV1
    fieldsV1:
      f:value: {}
    manager: longhorn-manager
    operation: Update
    time: "2022-09-08T18:48:13Z"
  name: default-replica-count
  namespace: longhorn-system
  resourceVersion: "4346"
  uid: c2ba018f-f523-48a5-b4ca-45667642fab0`
)

func Test_NullSetting(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(nullSetting)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		err = cleanupLonghornSettings(v)
		assert.NoError(err, "expected no error during object cleanup")
		val, valOK, err := unstructured.NestedString(v.Object, "value")
		assert.NoError(err, "expected no error looking up value")
		assert.True(valOK, "expected value to exist")
		assert.Equal("", val, "expected the value to be an empty string")

	}
}

func Test_Setting(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(normalSetting)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		err = cleanupLonghornSettings(v)
		assert.NoError(err, "expected no error during object cleanup")
		val, valOK, err := unstructured.NestedString(v.Object, "value")
		assert.NoError(err, "expected no error looking up value")
		assert.True(valOK, "expected value to exist")
		assert.Equal("3", val, "expected the value to be 3")

	}
}

func Test_MissingSetting(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(missingSetting)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		err = cleanupLonghornSettings(v)
		assert.NoError(err, "expected no error during object cleanup")
		val, valOK, err := unstructured.NestedString(v.Object, "value")
		assert.NoError(err, "expected no error looking up value")
		assert.True(valOK, "expected value to exist")
		assert.Equal("", val, "expected the value to be 3")

	}
}

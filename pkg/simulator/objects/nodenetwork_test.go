package objects

import (
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

var (
	nodeNetwork = `apiVersion: network.harvesterhci.io/v1beta1
kind: NodeNetwork
metadata:
  creationTimestamp: "2022-09-08T18:48:20Z"
  finalizers:
  - wrangler.cattle.io/harvester-nodenetwork-controller
  generation: 42
  labels:
    network.harvesterhci.io/nodename: p5820
  managedFields:
  - apiVersion: network.harvesterhci.io/v1beta1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .: {}
          v:"wrangler.cattle.io/harvester-nodenetwork-controller"null: {}
        f:labels:
          .: {}
          f:network.harvesterhci.io/nodename: {}
        f:ownerReferences:
          .: {}
          k:{"uid":"9b45a5e7-8a72-4396-94e9-36a20b432883"}: {}
      f:spec:
        .: {}
        f:nic: {}
        f:nodeName: {}
        f:type: {}
      f:status:
        .: {}
        f:conditions: {}
        f:networkLinkStatus:
          .: {}
          f:harvester-br0:
            .: {}
            f:index: {}
            f:ipv4Address: {}
            f:mac: {}
            f:promiscuous: {}
            f:routes: {}
            f:state: {}
            f:type: {}
          f:harvester-mgmt:
            .: {}
            f:index: {}
            f:ipv4Address: {}
            f:mac: {}
            f:masterIndex: {}
            f:state: {}
            f:type: {}
        f:nics: {}
    manager: harvester-network-controller
    operation: Update
    time: "2022-09-08T19:17:00Z"
  name: p5820-vlan
  ownerReferences:
  - apiVersion: v1
    kind: Node
    name: p5820
    uid: 9b45a5e7-8a72-4396-94e9-36a20b432883
  resourceVersion: "11271196"
  uid: c03c21d7-fe3e-4407-97ba-88abc6cfc560
spec:
  nic: harvester-mgmt
  nodeName: p5820
  type: vlan
status:
  conditions:
  - lastUpdateTime: "2022-09-08T19:17:00Z"
    status: "True"
    type: Ready
  networkLinkStatus:
    harvester-br0:
      index: 48
      mac: b4:96:91:6a:22:bc
      promiscuous: true
      state: up
      type: bridge
    harvester-mgmt:
      index: 5
      mac: b4:96:91:6a:22:bc
      masterIndex: 48
      state: up
      type: bond
  nics:
  - index: 2
    name: eno1
    state: down
    type: device
  - index: 3
    masterIndex: 5
    name: enp179s0f0
    state: up
    type: device
  - index: 4
    name: enp179s0f1
    state: down
    type: device
  - index: 5
    masterIndex: 48
    name: harvester-mgmt
    state: up
    type: bond
    usedByManagementNetwork: true
    usedByVlanNetwork: true`
)

func Test_NodeNetwork(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(nodeNetwork)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		err = cleanupNodeNetwork(v)
		assert.NoError(err, "expected no error during cleanup up of nodenetwork")
		status, statusOK, err := unstructured.NestedMap(v.Object, "status", "networkLinkStatus")
		assert.NoError(err, "expected no error while looking up ")
		assert.True(statusOK, "expected to find networkLinkStatus")
		for _, v := range status {
			vMap, vOK := v.(map[string]interface{})
			assert.True(vOK, "expected assertion to map[string]interface{} to be successful")
			assert.Equal("sim-generated", vMap["name"], "expected to find name to be sim-generated")
		}
	}
}

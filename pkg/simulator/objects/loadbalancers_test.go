package objects

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"
)

func TestLoadBalancers(t *testing.T) {
	tmpLoadBalancers, err := os.CreateTemp("/tmp", "loadbalancers")
	if err != nil {
		t.Fatalf("error generate temp file for block devices yaml: %v", err)
	}

	defer os.Remove(tmpLoadBalancers.Name())
	_, err = tmpLoadBalancers.Write([]byte(sampleLoadBalancers))
	if err != nil {
		t.Fatalf("error writing to temp file %s: %v", tmpLoadBalancers.Name(), err)
	}

	objs, err := GenerateObjects(tmpLoadBalancers.Name())
	if err != nil {
		t.Fatalf("error reading temp block device file %s %v", tmpLoadBalancers.Name(), err)
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
			t.Fatalf("error during object housekeeping: %v", err)
		}

		// check that there is a name always present in the loadbalancer spec
		listeners, ok, err := unstructured.NestedFieldCopy(unstructObj.Object, "spec", "listeners")
		if err != nil {
			t.Fatalf("error fetching spec.listeners for loadbalancer: %s %v ", unstructObj.GetName(), err)
		}

		if !ok {
			t.Fatalf("could not find spec.listeners for %s \n%v", unstructObj.GetName(), unstructObj)
		}

		listenersList, ok := listeners.([]interface{})
		if !ok {
			t.Fatalf("unable to assert listeners to []interface{}")
		}

		for _, v := range listenersList {
			listenerMap, ok := v.(map[string]interface{})
			if !ok {
				t.Fatalf("unable to assert listener to map[string]interface{}")
			}
			_, ok = listenerMap["name"]
			if !ok {
				t.Fatalf("unable to find key name in the listener backend %v", unstructObj)
			}
		}
	}
}

const sampleLoadBalancers = `
apiVersion: v1
items:
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: 3f730513-2fab-4270-8257-fd7053ed1276
    creationTimestamp: "2022-01-04T17:34:49Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 4
    labels:
      cloudprovider.harvesterhci.io/cluster: bhofmann-test
    name: bhofmann-test-default-nginxlb-f358377f
    namespace: default
    resourceVersion: "10863662"
    uid: 76b7114c-d9a3-4c8b-b4bd-cb150a8547c2
  spec:
    backendServers:
    - 10.65.0.109
    - 10.65.0.108
    - 10.65.0.111
    ipam: pool
    listeners:
    - backendPort: 30437
      name: "null"
      port: 80
      protocol: TCP
  status:
    address: 192.168.200.209
    conditions:
    - lastUpdateTime: "2022-01-04T17:34:49Z"
      status: "True"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: 087fc22f-7e11-4168-a330-e63274752185
    creationTimestamp: "2022-01-13T16:58:11Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 29
    labels:
      cloudprovider.harvesterhci.io/cluster: demo
    name: demo-default-nginx-loadbalancer-26a82b4e
    namespace: default
    resourceVersion: "21588258"
    uid: 4e011d98-b2d0-49e9-8db5-327ca84722a4
  spec:
    backendServers:
    - 10.65.0.128
    - 10.65.0.116
    - 10.65.0.127
    - 10.65.0.130
    - 10.65.0.186
    - 10.65.0.131
    ipam: pool
    listeners:
    - backendPort: 31924
      name: 80tcp80
      port: 80
      protocol: TCP
  status:
    address: 192.168.200.210
    conditions:
    - lastUpdateTime: "2022-01-13T16:58:11Z"
      status: "True"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: f4472ea9-ef2d-4dc6-9a90-002f26aee6a6
    creationTimestamp: "2022-01-06T17:57:00Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 168
    labels:
      cloudprovider.harvesterhci.io/cluster: demo
    name: demo-default-web-loadbalancer-834c8d16
    namespace: default
    resourceVersion: "21588165"
    uid: f5258951-622f-4442-8a0f-153f84448481
  spec:
    backendServers:
    - 10.65.0.116
    - 10.65.0.127
    - 10.65.0.130
    - 10.65.0.186
    - 10.65.0.131
    - 10.65.0.128
    ipam: pool
    listeners:
    - backendPort: 30424
      name: http
      port: 80
      protocol: TCP
  status:
    address: 192.168.200.216
    conditions:
    - lastUpdateTime: "2022-01-06T17:57:00Z"
      status: "True"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: 6d32a2fb-dfdd-4452-adfd-d962c11bf4a1
    creationTimestamp: "2022-01-07T11:12:19Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 124
    labels:
      cloudprovider.harvesterhci.io/cluster: demo
    name: demo-default-web2-ca0ef8ec
    namespace: default
    resourceVersion: "21588167"
    uid: 2150761b-9b62-4369-9a17-4d6c5474f2e3
  spec:
    backendServers:
    - 10.65.0.128
    - 10.65.0.116
    - 10.65.0.127
    - 10.65.0.130
    - 10.65.0.186
    - 10.65.0.131
    ipam: pool
    listeners:
    - backendPort: 30794
      name: http
      port: 80
      protocol: TCP
  status:
    address: 10.65.0.226
    conditions:
    - lastUpdateTime: "2022-01-07T11:12:19Z"
      status: "True"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: db9878d6-9306-44f3-9995-f440262b9074
    creationTimestamp: "2022-01-05T08:48:09Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 124
    labels:
      cloudprovider.harvesterhci.io/cluster: production-cluster
      manager: harvester-load-balancer
      operation: Update
      time: "2022-01-05T08:48:10Z"
    name: production-cluster-default-rke2-ingress-nginx-controller-f46f2e8a
    namespace: default
    resourceVersion: "10061454"
    uid: bac6d74c-f69b-4add-8fce-68b89659e2c7
  spec:
    backendServers:
    - 10.65.0.176
    - 10.65.0.177
    - 10.65.0.174
    ipam: pool
    listeners:
    - backendPort: 32541
      name: httpas
      port: 80
      protocol: TCP
  status:
    conditions:
    - lastUpdateTime: "2022-01-05T08:48:10Z"
      message: 'Service "production-cluster-default-rke2-ingress-nginx-controller-f46f2e8a"
        is invalid: metadata.name: Invalid value: "production-cluster-default-rke2-ingress-nginx-controller-f46f2e8a":
        must be no more than 63 characters'
      status: "False"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: bb8283d1-9da5-40a4-a579-58a5474cc8b6
    creationTimestamp: "2022-01-04T12:13:03Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 6
    labels:
      cloudprovider.harvesterhci.io/cluster: production-cluster
      manager: harvester-load-balancer
      operation: Update
      time: "2022-01-04T12:13:08Z"
    name: production-cluster-kube-system-ingress-325b46e2
    namespace: default
    resourceVersion: "10863588"
    uid: 38bb8f66-16a2-440e-8b6a-e8457da3fac2
  spec:
    backendServers:
    - 10.65.0.176
    - 10.65.0.177
    - 10.65.0.174
    ipam: pool
    listeners:
    - backendPort: 32395
      name: http
      port: 80
      protocol: TCP
    - backendPort: 30360
      name: https
      port: 443
      protocol: TCP
  status:
    address: 192.168.200.210
    conditions:
    - lastUpdateTime: "2022-01-04T12:13:03Z"
      status: "True"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: 0b24d2d2-6fbb-4b14-8d20-28390c52e3b2
    creationTimestamp: "2022-01-05T08:56:40Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 6
    labels:
      cloudprovider.harvesterhci.io/cluster: production-cluster
    name: production-cluster-kube-system-rke2-ingress-1e0fea5a
    namespace: default
    resourceVersion: "10863597"
    uid: b9b1579a-3b53-4237-8cc2-015ef7e26e93
  spec:
    backendServers:
    - 10.65.0.176
    - 10.65.0.177
    - 10.65.0.174
    ipam: pool
    listeners:
    - backendPort: 31069
      name: http
      port: 80
      protocol: TCP
  status:
    address: 192.168.200.212
    conditions:
    - lastUpdateTime: "2022-01-05T08:56:40Z"
      status: "True"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: 4ebf5e46-30e0-47b6-a8e5-1768df7777d8
    creationTimestamp: "2022-01-05T08:55:30Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 147
    labels:
      cloudprovider.harvesterhci.io/cluster: production-cluster
    name: production-cluster-kube-system-rke2-ingress-nginx-controller-54288d09
    namespace: default
    resourceVersion: "10062930"
    uid: 15e970e6-e74b-463c-a5d3-a4f499cebd65
  spec:
    backendServers:
    - 10.65.0.176
    - 10.65.0.177
    - 10.65.0.174
    ipam: dhcp
    listeners:
    - backendPort: 32041
      name: http
      port: 80
      protocol: TCP
  status:
    conditions:
    - lastUpdateTime: "2022-01-05T08:55:30Z"
      message: 'Service "production-cluster-kube-system-rke2-ingress-nginx-controller-54288d09"
        is invalid: metadata.name: Invalid value: "production-cluster-kube-system-rke2-ingress-nginx-controller-54288d09":
        must be no more than 63 characters'
      status: "False"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: b64fe596-36ae-4009-a29a-624eb2233419
    creationTimestamp: "2022-01-07T11:54:43Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 42
    labels:
      cloudprovider.harvesterhci.io/cluster: test-mgmt-net
      manager: harvester-load-balancer
      operation: Update
      time: "2022-01-07T11:54:48Z"
    name: test-mgmt-net-default-web1-af855da4
    namespace: default
    resourceVersion: "16698770"
    uid: 21bc858e-baf3-42c5-b624-83be0aa832b1
  spec:
    backendServers:
    - 192.168.200.192
    - 192.168.200.193
    - 192.168.200.194
    ipam: pool
    listeners:
    - backendPort: 30911
      name: httpo
      port: 80
      protocol: TCP
  status:
    address: 192.168.200.221
    conditions:
    - lastUpdateTime: "2022-01-07T11:54:43Z"
      status: "True"
      type: Ready
- apiVersion: loadbalancer.harvesterhci.io/v1alpha1
  kind: LoadBalancer
  metadata:
    annotations:
      cloudprovider.harvesterhci.io/service-uuid: c480bbc2-eb16-45bc-816c-9dec94a64d9e
    creationTimestamp: "2022-01-04T13:48:21Z"
    finalizers:
    - wrangler.cattle.io/harvester-lb-controller
    generation: 7
    labels:
      cloudprovider.harvesterhci.io/cluster: ubuntu
    name: ubuntu-default-wp-wordpress-7422040c
    namespace: default
    resourceVersion: "10863626"
    uid: 44c03c08-d2fa-4af4-94ae-dacffc2b366b
  spec:
    backendServers:
    - 10.65.0.107
    - 10.65.0.179
    - 10.65.0.178
    ipam: pool
    listeners:
    - backendPort: 31729
      name: http
      port: 80
      protocol: TCP
    - backendPort: 30966
      name: https
      port: 443
      protocol: TCP
  status:
    address: 192.168.200.214
    conditions:
    - lastUpdateTime: "2022-01-04T13:48:21Z"
      status: "True"
      type: Ready
kind: List
metadata:
  continue: "null"
  resourceVersion: "21747550"
`

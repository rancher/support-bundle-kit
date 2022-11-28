package objects

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	skipNamespace = `apiVersion: v1
kind: Namespace
metadata:
  annotations:
    psp.rke2.io/global-restricted: resolved
    psp.rke2.io/global-unrestricted: resolved
    psp.rke2.io/system-unrestricted: resolved
  creationTimestamp: "2022-11-15T19:56:25Z"
  labels:
    kubernetes.io/metadata.name: kube-system
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:kubernetes.io/metadata.name: {}
    manager: kube-apiserver
    operation: Update
    time: "2022-11-15T19:56:25Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .: {}
          f:psp.rke2.io/global-restricted: {}
          f:psp.rke2.io/global-unrestricted: {}
          f:psp.rke2.io/system-unrestricted: {}
    manager: rke2
    operation: Update
    time: "2022-11-15T19:56:28Z"
  name: kube-system
  resourceVersion: "226"
  uid: 376a2d6a-81e0-4d31-90d7-8fd249e45c24
spec:
  finalizers:
  - kubernetes
status:
  phase: Active`

	namespace = `apiVersion: v1
kind: Namespace
metadata:
  annotations:
    helm.sh/resource-policy: keep
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Namespace","metadata":{"annotations":{"helm.sh/resource-policy":"keep","meta.helm.sh/release-name":"rancher-monitoring","meta.helm.sh/release-namespace":"cattle-monitoring-system"},"labels":{"app.kubernetes.io/instance":"rancher-monitoring","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/part-of":"rancher-monitoring","heritage":"Helm","kubernetes.io/metadata.name":"cattle-dashboards","name":"cattle-dashboards","release":"rancher-monitoring"},"name":"cattle-dashboards"}}
    meta.helm.sh/release-name: rancher-monitoring
    meta.helm.sh/release-namespace: cattle-monitoring-system
    objectset.rio.cattle.io/id: default-mcc-rancher-monitoring-cattle-fleet-local-system
  creationTimestamp: "2022-11-15T19:59:20Z"
  labels:
    app.kubernetes.io/instance: rancher-monitoring
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/part-of: rancher-monitoring
    app.kubernetes.io/version: 100.1.0_up19.0.3
    chart: rancher-monitoring-100.1.0_up19.0.3
    heritage: Helm
    kubernetes.io/metadata.name: cattle-dashboards
    name: cattle-dashboards
    objectset.rio.cattle.io/hash: a8c87f2d01731fdad3b988f675a7c8a7da10d382
    release: rancher-monitoring
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .: {}
          f:helm.sh/resource-policy: {}
          f:kubectl.kubernetes.io/last-applied-configuration: {}
          f:meta.helm.sh/release-name: {}
          f:meta.helm.sh/release-namespace: {}
        f:labels:
          .: {}
          f:app.kubernetes.io/instance: {}
          f:app.kubernetes.io/managed-by: {}
          f:app.kubernetes.io/part-of: {}
          f:heritage: {}
          f:kubernetes.io/metadata.name: {}
          f:name: {}
          f:release: {}
    manager: kubectl-client-side-apply
    operation: Update
    time: "2022-11-15T19:59:20Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          f:objectset.rio.cattle.io/id: {}
        f:labels:
          f:app.kubernetes.io/version: {}
          f:chart: {}
          f:objectset.rio.cattle.io/hash: {}
    manager: fleetagent
    operation: Update
    time: "2022-11-15T20:01:28Z"
  name: cattle-dashboards
  resourceVersion: "5223"
  uid: 0113606f-7d42-4698-8c3c-7fab43275350
spec:
  finalizers:
  - kubernetes
status:
  phase: Active`

	skipEndpoint = `apiVersion: v1
kind: Endpoints
metadata:
  creationTimestamp: "2022-11-15T19:56:27Z"
  labels:
    endpointslice.kubernetes.io/skip-mirror: "true"
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:endpointslice.kubernetes.io/skip-mirror: {}
      f:subsets: {}
    manager: kube-apiserver
    operation: Update
    time: "2022-11-15T19:56:27Z"
  name: kubernetes
  namespace: default
  resourceVersion: "158058"
  uid: 3463cc4f-725c-4db5-9aca-5f61ca37f625
subsets:
- addresses:
  - ip: 192.168.3.21
  - ip: 192.168.3.22
  - ip: 192.168.3.23
  ports:
  - name: https
    port: 6443
    protocol: TCP`

	endpoint = `
apiVersion: v1
kind: Endpoints
metadata:
  annotations:
    endpoints.kubernetes.io/last-change-trigger-time: "2022-11-15T22:29:44Z"
  creationTimestamp: "2022-11-15T19:59:20Z"
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .: {}
          f:endpoints.kubernetes.io/last-change-trigger-time: {}
      f:subsets: {}
    manager: kube-controller-manager
    operation: Update
    time: "2022-11-15T20:55:00Z"
  name: harvester-cluster-repo
  namespace: cattle-system
  resourceVersion: "159827"
  uid: bfc304ec-f694-4617-b469-c7cd419f804a
subsets:
- addresses:
  - ip: 10.52.0.122
    nodeName: harv1
    targetRef:
      kind: Pod
      name: harvester-cluster-repo-5db86f484f-wqzk4
      namespace: cattle-system
      uid: fc61effe-10b3-440e-a5ae-124328ff38cd
  ports:
  - port: 80
    protocol: TCP`

	skipService = `apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2022-11-15T19:56:27Z"
  labels:
    component: apiserver
    provider: kubernetes
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:component: {}
          f:provider: {}
      f:spec:
        f:clusterIP: {}
        f:internalTrafficPolicy: {}
        f:ipFamilyPolicy: {}
        f:ports:
          .: {}
          k:{"port":443,"protocol":"TCP"}:
            .: {}
            f:name: {}
            f:port: {}
            f:protocol: {}
            f:targetPort: {}
        f:sessionAffinity: {}
        f:type: {}
    manager: kube-apiserver
    operation: Update
    time: "2022-11-15T19:56:27Z"
  name: kubernetes
  namespace: default
  resourceVersion: "199"
  uid: 34518def-1f0a-4773-a69d-f1a9856c3115
spec:
  clusterIP: 10.53.0.1
  clusterIPs:
  - 10.53.0.1
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 6443
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}`

	service = `apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"name":"harvester-cluster-repo","namespace":"cattle-system"},"spec":{"ports":[{"port":80,"protocol":"TCP","targetPort":80}],"selector":{"app":"harvester-cluster-repo"}}}
  creationTimestamp: "2022-11-15T19:59:20Z"
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .: {}
          f:kubectl.kubernetes.io/last-applied-configuration: {}
      f:spec:
        f:internalTrafficPolicy: {}
        f:ports:
          .: {}
          k:{"port":80,"protocol":"TCP"}:
            .: {}
            f:port: {}
            f:protocol: {}
            f:targetPort: {}
        f:selector: {}
        f:sessionAffinity: {}
        f:type: {}
    manager: kubectl-client-side-apply
    operation: Update
    time: "2022-11-15T19:59:20Z"
  name: harvester-cluster-repo
  namespace: cattle-system
  resourceVersion: "2238"
  uid: d0134ff3-071b-463e-958a-831316daa95c
spec:
  clusterIP: 10.53.35.228
  clusterIPs:
  - 10.53.35.228
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: harvester-cluster-repo
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}`

	skipPriorityClass = `  apiVersion: scheduling.k8s.io/v1
  description: Used for system critical pods that must run in the cluster, but can
    be moved to another node if necessary.
  kind: PriorityClass
  metadata:
    creationTimestamp: "2022-11-15T19:56:26Z"
    generation: 1
    managedFields:
    - apiVersion: scheduling.k8s.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:description: {}
        f:preemptionPolicy: {}
        f:value: {}
      manager: kube-apiserver
      operation: Update
      time: "2022-11-15T19:56:26Z"
    name: system-cluster-critical
    resourceVersion: "77"
    uid: 455a2ac0-8f04-4e0b-afb5-d3ae80e45fe5
  preemptionPolicy: PreemptLowerPriority
  value: 2e+09`

	priorityClass = `apiVersion: scheduling.k8s.io/v1
description: This priority class should be used for core kubevirt components only.
kind: PriorityClass
metadata:
  annotations:
    meta.helm.sh/release-name: harvester
    meta.helm.sh/release-namespace: harvester-system
    objectset.rio.cattle.io/id: default-mcc-harvester-cattle-fleet-local-system
  creationTimestamp: "2022-11-15T20:00:39Z"
  generation: 1
  labels:
    app.kubernetes.io/component: operator
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: virt-operator
    app.kubernetes.io/part-of: kubevirt-operator
    app.kubernetes.io/version: 0.30.x
    helm.sh/chart: kubevirt-operator-0.1.0
    helm.sh/release: harvester
    objectset.rio.cattle.io/hash: e852fa897f5eae59a44b4bfe186aad80b10b94b3
  managedFields:
  - apiVersion: scheduling.k8s.io/v1
    fieldsType: FieldsV1
    fieldsV1:
      f:description: {}
      f:metadata:
        f:annotations:
          .: {}
          f:meta.helm.sh/release-name: {}
          f:meta.helm.sh/release-namespace: {}
          f:objectset.rio.cattle.io/id: {}
        f:labels:
          .: {}
          f:app.kubernetes.io/component: {}
          f:app.kubernetes.io/managed-by: {}
          f:app.kubernetes.io/name: {}
          f:app.kubernetes.io/part-of: {}
          f:app.kubernetes.io/version: {}
          f:helm.sh/chart: {}
          f:helm.sh/release: {}
          f:objectset.rio.cattle.io/hash: {}
      f:preemptionPolicy: {}
      f:value: {}
    manager: fleetagent
    operation: Update
    time: "2022-11-15T20:00:39Z"
  name: kubevirt-cluster-critical
  resourceVersion: "4234"
  uid: 64666307-10ca-431a-9fcc-f0e741d3c51d
preemptionPolicy: PreemptLowerPriority
value: 1e+09`
)

func Test_SkipNamespace(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(skipNamespace)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.True(skipResources(v), "expected namespace to be skipped")
	}
}

func Test_Namespace(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(namespace)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.False(skipResources(v), "expected namespace to not be skipped")
	}
}

func Test_SkipEndpoint(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(skipEndpoint)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.True(skipResources(v), "expected endpoint to be skipped")
	}
}

func Test_Endpoint(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(endpoint)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.False(skipResources(v), "expected endpoint to not be skipped")
	}
}

func Test_SkipService(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(skipService)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.True(skipResources(v), "expected service to be skipped")
	}
}

func Test_Service(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(service)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.False(skipResources(v), "expected service to not be skipped")
	}
}

func Test_SkipPriorityClass(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(skipPriorityClass)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.True(skipResources(v), "expected priorityclass to be skipped")
	}
}

func Test_PriorityClass(t *testing.T) {
	assert := require.New(t)
	objs, err := GenerateUnstructuredObjectsFromString(priorityClass)
	assert.NoError(err, "expected no error during object generation")
	for _, v := range objs {
		assert.False(skipResources(v), "expected priorityclass to not be skipped")
	}
}

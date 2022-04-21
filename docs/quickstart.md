# support-bundle-kit simulator quick start guide

To get started with the simulator capability, download a harvester support bundle zip file.

For example:

```
drwxr-xr-x   5 gaurav gaurav     4096 Apr 19 18:19  supportbundle_101310ea-583b-4191-b33e-d265002214ea_2022-04-19T08-18-38Z
-rw-rw-r--   1 gaurav gaurav  5828575 Apr 19 18:19  supportbundle_101310ea-583b-4191-b33e-d265002214ea_2022-04-19T08-18-38Z.zip
```

change directory to the unzipped support bundle directory and run `support-bundle-kit simulator`

```shell
(⎈ |default:default)➜  supportbundle_101310ea-583b-4191-b33e-d265002214ea_2022-04-19T08-18-38Z support-bundle-kit simulator 
INFO[0000] Creating embedded etcd server                
{"level":"warn","ts":"2022-04-21T11:25:40.672+1000","caller":"etcdserver/util.go:121","msg":"failed to apply request","took":"656.527µs","request":"header:<ID:16984622604441375451 > lease_revoke:<id:6bb5804535cc52be>","response":"size:30","error":"lease not found"}
I0421 11:25:42.462055  688330 server.go:629] external host was not specified, using 192.168.1.129
W0421 11:25:42.462108  688330 authentication.go:507] AnonymousAuth is not allowed with the AlwaysAllow authorizer. Resetting AnonymousAuth to false. You should use a different authorizer
```

This will trigger the bootstrap of the simulator components and load the contents of the supportbundle into the simulator apiserver.

For `kubectl logs` to work, the `--bundle-path` should point to the correct bundle path, the default is `.`

The state of the simulator is stored in `$HOME/.sim`, where the users can find an `admin.kubeconfig`

This can now be used to interact with the cluster.

As part of the supportbundle collection, the bundle contains a ${nodename}.zip file for each node in the cluster in the `$SUPPORTBUNDLE/nodes` directory. 

The simulator will also load the contents of the same into a CRD named `NodeConfig` in `support-bundle-node-info` namespace.

The CRD spec is as follows:

```json

type NodeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              []NodeConfigSpec `json:"spec"`
}

type NodeConfigSpec struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}
```

The CRD contains the content of all config files present in the `$NODENAME.zip`.

In addition, the simulator will also create pods corresponding to each node in the cluster in the `support-bundle-node-info` namespace. Each system log file is represented by a container in the pod, and this allows the user to browse the node logs via kubectl as well.

For example:
```shell
(⎈ |default:default)➜  ~ kubectl get pods -n support-bundle-node-info
NAME              READY   STATUS    RESTARTS   AGE
harvester-pxe-1   8/8     Running   0          41h
harvester-pxe-2   8/8     Running   0          41h

```

A quick description of the pods shows the following:
```
(⎈ |default:default)➜  ~ kubectl describe pod harvester-pxe-1 -n support-bundle-node-info

Name:         harvester-pxe-1
Namespace:    support-bundle-node-info
Priority:     0
Node:         harvester-pxe-1/
Start Time:   Thu, 21 Apr 2022 11:32:47 +1000
Labels:       <none>
Annotations:  <none>
Status:       Running
IP:           
IPs:          <none>
Containers:
console:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
rancher-system-agent:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
wicked:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
rancherd:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
rke2-agent:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
kernel:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
rke2-server:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
iscsid:
Container ID:   
Image:          noimage
Image ID:       
Port:           <none>
Host Port:      <none>
State:          Running
Started:      Thu, 21 Apr 2022 11:32:47 +1000
Ready:          True
Restart Count:  0
Environment:    <none>
Mounts:
/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-prj4k (ro)
Conditions:
Type              Status
Ready             True
PodScheduled      True
Initialized       True
ContainersReady   True
Volumes:
kube-api-access-prj4k:
Type:                    Projected (a volume that contains injected data from multiple sources)
TokenExpirationSeconds:  3607
ConfigMapName:           kube-root-ca.crt
ConfigMapOptional:       <nil>
DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:                      <none>
```

Users can now browse the node logs as follows:

```shell
(⎈ |default:default)➜  ~ kubectl  logs harvester-pxe-1 -c iscsid -n support-bundle-node-info 
-- Logs begin at Tue 2022-04-19 06:59:30 UTC, end at Tue 2022-04-19 08:19:00 UTC. --
Apr 19 06:59:30 localhost systemd[1]: Starting Open-iSCSI...
Apr 19 06:59:30 localhost systemd[1]: Started Open-iSCSI.
Apr 19 06:59:34 harvester-pxe-1 systemd[1]: Stopping Open-iSCSI...
Apr 19 06:59:34 harvester-pxe-1 iscsid[332]: iscsid shutting down.
Apr 19 06:59:34 harvester-pxe-1 systemd[1]: iscsid.service: Succeeded.
Apr 19 06:59:34 harvester-pxe-1 systemd[1]: Stopped Open-iSCSI.
Apr 19 07:03:55 harvester-pxe-1 systemd[1]: Starting Open-iSCSI...
Apr 19 07:03:55 harvester-pxe-1 systemd[1]: Started Open-iSCSI.
Apr 19 07:03:55 harvester-pxe-1 iscsid[23921]: iscsid: Connection1:0 to [target: iqn.2019-10.io.longhorn:pvc-164332e8-99ee-4acc-8439-9509074886f9, portal: 10.52.0.36,3260] through [iface: default] is operational now
Apr 19 07:03:58 harvester-pxe-1 iscsid[23921]: iscsid: Connection2:0 to [target: iqn.2019-10.io.longhorn:pvc-d77f3f5b-24cf-4afe-84d5-b109007422d5, portal: 10.52.0.36,3260] through [iface: default] is operational now

```
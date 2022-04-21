# Support Bundle Kit [![Build Status](http://drone-publish.rancher.io/api/badges/rancher/support-bundle-kit/status.svg)](http://drone-publish.rancher.io/rancher/support-bundle-kit)

**working in progress**

This project contains support bundle scripts and utilities for applications running on top of Kubernetes

- `harvester-sb-collector.sh`: The script is used to collect k3os node logs. It can be run in a container with the host log folder mapped or be run on the host directly.
- `support-bundle-utils`: This application contains several commands:
  - `manager`: start a support bundle kit manager. The manager does these works:
    - It collects the cluster bundle, including YAML manifests and pod logs.
    - It collects external bundles. e.g., Longhorn support bundle.
    - It starts a web server and waits for bundle downloading and uploading.
    - It starts a daemonset on each node. The agents in the daemonset collect node bundles and push them back to the manager.

    The manager is designed to be spawned as a Kubernetes deployment by the application. But it can also be deployed manually from a manifest file. Please check [standalone mode](./docs/standalone.md) for more information.
  - `simulator`: the command allows users to simulate an end user environment by loading the support bundle into a minimal apiserver allowing end users to browse the objects and logs from the support bundle. It will do the following things
    - It runs an embedded etcd server
    - It runs a minimal apiserver only
    - It runs a minimal kubelet
  

## Support bundle contents

The Harvester support bundle is structured as the following layout:

```yaml
- [logs]            # pod logs, organized by namespaces
  - [namespace1]
    - [pod1]
     - container1.log
    - [pod2]
  - [namespace2]
    - [pod1]

- [yamls]           # definition of resources
  - [cluster]        # cluster scope
    - [kubernetes]    # Kubernetes resources
      - nodes.yaml
      - volumeattachments.yaml
      - nodemetrics.yaml
    - [harvester]     # Harvester custom resources
      - settings.yaml
      - users.yaml
  - [namespaced]     # namespaced scope
    - [default]       # namespace `default`
      - [kubernetes]   # Kubernetes resources
        - pods.yaml
        - jobs.yaml
        - ...
      - [harvester]    # Harvester custom resources
        - keypairs.yaml
        - virtualmachineimages.yaml
        - ...
      - [cdi]          # cdi.kubevirt.io custom resources
        - datavolumes.yaml
      - [kubevirt]     # kubevirt.io custom resources
        - virtualmachines.yaml
        - virtualmachineinstancemigrations.yaml
        - ...
    - [harvester-system]
        - ...
    - [kube-system]
        - ...
    - [cattle-system]
        - ...

- [external]        # External support bundles
  - longhorn-support-bundle_d2f32c7f-6605-4a3b-8571-521856e64233_2021-05-05T03-28-37Z.zip

- [nodes1]          # Node support bundles
  - node1.zip
  - node2.zip
  - ...
```

## Simulator command

The simulator command loads the contents of the support bundle into the apiserver, and updates the status of objects to reflect the contents of the support bundle.

The bundle-path should be the location of the extracted support bundle

The simulator command expects the following flags

```
Usage:
  support-bundle-kit simulator [flags]

Flags:
      --bundle-path string   location to support bundle. default is . (default ".")
  -h, --help                 help for simulator
      --reset                reset sim-home, will clear the contents and start a clean etcd + apiserver instance
      --sim-home string      default home directory where sim stores its configuration. default is $HOME/.sim (default "$HOME/.sim")
      --skip-load            skip load / re-load of bundle. this will ensure current etcd contents are only accessible

```

Known Issues: 
The following are known issues with the simulator at the moment:
* creationTimepstamps of objects are reset, however the original creationTimestamp is copied into the object annoations.
* to enable log parsing the node addresses are updated to localhost, to point them to the in process kubelet. The original addresses are again copied into annotations for future reference.
* APIServices are skipped during the load processes as api aggregation cannot be replicated at the moment.

Additional information can be found in the [QuickStart Guide](./docs/quickstart.md)
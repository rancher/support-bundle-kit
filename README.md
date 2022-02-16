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

# Harvester support bundle utils

This project contains support bundle scripts and utilities for Harvester.

- `bin/harvester-sb-collector.sh`: The script is used to collect k3os node logs. It can be run in a container with host log folder mapped or be run on host.
- `support-bundle-utils`: This application contains serveral commands:
  - `manager`: start a Harvester support bundle manager. The manager does these works: 
    - It collects cluster bundle, including YAML manifests and pod logs.
    - It collects external bundles. e.g., Longhorn support bundle.
    - It starts a web server and wait for bundle downloading and uploading.
    - It starts a daemonset on each node. The agents in the daemonset collect node bundles and push them back to the manager.



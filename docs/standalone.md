# Standalone mode

To deploy the support bundle manager manually, please use the [sample manifest file](../deploy/manifests/support-bundle-manager.yaml).


Here is an example:

```
(with a Harvester Kubernetes cluster context)

$ wget https://raw.githubusercontent.com/harvester/support-bundle-utils/master/deploy/manifests/support-bundle-manager.yaml

# Edit the file if needed, the user might want to change the supportbundle name (default: "sample").

# Create the manager deployment and service
$ kubectl create -f support-bundle-manager.yaml
service/support-bundle-sample created


# Check if the supportbundle file is generated
$ kubectl logs -n harvester-system deployments/supportbundle-manager-sample
...
time="2021-05-24T04:40:56Z" level=info msg="succeed to run phase packaging. Progress (80)."
time="2021-05-24T04:40:56Z" level=info msg="running phase done"
time="2021-05-24T04:40:56Z" level=info msg="support bundle /tmp/harvester-support-bundle/harvester-supportbundle_2d3a9c33-e6c3-4c56-b747-3272326374ba_2021-05-24T04-40-38Z.zip ready to download"
time="2021-05-24T04:40:56Z" level=info msg="succeed to run phase done. Progress (100)."


# Get the supportbundle manager pod:
$ kubectl get pods --selector "app=support-bundle-manager" -n harvester-system
NAME                                            READY   STATUS    RESTARTS   AGE
supportbundle-manager-sample-8477487f46-2lnvq   1/1     Running   0          3m15s

# Copy out the bundle to `/tmp/bundle.zip`
$ kubectl cp harvester-system/supportbundle-manager-sample-8477487f46-2lnvq:/tmp/harvester-support-bundle/harvester-supportbundle_2d3a9c33-e6c3-4c56-b747-3272326374ba_2021-05-24T04-40-38Z.zip /tmp/bundle.zip
```

The bundle is copied and the user can delete the sample deployment to free up resources:

```
$ kubectl delete -f support-bundle-manager.yaml
```

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
deployment.apps/supportbundle-manager-sample created


# Check if the supportbundle file is generated
$ kubectl logs -n harvester-system deployments/supportbundle-manager-sample
...
time="2021-05-05T06:31:50Z" level=info msg="all node bundles are received."
time="2021-05-05T06:31:50Z" level=info msg="support bundle /tmp/harvester-support-bundle/harvester-supportbundle_6c3992c4-2c8c-4444-9cf3-4422f57d88dd_2021-05-05T06-31-31Z.zip ready for downloading"


# Get the service cluster IP
$ kubectl get service support-bundle-sample -n harvester-system
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
support-bundle-sample   ClusterIP   10.53.130.149   <none>        8080/TCP   3m59s


# Download the bundle. The service is exported with ClusterIP type, please download it on one of Harvester nodes.
$ wget http://10.53.130.149:8080/bundle -O bundle.zip
Connecting to 10.53.130.149:8080 (10.53.130.149:8080)
saving to 'bundle.zip'
bundle.zip           100% |*********************************************| 2122k  0:00:00 ETA
'bundle.zip' saved
```

**NOTE**: Because the service doesn't have any access control, the user is encouraged to delete the sample deployment after downloading the bundle. e.g., 

```
$ kubectl delete -f support-bundle-manager.yaml
```

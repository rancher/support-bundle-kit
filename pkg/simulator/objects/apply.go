package objects

import (
	"context"
	"fmt"
	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"time"
)

type ObjectManager struct {
	ctx    context.Context
	path   string
	config *rest.Config
	kc     *kubernetes.Clientset
	dc     dynamic.Interface
}

const (
	simCreationTimeStamp = "sim.harvesterhci.io/creationTimestamp"
	simLabelPrefix       = "sim.harvesterhci.io/"
)

// currently we are skipping certain objects during bundle processing
// this map helps speed up the process and is easier to maintain
var (
	skippedGroups = map[string]bool{
		"events.k8s.io":                true,
		"admissionregistration.k8s.io": true,
		"apiregistration.k8s.io":       true,
		"metrics.k8s.io":               true,
	}

	skippedKinds = map[string]bool{
		"ComponentStatus": true,
	}
)

// NewObjectManager is a wrapper around apply and support bundle path
func NewObjectManager(ctx context.Context, config *rest.Config, path string) (*ObjectManager, error) {

	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kc, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &ObjectManager{
		ctx:    ctx,
		path:   path,
		config: config,
		kc:     kc,
		dc:     dclient,
	}, nil
}

// CreateUnstructuredClusterObjects will use the dynamic client to create all
// cluster scoped objects from the support bundle
func (o *ObjectManager) CreateUnstructuredClusterObjects() error {
	crds, clusterObjs, err := GenerateClusterScopedRuntimeObjects(o.path)
	if err != nil {
		return err
	}

	// apply CRDs first
	err = o.ApplyObjects(crds, false, nil)
	if err != nil {
		return err
	}

	err = o.ApplyObjects(clusterObjs, true, nil)

	if err != nil {
		return err
	}
	return nil
}

// CreateUnstructuredObjects will use the dynamic client to create all
// namespace scoped objects from the support bundle
func (o *ObjectManager) CreateUnstructuredObjects() error {
	nonpods, pods, err := GenerateNamespacedRuntimeObjects(o.path)
	if err != nil {
		return err
	}

	// apply non pods first
	err = o.ApplyObjects(nonpods, true, nil)
	if err != nil {
		return err
	}

	err = o.ApplyObjects(pods, true, nil)
	if err != nil {
		return err
	}
	return nil
}

// ApplyObjects is a wrapper to convert runtime.Objects to unstructured.Unstructured, perform some housekeeping before submitting the same to apiserver
func (o *ObjectManager) ApplyObjects(objs []runtime.Object, patchStatus bool, skipGVR *schema.GroupVersionResource) error {
	var dr dynamic.ResourceInterface
	var resp *unstructured.Unstructured
	for _, v := range objs {
		unstructuredObj, err := wranglerunstructured.ToUnstructured(v)
		if err != nil {
			return fmt.Errorf("error converting obj to unstructured %v", err)
		}

		// skip objects that dont need to be processed //

		if skippedGroups[unstructuredObj.GroupVersionKind().Group] || skippedKinds[unstructuredObj.GetKind()] {
			continue
		}

		err = cleanupObjects(unstructuredObj.Object)
		if err != nil {
			return err
		}

		//GVK specific cleanup needed before objects can be created
		err = objectHousekeeping(unstructuredObj)
		if err != nil {
			return fmt.Errorf("error during housekeeping on objects %v", err)
		}

		restMapping, err := findGVR(v.GetObjectKind().GroupVersionKind(), o.config)
		if err != nil {
			return fmt.Errorf("error looking up GVR %v for object %v", err, unstructuredObj)
		}

		if skipGVR != nil && restMapping.Resource == *skipGVR {
			continue
		}
		if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
			dr = o.dc.Resource(restMapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dr = o.dc.Resource(restMapping.Resource)
		}

		resp, err = dr.Get(o.ctx, unstructuredObj.GetName(), metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				resp, err = dr.Create(o.ctx, unstructuredObj, metav1.CreateOptions{})
				if err != nil {
					logrus.Errorf("error during creation of resource %s with gvr %s", unstructuredObj.GetName(), restMapping.Resource.String())
					logrus.Error(unstructuredObj.Object)
					return fmt.Errorf("error during creation of unstructured resource %v", err)
				}
			} else {
				return fmt.Errorf("error looking up object before creating the same %v", err)
			}
		}

		if patchStatus {
			// we will patch the status here later
			status, ok, err := unstructured.NestedFieldCopy(unstructuredObj.Object, "status")
			if err != nil {
				return fmt.Errorf("error looking for status field: %v", err)
			}
			if ok {
				/*processedObj, err := dr.Get(o.ctx, resp.GetName(), metav1.GetOptions{})
				if err != nil && !apierrors.IsNotFound(err) {
					return fmt.Errorf("error looking up resource %s with gvr %v with error %v", unstructuredObj.GetName(), unstructuredObj.GroupVersionKind(), err)
				}*/
				unstructured.SetNestedField(resp.Object, status, "status")
				_, err = dr.UpdateStatus(o.ctx, resp, metav1.UpdateOptions{})
				// update sometimes returns this object not found error
				// the 404 lookup is to try and work around the same.
				if err != nil && !apierrors.IsNotFound(err) {
					return fmt.Errorf("error updating status on resource %s with gvr %v with error %v", resp.GetName(), resp.GroupVersionKind(), err)
				}
			}

		}
	}
	return nil
}

// objectHousekeeping will perform some common housekeeping tasks based on GVR
// needed to keep the apiserver happy since we are dealing with exported CRDs
func objectHousekeeping(obj *unstructured.Unstructured) error {
	// Common housekeeping performed on all objects
	// need to clear resource version before apply.
	// this will be added as an annotation
	annotations, ok, err := unstructured.NestedStringMap(obj.Object, "metadata", "annotations")
	if err != nil {
		return err
	}
	if !ok {
		annotations = make(map[string]string)
	}

	orgCreationTimestamp, ok, err := unstructured.NestedString(obj.Object, "metadata", "creationTimestamp")
	if err != nil {
		return err
	}
	if ok {
		annotations[simCreationTimeStamp] = orgCreationTimestamp
		unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
		err = unstructured.SetNestedStringMap(obj.Object, annotations, "metadata", "annotations")
		if err != nil {
			return err
		}
	}

	switch obj.GetKind() {
	case "Ingress":
		// Ingress specific housekeeping
		err = ingressCleanup(obj)
	case "Job", "Batch":
		err = jobCleanup(obj)
	case "APIService":
		err = apiServiceCleanup(obj)
	case "Node":
		err = nodeCleanup(obj)
	case "LoadBalancer":
		err = loadBalancerCleanup(obj)
	case "BlockDevice":
		err = blockDevicesCleanup(obj)
	}
	return err
}

// wrapper to lookup GVR for usage with dynamic client
func findGVR(gvk schema.GroupVersionKind, cfg *rest.Config) (*meta.RESTMapping, error) {

	// DiscoveryClient queries API server about the resources
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}

//verifyObj is a helper method used to verify objects to make it easier to test
func verifyObject(obj *unstructured.Unstructured, fn func(interface{}) (bool, error), keys ...string) (ok bool, err error) {
	tmpObj, ok, err := unstructured.NestedFieldCopy(obj.Object, keys...)
	// if not found or no key present
	if err != nil || !ok {
		return ok, err
	}

	if fn != nil {
		return fn(tmpObj)
	}

	return ok, err
}

// cleanupObjects will clean up all "null" strings that appear in
// support bundles.
func cleanupObjects(obj map[string]interface{}) error {
	// key: null is a valid value in prometheusrules, hence that is ignored from this cleanup
	for key, value := range obj {
		if v, ok := value.(string); ok && v == "null" && key != "key" {
			delete(obj, key)
		}

		if _, ok := value.([]string); ok {
			continue
		}

		if key == "resourceVersion" {
			delete(obj, key)
		}

		if valArr, ok := value.([]interface{}); ok {
			var newArr []interface{}
			for _, v := range valArr {
				newMap, innerOK := v.(map[string]interface{})
				if innerOK {
					err := cleanupObjects(newMap)
					if err != nil {
						return err
					}
					newArr = append(newArr, newMap)
				}

			}
			if len(newArr) != 0 {
				obj[key] = newArr
			}
		}

		if valMap, ok := value.(map[string]interface{}); ok {
			err := cleanupObjects(valMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// FetchObject will use the dynamic client to fetch runtime.Object from apiserver.
func (o *ObjectManager) FetchObject(obj runtime.Object) (*unstructured.Unstructured, error) {
	var dr dynamic.ResourceInterface
	unstructObj, err := wranglerunstructured.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	restMapping, err := findGVR(unstructObj.GroupVersionKind(), o.config)
	if err != nil {
		return nil, err
	}
	if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = o.dc.Resource(restMapping.Resource).Namespace(unstructObj.GetNamespace())
	} else {
		dr = o.dc.Resource(restMapping.Resource)
	}

	return dr.Get(o.ctx, unstructObj.GetName(), metav1.GetOptions{})
}

// WaitForNamespaces ensures apiserver is ready and namespaces can be listed before it times out
func (o *ObjectManager) WaitForNamespaces(timeout time.Duration) error {
	now := time.Now()
	for currtime := now; currtime.Before(now.Add(timeout)); {
		// ignore errors
		ns, _ := o.kc.CoreV1().Namespaces().List(o.ctx, metav1.ListOptions{})
		if ns != nil && len(ns.Items) != 0 {
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timed out waiting for apiserver to be ready")
}

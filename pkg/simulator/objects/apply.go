package objects

import (
	"context"
	"fmt"
	"sync"
	"time"

	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	supportbundlekit "github.com/rancher/support-bundle-kit/pkg/simulator/apis/supportbundlekit.io/v1"
	"github.com/rancher/support-bundle-kit/pkg/simulator/crd"
)

type ProgressHandler func(int, int)

type ObjectManager struct {
	ctx        context.Context
	path       string
	config     *rest.Config
	kc         *kubernetes.Clientset
	dc         dynamic.Interface
	failedObjs []supportbundlekit.FailedObjectSpec
}

const (
	simCreationTimeStamp = "sim.harvesterhci.io/creationTimestamp"
	simLabelPrefix       = "sim.harvesterhci.io/"
)

// currently we are skipping certain objects during bundle processing
// this map helps speed up the process and is easier to maintain
var (
	skippedGroups = map[string]bool{
		"admissionregistration.k8s.io": true,
		"apiregistration.k8s.io":       true,
		"metrics.k8s.io":               true,
	}

	skippedKinds = map[string]bool{
		"ComponentStatus":   true,
		"PodSecurityPolicy": true,
	}

	cacheMap  = make(map[schema.GroupVersionKind]*meta.RESTMapping)
	cacheLock = new(sync.Mutex)
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

	progressMgr := NewProgressManager("Step 1/4: Cluster CRDs")

	// add simulator CRDs
	simCRDs, err := crd.Objects(false)
	if err != nil {
		return fmt.Errorf("error generating Simulator CRD objects: %v", err)
	}
	crds = append(crds, simCRDs...)

	// apply CRDs first
	err = o.ApplyObjects(crds, false, nil, progressMgr.progress)
	if err != nil {
		return err
	}

	// TODO: check all CRDs are created
	logrus.Info("Sleeping for 5 seconds before applying cluster objects")
	time.Sleep(5 * time.Second)
	logrus.Info("Sleeping done")

	progressMgr = NewProgressManager("Step 1/4: Cluster objects")
	err = o.ApplyObjects(clusterObjs, true, nil, progressMgr.progress)

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

	progressMgr := NewProgressManager("Step 2/4: Namespaced non-pods")
	// apply non pods first
	err = o.ApplyObjects(nonpods, true, nil, progressMgr.progress)
	if err != nil {
		return err
	}

	progressMgr = NewProgressManager("Step 2/4: Namespaced pods and events")
	err = o.ApplyObjects(pods, true, nil, progressMgr.progress)
	if err != nil {
		return err
	}
	return nil
}

// ApplyObjects is a wrapper to convert runtime.Objects to unstructured.Unstructured, perform some housekeeping before submitting the same to apiserver
func (o *ObjectManager) ApplyObjects(objs []runtime.Object, patchStatus bool, skipGVR *schema.GroupVersionResource, progressHandler ProgressHandler) error {
	var dr dynamic.ResourceInterface
	var resp *unstructured.Unstructured
	total := len(objs)
	for i, v := range objs {
		if progressHandler != nil {
			progressHandler(i+1, total)
		}

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
			return fmt.Errorf("error during housekeeping on objects %v, error: %v", unstructuredObj, err)
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

		var skipPatchStatus bool
		resp, err = dr.Create(o.ctx, unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				resp, err = dr.Get(o.ctx, unstructuredObj.GetName(), metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("error looking up object %v", err)
				}
			} else {
				logrus.WithError(err).Errorf("error during creation of resource %s with gvr %s", unstructuredObj.GetName(), restMapping.Resource.String())
				o.addToFailedObjects(unstructuredObj, err)
				// no need to patch status when object errors out
				skipPatchStatus = true
			}
		}

		if patchStatus && !skipPatchStatus {
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
				err = unstructured.SetNestedField(resp.Object, status, "status")
				if err != nil {
					return err
				}
				_, err = dr.UpdateStatus(o.ctx, resp, metav1.UpdateOptions{})
				// update sometimes returns this object not found error
				// the 404 lookup is to try and work around the same.
				if err != nil && !apierrors.IsNotFound(err) {
					o.addToFailedObjects(unstructuredObj, err)
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

	metadataMap, ok, err := unstructured.NestedMap(obj.Object, "metadata")
	if err != nil {
		return err
	}

	if ok {
		orgCreationTimestamp, innerOK := metadataMap["creationTimestamp"]
		if innerOK && orgCreationTimestamp != nil {
			annotations[simCreationTimeStamp] = orgCreationTimestamp.(string)
			unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
			err = unstructured.SetNestedStringMap(obj.Object, annotations, "metadata", "annotations")
			if err != nil {
				return err
			}
		}
	}

	switch obj.GetKind() {
	case "Job", "Batch":
		err = jobCleanup(obj)
	case "APIService":
		err = apiServiceCleanup(obj)
	case "LoadBalancer":
		err = loadBalancerCleanup(obj)
	case "BlockDevice":
		err = blockDevicesCleanup(obj)
	case "Secret":
		err = cleanupSecret(obj)
	case "Event":
		err = cleanupEvent(obj)
	case "Ingress":
		err = cleanupIngress(obj)
	case "CustomResourceDefinition":
		err = cleanupCRDConversion(obj)
	}
	return err
}

// wrapper to lookup GVR for usage with dynamic client
func findGVR(gvk schema.GroupVersionKind, cfg *rest.Config) (*meta.RESTMapping, error) {

	cacheLock.Lock()
	defer cacheLock.Unlock()
	existingMapping, ok := cacheMap[gvk]
	if ok {
		return existingMapping, nil
	}
	// DiscoveryClient queries API server about the resources
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	newMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	cacheMap[gvk] = newMapping
	return newMapping, nil
}

// verifyObj is a helper method used to verify objects to make it easier to test
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
	err := wait.PollUntilContextTimeout(o.ctx, 5*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		ns, _ := o.kc.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if ns != nil && len(ns.Items) != 0 {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("timed out waiting for apiserver to be ready: %v", err)
	}

	return nil
}

func (o *ObjectManager) CreatedFailedObjectsList() error {
	failedObject := supportbundlekit.FailedObject{
		ObjectMeta: metav1.ObjectMeta{
			Name: "failedobjects",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "FailedObject",
			APIVersion: "supportbundlekit.io/v1",
		},
	}
	progressMgr := NewProgressManager("Step 4/4: CreatedFailedObjectsList")
	failedObject.Spec = o.failedObjs
	return o.ApplyObjects([]runtime.Object{&failedObject}, false, nil, progressMgr.progress)
}

func (o *ObjectManager) addToFailedObjects(obj *unstructured.Unstructured, err error) {
	fObj := supportbundlekit.FailedObjectSpec{
		GVK:       obj.GroupVersionKind().String(),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Error:     err.Error(),
	}

	o.failedObjs = append(o.failedObjs, fObj)
}

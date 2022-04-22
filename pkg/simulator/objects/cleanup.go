package objects

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
	"time"
)

// ingressCleanup object specific cleanup
func ingressCleanup(obj *unstructured.Unstructured) error {
	obj.SetAPIVersion("networking.k8s.io/v1")
	return nil
}

// jobCleanup performs job specific cleanup
func jobCleanup(obj *unstructured.Unstructured) error {
	labels := obj.GetLabels()
	copyKeyValue(labels, "controller-uid", simLabelPrefix+"controller-uid")
	obj.SetLabels(labels)

	unstructured.RemoveNestedField(obj.Object, "spec", "template", "metadata", "labels")
	unstructured.RemoveNestedField(obj.Object, "spec", "selector")
	return nil
}

// copyKeyValue is a helper function to copy a kv and create a new kv in the map with new key but same value. Needed to help maintain resource versions when needed
func copyKeyValue(obj map[string]string, key string, newKey string) {
	v, ok := obj[key]
	if ok {
		obj[newKey] = v
		delete(obj, key)
	}
}

// cleans the apiService to point to no-where
func apiServiceCleanup(obj *unstructured.Unstructured) error {
	unstructured.RemoveNestedField(obj.Object, "spec", "service")
	unstructured.RemoveNestedField(obj.Object, "spec", "caBundle")
	unstructured.RemoveNestedField(obj.Object, "spec", "insecureSkipTLSVerify")
	return nil
}

// nodeCleanup patches the address in node to localhost and saves existing address
// into annotations
func nodeCleanup(obj *unstructured.Unstructured) error {
	if obj.GroupVersionKind().Group != "" || obj.GroupVersionKind().Version != "v1" {
		// kind Node may be present in other GVK
		// this ensures we patch nothing else
		return nil
	}

	status, ok, err := unstructured.NestedFieldCopy(obj.Object, "status")
	if err != nil {
		return err
	}
	if !ok {
		return nil //nothing to do
	}

	statusMap, ok := status.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unable to assert status as a map[string]interface{}")
	}
	addresses, ok := statusMap["addresses"]
	if !ok {
		return nil // no addresses present. nothing to patch
	}

	addressList, ok := addresses.([]interface{})
	if !ok {
		return fmt.Errorf("unable to assert addresses into []interface{}. current values %v", addresses)
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	var newAddresses []interface{}
	for _, address := range addressList {
		addressMap, ok := address.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to assert address into map[string]string")
		}
		t := addressMap["type"]
		a := addressMap["address"]
		addressMap["address"] = "localhost"
		annotations[fmt.Sprintf("%soriginal-%s", simLabelPrefix, t)] = a.(string)
		newAddresses = append(newAddresses, addressMap)
	}

	statusMap["addresses"] = newAddresses
	obj.SetAnnotations(annotations)
	return unstructured.SetNestedField(obj.Object, statusMap, "status")
}

// loadBalancerCleanup will cleanup loadbalancer.harvesterhci.io objects
// the backend name is not a mandatory object, however a value of "null" is cleaned up
// this cleanup method adds it back
func loadBalancerCleanup(obj *unstructured.Unstructured) error {
	listeners, ok, err := unstructured.NestedFieldCopy(obj.Object, "spec", "listeners")
	if err != nil {
		return fmt.Errorf("unable to fetch listeners from the object %v", err)
	}

	// no listeners defined.. nothing to do
	if !ok {
		return nil
	}

	var patchedListeners []interface{}
	listenerList, ok := listeners.([]interface{})
	if !ok {
		return fmt.Errorf("unable to assert listeners interface as []interface{}")
	}

	for _, v := range listenerList {
		listenerMap, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to assert listener into a map[string]interface{}")
		}

		_, ok = listenerMap["name"]
		if !ok {
			listenerMap["name"] = "null"
		}

		patchedListeners = append(patchedListeners, listenerMap)
	}

	// patch the object with updated values
	return unstructured.SetNestedField(obj.Object, patchedListeners, "spec", "listeners")
}

// blockDevices cleanup will cleanup objects with invalid mountPoint details
// some examples have mountPoint as "null" and this gets removed, which breaks
// crd validation. this method as a simple work around introduces it back
func blockDevicesCleanup(obj *unstructured.Unstructured) error {
	err := cleanupDeviceOrStatus(obj, "spec", "fileSystem", "mountPoint")
	if err != nil {
		return err
	}

	err = cleanupDeviceOrStatus(obj, "status", "deviceStatus", "fileSystem", "mountPoint")
	if err != nil {
		return err
	}
	return cleanupDeviceOrStatus(obj, "status", "deviceStatus", "fileSystem", "type")
}

func cleanupDeviceOrStatus(obj *unstructured.Unstructured, fields ...string) error {
	fieldValue := strings.Join(fields, ".")
	_, ok, err := unstructured.NestedString(obj.Object, fields...)

	if err != nil {

		return fmt.Errorf("unable to fetch fields %s from object %v", fieldValue, err)
	}

	if !ok {
		return unstructured.SetNestedField(obj.Object, "null", fields...)
	}

	return nil
}

// cleanupSecret is needed to clean up secrets which have no data
// and are represented as a string rather than a map[string]string
func cleanupSecret(obj *unstructured.Unstructured) error {
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	return nil
}

// cleanupEvents will remove the fields firstTimestamp, lastTimestamp from the core Event
func cleanupEvent(obj *unstructured.Unstructured) error {
	if obj.GroupVersionKind().Group == "events.k8s.io" {
		unstructured.RemoveNestedField(obj.Object, "deprecatedFirstTimestamp")
		unstructured.RemoveNestedField(obj.Object, "deprecatedLastTimestamp")
		unstructured.RemoveNestedField(obj.Object, "deprecatedCount")
		unstructured.RemoveNestedField(obj.Object, "deprecatedSource")
		unstructured.RemoveNestedField(obj.Object, "series")
	} else {
		// cleanup corev1 Events
		unstructured.RemoveNestedField(obj.Object, "firstTimestamp")
		unstructured.RemoveNestedField(obj.Object, "lastTimestamp")
		unstructured.RemoveNestedField(obj.Object, "count")
		unstructured.RemoveNestedField(obj.Object, "source")
	}

	unstructured.RemoveNestedField(obj.Object, "series")

	orgEventTime, eventOk, err := unstructured.NestedString(obj.Object, "eventTime")
	if err != nil {
		return err
	}

	// apply an eventTime if none is present or is empty
	if !eventOk || orgEventTime == "" {
		creationTimeStamp, ok, err := unstructured.NestedString(obj.Object, "metadata", "creationTimestamp")
		if err != nil {
			return err
		}

		// create a new time or convert existing time to UnixMicro
		var tmpTime time.Time
		if !ok {
			tmpTime = time.Now()
		} else {
			tmpTime, err = time.Parse(time.RFC3339, creationTimeStamp)
			if err != nil {
				return err
			}
		}
		creationTimeStamp = tmpTime.Format(metav1.RFC3339Micro)
		err = unstructured.SetNestedField(obj.Object, creationTimeStamp, "eventTime")
		if err != nil {
			return err
		}
	}

	if err := checkAndSetDefaultValue(obj, []string{"reportingController"}, "sim-generated"); err != nil {
		return err
	}

	if err := checkAndSetDefaultValue(obj, []string{"reportingInstance"}, "sim-generated"); err != nil {
		return err
	}

	if err := checkAndSetDefaultValue(obj, []string{"action"}, "sim-generated"); err != nil {
		return err
	}
	return nil
}

func checkAndSetDefaultValue(obj *unstructured.Unstructured, field []string, defaultVal string) error {
	val, ok, err := unstructured.NestedString(obj.Object, field...)
	if err != nil {
		return err
	}

	if !ok || val == "" {
		val = defaultVal
		return unstructured.SetNestedField(obj.Object, val, field...)
	}

	return nil
}

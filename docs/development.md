# Development instructions

The support-bundle-kit simulator functionality tweaks a few objects in the upstream k8s.io code based which is available in the vendor directory.

The following instructions need to be followed to ensure that vendoring updates refactor these changes.

We have patched a few items in the code to allow the `simulator` functionality:

## creationTimestamp support
To ensure that the `creationTimestamp` is honored from the exported objects we have disabled the validation on `creationTimestamp` the following changes have been made:

* `k8s.io/apimachinery/pkg/api/validation/objectmeta.go`
```go
	allErrs = append(allErrs, ValidateImmutableField(newMeta.GetName(), oldMeta.GetName(), fldPath.Child("name"))...)
	allErrs = append(allErrs, ValidateImmutableField(newMeta.GetNamespace(), oldMeta.GetNamespace(), fldPath.Child("namespace"))...)
	allErrs = append(allErrs, ValidateImmutableField(newMeta.GetUID(), oldMeta.GetUID(), fldPath.Child("uid"))...)
	//Disabled to ensure the objectMeta from support bundle is honored while importing the object
	//allErrs = append(allErrs, ValidateImmutableField(newMeta.GetCreationTimestamp(), oldMeta.GetCreationTimestamp(), fldPath.Child("creationTimestamp"))...)
	allErrs = append(allErrs, ValidateImmutableField(newMeta.GetDeletionTimestamp(), oldMeta.GetDeletionTimestamp(), fldPath.Child("deletionTimestamp"))...)
	allErrs = append(allErrs, ValidateImmutableField(newMeta.GetDeletionGracePeriodSeconds(), oldMeta.GetDeletionGracePeriodSeconds(), fldPath.Child("deletionGracePeriodSeconds"))...)
	allErrs = append(allErrs, ValidateImmutableField(newMeta.GetClusterName(), oldMeta.GetClusterName(), fldPath.Child("clusterName"))...)

	allErrs = append(allErrs, v1validation.ValidateLabels(newMeta.GetLabels(), fldPath.Child("labels"))...)
	allErrs = append(allErrs, ValidateAnnotations(newMeta.GetAnnotations(), fldPath.Child("annotations"))...)
	allErrs = append(allErrs, ValidateOwnerReferences(newMeta.GetOwnerReferences(), fldPath.Child("ownerReferences"))...)
	allErrs = append(allErrs, v1validation.ValidateManagedFields(newMeta.GetManagedFields(), fldPath.Child("managedFields"))...)
```

* `k8s.io/apiserver/pkg/registry/rest/meta.go`
```go
// FillObjectMetaSystemFields populates fields that are managed by the system on ObjectMeta.
func FillObjectMetaSystemFields(meta metav1.Object) {
	if meta.GetCreationTimestamp().String() == "" {
		meta.SetCreationTimestamp(metav1.Now())
	}
	meta.SetUID(uuid.NewUUID())
	meta.SetSelfLink("")
}
```

## virtual kubelet log support
The `support-bundle-kit simulator` runs a minimal virtual-kubelet to support log streaming from the support bundle.
The simulator listens on localhost, to ensure kubectl and other cli tooling works natively, 
the routes in the apiserver have been patched to update NodeAddress to localhost.

* `k8s.io/kubernetes/pkg/registry/core/pod/strategy.go`

```go
	nodeInfo, err := connInfo.GetConnectionInfo(ctx, nodeName)
	if err != nil {
		return nil, nil, err
	}

	//patch to allow the node to point to localhost to ensure virtual kubelet can return logs
	nodeInfo.Hostname = "localhost"

	params := url.Values{}
```
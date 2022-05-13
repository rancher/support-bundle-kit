package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              []NodeConfigSpec `json:"spec"`
}

type NodeConfigSpec struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FailedObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              []FailedObjectSpec `json:"spec"`
}

type FailedObjectSpec struct {
	GVK       string `json:"gvk"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Error     string `json:"error"`
}

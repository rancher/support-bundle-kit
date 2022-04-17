package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeConfig struct {
	metav1.TypeMeta   `json:"inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              []NodeConfigSpec `json:"spec"`
}

type NodeConfigSpec struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

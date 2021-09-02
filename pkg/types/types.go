// shared types for support bundle controller and manager
package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StateNone       = ""
	StateGenerating = "generating"
	StateError      = "error"
	StateReady      = "ready"

	// labels
	SupportBundleLabelKey       = "rancher/supportbundle"
	SupportBundleNodeLabelValue = "true"
	DrainKey                    = "kubevirt.io/drain"

	SupportBundleManager = "support-bundle-manager"
	SupportBundleAgent   = "support-bundle-agent"
)

type ManagerStatus struct {
	Phase        string
	Error        bool
	ErrorMessage string
	Progress     int
	FileName     string
	FileSize     int64
}

type SupportBundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SupportBundleSpec   `json:"spec,omitempty"`
	Status SupportBundleStatus `json:"status,omitempty"`
}

type SupportBundleSpec struct {
	IssueURL    string `json:"issueURL"`
	Description string `json:"description"`
}

type SupportBundleStatus struct {
	State    SupportBundleState `json:"state,omitempty"`
	Progress int                `json:"progress,omitempty"`
	FileName string             `json:"filename,omitempty"`
	FileSize int64              `json:"filesize,omitempty"`
}

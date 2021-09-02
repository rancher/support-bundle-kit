// shared types for support bundle controller and manager
package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SupportBundleState string

const (
	SupportBundleStateNone       = SupportBundleState("")
	SupportBundleStateGenerating = SupportBundleState("generating")
	SupportBundleStateReady      = SupportBundleState("ready")
	SupportBundleStateError      = SupportBundleState("error")

	// labels
	SupportBundleLabelKey       = "rancher/supportbundle"
	SupportBundleNodeLabelValue = "true"
	DrainKey                    = "kubevirt.io/drain"

	SupportBundleManager = "support-bundle-manager"
	SupportBundleAgent   = "support-bundle-agent"
)

type ManagerPhase string

const (
	ManagerPhaseInit          = ManagerPhase("init")
	ManagerPhaseClusterBundle = ManagerPhase("cluster bundle")
	ManagerPhaseNodeBundle    = ManagerPhase("node bundle")
	ManagerPhasePackaging     = ManagerPhase("package")
	ManagerPhaseDone          = ManagerPhase("done")
)

type ManagerStatus struct {
	Phase        ManagerPhase
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
	FileName string             `json:"fileName,omitempty"`
	FileSize int64              `json:"fileSize,omitempty"`
}

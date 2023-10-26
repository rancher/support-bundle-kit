package manager

import "github.com/rancher/support-bundle-kit/pkg/types"

const (
	PhaseInit          = "start"
	PhaseClusterBundle = "cluster"
	PhaseNodeBundle    = "node"
	PhasePackaging     = "packaging"
	PhaseDone          = "done"

	BundleVersion = "0.1.0"
)

type BundleMeta struct {
	BundleName           string `json:"projectName"`
	BundleVersion        string `json:"bundleVersion"`
	KubernetesVersion    string `json:"kubernetesVersion"`
	ProjectNamespaceUUID string `json:"projectNamspaceUUID"`
	BundleCreatedAt      string `json:"bundleCreatedAt"`
	IssueURL             string `json:"issueURL"`
	IssueDescription     string `json:"issueDescription"`
}

type StateStoreInterface interface {
	GetState(namespace, supportbundle string) (types.SupportBundleState, error)
}

package manager

import harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

const (
	PhaseInit          = "start"
	PhaseClusterBundle = "cluster"
	PhaseNodeBundle    = "node"
	PhasePackaging     = "packaging"
	PhaseDone          = "done"

	BundleVersion = "0.1.0"
)

type BundleMeta struct {
	ProjectName          string `json:"projectName"`
	ProjectVersion       string `json:"projectVersion"`
	BundleVersion        string `json:"bundleVersion"`
	KubernetesVersion    string `json:"kubernetesVersion"`
	ProjectNamespaceUUID string `json:"projectNamspaceUUID"`
	BundleCreatedAt      string `json:"bundleCreatedAt"`
	IssueURL             string `json:"issueURL"`
	IssueDescription     string `json:"issueDescription"`
}

type StateStoreInterface interface {
	GetSupportBundle(namespace, supportbundle string) (*harvesterv1.SupportBundle, error)
	GetState(namespace, supportbundle string) (string, error)
}

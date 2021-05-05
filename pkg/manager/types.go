package manager

import harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

const (
	StateNone        = ""
	StateGenerating  = "generating"
	StateManagerDone = "managerdone"
	StateAgentDone   = "agentdone"
	StateError       = "error"
	StateReady       = "ready"

	HarvesterNodeLabelKey   = "harvesterhci.io/managed"
	HarvesterNodeLabelValue = "true"
	SupportBundleLabelKey   = "harvesterhci.io/supportbundle"
	DrainKey                = "kubevirt.io/drain"

	AppManager = "support-bundle-manager"
	AppAgent   = "support-bundle-agent"

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

	Done(namespace, supportbundle, filename string, filesize int64) error

	SetError(namespace, supportbundle string, er error) error
}

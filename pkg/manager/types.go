package manager

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
)

type BundleMeta struct {
	ProjectName          string `json:"projectName"`
	ProjectVersion       string `json:"projectVersion"`
	KubernetesVersion    string `json:"kubernetesVersion"`
	ProjectNamespaceUUID string `json:"projectNamspaceUUID"`
	BundleCreatedAt      string `json:"bundleCreatedAt"`
	IssueURL             string `json:"issueURL"`
	IssueDescription     string `json:"issueDescription"`
}

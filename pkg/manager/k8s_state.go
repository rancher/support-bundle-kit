package manager

import (
	harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/support-bundle-utils/pkg/manager/client"
)

type K8sStore struct {
	client *client.HarvesterClient
}

func NewK8sStore(c *client.HarvesterClient) *K8sStore {
	return &K8sStore{
		client: c,
	}
}

func (s *K8sStore) GetSupportBundle(namespace, supportbundle string) (*harvesterv1.SupportBundle, error) {
	return s.client.GetSupportBundle(namespace, supportbundle)
}

func (s *K8sStore) GetState(namespace, supportbundle string) (string, error) {
	return s.client.GetSupportBundleState(namespace, supportbundle)
}

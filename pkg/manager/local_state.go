package manager

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/rancher/support-bundle-kit/pkg/types"
)

type LocalStore struct {
	sbs map[string]*types.SupportBundle
}

// NewLocalStore creates a local state store with one supportbundle
func NewLocalStore(namespace, supportbundle string) *LocalStore {
	sbs := map[string]*types.SupportBundle{
		getSupportBundleKey(namespace, supportbundle): {
			Status: types.SupportBundleStatus{
				State: types.SupportBundleStateGenerating,
			},
		},
	}
	logrus.Debugf("Create a local state store. (%s/%s)", namespace, supportbundle)
	return &LocalStore{
		sbs: sbs,
	}
}

func getSupportBundleKey(namespace, supportbundle string) string {
	return fmt.Sprintf("%s-%s", namespace, supportbundle)
}

func (s *LocalStore) getSb(namespace, supportbundle string) (*types.SupportBundle, error) {
	key := getSupportBundleKey(namespace, supportbundle)
	if _, ok := s.sbs[key]; !ok {
		return nil, fmt.Errorf("supportbundle %s is not found", supportbundle)
	}
	return s.sbs[key], nil
}

func (s *LocalStore) GetSupportBundle(namespace, supportbundle string) (*types.SupportBundle, error) {
	logrus.Debugf("Get supportbundle %s/%s", namespace, supportbundle)
	return s.getSb(namespace, supportbundle)
}

func (s *LocalStore) GetState(namespace, supportbundle string) (types.SupportBundleState, error) {
	sb, err := s.getSb(namespace, supportbundle)
	if err != nil {
		return "", err
	}
	logrus.Debugf("Get supportbundle %s/%s state %s", namespace, supportbundle, sb.Status.State)
	return sb.Status.State, nil
}

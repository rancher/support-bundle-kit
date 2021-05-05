package manager

import (
	"fmt"

	harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/sirupsen/logrus"
)

type LocalStore struct {
	sbs map[string]*harvesterv1.SupportBundle
}

// NewLocalStore creates a local state store with one supportbundle
func NewLocalStore(namespace, supportbundle string) *LocalStore {
	sbs := map[string]*harvesterv1.SupportBundle{
		getSupportBundleKey(namespace, supportbundle): {
			Status: harvesterv1.SupportBundleStatus{
				State: StateGenerating,
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

func (s *LocalStore) getSb(namespace, supportbundle string) (*harvesterv1.SupportBundle, error) {
	key := getSupportBundleKey(namespace, supportbundle)
	if _, ok := s.sbs[key]; !ok {
		return nil, fmt.Errorf("supportbundle %s is not found", supportbundle)
	}
	return s.sbs[key], nil
}

func (s *LocalStore) GetSupportBundle(namespace, supportbundle string) (*harvesterv1.SupportBundle, error) {
	logrus.Debugf("Get supportbundle %s/%s", namespace, supportbundle)
	return s.getSb(namespace, supportbundle)
}

func (s *LocalStore) GetState(namespace, supportbundle string) (string, error) {
	sb, err := s.getSb(namespace, supportbundle)
	if err != nil {
		return "", err
	}
	logrus.Debugf("Get supportbundle %s/%s state %s", namespace, supportbundle, sb.Status.State)
	return sb.Status.State, nil
}

func (s *LocalStore) Done(namespace, supportbundle, filename string, filesize int64) error {
	sb, err := s.getSb(namespace, supportbundle)
	if err != nil {
		return err
	}
	logrus.Debugf("Mark supportbundle %s/%s done", namespace, supportbundle)
	sb.Status.State = StateReady
	return nil
}

func (s *LocalStore) SetError(namespace, supportbundle string, er error) (err error) {
	sb, err := s.getSb(namespace, supportbundle)
	if err != nil {
		return err
	}
	logrus.Debugf("Mark supportbundle %s/%s error", namespace, supportbundle)
	sb.Status.State = StateError
	return nil
}

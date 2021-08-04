package manager

import (
	"sync"

	"github.com/rancher/support-bundle-kit/pkg/types"
)

type ManagerStatus struct {
	sync.RWMutex
	types.ManagerStatus
}

func (s *ManagerStatus) SetPhase(phase string) {
	s.Lock()
	defer s.Unlock()
	s.Phase = phase
}

func (s *ManagerStatus) SetError(message string) {
	s.Lock()
	defer s.Unlock()
	s.Error = true
	s.ErrorMessage = message
}

func (s *ManagerStatus) SetProgress(progress int) {
	s.Lock()
	defer s.Unlock()
	s.Progress = progress
}

func (s *ManagerStatus) SetFileinfo(filename string, filesize int64) {
	s.Lock()
	defer s.Unlock()
	s.Filename = filename
	s.Filesize = filesize
}

package manager

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/rancher/support-bundle-kit/pkg/types"
)

func TestParseToleration(t *testing.T) {
	type testCase struct {
		input string

		expectedToleration *corev1.Toleration
		expectError        bool
	}
	testCases := map[string]testCase{
		"valid key:NoSchedule": {
			input: "key:NoSchedule",
			expectedToleration: &corev1.Toleration{
				Key:      "key",
				Value:    "",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
			expectError: false,
		},
		"valid key=value:NoExecute": {
			input: "key=value:NoExecute",
			expectedToleration: &corev1.Toleration{
				Key:      "key",
				Value:    "value",
				Operator: corev1.TolerationOpEqual,
				Effect:   corev1.TaintEffectNoExecute,
			},
			expectError: false,
		},
		"valid key=value:PreferNoSchedule": {
			input: "key=value:PreferNoSchedule",
			expectedToleration: &corev1.Toleration{
				Key:      "key",
				Value:    "value",
				Operator: corev1.TolerationOpEqual,
				Effect:   corev1.TaintEffectPreferNoSchedule,
			},
			expectError: false,
		},
		"invalid key:InvalidEffect": {
			input:              "key:InvalidEffect",
			expectedToleration: nil,
			expectError:        true,
		},
		"invalid key=value=NoSchedule": {
			input:              "key=value=NoSchedule",
			expectedToleration: nil,
			expectError:        true,
		},
	}

	for name, test := range testCases {
		fmt.Printf("testing %v\n", name)

		toleration, err := parseToleration(test.input)
		if !reflect.DeepEqual(toleration, test.expectedToleration) {
			t.Errorf("unexpected toleration:\nGot: %v\nWant: %v", toleration, test.expectedToleration)
		}

		if test.expectError && err == nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestRunAllPhases(t *testing.T) {
	tests := []struct {
		name             string
		requiredPhases   []RunPhase
		optionalPhases   []RunPhase
		postPhases       []RunPhase
		expectedError    bool
		expectedProgress int
	}{
		{
			name: "All pass",
			requiredPhases: []RunPhase{
				{Name: types.ManagerPhaseInit, Run: func() error { return nil }},
				{Name: types.ManagerPhaseClusterBundle, Run: func() error { return nil }},
			},
			optionalPhases: []RunPhase{
				{Name: types.ManagerPhasePrometheusBundle, Run: func() error { return nil }},
			},
			postPhases: []RunPhase{
				{Name: types.ManagerPhasePackaging, Run: func() error { return nil }},
				{Name: types.ManagerPhaseDone, Run: func() error { return nil }},
			},
			expectedError:    false,
			expectedProgress: 100,
		},
		{
			name: "First Required phase error",
			requiredPhases: []RunPhase{
				{Name: types.ManagerPhaseInit, Run: func() error { return errors.New("required phase error") }},
				{Name: types.ManagerPhaseClusterBundle, Run: func() error { return nil }},
			},
			optionalPhases: []RunPhase{
				{Name: types.ManagerPhasePrometheusBundle, Run: func() error { return nil }},
			},
			postPhases: []RunPhase{
				{Name: types.ManagerPhasePackaging, Run: func() error { return nil }},
				{Name: types.ManagerPhaseDone, Run: func() error { return nil }},
			},
			expectedError:    true,
			expectedProgress: 0,
		},
		{
			name: "Second Required phase error",
			requiredPhases: []RunPhase{
				{Name: types.ManagerPhaseInit, Run: func() error { return nil }},
				{Name: types.ManagerPhaseClusterBundle, Run: func() error { return errors.New("required phase error") }},
			},
			optionalPhases: []RunPhase{
				{Name: types.ManagerPhasePrometheusBundle, Run: func() error { return nil }},
			},
			postPhases: []RunPhase{
				{Name: types.ManagerPhasePackaging, Run: func() error { return nil }},
				{Name: types.ManagerPhaseDone, Run: func() error { return nil }},
			},
			expectedError:    true,
			expectedProgress: 20,
		},
		{
			name: "Optional phase error",
			requiredPhases: []RunPhase{
				{Name: types.ManagerPhaseInit, Run: func() error { return nil }},
				{Name: types.ManagerPhaseClusterBundle, Run: func() error { return nil }},
			},
			optionalPhases: []RunPhase{
				{Name: types.ManagerPhasePrometheusBundle, Run: func() error { return errors.New("optional phase error") }},
			},
			postPhases: []RunPhase{
				{Name: types.ManagerPhasePackaging, Run: func() error { return nil }},
				{Name: types.ManagerPhaseDone, Run: func() error { return nil }},
			},
			expectedError:    false,
			expectedProgress: 100,
		},
		{
			name: "Final Post phase error",
			requiredPhases: []RunPhase{
				{Name: types.ManagerPhaseInit, Run: func() error { return nil }},
				{Name: types.ManagerPhaseClusterBundle, Run: func() error { return nil }},
			},
			optionalPhases: []RunPhase{
				{Name: types.ManagerPhasePrometheusBundle, Run: func() error { return nil }},
			},
			postPhases: []RunPhase{
				{Name: types.ManagerPhasePackaging, Run: func() error { return nil }},
				{Name: types.ManagerPhaseDone, Run: func() error { return errors.New("post phase error") }},
			},
			expectedError:    true,
			expectedProgress: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SupportBundleManager{
				status: ManagerStatus{},
			}
			m.runAllPhases(tt.requiredPhases, tt.optionalPhases, tt.postPhases)
			assert.Equal(t, tt.expectedError, m.status.Error, "expected error %v, got %v")
			assert.Equal(t, tt.expectedProgress, m.status.Progress, "expected progress %v, got %v")
		})
	}
}

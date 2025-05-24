// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/common/config"
	"github.com/stretchr/testify/require"
)

// MockConfig implements config.Config for testing.
type MockConfig struct {
	keyDirValue string
}

func (m *MockConfig) Get(key string) interface{} {
	if key == "signing.key_directory" {
		return m.keyDirValue
	}

	return nil
}

func (m *MockConfig) GetString(key string) string {
	if key == "signing.key_directory" {
		return m.keyDirValue
	}

	return ""
}

func (m *MockConfig) GetBool(_ string) bool {
	return false
}

func (m *MockConfig) GetInt(_ string) int {
	return 0
}

func (m *MockConfig) GetStringSlice(_ string) []string {
	return nil
}

// Verify MockConfig implements the interface.
var _ config.Config = (*MockConfig)(nil)

// TestWithConfiguration tests the WithConfiguration option.
func TestWithConfiguration(t *testing.T) {
	// Create test cases
	tests := []struct {
		name       string
		configPath string
		expectPath string
	}{
		{
			name:       "Custom directory",
			configPath: "/custom/keys",
			expectPath: "/custom/keys",
		},
		{
			name:       "Empty path uses default",
			configPath: "",
			expectPath: "", // Uses default (blank in test)
		},
	}

	// Run test cases
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create mock config
			mockCfg := &MockConfig{keyDirValue: testCase.configPath}

			// Create adapter with configuration
			adapter := NewVerificationAdapterWithOptions(WithConfiguration(mockCfg))

			// Check key directory via repository
			require.Equal(t, testCase.expectPath, adapter.repository.GetKeyDirectory())
		})
	}
}

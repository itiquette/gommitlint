// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestResolvePath tests the ResolvePath function with various inputs.
func TestResolvePath(t *testing.T) {
	// Create a mock config
	mockConfig := &MockConfig{
		stringValues: map[string]string{
			"empty.path":     "",
			"relative.path":  "test/path",
			"absolute.path":  "/absolute/test/path",
			"incorrect.path": "incorrect/path",
		},
	}

	workingDir, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name         string
		key          string
		defaultPath  string
		expectedPath string
	}{
		{
			name:         "Empty path returns default",
			key:          "empty.path",
			defaultPath:  "/default/path",
			expectedPath: "/default/path",
		},
		{
			name:         "Relative path is joined with working directory",
			key:          "relative.path",
			defaultPath:  "/default/path",
			expectedPath: filepath.Join(workingDir, "test/path"),
		},
		{
			name:         "Absolute path is returned as is",
			key:          "absolute.path",
			defaultPath:  "/default/path",
			expectedPath: "/absolute/test/path",
		},
		{
			name:         "NonExistent key returns default",
			key:          "nonexistent.key",
			defaultPath:  "/default/path",
			expectedPath: "/default/path",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := ResolvePath(mockConfig, testCase.key, testCase.defaultPath)
			require.Equal(t, testCase.expectedPath, result)
		})
	}
}

// TestSafeJoin moved to fsutils package - path joining operations are centralized there

// MockConfig is a simple implementation of the Config interface for testing.
type MockConfig struct {
	stringValues map[string]string
}

func (m *MockConfig) Get(key string) interface{} {
	if val, ok := m.stringValues[key]; ok {
		return val
	}

	return nil
}

func (m *MockConfig) GetString(key string) string {
	if val, ok := m.stringValues[key]; ok {
		return val
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

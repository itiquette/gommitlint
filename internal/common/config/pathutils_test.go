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

// TestEnsureDirectory tests the EnsureDirectory function.
func TestEnsureDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		perm        os.FileMode
		expectError bool
		setup       func(string) // Setup function to run before the test
	}{
		{
			name:        "Empty path returns error",
			path:        "",
			perm:        0755,
			expectError: true,
		},
		{
			name:        "Creates directory successfully",
			path:        filepath.Join(tempDir, "new_dir"),
			perm:        0755,
			expectError: false,
		},
		{
			name:        "Does nothing when directory exists",
			path:        filepath.Join(tempDir, "existing_dir"),
			perm:        0755,
			expectError: false,
			setup: func(path string) {
				err := os.MkdirAll(path, 0755)
				require.NoError(t, err)
			},
		},
		{
			name:        "Creates nested directories",
			path:        filepath.Join(tempDir, "parent/child/grandchild"),
			perm:        0755,
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.path)
			}

			err := EnsureDirectory(testCase.path, testCase.perm)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Verify directory exists
				info, err := os.Stat(testCase.path)
				require.NoError(t, err)
				require.True(t, info.IsDir())
			}
		})
	}
}

// TestResolveFilePath tests the ResolveFilePath function.
func TestResolveFilePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create a mock config
	mockConfig := &MockConfig{
		stringValues: map[string]string{
			"valid.file":   filepath.Join(tempDir, "config/file.txt"),
			"invalid.file": "", // Empty path should use default
		},
	}

	tests := []struct {
		name        string
		key         string
		defaultPath string
		perm        os.FileMode
		expectError bool
	}{
		{
			name:        "Creates parent directory for file path",
			key:         "valid.file",
			defaultPath: filepath.Join(tempDir, "default/file.txt"),
			perm:        0755,
			expectError: false,
		},
		{
			name:        "Uses default path when key is empty",
			key:         "invalid.file",
			defaultPath: filepath.Join(tempDir, "default/file.txt"),
			perm:        0755,
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			path, err := ResolveFilePath(mockConfig, testCase.key, testCase.defaultPath, testCase.perm)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify parent directory was created
				dir := filepath.Dir(path)
				info, err := os.Stat(dir)
				require.NoError(t, err)
				require.True(t, info.IsDir())
			}
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

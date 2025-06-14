// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateConfigPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectError  bool
		errorPattern string
		description  string
	}{
		{
			name:        "valid relative path",
			path:        "config.yaml",
			expectError: false,
			description: "should accept valid relative paths",
		},
		{
			name:        "valid nested path",
			path:        "configs/app.yaml",
			expectError: false,
			description: "should accept valid nested paths",
		},
		{
			name:         "empty path",
			path:         "",
			expectError:  true,
			errorPattern: "empty path",
			description:  "should reject empty paths",
		},
		{
			name:         "null byte injection",
			path:         "config.yaml\x00",
			expectError:  true,
			errorPattern: "null byte",
			description:  "should reject paths with null bytes",
		},
		{
			name:         "control character",
			path:         "config\n.yaml",
			expectError:  true,
			errorPattern: "control character",
			description:  "should reject paths with control characters",
		},
		{
			name:        "tab character allowed",
			path:        "config\t.yaml",
			expectError: false,
			description: "should allow tab characters",
		},
		{
			name:         "very long path",
			path:         strings.Repeat("a", 1001),
			expectError:  true,
			errorPattern: "path too long",
			description:  "should reject excessively long paths",
		},
		{
			name:         "absolute path",
			path:         "/etc/passwd",
			expectError:  true,
			errorPattern: "absolute path not allowed",
			description:  "should reject absolute paths",
		},
		{
			name:         "path traversal",
			path:         "../config.yaml",
			expectError:  true,
			errorPattern: "path traversal",
			description:  "should reject path traversal attempts",
		},
		{
			name:         "complex traversal",
			path:         "../../etc/passwd",
			expectError:  true,
			errorPattern: "path traversal",
			description:  "should reject complex traversal attempts",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := validateConfigPath(testCase.path)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorPattern != "" {
					require.Contains(t, err.Error(), testCase.errorPattern, testCase.description)
				}

				require.Empty(t, result)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, result)
				require.True(t, filepath.IsAbs(result), "should return absolute path")
			}
		})
	}
}

func TestValidateConfigFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		fileMode     os.FileMode
		expectError  bool
		errorPattern string
		description  string
	}{
		{
			name:        "secure permissions",
			fileMode:    0600,
			expectError: false,
			description: "should accept secure file permissions",
		},
		{
			name:        "read-only permissions",
			fileMode:    0400,
			expectError: false,
			description: "should accept read-only permissions",
		},
		{
			name:         "world-writable",
			fileMode:     0666,
			expectError:  true,
			errorPattern: "world-writable",
			description:  "should reject world-writable files",
		},
		{
			name:         "group-writable",
			fileMode:     0620,
			expectError:  true,
			errorPattern: "group-writable",
			description:  "should reject group-writable files",
		},
		{
			name:         "executable file",
			fileMode:     0700,
			expectError:  true,
			errorPattern: "should not be executable",
			description:  "should reject executable files",
		},
		{
			name:         "world-executable",
			fileMode:     0755,
			expectError:  true,
			errorPattern: "should not be executable",
			description:  "should reject world-executable files",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create test file with specific permissions
			testFile := filepath.Join(tmpDir, "test-"+testCase.name+".yaml")
			content := "gommitlint:\n  message:\n    subject:\n      max_length: 50"

			err := os.WriteFile(testFile, []byte(content), testCase.fileMode)
			require.NoError(t, err)

			// Test permission validation
			err = validateConfigFilePermissions(testFile)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorPattern != "" {
					require.Contains(t, err.Error(), testCase.errorPattern, testCase.description)
				}
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestSecureConfigPathValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to tmpDir so relative paths work correctly
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid config file
	validConfigFile := "valid-config.yaml"
	configContent := "gommitlint:\n  message:\n    subject:\n      max_length: 50"
	err = os.WriteFile(validConfigFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Create an insecure config file
	insecureConfigFile := "insecure-config.yaml"
	err = os.WriteFile(insecureConfigFile, []byte(configContent), 0666) //nolint:gosec // Intentionally insecure for testing
	require.NoError(t, err)

	tests := []struct {
		name         string
		configPath   string
		expectError  bool
		errorPattern string
		description  string
	}{
		{
			name:        "valid secure config",
			configPath:  validConfigFile,
			expectError: false,
			description: "should accept valid secure config files",
		},
		{
			name:         "insecure permissions",
			configPath:   insecureConfigFile,
			expectError:  true,
			errorPattern: "insecure permissions",
			description:  "should reject files with insecure permissions",
		},
		{
			name:         "non-existent file",
			configPath:   "nonexistent.yaml",
			expectError:  true,
			errorPattern: "not found",
			description:  "should reject non-existent files",
		},
		{
			name:         "path traversal",
			configPath:   "../../../etc/passwd",
			expectError:  true,
			errorPattern: "path traversal",
			description:  "should reject path traversal attempts",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := secureConfigPathValidation(testCase.configPath)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorPattern != "" {
					require.Contains(t, err.Error(), testCase.errorPattern, testCase.description)
				}
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestIsConfigPathSecure(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to tmpDir so relative paths work correctly
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a valid config file
	validConfigFile := "valid-config.yaml"
	configContent := "gommitlint:\n  message:\n    subject:\n      max_length: 50"
	err = os.WriteFile(validConfigFile, []byte(configContent), 0600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		configPath  string
		expected    bool
		description string
	}{
		{
			name:        "valid secure config",
			configPath:  validConfigFile,
			expected:    true,
			description: "should return true for valid secure configs",
		},
		{
			name:        "path traversal",
			configPath:  "../../../etc/passwd",
			expected:    false,
			description: "should return false for path traversal",
		},
		{
			name:        "non-existent file",
			configPath:  "nonexistent.yaml",
			expected:    false,
			description: "should return false for non-existent files",
		},
		{
			name:        "empty path",
			configPath:  "",
			expected:    false,
			description: "should return false for empty paths",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := isConfigPathSecure(testCase.configPath)
			require.Equal(t, testCase.expected, result, testCase.description)
		})
	}
}

func TestSanitizeConfigPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "clean path",
			input:       "config.yaml",
			expected:    "config.yaml",
			description: "should return clean paths unchanged",
		},
		{
			name:        "empty path",
			input:       "",
			expected:    "",
			description: "should handle empty paths",
		},
		{
			name:        "null byte removal",
			input:       "config.yaml\x00malicious",
			expected:    "config.yamlmalicious",
			description: "should remove null bytes",
		},
		{
			name:        "control character removal",
			input:       "config\n.yaml",
			expected:    "config.yaml",
			description: "should remove control characters",
		},
		{
			name:        "tab preservation",
			input:       "config\t.yaml",
			expected:    "config\t.yaml",
			description: "should preserve tab characters",
		},
		{
			name:        "path cleaning",
			input:       "./config/../config.yaml",
			expected:    "config.yaml",
			description: "should clean path components",
		},
		{
			name:        "length limiting",
			input:       strings.Repeat("a", 1100),
			expected:    strings.Repeat("a", 1000),
			description: "should limit path length",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := sanitizeConfigPath(testCase.input)
			require.Equal(t, testCase.expected, result, testCase.description)
		})
	}
}

func TestConfigPathValidationError(t *testing.T) {
	err := ConfigPathValidationError{
		Path:   "../../../etc/passwd",
		Reason: "path traversal detected",
	}

	expected := "invalid config path '../../../etc/passwd': path traversal detected"
	require.Equal(t, expected, err.Error())
}

func TestSymlinkValidation(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping symlink tests when running as root")
	}

	tmpDir := t.TempDir()

	// Change to tmpDir to use relative paths
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a real config file
	realConfig := "real-config.yaml"
	configContent := "gommitlint:\n  message:\n    subject:\n      max_length: 50"
	err = os.WriteFile(realConfig, []byte(configContent), 0600)
	require.NoError(t, err)

	// Create a symlink to the real config
	symlinkConfig := "symlink-config.yaml"
	err = os.Symlink(realConfig, symlinkConfig)
	require.NoError(t, err)

	// Test that symlink is rejected
	_, err = validateConfigPath(symlinkConfig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "symlink not allowed")
}

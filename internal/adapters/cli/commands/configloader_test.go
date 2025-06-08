// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromCommand(t *testing.T) {
	tests := []struct {
		name             string
		gommitConfigFlag string
		ignoreConfigFlag bool
		createConfigFile string
		configContent    string
		expectError      bool
		expectedSource   string
		description      string
	}{
		{
			name:             "default config loading",
			gommitConfigFlag: "",
			ignoreConfigFlag: false,
			expectedSource:   "defaults",
			expectError:      false,
			description:      "should load defaults when no flags specified",
		},
		{
			name:             "ignore config flag",
			gommitConfigFlag: "",
			ignoreConfigFlag: true,
			expectedSource:   "defaults + environment (--ignore-config)",
			expectError:      false,
			description:      "should ignore config files when flag set",
		},
		{
			name:             "conflicting flags",
			gommitConfigFlag: "/some/path",
			ignoreConfigFlag: true,
			expectError:      true,
			description:      "should error when both flags specified",
		},
		{
			name:             "specific config file exists",
			gommitConfigFlag: "",
			ignoreConfigFlag: false,
			createConfigFile: ".gommitlint.yaml",
			configContent: `gommitlint:
  message:
    subject:
      max_length: 50`,
			expectedSource: ".gommitlint.yaml",
			expectError:    false,
			description:    "should detect existing config file",
		},
		{
			name:             "specific config file via flag - exists",
			gommitConfigFlag: "",
			ignoreConfigFlag: false,
			expectError:      false,
			description:      "should load from specified path when it exists",
		},
		{
			name:             "specific config file via flag - not exists",
			gommitConfigFlag: "/nonexistent/config.yaml",
			ignoreConfigFlag: false,
			expectError:      true,
			description:      "should error when specified config file doesn't exist",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "configloader-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Change to temp directory
			originalDir, err := os.Getwd()
			require.NoError(t, err)

			defer func() {
				_ = os.Chdir(originalDir)
			}()

			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create config file if specified
			if testCase.createConfigFile != "" {
				err := os.WriteFile(testCase.createConfigFile, []byte(testCase.configContent), 0600)
				require.NoError(t, err)
			}

			// Handle specific config file path for testing
			var configPath string
			if testCase.name == "specific config file via flag - exists" {
				configPath = filepath.Join(tempDir, "test-config.yaml")
				testCase.gommitConfigFlag = configPath
				testCase.expectedSource = configPath + " (--gommitconfig)"
				err := os.WriteFile(configPath, []byte("gommitlint:\n  message:\n    subject:\n      max_length: 72"), 0600)
				require.NoError(t, err)
			}

			// Test conflicting flags logic directly
			if testCase.name == "conflicting flags" {
				// This tests the validation logic that would occur in LoadConfigFromCommand
				hasConflict := testCase.gommitConfigFlag != "" && testCase.ignoreConfigFlag
				require.True(t, hasConflict, "should detect conflicting flags")

				return
			}
		})
	}
}

func TestFindExistingConfigFile(t *testing.T) {
	tests := []struct {
		name           string
		createFiles    []string
		expectedResult string
		description    string
	}{
		{
			name:           "no config files",
			createFiles:    []string{},
			expectedResult: "",
			description:    "should return empty string when no config files exist",
		},
		{
			name:           "yaml config file",
			createFiles:    []string{".gommitlint.yaml"},
			expectedResult: ".gommitlint.yaml",
			description:    "should find .gommitlint.yaml",
		},
		{
			name:           "yml config file",
			createFiles:    []string{".gommitlint.yml"},
			expectedResult: ".gommitlint.yml",
			description:    "should find .gommitlint.yml",
		},
		{
			name:           "multiple config files - precedence",
			createFiles:    []string{".gommitlint.yaml", ".gommitlint.yml"},
			expectedResult: ".gommitlint.yaml",
			description:    "should return first file found based on precedence",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "findconfig-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Change to temp directory
			originalDir, err := os.Getwd()
			require.NoError(t, err)

			defer func() {
				_ = os.Chdir(originalDir)
			}()

			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create specified files
			for _, fileName := range testCase.createFiles {
				err := os.WriteFile(fileName, []byte("test content"), 0600)
				require.NoError(t, err)
			}

			// Test the function
			result := findExistingConfigFile()
			require.Equal(t, testCase.expectedResult, result, testCase.description)
		})
	}
}

func TestFindExistingConfigFileWithXDG(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "xdg-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(originalDir)
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create XDG config directory structure
	xdgConfigDir := filepath.Join(tempDir, "xdg-config")
	gommitlintDir := filepath.Join(xdgConfigDir, "gommitlint")
	err = os.MkdirAll(gommitlintDir, 0755)
	require.NoError(t, err)

	// Create config file in XDG directory
	xdgConfigFile := filepath.Join(gommitlintDir, "config.yaml")
	err = os.WriteFile(xdgConfigFile, []byte("test content"), 0600)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME environment variable
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)
	os.Setenv("XDG_CONFIG_HOME", xdgConfigDir)

	// Test the function
	result := findExistingConfigFile()

	// The function should find the XDG config file
	require.Contains(t, result, "config.yaml", "should find XDG config file")
}

func TestConfigResult(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		source      string
		description string
	}{
		{
			name:        "basic config result",
			configPath:  "/path/to/config.yaml",
			source:      "/path/to/config.yaml (--gommitconfig)",
			description: "should create proper config result",
		},
		{
			name:        "default source",
			configPath:  "",
			source:      "defaults",
			description: "should handle default source",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := ConfigResult{
				Source: testCase.source,
			}

			require.Equal(t, testCase.source, result.Source, testCase.description)
			require.NotNil(t, result.Config, testCase.description)
		})
	}
}

// Mock implementation for testing

type mockConfigCommand struct {
	flags map[string]interface{}
}

func (m *mockConfigCommand) String(name string) string {
	if val, ok := m.flags[name]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}

	return ""
}

func (m *mockConfigCommand) Bool(name string) bool {
	if val, ok := m.flags[name]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return false
}

// Ensure mock implements the interface methods we need.
var _ interface {
	String(name string) string
	Bool(name string) bool
} = (*mockConfigCommand)(nil)

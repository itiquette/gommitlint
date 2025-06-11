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

// TestGetConfigSearchPaths tests the config search path logic.
func TestGetConfigSearchPaths(t *testing.T) {
	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		// Save original value
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if originalXDG != "" {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Unset XDG_CONFIG_HOME
		os.Unsetenv("XDG_CONFIG_HOME")

		paths := getConfigSearchPaths()
		expected := []string{
			".gommitlint.yaml",
			".gommitlint.yml",
		}

		require.Equal(t, expected, paths)
	})

	t.Run("with XDG_CONFIG_HOME but no gommitlint directory", func(t *testing.T) {
		// Save original value
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if originalXDG != "" {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Set XDG_CONFIG_HOME to a temp directory without gommitlint subdirectory
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		paths := getConfigSearchPaths()
		expected := []string{
			".gommitlint.yaml",
			".gommitlint.yml",
		}

		require.Equal(t, expected, paths)
	})

	t.Run("with XDG_CONFIG_HOME and gommitlint directory", func(t *testing.T) {
		// Save original value
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if originalXDG != "" {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Create temp directory with gommitlint subdirectory
		tmpDir := t.TempDir()
		gommitlintDir := filepath.Join(tmpDir, "gommitlint")
		err := os.MkdirAll(gommitlintDir, 0755)
		require.NoError(t, err)

		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		paths := getConfigSearchPaths()
		expected := []string{
			".gommitlint.yaml",
			".gommitlint.yml",
			filepath.Join(gommitlintDir, "config.yaml"),
			filepath.Join(gommitlintDir, "config.yml"),
		}

		require.Equal(t, expected, paths)
	})
}

// TestFindFirstExistingConfigFile tests finding config files.
func TestFindFirstExistingConfigFile(t *testing.T) {
	// Save original XDG_CONFIG_HOME
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	t.Run("no config files exist", func(t *testing.T) {
		// Change to empty temp directory
		tmpDir := t.TempDir()

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)

		// Unset XDG_CONFIG_HOME
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		require.Empty(t, result)
	})

	t.Run("finds .gommitlint.yaml in current directory", func(t *testing.T) {
		// Create temp directory with config file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, ".gommitlint.yaml")
		err := os.WriteFile(configFile, []byte("test"), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)

		// Unset XDG_CONFIG_HOME
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		require.Equal(t, ".gommitlint.yaml", result)
	})

	t.Run("finds .gommitlint.yml in current directory", func(t *testing.T) {
		// Create temp directory with .yml config file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, ".gommitlint.yml")
		err := os.WriteFile(configFile, []byte("test"), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)

		// Unset XDG_CONFIG_HOME
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		require.Equal(t, ".gommitlint.yml", result)
	})

	t.Run("prioritizes .yaml over .yml in current directory", func(t *testing.T) {
		// Create temp directory with both config files
		tmpDir := t.TempDir()
		yamlFile := filepath.Join(tmpDir, ".gommitlint.yaml")
		ymlFile := filepath.Join(tmpDir, ".gommitlint.yml")

		err := os.WriteFile(yamlFile, []byte("yaml"), 0600)
		require.NoError(t, err)
		err = os.WriteFile(ymlFile, []byte("yml"), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)

		// Unset XDG_CONFIG_HOME
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		require.Equal(t, ".gommitlint.yaml", result)
	})

	t.Run("finds XDG config when no current directory config", func(t *testing.T) {
		// Create empty current directory
		tmpDir := t.TempDir()

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)

		// Create XDG config directory and file
		xdgDir := t.TempDir()
		gommitlintDir := filepath.Join(xdgDir, "gommitlint")
		err := os.MkdirAll(gommitlintDir, 0755)
		require.NoError(t, err)

		xdgConfigFile := filepath.Join(gommitlintDir, "config.yaml")
		err = os.WriteFile(xdgConfigFile, []byte("xdg config"), 0600)
		require.NoError(t, err)

		os.Setenv("XDG_CONFIG_HOME", xdgDir)

		result := findFirstExistingConfigFile()
		require.Equal(t, xdgConfigFile, result)
	})

	t.Run("prioritizes current directory over XDG", func(t *testing.T) {
		// Create current directory with config
		tmpDir := t.TempDir()
		currentConfigFile := filepath.Join(tmpDir, ".gommitlint.yaml")
		err := os.WriteFile(currentConfigFile, []byte("current"), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)

		// Create XDG config directory and file
		xdgDir := t.TempDir()
		gommitlintDir := filepath.Join(xdgDir, "gommitlint")
		err = os.MkdirAll(gommitlintDir, 0755)
		require.NoError(t, err)

		xdgConfigFile := filepath.Join(gommitlintDir, "config.yaml")
		err = os.WriteFile(xdgConfigFile, []byte("xdg config"), 0600)
		require.NoError(t, err)

		os.Setenv("XDG_CONFIG_HOME", xdgDir)

		result := findFirstExistingConfigFile()
		require.Equal(t, ".gommitlint.yaml", result)
	})
}

// TestLoadConfigFromPath tests loading config from specific path.
func TestLoadConfigFromPath(t *testing.T) {
	t.Run("loads valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "test-config.yaml")

		configContent := `gommitlint:
  rules:
    enabled:
      - subject
      - conventional
`
		err := os.WriteFile(configFile, []byte(configContent), 0600)
		require.NoError(t, err)

		cfg, err := LoadConfigFromPath(configFile)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Contains(t, cfg.Rules.Enabled, "subject")
		require.Contains(t, cfg.Rules.Enabled, "conventional")
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		cfg, err := LoadConfigFromPath("/non/existent/config.yaml")
		require.NoError(t, err)
		// Should return default config when file doesn't exist
		require.NotNil(t, cfg)
	})

	t.Run("handles invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "invalid-config.yaml")

		invalidYAML := `gommitlint:
  rules:
    - invalid: yaml: structure
`
		err := os.WriteFile(configFile, []byte(invalidYAML), 0600)
		require.NoError(t, err)

		cfg, err := LoadConfigFromPath(configFile)
		require.NoError(t, err)
		// Should return default config when YAML is invalid
		require.NotNil(t, cfg)
	})
}

// TestApplyRulePriority tests rule priority logic.
func TestApplyRulePriority(t *testing.T) {
	tests := []struct {
		name            string
		enabled         []string
		disabled        []string
		expectedEnabled []string
	}{
		{
			name:            "no conflicts",
			enabled:         []string{"subject", "conventional"},
			disabled:        []string{"spell"},
			expectedEnabled: []string{"subject", "conventional"},
		},
		{
			name:            "disabled takes precedence",
			enabled:         []string{"subject", "conventional", "spell"},
			disabled:        []string{"spell"},
			expectedEnabled: []string{"subject", "conventional"},
		},
		{
			name:            "multiple conflicts",
			enabled:         []string{"subject", "conventional", "spell", "jirareference"},
			disabled:        []string{"spell", "jirareference"},
			expectedEnabled: []string{"subject", "conventional"},
		},
		{
			name:            "empty lists",
			enabled:         nil,
			disabled:        nil,
			expectedEnabled: nil,
		},
		{
			name:            "all disabled",
			enabled:         []string{"subject", "conventional"},
			disabled:        []string{"subject", "conventional"},
			expectedEnabled: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with test data
			cfg := LoadDefaultConfig()
			cfg.Rules.Enabled = testCase.enabled
			cfg.Rules.Disabled = testCase.disabled

			result := applyRulePriority(cfg)
			require.Equal(t, testCase.expectedEnabled, result.Rules.Enabled)
			require.Equal(t, testCase.disabled, result.Rules.Disabled)
		})
	}
}

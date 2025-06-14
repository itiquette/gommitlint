// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"os"
	"path/filepath"
	"strings"
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
			".gommitlint.toml",
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
			".gommitlint.toml",
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
			".gommitlint.toml",
			filepath.Join(gommitlintDir, "config.yaml"),
			filepath.Join(gommitlintDir, "config.yml"),
			filepath.Join(gommitlintDir, "config.toml"),
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

// TestLoadConfigFromPath tests loading config from specific path (format-agnostic tests).
func TestLoadConfigFromPath(t *testing.T) {
	t.Run("handles non-existent file", func(t *testing.T) {
		cfg, err := LoadConfigFromPath("/non/existent/config.yaml")
		require.NoError(t, err)
		// Should return default config when file doesn't exist
		require.NotNil(t, cfg)
	})

	t.Run("handles invalid file extension defaults to YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.unknown")

		// Even with unknown extension, should default to YAML parser
		configContent := `gommitlint:
  output: json
`
		err := os.WriteFile(configFile, []byte(configContent), 0600)
		require.NoError(t, err)

		cfg, err := LoadConfigFromPath(configFile)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "json", cfg.Output)
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

// TestEnvironmentVariablesIgnored verifies that environment variables no longer affect configuration.
func TestEnvironmentVariablesIgnored(t *testing.T) {
	// Save original environment variables
	originalVars := map[string]string{
		"GOMMITLINT_OUTPUT":                    os.Getenv("GOMMITLINT_OUTPUT"),
		"GOMMITLINT_SUBJECT_MAX_LENGTH":        os.Getenv("GOMMITLINT_SUBJECT_MAX_LENGTH"),
		"GOMMITLINT_CONVENTIONAL_TYPES":        os.Getenv("GOMMITLINT_CONVENTIONAL_TYPES"),
		"GOMMITLINT_REPO_REFERENCE_BRANCH":     os.Getenv("GOMMITLINT_REPO_REFERENCE_BRANCH"),
		"GOMMITLINT_SIGNING_REQUIRE_SIGNATURE": os.Getenv("GOMMITLINT_SIGNING_REQUIRE_SIGNATURE"),
		"GOMMITLINT_RULES_ENABLED":             os.Getenv("GOMMITLINT_RULES_ENABLED"),
		"GOMMITLINT_RULES_DISABLED":            os.Getenv("GOMMITLINT_RULES_DISABLED"),
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range originalVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("environment variables are ignored", func(t *testing.T) {
		// Set environment variables that would have affected config in the past
		os.Setenv("GOMMITLINT_OUTPUT", "should-be-ignored")
		os.Setenv("GOMMITLINT_SUBJECT_MAX_LENGTH", "999")
		os.Setenv("GOMMITLINT_CONVENTIONAL_TYPES", "ignored,values")
		os.Setenv("GOMMITLINT_REPO_REFERENCE_BRANCH", "ignored-branch")
		os.Setenv("GOMMITLINT_SIGNING_REQUIRE_SIGNATURE", "true")
		os.Setenv("GOMMITLINT_RULES_ENABLED", "ignored-rule")
		os.Setenv("GOMMITLINT_RULES_DISABLED", "another-ignored-rule")

		// Load default config
		cfg := LoadDefaultConfig()

		// Environment variables should not affect the config
		require.NotEqual(t, "should-be-ignored", cfg.Output)
		require.NotEqual(t, 999, cfg.Message.Subject.MaxLength)
		require.NotContains(t, cfg.Conventional.Types, "ignored")
		require.NotEqual(t, "ignored-branch", cfg.Repo.ReferenceBranch)
		require.NotContains(t, cfg.Rules.Enabled, "ignored-rule")
		require.NotContains(t, cfg.Rules.Disabled, "another-ignored-rule")
	})

	t.Run("LoadConfig ignores environment variables", func(t *testing.T) {
		// Set environment variables
		os.Setenv("GOMMITLINT_OUTPUT", "env-output")
		os.Setenv("GOMMITLINT_RULES_ENABLED", "env-rule")

		// Load config (should only use defaults and file config, not env)
		cfg, err := LoadConfig()
		require.NoError(t, err)

		// Environment variables should not affect the result
		require.NotEqual(t, "env-output", cfg.Output)
		require.NotContains(t, cfg.Rules.Enabled, "env-rule")
	})

	t.Run("LoadConfigFromPath ignores environment variables", func(t *testing.T) {
		// Create a config file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "test.yaml")
		configContent := `gommitlint:
  output: json
  rules:
    enabled:
      - file-rule`
		err := os.WriteFile(configFile, []byte(configContent), 0600)
		require.NoError(t, err)

		// Set conflicting environment variables
		os.Setenv("GOMMITLINT_OUTPUT", "env-output")
		os.Setenv("GOMMITLINT_RULES_ENABLED", "env-rule")

		// Load config from file
		cfg, err := LoadConfigFromPath(configFile)
		require.NoError(t, err)

		// Should use file config, not environment variables
		require.Equal(t, "json", cfg.Output)
		require.Contains(t, cfg.Rules.Enabled, "file-rule")
		require.NotContains(t, cfg.Rules.Enabled, "env-rule")
	})
}

// TestLoadConfigWithRepoPath tests config loading with repository path.
func TestLoadConfigWithRepoPath(t *testing.T) {
	// Create test directories
	currentDir := t.TempDir()
	repoDir := t.TempDir()

	// Change to current directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()

	err := os.Chdir(currentDir)
	require.NoError(t, err)

	// Create config in current directory
	currentConfigContent := `gommitlint:
  output: json`
	err = os.WriteFile(filepath.Join(currentDir, ".gommitlint.yaml"), []byte(currentConfigContent), 0600)
	require.NoError(t, err)

	// Create config in repo directory
	repoConfigContent := `gommitlint:
  output: text`
	err = os.WriteFile(filepath.Join(repoDir, ".gommitlint.yaml"), []byte(repoConfigContent), 0600)
	require.NoError(t, err)

	tests := []struct {
		name           string
		repoPath       string
		expectedOutput string
		description    string
	}{
		{
			name:           "empty repo path uses current directory",
			repoPath:       "",
			expectedOutput: "json",
			description:    "should use current directory config when no repo path",
		},
		{
			name:           "repo path uses repository directory",
			repoPath:       repoDir,
			expectedOutput: "text",
			description:    "should use repository directory config when repo path specified",
		},
		{
			name:           "non-existent repo path uses defaults",
			repoPath:       "/nonexistent/path",
			expectedOutput: "text",
			description:    "should use defaults when repo path doesn't exist",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg, err := LoadConfigWithRepoPath(testCase.repoPath)
			require.NoError(t, err, testCase.description)
			require.Equal(t, testCase.expectedOutput, cfg.Output, testCase.description)
		})
	}
}

// TestGetConfigSearchPathsForRepo tests repository-specific config search paths.
func TestGetConfigSearchPathsForRepo(t *testing.T) {
	// Save original XDG_CONFIG_HOME
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	tests := []struct {
		name          string
		repoPath      string
		setupXDG      func()
		expectedPaths []string
		description   string
	}{
		{
			name:     "empty repo path uses current directory",
			repoPath: "",
			setupXDG: func() { os.Unsetenv("XDG_CONFIG_HOME") },
			expectedPaths: []string{
				".gommitlint.yaml",
				".gommitlint.yml",
				".gommitlint.toml",
			},
			description: "should search in current directory when no repo path",
		},
		{
			name:     "repo path uses repository directory",
			repoPath: "/path/to/repo",
			setupXDG: func() { os.Unsetenv("XDG_CONFIG_HOME") },
			expectedPaths: []string{
				"/path/to/repo/.gommitlint.yaml",
				"/path/to/repo/.gommitlint.yml",
				"/path/to/repo/.gommitlint.toml",
			},
			description: "should search in repository directory when repo path specified",
		},
		{
			name:     "includes XDG paths with repo path",
			repoPath: "/repo",
			setupXDG: func() {
				tmpDir := t.TempDir()
				gommitlintDir := filepath.Join(tmpDir, "gommitlint")
				err := os.MkdirAll(gommitlintDir, 0755)
				require.NoError(t, err)
				os.Setenv("XDG_CONFIG_HOME", tmpDir)
			},
			expectedPaths: []string{
				"/repo/.gommitlint.yaml",
				"/repo/.gommitlint.yml",
				"/repo/.gommitlint.toml",
			},
			description: "should include both repo and XDG paths",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupXDG()

			paths := getConfigSearchPathsForRepo(testCase.repoPath)

			// Check that expected paths are present
			for _, expectedPath := range testCase.expectedPaths {
				require.Contains(t, paths, expectedPath, testCase.description)
			}

			// For XDG test, verify XDG paths are also included
			if testCase.name == "includes XDG paths with repo path" {
				xdgFound := false

				for _, path := range paths {
					if strings.Contains(path, "gommitlint/config.") {
						xdgFound = true

						break
					}
				}

				require.True(t, xdgFound, "should include XDG config paths")
			}
		})
	}
}

// TestFindFirstExistingConfigFileInRepo tests repository-specific config file discovery.
func TestFindFirstExistingConfigFileInRepo(t *testing.T) {
	// Create test directories
	currentDir := t.TempDir()
	repoDir := t.TempDir()

	// Change to current directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()

	err := os.Chdir(currentDir)
	require.NoError(t, err)

	// Create config in current directory
	err = os.WriteFile(filepath.Join(currentDir, ".gommitlint.yaml"), []byte("current"), 0600)
	require.NoError(t, err)

	// Create config in repo directory
	err = os.WriteFile(filepath.Join(repoDir, ".gommitlint.yml"), []byte("repo"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name         string
		repoPath     string
		expectedFile string
		description  string
	}{
		{
			name:         "finds config in current directory",
			repoPath:     "",
			expectedFile: ".gommitlint.yaml",
			description:  "should find config in current directory when no repo path",
		},
		{
			name:         "finds config in repository directory",
			repoPath:     repoDir,
			expectedFile: filepath.Join(repoDir, ".gommitlint.yml"),
			description:  "should find config in repository directory when repo path specified",
		},
		{
			name:         "returns empty for non-existent path",
			repoPath:     "/nonexistent/path",
			expectedFile: "",
			description:  "should return empty when no config found",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := findFirstExistingConfigFileInRepo(testCase.repoPath)
			require.Equal(t, testCase.expectedFile, result, testCase.description)
		})
	}
}

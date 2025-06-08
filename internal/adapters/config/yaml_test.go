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

func TestLoadFileConfig_YAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		validate    func(*testing.T, *require.Assertions, interface{})
		description string
	}{
		{
			name: "valid YAML config with rules",
			yamlContent: `gommitlint:
  rules:
    enabled:
      - subject
      - conventional
    disabled:
      - spell
      - jirareference`,
			expectError: false,
			validate: func(_ *testing.T, req *require.Assertions, result interface{}) {
				req.NotNil(result)
			},
			description: "should load basic YAML config with rules",
		},
		{
			name: "comprehensive YAML config",
			yamlContent: `gommitlint:
  output: json
  message:
    subject:
      max_length: 80
      case: lower
      require_imperative: true
      forbid_endings:
        - "."
        - "!"
    body:
      min_length: 10
      min_lines: 2
      allow_signoff_only: false
      require_signoff: true
  conventional:
    require_scope: true
    types:
      - feat
      - fix
      - docs
      - test
    scopes:
      - api
      - ui
      - core
    allow_breaking: true
    max_description_length: 100
  signing:
    require_signature: true
    require_verification: false
    require_multi_signoff: false
    key_directory: "/path/to/keys"
    allowed_signers:
      - alice@example.com
      - bob@example.com
  repo:
    max_commits_ahead: 5
    reference_branch: main
    allow_merge_commits: false
  jira:
    project_prefixes:
      - PROJ
      - TEST
    require_in_body: true
    require_in_subject: false
    ignore_ticket_patterns:
      - "IGNORE-.*"
  spell:
    ignore_words:
      - gommitlint
      - refactor
    locale: en_US
  rules:
    enabled:
      - subject
      - conventional
      - signing
    disabled:
      - spell`,
			expectError: false,
			validate: func(_ *testing.T, req *require.Assertions, result interface{}) {
				req.NotNil(result)
			},
			description: "should load comprehensive YAML config with all sections",
		},
		{
			name: "invalid YAML syntax",
			yamlContent: `gommitlint:
  rules:
    - invalid: yaml: structure`,
			expectError: true,
			validate: func(_ *testing.T, _ *require.Assertions, _ interface{}) {
				// Should return empty config on parse error
			},
			description: "should handle invalid YAML syntax gracefully",
		},
		{
			name:        "empty YAML file",
			yamlContent: ``,
			expectError: true,
			validate: func(_ *testing.T, _ *require.Assertions, _ interface{}) {
				// Should return empty config for empty file
			},
			description: "should handle empty YAML file",
		},
		{
			name: "YAML with only gommitlint section",
			yamlContent: `gommitlint:
  output: text`,
			expectError: false,
			validate: func(_ *testing.T, req *require.Assertions, result interface{}) {
				req.NotNil(result)
			},
			description: "should load minimal YAML config",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req := require.New(t)

			// Create temporary YAML file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "test.yaml")
			err := os.WriteFile(configFile, []byte(testCase.yamlContent), 0600)
			req.NoError(err)

			// Load config
			cfg := LoadFileConfig(configFile)

			// Basic validation - config should not be completely empty unless there's an error
			if !testCase.expectError {
				// For success cases, validate the config loaded properly
				testCase.validate(t, req, cfg)
			}
		})
	}
}

func TestConfigSearchPaths_YAML(t *testing.T) {
	t.Run("includes YAML files in search paths", func(t *testing.T) {
		// Save original XDG_CONFIG_HOME
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if originalXDG != "" {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Unset XDG_CONFIG_HOME for this test
		os.Unsetenv("XDG_CONFIG_HOME")

		paths := getConfigSearchPaths()

		// Should include YAML extensions
		require.Contains(t, paths, ".gommitlint.yaml")
		require.Contains(t, paths, ".gommitlint.yml")

		// Should still include TOML files
		require.Contains(t, paths, ".gommitlint.toml")

		// Verify order: YAML files first, then TOML
		require.Equal(t, []string{
			".gommitlint.yaml",
			".gommitlint.yml",
			".gommitlint.toml",
		}, paths)
	})

	t.Run("includes YAML in XDG config paths", func(t *testing.T) {
		// Save original XDG_CONFIG_HOME
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

		// Should include YAML in XDG paths
		expectedYAMLPath := filepath.Join(gommitlintDir, "config.yaml")
		require.Contains(t, paths, expectedYAMLPath)

		expectedYMLPath := filepath.Join(gommitlintDir, "config.yml")
		require.Contains(t, paths, expectedYMLPath)
	})
}

func TestFindFirstExistingConfigFile_YAML(t *testing.T) {
	// Save original XDG_CONFIG_HOME
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	t.Run("finds YAML config file", func(t *testing.T) {
		// Create temp directory with YAML config
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, ".gommitlint.yaml")
		yamlContent := `gommitlint:
  rules:
    enabled:
      - subject`
		err := os.WriteFile(configFile, []byte(yamlContent), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		require.Equal(t, ".gommitlint.yaml", result)
	})

	t.Run("prioritizes .yaml over .yml", func(t *testing.T) {
		// Create temp directory with both .yaml and .yml configs
		tmpDir := t.TempDir()

		yamlFile := filepath.Join(tmpDir, ".gommitlint.yaml")
		err := os.WriteFile(yamlFile, []byte("gommitlint:\n  output: yaml"), 0600)
		require.NoError(t, err)

		ymlFile := filepath.Join(tmpDir, ".gommitlint.yml")
		err = os.WriteFile(ymlFile, []byte("gommitlint:\n  output: yml"), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		// Should find .yaml first due to search order
		require.Equal(t, ".gommitlint.yaml", result)
	})

	t.Run("finds .yml when no .yaml exists", func(t *testing.T) {
		// Create temp directory with only .yml config
		tmpDir := t.TempDir()

		ymlFile := filepath.Join(tmpDir, ".gommitlint.yml")
		err := os.WriteFile(ymlFile, []byte("gommitlint:\n  output: yml"), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		require.Equal(t, ".gommitlint.yml", result)
	})

	t.Run("prioritizes YAML over TOML", func(t *testing.T) {
		// Create temp directory with both YAML and TOML configs
		tmpDir := t.TempDir()

		yamlFile := filepath.Join(tmpDir, ".gommitlint.yaml")
		err := os.WriteFile(yamlFile, []byte("gommitlint:\n  output: yaml"), 0600)
		require.NoError(t, err)

		tomlFile := filepath.Join(tmpDir, ".gommitlint.toml")
		err = os.WriteFile(tomlFile, []byte("[gommitlint]\noutput = \"toml\""), 0600)
		require.NoError(t, err)

		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()

		_ = os.Chdir(tmpDir)
		os.Unsetenv("XDG_CONFIG_HOME")

		result := findFirstExistingConfigFile()
		// Should find YAML first due to search order
		require.Equal(t, ".gommitlint.yaml", result)
	})
}

func TestLoadConfigFromPath_YAML(t *testing.T) {
	t.Run("loads YAML config from specific path", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "custom-config.yaml")

		yamlContent := `gommitlint:
  rules:
    enabled:
      - subject
      - conventional
    disabled:
      - spell
  message:
    subject:
      max_length: 72`

		err := os.WriteFile(configFile, []byte(yamlContent), 0600)
		require.NoError(t, err)

		cfg, err := LoadConfigFromPath(configFile)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Verify rules were loaded
		require.Contains(t, cfg.Rules.Enabled, "subject")
		require.Contains(t, cfg.Rules.Enabled, "conventional")
		require.Contains(t, cfg.Rules.Disabled, "spell")

		// Verify subject config was loaded
		require.Equal(t, 72, cfg.Message.Subject.MaxLength)
	})
}

func TestYAMLCompatibilityWithTOML(t *testing.T) {
	t.Run("YAML and TOML produce equivalent configs", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create equivalent YAML config
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		yamlContent := `gommitlint:
  output: json
  rules:
    enabled:
      - subject
      - conventional
    disabled:
      - spell
  message:
    subject:
      max_length: 80
      case: lower`

		err := os.WriteFile(yamlFile, []byte(yamlContent), 0600)
		require.NoError(t, err)

		// Create equivalent TOML config
		tomlFile := filepath.Join(tmpDir, "config.toml")
		tomlContent := `[gommitlint]
output = "json"

[gommitlint.rules]
enabled = ["subject", "conventional"]
disabled = ["spell"]

[gommitlint.message.subject]
max_length = 80
case = "lower"`

		err = os.WriteFile(tomlFile, []byte(tomlContent), 0600)
		require.NoError(t, err)

		// Load both configs
		yamlCfg := LoadFileConfig(yamlFile)
		tomlCfg := LoadFileConfig(tomlFile)

		// They should be equivalent
		require.Equal(t, yamlCfg.Output, tomlCfg.Output)
		require.Equal(t, yamlCfg.Rules.Enabled, tomlCfg.Rules.Enabled)
		require.Equal(t, yamlCfg.Rules.Disabled, tomlCfg.Rules.Disabled)
		require.Equal(t, yamlCfg.Message.Subject.MaxLength, tomlCfg.Message.Subject.MaxLength)
		require.Equal(t, yamlCfg.Message.Subject.Case, tomlCfg.Message.Subject.Case)
	})
}

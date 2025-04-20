// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/defaults"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected interface{}
	}{
		{
			name:     "subject max length",
			field:    "subject.maxLength",
			expected: defaults.SubjectMaxLengthDefault,
		},
		{
			name:     "subject case",
			field:    "subject.case",
			expected: defaults.SubjectCaseDefault,
		},
		{
			name:     "subject invalid suffixes",
			field:    "subject.invalidSuffixes",
			expected: defaults.SubjectInvalidSuffixesDefault,
		},
		{
			name:     "subject imperative",
			field:    "subject.imperative",
			expected: defaults.SubjectImperativeDefault,
		},
		{
			name:     "jira required",
			field:    "jira.required",
			expected: defaults.JIRARequiredDefault,
		},
		{
			name:     "jira pattern",
			field:    "jira.pattern",
			expected: defaults.JIRAPatternDefault,
		},
		{
			name:     "body required",
			field:    "body.required",
			expected: defaults.BodyRequiredDefault,
		},
		{
			name:     "conventional commit required",
			field:    "conventional.required",
			expected: defaults.ConventionalCommitRequiredDefault,
		},
		{
			name:     "conventional commit max desc length",
			field:    "conventional.maxDescLength",
			expected: defaults.ConventionalCommitMaxDescLengthDefault,
		},
		{
			name:     "conventional commit types",
			field:    "conventional.types",
			expected: defaults.ConventionalCommitTypesDefault,
		},
		{
			name:     "spellcheck enabled",
			field:    "spellcheck.enabled",
			expected: defaults.SpellcheckEnabledDefault,
		},
		{
			name:     "spellcheck locale",
			field:    "spellcheck.locale",
			expected: defaults.SpellcheckLocaleDefault,
		},
		{
			name:     "signature required",
			field:    "signature.required",
			expected: defaults.SignatureRequiredDefault,
		},
		{
			name:     "sign-off required",
			field:    "signoff",
			expected: defaults.SignOffRequiredDefault,
		},
		{
			name:     "n commits ahead",
			field:    "nCommitsAhead",
			expected: defaults.NCommitsAheadDefault,
		},
		{
			name:     "ignore merge commits",
			field:    "ignoreMergeCommits",
			expected: defaults.IgnoreMergeCommitsDefault,
		},
	}

	// Get default configuration
	defaultConfig := config.DefaultConfiguration()

	// Basic validation
	require.NotNil(t, defaultConfig, "DefaultConfiguration() should return non-nil")
	require.NotNil(t, defaultConfig.GommitConf, "GommitConf should not be nil")
	require.NotNil(t, defaultConfig.GommitConf.Subject, "Subject should not be nil")
	require.NotNil(t, defaultConfig.GommitConf.Body, "Body should not be nil")
	require.NotNil(t, defaultConfig.GommitConf.ConventionalCommit, "ConventionalCommit should not be nil")
	require.NotNil(t, defaultConfig.GommitConf.SpellCheck, "SpellCheck should not be nil")
	require.NotNil(t, defaultConfig.GommitConf.Signature, "Signature should not be nil")

	// Check each field against expected default value
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			switch testCase.field {
			case "subject.maxLength":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.Subject.MaxLength,
					"Default subject.maxLength should match expected value")
			case "subject.case":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.Subject.Case,
					"Default subject.case should match expected value")
			case "subject.invalidSuffixes":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.Subject.InvalidSuffixes,
					"Default subject.invalidSuffixes should match expected value")
			case "subject.imperative":
				require.NotNil(t, defaultConfig.GommitConf.Subject.Imperative,
					"Subject.Imperative should not be nil")
				require.Equal(t, testCase.expected, *defaultConfig.GommitConf.Subject.Imperative,
					"Default subject.imperative should match expected value")
			case "jira.required":
				require.NotNil(t, defaultConfig.GommitConf.Subject.Jira,
					"Subject.Jira should not be nil")
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.Subject.Jira.Required,
					"Default jira.required should match expected value")
			case "jira.pattern":
				require.NotNil(t, defaultConfig.GommitConf.Subject.Jira,
					"Subject.Jira should not be nil")
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.Subject.Jira.Pattern,
					"Default jira.pattern should match expected value")
			case "body.required":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.Body.Required,
					"Default body.required should match expected value")
			case "conventional.required":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.ConventionalCommit.Required,
					"Default conventional.required should match expected value")
			case "conventional.maxDescLength":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.ConventionalCommit.MaxDescriptionLength,
					"Default conventional.maxDescLength should match expected value")
			case "conventional.types":
				require.ElementsMatch(t, testCase.expected, defaultConfig.GommitConf.ConventionalCommit.Types,
					"Default conventional.types should match expected value")
			case "spellcheck.enabled":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.SpellCheck.Enabled,
					"Default spellcheck.enabled should match expected value")
			case "spellcheck.locale":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.SpellCheck.Locale,
					"Default spellcheck.locale should match expected value")
			case "signature.required":
				require.Equal(t, testCase.expected, defaultConfig.GommitConf.Signature.Required,
					"Default signature.required should match expected value")
			case "signoff":
				require.NotNil(t, defaultConfig.GommitConf.SignOffRequired,
					"SignOffRequired should not be nil")
				require.Equal(t, testCase.expected, *defaultConfig.GommitConf.SignOffRequired,
					"Default signOffRequired should match expected value")
			case "nCommitsAhead":
				require.NotNil(t, defaultConfig.GommitConf.NCommitsAhead,
					"NCommitsAhead should not be nil")
				require.Equal(t, testCase.expected, *defaultConfig.GommitConf.NCommitsAhead,
					"Default nCommitsAhead should match expected value")
			case "ignoreMergeCommits":
				require.NotNil(t, defaultConfig.GommitConf.IgnoreMergeCommits,
					"IgnoreMergeCommits should not be nil")
				require.Equal(t, testCase.expected, *defaultConfig.GommitConf.IgnoreMergeCommits,
					"Default ignoreMergeCommits should match expected value")
			default:
				t.Fatalf("Test case for field %s is not implemented", testCase.field)
			}
		})
	}

	// Additional check for reference field which has a hardcoded default
	require.Equal(t, "main", defaultConfig.GommitConf.Reference,
		"Default Reference should be 'main'")
}

func TestDefaultConfigPaths(t *testing.T) {
	// We can't test the unexported getDefaultConfigPaths function directly,
	// but we can test the path loading behavior through the manager
	tests := []struct {
		name           string
		setupXDG       bool
		setupProject   bool
		expectLoaded   bool
		expectedSource string
	}{
		{
			name:         "no config files",
			setupXDG:     false,
			setupProject: false,
			expectLoaded: false,
		},
		{
			name:           "only project config",
			setupXDG:       false,
			setupProject:   true,
			expectLoaded:   true,
			expectedSource: ".gommitlint.yaml",
		},
		{
			name:           "only XDG config",
			setupXDG:       true,
			setupProject:   false,
			expectLoaded:   true,
			expectedSource: "gommitlint.yaml", // Will be in XDG path
		},
		{
			name:           "both configs",
			setupXDG:       true,
			setupProject:   true,
			expectLoaded:   true,
			expectedSource: ".gommitlint.yaml", // Project has higher precedence
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory structure
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, "project")
			xdgDir := filepath.Join(tmpDir, ".config", "gommitlint")

			// Create directories
			err := os.MkdirAll(projectDir, 0755)
			require.NoError(t, err, "Failed to create project directory")

			err = os.MkdirAll(xdgDir, 0755)
			require.NoError(t, err, "Failed to create XDG directory")

			// Get original working directory to restore later
			originalDir, err := os.Getwd()
			require.NoError(t, err, "Failed to get current directory")
			defer os.Chdir(originalDir) //nolint

			// Save original XDG_CONFIG_HOME to restore later
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

			// Set XDG_CONFIG_HOME to our test directory
			os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

			// Create config files based on test case
			if testCase.setupXDG {
				xdgConfig := `gommitlint:
  subject:
    max-length: 60
`
				err := os.WriteFile(filepath.Join(xdgDir, "gommitlint.yaml"), []byte(xdgConfig), 0600)
				require.NoError(t, err, "Failed to write XDG config file")
			}

			if testCase.setupProject {
				projectConfig := `gommitlint:
  subject:
    max-length: 70
`
				err := os.WriteFile(filepath.Join(projectDir, ".gommitlint.yaml"), []byte(projectConfig), 0600)
				require.NoError(t, err, "Failed to write project config file")
			}

			// Change to project directory
			err = os.Chdir(projectDir)
			require.NoError(t, err, "Failed to change to project directory")

			// Create new manager which should load the appropriate config files
			manager, err := config.New()
			require.NoError(t, err, "Manager creation should not fail")

			// Check if a config file was loaded
			require.Equal(t, testCase.expectLoaded, manager.WasLoadedFromFile(),
				"WasLoadedFromFile() should match expectLoaded")

			// If a config was loaded, check if it's the correct one
			if testCase.expectLoaded {
				sourcePath := manager.GetSourcePath()
				require.Contains(t, sourcePath, testCase.expectedSource,
					"Source path should contain expected filename")

				// Also check the actual loaded values
				cfg := manager.GetConfig()
				require.NotNil(t, cfg, "Config should not be nil")
				require.NotNil(t, cfg.GommitConf, "GommitConf should not be nil")
				require.NotNil(t, cfg.GommitConf.Subject, "Subject should not be nil")

				if testCase.setupProject {
					require.Equal(t, 70, cfg.GommitConf.Subject.MaxLength,
						"Subject.MaxLength should match project config value")
				} else if testCase.setupXDG {
					require.Equal(t, 60, cfg.GommitConf.Subject.MaxLength,
						"Subject.MaxLength should match XDG config value")
				}
			}
		})
	}
}

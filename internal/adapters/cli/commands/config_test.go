// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/config"
	"github.com/itiquette/gommitlint/internal/domain"
	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewConfigCommand(t *testing.T) {
	cmd := NewConfigCommand()

	require.Equal(t, "config", cmd.Name)
	require.Equal(t, "Configuration operations", cmd.Usage)
	require.NotEmpty(t, cmd.Description)
	require.Len(t, cmd.Commands, 2)

	// Check subcommands
	showCmd := cmd.Commands[0]
	require.Equal(t, "show", showCmd.Name)
	require.Equal(t, "Show effective configuration", showCmd.Usage)
	require.NotNil(t, showCmd.Action)

	initCmd := cmd.Commands[1]
	require.Equal(t, "init", initCmd.Name)
	require.Equal(t, "Generate complete configuration file template", initCmd.Usage)
	require.NotNil(t, initCmd.Action)
}

func TestBuildEffectiveConfig(t *testing.T) {
	tests := []struct {
		name         string
		cfg          configTypes.Config
		commitRules  []domain.CommitRule
		repoRules    []domain.RepositoryRule
		configSource string
		description  string
	}{
		{
			name:         "empty configuration",
			cfg:          configTypes.Config{},
			commitRules:  []domain.CommitRule{},
			repoRules:    []domain.RepositoryRule{},
			configSource: "default",
			description:  "should handle empty configuration",
		},
		{
			name: "configuration with rules",
			cfg: configTypes.Config{
				Message: configTypes.MessageConfig{
					Subject: configTypes.SubjectConfig{MaxLength: 50},
				},
			},
			commitRules: []domain.CommitRule{
				mockCommitRule{name: "subject"},
				mockCommitRule{name: "conventional"},
			},
			repoRules: []domain.RepositoryRule{
				mockRepoRule{name: "branchahead"},
			},
			configSource: ".gommitlint.yaml",
			description:  "should include all rules in enabled rules list",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			effective := BuildEffectiveConfig(
				testCase.cfg,
				testCase.commitRules,
				testCase.repoRules,
				testCase.configSource,
			)

			require.Equal(t, testCase.cfg, effective.Config)
			require.Equal(t, testCase.configSource, effective.ConfigSource)

			expectedRulesCount := len(testCase.commitRules) + len(testCase.repoRules)
			require.Len(t, effective.EnabledRules, expectedRulesCount)

			// Check that all rule names are present
			ruleNameMap := make(map[string]bool)
			for _, ruleName := range effective.EnabledRules {
				ruleNameMap[ruleName] = true
			}

			for _, rule := range testCase.commitRules {
				require.True(t, ruleNameMap[rule.Name()], "commit rule %s should be in enabled rules", rule.Name())
			}

			for _, rule := range testCase.repoRules {
				require.True(t, ruleNameMap[rule.Name()], "repo rule %s should be in enabled rules", rule.Name())
			}
		})
	}
}

func TestPrintConfigJSON(t *testing.T) {
	tests := []struct {
		name            string
		effectiveConfig EffectiveConfig
		description     string
	}{
		{
			name: "basic configuration",
			effectiveConfig: EffectiveConfig{
				Config: configTypes.Config{
					Message: configTypes.MessageConfig{
						Subject: configTypes.SubjectConfig{MaxLength: 50},
					},
				},
				EnabledRules: []string{"subject", "conventional"},
				ConfigSource: "default",
			},
			description: "should output valid JSON",
		},
		{
			name: "empty configuration",
			effectiveConfig: EffectiveConfig{
				Config:       configTypes.Config{},
				EnabledRules: []string{},
				ConfigSource: "default",
			},
			description: "should handle empty configuration",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a temporary file for testing
			tmpFile, err := os.CreateTemp("", "config-test-*.json")
			require.NoError(t, err)

			defer func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			}()

			exitCode, err := printConfigJSON(testCase.effectiveConfig, tmpFile)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)

			// Read back the file content
			_, err = tmpFile.Seek(0, 0)
			require.NoError(t, err)

			content := make([]byte, 10240)
			n, _ := tmpFile.Read(content)
			jsonOutput := string(content[:n])

			// Verify it's valid JSON
			var parsed EffectiveConfig
			err = json.Unmarshal([]byte(jsonOutput), &parsed)
			require.NoError(t, err, testCase.description)

			// Verify content matches
			require.Equal(t, testCase.effectiveConfig.ConfigSource, parsed.ConfigSource)
			require.Equal(t, testCase.effectiveConfig.EnabledRules, parsed.EnabledRules)
		})
	}
}

func TestPrintConfigYAML(t *testing.T) {
	tests := []struct {
		name            string
		effectiveConfig EffectiveConfig
		description     string
	}{
		{
			name: "basic configuration",
			effectiveConfig: EffectiveConfig{
				Config: configTypes.Config{
					Message: configTypes.MessageConfig{
						Subject: configTypes.SubjectConfig{MaxLength: 50},
					},
				},
				EnabledRules: []string{"subject"},
				ConfigSource: "default",
			},
			description: "should output valid YAML",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a temporary file for testing
			tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
			require.NoError(t, err)

			defer func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			}()

			exitCode, err := printConfigYAML(testCase.effectiveConfig, tmpFile)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)

			// Read back the file content
			_, err = tmpFile.Seek(0, 0)
			require.NoError(t, err)

			content := make([]byte, 10240)
			n, _ := tmpFile.Read(content)
			yamlOutput := string(content[:n])

			// Verify it contains expected content
			require.Contains(t, yamlOutput, "# Gommitlint configuration file")
			require.Contains(t, yamlOutput, "gommitlint:")
			require.Contains(t, yamlOutput, testCase.effectiveConfig.ConfigSource)

			// Verify it's valid YAML
			var parsed map[string]interface{}
			err = yaml.Unmarshal([]byte(yamlOutput), &parsed)
			require.NoError(t, err, testCase.description)

			// Check for gommitlint root key
			require.Contains(t, parsed, "gommitlint")
		})
	}
}

func TestPrintConfigText(t *testing.T) {
	tests := []struct {
		name            string
		effectiveConfig EffectiveConfig
		expectedContent []string
		description     string
	}{
		{
			name: "basic configuration",
			effectiveConfig: EffectiveConfig{
				Config: configTypes.Config{
					Message: configTypes.MessageConfig{
						Subject: configTypes.SubjectConfig{
							MaxLength:         50,
							Case:              "any",
							RequireImperative: false,
						},
						Body: configTypes.BodyConfig{
							Required:  false,
							MinLength: 0,
						},
					},
					Conventional: configTypes.ConventionalConfig{
						RequireScope:         false,
						MaxDescriptionLength: 100,
						AllowBreaking:        true,
					},
					Signature: configTypes.SignatureConfig{
						Required:     false,
						VerifyFormat: false,
					},
					Identity: configTypes.IdentityConfig{
						AllowedAuthors: []string{},
					},
					Repo: configTypes.RepoConfig{
						MaxCommitsAhead:   10,
						ReferenceBranch:   "main",
						AllowMergeCommits: true,
					},
					Jira: configTypes.JiraConfig{
						RequireInSubject: false,
						RequireInBody:    false,
					},
					Spell: configTypes.SpellConfig{
						Locale: "en_US",
					},
					Output: "text",
				},
				EnabledRules: []string{"subject", "conventional"},
				ConfigSource: "default",
			},
			expectedContent: []string{
				"Gommitlint Configuration",
				"Configuration Source: default",
				"Enabled Rules (2):",
				"- conventional",
				"- subject",
				"Subject Max Length: 50",
				"Subject Case: any",
				"Require Imperative: false",
				"Body Min Length: 0",
				"Conventional Commit Configuration:",
				"Max Description Length: 100",
				"Signature Configuration:",
				"Repository Configuration:",
				"JIRA Configuration:",
				"Spell Configuration:",
				"Output Configuration:",
			},
			description: "should output comprehensive text format",
		},
		{
			name: "empty rules",
			effectiveConfig: EffectiveConfig{
				Config:       configTypes.Config{},
				EnabledRules: []string{},
				ConfigSource: "empty",
			},
			expectedContent: []string{
				"Enabled Rules (0):",
				"(none)",
				"Configuration Source: empty",
			},
			description: "should handle empty rules list",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a temporary file for testing
			tmpFile, err := os.CreateTemp("", "config-test-*.txt")
			require.NoError(t, err)

			defer func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			}()

			exitCode, err := printConfigText(testCase.effectiveConfig, tmpFile)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)

			// Read back the file content
			_, err = tmpFile.Seek(0, 0)
			require.NoError(t, err)

			content := make([]byte, 10240)
			n, _ := tmpFile.Read(content)
			textOutput := string(content[:n])

			// Verify expected content is present
			for _, expected := range testCase.expectedContent {
				require.Contains(t, textOutput, expected, "%s should contain '%s'", testCase.description, expected)
			}
		})
	}
}

func TestPrintConfigInitYAML(t *testing.T) {
	tests := []struct {
		name        string
		cfg         configTypes.Config
		description string
	}{
		{
			name:        "default configuration",
			cfg:         config.LoadDefaultConfig(),
			description: "should generate complete template",
		},
		{
			name: "custom configuration",
			cfg: configTypes.Config{
				Message: configTypes.MessageConfig{
					Subject: configTypes.SubjectConfig{MaxLength: 72},
				},
			},
			description: "should handle custom configuration",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a temporary file for testing
			tmpFile, err := os.CreateTemp("", "config-init-test-*.yaml")
			require.NoError(t, err)

			defer func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			}()

			err = printConfigInitYAML(testCase.cfg, tmpFile)
			require.NoError(t, err)

			// Read back the file content
			_, err = tmpFile.Seek(0, 0)
			require.NoError(t, err)

			content := make([]byte, 10240)
			n, _ := tmpFile.Read(content)
			yamlOutput := string(content[:n])

			// Verify expected header comments
			require.Contains(t, yamlOutput, "# Gommitlint Configuration Template")
			require.Contains(t, yamlOutput, "# Generated by: gommitlint config init")
			require.Contains(t, yamlOutput, "gommitlint:")
			require.Contains(t, yamlOutput, "# Configuration complete!")

			// Verify it's valid YAML
			var parsed map[string]interface{}
			// Remove comments for YAML parsing
			yamlLines := strings.Split(yamlOutput, "\n")

			var yamlWithoutComments []string

			for _, line := range yamlLines {
				if !strings.HasPrefix(strings.TrimSpace(line), "#") && strings.TrimSpace(line) != "" {
					yamlWithoutComments = append(yamlWithoutComments, line)
				}
			}

			cleanYAML := strings.Join(yamlWithoutComments, "\n")

			err = yaml.Unmarshal([]byte(cleanYAML), &parsed)
			require.NoError(t, err, testCase.description)

			// Check for gommitlint root key
			require.Contains(t, parsed, "gommitlint")
		})
	}
}

// Mock implementations for testing

type mockCommitRule struct {
	name string
}

func (m mockCommitRule) Name() string {
	return m.name
}

func (m mockCommitRule) Validate(_ domain.Commit, _ configTypes.Config) []domain.ValidationError {
	return nil
}

type mockRepoRule struct {
	name string
}

func (m mockRepoRule) Name() string {
	return m.name
}

func (m mockRepoRule) Validate(_ domain.Commit, _ domain.Repository, _ configTypes.Config) []domain.ValidationError {
	return nil
}

// Ensure mocks implement the interfaces.
var _ domain.CommitRule = (*mockCommitRule)(nil)
var _ domain.RepositoryRule = (*mockRepoRule)(nil)

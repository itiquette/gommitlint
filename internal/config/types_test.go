// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestConfigStructsUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		yamlConfig  string
		fieldChecks map[string]interface{}
	}{
		{
			name: "full config",
			yamlConfig: `
gommitlint:
  subject:
    max-length: 72
    case: lower
    imperative: true
    invalid-suffixes: ".,;"
    jira:
      required: true
      pattern: "[A-Z]+-[0-9]+"
      bodyref: false
      keys:
        - PROJECT
        - FEATURE
      strict: true
  body:
    required: true
  conventional-commit:
    required: true
    types:
      - feat
      - fix
      - docs
    scopes:
      - ui
      - api
      - docs
    max-description-length: 100
  spellcheck:
    locale: en-US
    enabled: true
  signature:
    required: true
    identity:
      public-key-uri: "/path/to/keys.asc"
  sign-off: true
  n-commits-ahead: true
  ignore-merge-commit: true
  reference: "develop"
`,
			fieldChecks: map[string]interface{}{
				"subject.maxLength":          72,
				"subject.case":               "lower",
				"subject.imperative":         true,
				"subject.invalidSuffixes":    ".,;",
				"jira.required":              true,
				"jira.pattern":               "[A-Z]+-[0-9]+",
				"jira.bodyRef":               false,
				"jira.projects":              []string{"PROJECT", "FEATURE"},
				"jira.strict":                true,
				"body.required":              true,
				"conventional.required":      true,
				"conventional.types":         []string{"feat", "fix", "docs"},
				"conventional.scopes":        []string{"ui", "api", "docs"},
				"conventional.maxDescLength": 100,
				"spellcheck.locale":          "en-US",
				"spellcheck.enabled":         true,
				"signature.required":         true,
				"signature.identity.uri":     "/path/to/keys.asc",
				"signoff":                    true,
				"nCommitsAhead":              true,
				"ignoreMergeCommits":         true,
				"reference":                  "develop",
			},
		},
		{
			name: "minimal config",
			yamlConfig: `
gommitlint:
  subject:
    max-length: 50
  body:
    required: false
`,
			fieldChecks: map[string]interface{}{
				"subject.maxLength": 50,
				"body.required":     false,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Use koanf to parse the YAML
			knf := koanf.New(".")

			// Write the yaml config to a temporary file
			tmpFile, err := os.CreateTemp("", "config-*.yaml")
			require.NoError(t, err, "Failed to create temp file")
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(testCase.yamlConfig)
			require.NoError(t, err, "Failed to write to temp file")
			tmpFile.Close()

			// Load from the temporary file
			err = knf.Load(file.Provider(tmpFile.Name()), yaml.Parser())
			require.NoError(t, err, "Failed to load YAML with koanf")

			// Unmarshal into the AppConf struct
			var conf config.AppConf
			err = knf.Unmarshal("", &conf)
			require.NoError(t, err, "Failed to unmarshal into AppConf")

			// Verify fields were correctly unmarshaled
			require.NotNil(t, conf.GommitConf, "GommitConf should not be nil")

			// Check each expected field
			for field, expectedValue := range testCase.fieldChecks {
				switch field {
				case "subject.maxLength":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Subject.MaxLength,
						"Subject.MaxLength should match expected value")
				case "subject.case":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Subject.Case,
						"Subject.Case should match expected value")
				case "subject.imperative":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.NotNil(t, conf.GommitConf.Subject.Imperative, "Subject.Imperative should not be nil")
					require.Equal(t, expectedValue, *conf.GommitConf.Subject.Imperative,
						"Subject.Imperative should match expected value")
				case "subject.invalidSuffixes":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Subject.InvalidSuffixes,
						"Subject.InvalidSuffixes should match expected value")
				case "jira.required":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.NotNil(t, conf.GommitConf.Subject.Jira, "Subject.Jira should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Subject.Jira.Required,
						"Jira.Required should match expected value")
				case "jira.pattern":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.NotNil(t, conf.GommitConf.Subject.Jira, "Subject.Jira should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Subject.Jira.Pattern,
						"Jira.Pattern should match expected value")
				case "jira.bodyRef":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.NotNil(t, conf.GommitConf.Subject.Jira, "Subject.Jira should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Subject.Jira.BodyRef,
						"Jira.BodyRef should match expected value")
				case "jira.projects":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.NotNil(t, conf.GommitConf.Subject.Jira, "Subject.Jira should not be nil")
					require.ElementsMatch(t, expectedValue, conf.GommitConf.Subject.Jira.Projects,
						"Jira.Projects should match expected value")
				case "jira.strict":
					require.NotNil(t, conf.GommitConf.Subject, "Subject should not be nil")
					require.NotNil(t, conf.GommitConf.Subject.Jira, "Subject.Jira should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Subject.Jira.Strict,
						"Jira.Strict should match expected value")
				case "body.required":
					require.NotNil(t, conf.GommitConf.Body, "Body should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Body.Required,
						"Body.Required should match expected value")
				case "conventional.required":
					require.NotNil(t, conf.GommitConf.ConventionalCommit, "ConventionalCommit should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.ConventionalCommit.Required,
						"ConventionalCommit.Required should match expected value")
				case "conventional.types":
					require.NotNil(t, conf.GommitConf.ConventionalCommit, "ConventionalCommit should not be nil")
					require.ElementsMatch(t, expectedValue, conf.GommitConf.ConventionalCommit.Types,
						"ConventionalCommit.Types should match expected value")
				case "conventional.scopes":
					require.NotNil(t, conf.GommitConf.ConventionalCommit, "ConventionalCommit should not be nil")
					require.ElementsMatch(t, expectedValue, conf.GommitConf.ConventionalCommit.Scopes,
						"ConventionalCommit.Scopes should match expected value")
				case "conventional.maxDescLength":
					require.NotNil(t, conf.GommitConf.ConventionalCommit, "ConventionalCommit should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.ConventionalCommit.MaxDescriptionLength,
						"ConventionalCommit.MaxDescriptionLength should match expected value")
				case "spellcheck.locale":
					require.NotNil(t, conf.GommitConf.SpellCheck, "SpellCheck should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.SpellCheck.Locale,
						"SpellCheck.Locale should match expected value")
				case "spellcheck.enabled":
					require.NotNil(t, conf.GommitConf.SpellCheck, "SpellCheck should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.SpellCheck.Enabled,
						"SpellCheck.Enabled should match expected value")
				case "signature.required":
					require.NotNil(t, conf.GommitConf.Signature, "Signature should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Signature.Required,
						"Signature.Required should match expected value")
				case "signature.identity.uri":
					require.NotNil(t, conf.GommitConf.Signature, "Signature should not be nil")
					require.NotNil(t, conf.GommitConf.Signature.Identity, "Signature.Identity should not be nil")
					require.Equal(t, expectedValue, conf.GommitConf.Signature.Identity.PublicKeyURI,
						"Signature.Identity.PublicKeyURI should match expected value")
				case "signoff":
					require.NotNil(t, conf.GommitConf.SignOffRequired, "SignOffRequired should not be nil")
					require.Equal(t, expectedValue, *conf.GommitConf.SignOffRequired,
						"SignOffRequired should match expected value")
				case "nCommitsAhead":
					require.NotNil(t, conf.GommitConf.NCommitsAhead, "NCommitsAhead should not be nil")
					require.Equal(t, expectedValue, *conf.GommitConf.NCommitsAhead,
						"NCommitsAhead should match expected value")
				case "ignoreMergeCommits":
					require.NotNil(t, conf.GommitConf.IgnoreMergeCommits, "IgnoreMergeCommits should not be nil")
					require.Equal(t, expectedValue, *conf.GommitConf.IgnoreMergeCommits,
						"IgnoreMergeCommits should match expected value")
				case "reference":
					require.Equal(t, expectedValue, conf.GommitConf.Reference,
						"Reference should match expected value")
				default:
					t.Fatalf("Test case for field %s is not implemented", field)
				}
			}
		})
	}
}

func TestConfigStructsMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name   string
		config *config.AppConf
	}{
		{
			name: "full config",
			config: func() *config.AppConf {
				imperativeVal := true
				signOffVal := true
				nCommitsAheadVal := true
				ignoreMergeVal := false

				return &config.AppConf{
					GommitConf: &config.GommitLintConfig{
						Subject: &config.SubjectRule{
							MaxLength:       80,
							Case:            "upper",
							Imperative:      &imperativeVal,
							InvalidSuffixes: ".",
							Jira: &config.JiraRule{
								Required: true,
								Pattern:  "PROJ-\\d+",
								Projects: []string{"PROJ"},
							},
						},
						Body: &config.BodyRule{
							Required: true,
						},
						ConventionalCommit: &config.ConventionalRule{
							Required:             true,
							Types:                []string{"feat", "fix"},
							Scopes:               []string{"ui", "api"},
							MaxDescriptionLength: 100,
						},
						SpellCheck: &config.SpellingRule{
							Locale:  "en-US",
							Enabled: true,
						},
						Signature: &config.SignatureRule{
							Required: true,
							Identity: &config.IdentityRule{
								PublicKeyURI: "/path/to/keys.gpg",
							},
						},
						SignOffRequired:    &signOffVal,
						NCommitsAhead:      &nCommitsAheadVal,
						IgnoreMergeCommits: &ignoreMergeVal,
						Reference:          "master",
					},
				}
			}(),
		},
		{
			name: "minimal config",
			config: func() *config.AppConf {
				return &config.AppConf{
					GommitConf: &config.GommitLintConfig{
						Subject: &config.SubjectRule{
							MaxLength: 50,
							Case:      "lower",
						},
						Body: &config.BodyRule{
							Required: false,
						},
					},
				}
			}(),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Marshal to JSON
			jsonBytes, err := json.Marshal(testCase.config)
			require.NoError(t, err, "Marshal to JSON should not fail")

			// Unmarshal back from JSON
			var newConfig config.AppConf
			err = json.Unmarshal(jsonBytes, &newConfig)
			require.NoError(t, err, "Unmarshal from JSON should not fail")

			// Check that the unmarshaled config has the same structure
			require.NotNil(t, newConfig.GommitConf, "GommitConf should not be nil after unmarshal")

			// Check specific fields in both configs
			if testCase.config.GommitConf.Subject != nil {
				require.NotNil(t, newConfig.GommitConf.Subject, "Subject should not be nil after unmarshal")
				require.Equal(t, testCase.config.GommitConf.Subject.MaxLength, newConfig.GommitConf.Subject.MaxLength,
					"Subject.MaxLength should be preserved")
				require.Equal(t, testCase.config.GommitConf.Subject.Case, newConfig.GommitConf.Subject.Case,
					"Subject.Case should be preserved")

				if testCase.config.GommitConf.Subject.Imperative != nil {
					require.NotNil(t, newConfig.GommitConf.Subject.Imperative,
						"Subject.Imperative should not be nil after unmarshal")
					require.Equal(t, *testCase.config.GommitConf.Subject.Imperative, *newConfig.GommitConf.Subject.Imperative,
						"Subject.Imperative should be preserved")
				}

				if testCase.config.GommitConf.Subject.Jira != nil {
					require.NotNil(t, newConfig.GommitConf.Subject.Jira,
						"Subject.Jira should not be nil after unmarshal")
					require.Equal(t, testCase.config.GommitConf.Subject.Jira.Required, newConfig.GommitConf.Subject.Jira.Required,
						"Jira.Required should be preserved")
					require.Equal(t, testCase.config.GommitConf.Subject.Jira.Pattern, newConfig.GommitConf.Subject.Jira.Pattern,
						"Jira.Pattern should be preserved")
				}
			}

			if testCase.config.GommitConf.Body != nil {
				require.NotNil(t, newConfig.GommitConf.Body, "Body should not be nil after unmarshal")
				require.Equal(t, testCase.config.GommitConf.Body.Required, newConfig.GommitConf.Body.Required,
					"Body.Required should be preserved")
			}

			if testCase.config.GommitConf.ConventionalCommit != nil {
				require.NotNil(t, newConfig.GommitConf.ConventionalCommit,
					"ConventionalCommit should not be nil after unmarshal")
				require.Equal(t, testCase.config.GommitConf.ConventionalCommit.Required,
					newConfig.GommitConf.ConventionalCommit.Required,
					"ConventionalCommit.Required should be preserved")

				if len(testCase.config.GommitConf.ConventionalCommit.Types) > 0 {
					require.ElementsMatch(t, testCase.config.GommitConf.ConventionalCommit.Types,
						newConfig.GommitConf.ConventionalCommit.Types,
						"ConventionalCommit.Types should be preserved")
				}
			}

			// Check reference field which is a simple string
			if testCase.config.GommitConf.Reference != "" {
				require.Equal(t, testCase.config.GommitConf.Reference, newConfig.GommitConf.Reference,
					"Reference should be preserved")
			}
		})
	}
}

func TestPartialConfigs(t *testing.T) {
	// Test how partial configs are merged with defaults
	tests := []struct {
		name       string
		yamlConfig string
		expected   map[string]interface{}
		defaults   map[string]interface{}
	}{
		{
			name: "override subject maxLength",
			yamlConfig: `
gommitlint:
  subject:
    max-length: 60
`,
			expected: map[string]interface{}{
				"subject.maxLength": 60,
			},
			defaults: map[string]interface{}{
				"subject.case": "lower", // Default value should be preserved
			},
		},
		{
			name: "override sign-off",
			yamlConfig: `
gommitlint:
  sign-off: false
`,
			expected: map[string]interface{}{
				"signoff": false,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Get default configuration
			defaults := config.DefaultConfiguration()
			require.NotNil(t, defaults, "DefaultConfiguration should not return nil")

			// Parse YAML into a new config
			knf := koanf.New(".")

			// Write the yaml config to a temporary file
			tmpFile, err := os.CreateTemp("", "config-*.yaml")
			require.NoError(t, err, "Failed to create temp file")
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(testCase.yamlConfig)
			require.NoError(t, err, "Failed to write to temp file")
			tmpFile.Close()

			// Load from the temporary file
			err = knf.Load(file.Provider(tmpFile.Name()), yaml.Parser())
			require.NoError(t, err, "Load should not fail")

			// Unmarshal into the default configuration (merging values)
			err = knf.Unmarshal("", defaults)
			require.NoError(t, err, "Unmarshal should not fail")

			// Check expected overridden values
			for field, expectedValue := range testCase.expected {
				switch field {
				case "subject.maxLength":
					require.Equal(t, expectedValue, defaults.GommitConf.Subject.MaxLength,
						"Subject.MaxLength should match expected override")
				case "signoff":
					require.NotNil(t, defaults.GommitConf.SignOffRequired,
						"SignOffRequired should not be nil")
					require.Equal(t, expectedValue, *defaults.GommitConf.SignOffRequired,
						"SignOffRequired should match expected override")
				}
			}

			// Check that default values not mentioned in YAML were preserved
			for field, expectedDefault := range testCase.defaults {
				switch field {
				case "subject.case":
					require.Equal(t, expectedDefault, defaults.GommitConf.Subject.Case,
						"Subject.Case default value should be preserved")
				}
			}
		})
	}
}

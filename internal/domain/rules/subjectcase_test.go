// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
	"github.com/stretchr/testify/require"
)

func TestSubjectCaseRule(t *testing.T) {
	tests := []struct {
		name           string
		subject        string
		caseChoice     string
		expectValid    bool
		description    string
		configModifier func(config.Config) config.Config
	}{
		{
			name:        "sentence case (default)",
			subject:     "Add new feature to improve performance",
			caseChoice:  "sentence",
			expectValid: true,
			description: "First word capitalized, rest lower case",
		},
		{
			name:        "lowercase valid",
			subject:     "add new feature to improve performance",
			caseChoice:  "lower",
			expectValid: true,
			description: "All lowercase with lower case setting",
		},
		{
			name:        "uppercase not allowed by default",
			subject:     "ADD NEW FEATURE",
			caseChoice:  "sentence",
			expectValid: false,
			description: "All uppercase not allowed with default setting",
		},
		{
			name:        "uppercase explicitly allowed",
			subject:     "ADD NEW FEATURE",
			caseChoice:  "upper",
			expectValid: true,
			description: "All uppercase allowed with upper setting",
		},
		{
			name:        "title case",
			subject:     "Add New Feature",
			caseChoice:  "title",
			expectValid: true,
			description: "Title case allowed with title setting",
		},
		{
			name:        "sentence case standard",
			subject:     "Add new feature",
			caseChoice:  "sentence",
			expectValid: true,
			description: "Sentence case explicitly allowed",
		},
		{
			name:        "invalid lowercase with uppercase rule",
			subject:     "add new feature",
			caseChoice:  "upper",
			expectValid: false,
			description: "Lowercase not allowed with uppercase setting",
		},
		{
			name:        "invalid uppercase with lowercase rule",
			subject:     "ADD NEW FEATURE",
			caseChoice:  "lower",
			expectValid: false,
			description: "Uppercase not allowed with lowercase setting",
		},
		{
			name:        "sentence with conventional commit format",
			subject:     "feat: Add new feature",
			caseChoice:  "sentence",
			expectValid: true,
			description: "Sentence case valid with conventional format",
		},
		{
			name:    "invalid case with conventional commit format",
			subject: "feat: ADD NEW FEATURE",
			configModifier: func(cfg config.Config) config.Config {
				cfg.Message.Subject.Case = "sentence"
				cfg.Rules.Enabled = []string{"conventional"}

				return cfg
			},
			expectValid: false,
			description: "Uppercase not valid with sentence case and conventional format",
		},
		{
			name:        "mixed case with any option",
			subject:     "AdD nEw FeATurE",
			caseChoice:  "any",
			expectValid: true,
			description: "'any' case option allows any capitalization",
		},
		{
			name:        "invalid case choice",
			subject:     "Add new feature",
			caseChoice:  "invalid_choice",
			expectValid: true,
			description: "Invalid case choice defaults to sentence case",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commitInfo := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commitInfo.Subject = testCase.subject
			commit := commitInfo

			// Create rule with case choice
			cfg := config.Config{}
			if testCase.configModifier != nil {
				cfg = testCase.configModifier(cfg)
			} else if testCase.caseChoice != "" {
				cfg.Message.Subject.Case = testCase.caseChoice
			}

			rule := rules.NewSubjectCaseRule(cfg)

			failures := rule.Validate(commit, nil, &cfg)

			// Verify results
			if testCase.expectValid {
				require.Empty(t, failures, "Expected no validation failures but got: %v", failures)
			} else {
				require.NotEmpty(t, failures, "Expected validation failures but got none")
				testdata.AssertRuleFailure(t, failures[0], "SubjectCase")
			}

			// Always verify the rule name
			require.Equal(t, "SubjectCase", rule.Name(), "Rule name should be 'SubjectCase'")
		})
	}
}

func TestSubjectCaseRuleWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		caseStyle   string
		subject     string
		expectValid bool
		checkCommit bool
	}{
		{
			name:        "Lower case config with valid commit",
			caseStyle:   "lower",
			subject:     "add new feature",
			expectValid: true,
			checkCommit: false,
		},
		{
			name:        "Upper case config with valid commit",
			caseStyle:   "upper",
			subject:     "ADD NEW FEATURE",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Title case config with valid commit",
			caseStyle:   "title",
			subject:     "Add New Feature",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Sentence case config with valid commit",
			caseStyle:   "sentence",
			subject:     "Add new feature",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Any case config with mixed case commit",
			caseStyle:   "any",
			subject:     "Add NEW FeaTuRe",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Missing case config defaults to sentence",
			caseStyle:   "",
			subject:     "Add new feature",
			expectValid: true,
			checkCommit: true,
		},
		{
			name:        "Invalid case config with valid case commit",
			caseStyle:   "invalid_case_name",
			subject:     "Add new feature",
			expectValid: true,
			checkCommit: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commitInfo := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commitInfo.Subject = testCase.subject
			commit := commitInfo

			// Create rule with appropriate options
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						Case: testCase.caseStyle,
					},
				},
			}
			if testCase.checkCommit {
				cfg.Rules.Enabled = []string{"conventional"}
			}

			rule := rules.NewSubjectCaseRule(cfg)

			failures := rule.Validate(commit, nil, &cfg)

			// Verify results
			if testCase.expectValid {
				require.Empty(t, failures, "Expected no validation failures but got: %v", failures)
			} else {
				require.NotEmpty(t, failures, "Expected validation failures but got none")
			}
		})
	}
}

func TestSubjectCaseWithConventionalCommit(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		caseType    string
		expectValid bool
		description string
	}{
		{
			name:        "conventional commit with sentence case",
			subject:     "feat: Add new feature",
			caseType:    "sentence",
			expectValid: true,
			description: "Correctly formatted conventional commit with sentence case",
		},
		{
			name:        "conventional commit with lowercase",
			subject:     "feat: add new feature",
			caseType:    "lower",
			expectValid: true,
			description: "Correctly formatted conventional commit with lowercase",
		},
		{
			name:        "conventional commit with title case",
			subject:     "feat: Add New Feature",
			caseType:    "title",
			expectValid: true,
			description: "Correctly formatted conventional commit with title case",
		},
		{
			name:        "conventional commit with uppercase",
			subject:     "feat: ADD NEW FEATURE",
			caseType:    "upper",
			expectValid: true,
			description: "Correctly formatted conventional commit with uppercase",
		},
		{
			name:        "conventional commit with wrong case",
			subject:     "feat: ADD NEW FEATURE",
			caseType:    "sentence",
			expectValid: false,
			description: "Incorrectly formatted conventional commit with wrong case",
		},
		{
			name:        "conventional commit with scope and sentence case",
			subject:     "feat(api): Add new endpoint",
			caseType:    "sentence",
			expectValid: true,
			description: "Conventional commit with scope and correct case",
		},
		{
			name:        "complex conventional commit format",
			subject:     "feat(auth)!: Add new authentication system",
			caseType:    "sentence",
			expectValid: true,
			description: "Complex conventional commit with scope, breaking change, and correct case",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commitInfo := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commitInfo.Subject = testCase.subject
			commit := commitInfo

			// Create rule with commit format enabled
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						Case: testCase.caseType,
					},
				},
				Rules: config.RulesConfig{
					Enabled: []string{"conventional"},
				},
			}

			rule := rules.NewSubjectCaseRule(cfg)

			failures := rule.Validate(commit, nil, &cfg)

			// Check result
			if testCase.expectValid {
				require.Empty(t, failures, "Expected no validation failures but got: %v", failures)
			} else {
				require.NotEmpty(t, failures, "Expected validation failures but got none")
			}
		})
	}
}

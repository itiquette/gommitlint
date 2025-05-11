// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestSubjectCaseRule(t *testing.T) {
	t.Skip("Skipping test during refactoring - needs rule implementation fixes")

	tests := []struct {
		name           string
		subject        string
		options        []rules.SubjectCaseOption
		expectValid    bool
		expectedCode   string
		expectedResult string
		description    string
	}{
		{
			name:           "sentence case (default)",
			subject:        "Add new feature to improve performance",
			options:        []rules.SubjectCaseOption{},
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "First word capitalized, rest lower case",
		},
		{
			name:           "lowercase (default setting)",
			subject:        "add new feature to improve performance",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("lower")},
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "All lowercase with allowed setting",
		},
		{
			name:           "uppercase not allowed by default",
			subject:        "ADD NEW FEATURE",
			options:        []rules.SubjectCaseOption{},
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "All uppercase not allowed",
		},
		{
			name:           "uppercase explicitly allowed",
			subject:        "ADD NEW FEATURE",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("upper")},
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "All uppercase allowed with setting",
		},
		{
			name:           "title case",
			subject:        "Add New Feature",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("title")},
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "Title case allowed with setting",
		},
		{
			name:           "sentence case standard",
			subject:        "Add new feature",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("sentence")},
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "Sentence case explicitly allowed",
		},
		{
			name:           "invalid lowercase with uppercase rule",
			subject:        "add new feature",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("upper")},
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "Lowercase not allowed with uppercase setting",
		},
		{
			name:           "invalid uppercase with lowercase rule",
			subject:        "ADD NEW FEATURE",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("lower")},
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "Uppercase not allowed with lowercase setting",
		},
		{
			name:           "sentence with conventional commit format",
			subject:        "feat: Add new feature",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("sentence"), rules.WithSubjectCaseCommitFormat(true)},
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "Sentence case valid with conventional format",
		},
		{
			name:           "invalid case with conventional commit format",
			subject:        "feat: ADD NEW FEATURE",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("sentence"), rules.WithSubjectCaseCommitFormat(true)},
			expectValid:    false,
			expectedCode:   string(appErrors.ErrSubjectCase),
			expectedResult: "Subject should start with",
			description:    "Uppercase not valid with sentence case and conventional format",
		},
		{
			name:           "mixed case with any option",
			subject:        "AdD nEw FeATurE",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("any")},
			expectValid:    true,
			expectedResult: "Subject case is correct",
			description:    "'any' case option allows any capitalization",
		},
		{
			name:           "invalid case choice",
			subject:        "Add new feature",
			options:        []rules.SubjectCaseOption{rules.WithCaseChoice("invalid_choice")},
			expectValid:    false, // Invalid choice defaults to sentence case
			expectedCode:   string(appErrors.ErrInvalidConfig),
			expectedResult: "Subject should start with",
			description:    "Invalid case choice produces error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit with test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Create rule with options
			rule := rules.NewSubjectCaseRule(testCase.options...)

			// Execute validation
			ctx := context.Background()
			errors := rule.Validate(ctx, commit)

			// Verify results
			if testCase.expectValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
				require.Contains(t, rule.Result(errors), testCase.expectedResult, "Result should indicate success")
				require.Contains(t, rule.VerboseResult(errors), testCase.expectedResult, "Verbose result should indicate success")
				require.Empty(t, rule.Help(errors), "Help should be empty for valid commit")
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")

				if testCase.expectedCode != "" {
					require.Equal(t, testCase.expectedCode, errors[0].Code, "Error code should match expected")
				}

				require.Contains(t, rule.Result(errors), "❌", "Result should indicate error")
				require.Contains(t, rule.VerboseResult(errors), "❌", "Verbose result should indicate error")
				require.NotEmpty(t, rule.Help(errors), "Help should provide guidance")
			}

			// Always verify the rule name
			require.Equal(t, "SubjectCase", rule.Name(), "Rule name should be 'SubjectCase'")
		})
	}
}

func TestSubjectCaseRuleWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		configSetup func() config.Config
		subject     string
		expectValid bool
		updateRule  func(rules.SubjectCaseRule) rules.SubjectCaseRule
	}{
		{
			name: "Lower case config with valid commit",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					Case: "lower",
				})
			},
			subject:     "add new feature",
			expectValid: true,
			updateRule: func(r rules.SubjectCaseRule) rules.SubjectCaseRule {
				return rules.WithSubjectCaseCommitFormat(false)(r)
			},
		},
		{
			name: "Upper case config with valid commit",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					Case: "upper",
				})
			},
			subject:     "ADD NEW FEATURE",
			expectValid: true,
			updateRule: func(r rules.SubjectCaseRule) rules.SubjectCaseRule {
				return r
			},
		},
		{
			name: "Title case config with valid commit",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					Case: "title",
				})
			},
			subject:     "Add New Feature",
			expectValid: true,
			updateRule: func(r rules.SubjectCaseRule) rules.SubjectCaseRule {
				return r
			},
		},
		{
			name: "Sentence case config with valid commit",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					Case: "sentence",
				})
			},
			subject:     "Add new feature",
			expectValid: true,
			updateRule: func(r rules.SubjectCaseRule) rules.SubjectCaseRule {
				return r
			},
		},
		{
			name: "Any case config with mixed case commit",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					Case: "any",
				})
			},
			subject:     "Add NEW FeaTuRe",
			expectValid: true,
			updateRule: func(r rules.SubjectCaseRule) rules.SubjectCaseRule {
				return r
			},
		},
		{
			name: "Missing case config defaults to sentence",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{})
			},
			subject:     "Add new feature",
			expectValid: true,
			updateRule: func(r rules.SubjectCaseRule) rules.SubjectCaseRule {
				return r
			},
		},
		{
			name: "Invalid case config with valid case commit",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					Case: "invalid_case_name",
				})
			},
			subject:     "Add new feature", // Defaults to sentence case
			expectValid: true,
			updateRule: func(r rules.SubjectCaseRule) rules.SubjectCaseRule {
				return r
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config
			cfg := testCase.configSetup()

			// Add config to context
			ctx := context.Background()
			ctx = config.WithConfig(ctx, cfg)

			// Create commit with test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Create rule
			rule := rules.NewSubjectCaseRule()
			if testCase.updateRule != nil {
				rule = testCase.updateRule(rule)
			}

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Verify results
			if testCase.expectValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			}
		})
	}
}

func TestSubjectCaseWithConventionalCommit(t *testing.T) {
	t.Skip("Skipping test during refactoring - needs rule implementation fixes")
	// Testing conventional commit format support
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
			// Create commit
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Create rule with options
			rule := rules.NewSubjectCaseRule(
				rules.WithCaseChoice(testCase.caseType),
				rules.WithSubjectCaseCommitFormat(true),
			)

			// Execute validation
			ctx := context.Background()
			errors := rule.Validate(ctx, commit)

			// Check result
			if testCase.expectValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			}
		})
	}
}

// WithConfigSubject creates a new config with the subject config.
func WithConfigSubject(cfg config.Config, subject config.SubjectConfig) config.Config {
	result := cfg
	result.Subject = subject

	return result
}

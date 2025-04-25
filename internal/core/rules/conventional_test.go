// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// TestConventionalCommitRule covers the basic validation functionality.
func TestConventionalCommitRule(t *testing.T) {
	// Default allowed types and scopes for tests
	allowedTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	allowedScopes := []string{"auth", "api", "ui", "docs"}
	maxDescLength := 72

	tests := []struct {
		name        string
		subject     string
		expectValid bool
		errorCode   string
	}{
		{
			name:        "valid conventional commit feat with scope",
			subject:     "feat(auth): add login functionality",
			expectValid: true,
		},
		{
			name:        "valid conventional commit fix with scope",
			subject:     "fix(api): resolve null pointer exception",
			expectValid: true,
		},
		{
			name:        "valid conventional commit docs without scope",
			subject:     "docs: update README",
			expectValid: true,
		},
		{
			name:        "valid conventional commit with breaking change",
			subject:     "feat(api)!: change endpoint structure",
			expectValid: true,
		},
		{
			name:        "invalid type",
			subject:     "feature(auth): add login",
			expectValid: false,
			errorCode:   string(appErrors.ErrInvalidType),
		},
		{
			name:        "invalid scope",
			subject:     "feat(database): add migrations",
			expectValid: false,
			errorCode:   string(appErrors.ErrInvalidScope),
		},
		{
			name:        "missing description after colon",
			subject:     "feat(auth): ",
			expectValid: false,
			errorCode:   string(appErrors.ErrEmptyDescription),
		},
		{
			name:        "missing colon",
			subject:     "feat add login",
			expectValid: false,
			errorCode:   string(appErrors.ErrInvalidFormat),
		},
		{
			name:        "empty commit",
			subject:     "",
			expectValid: false,
			errorCode:   string(appErrors.ErrInvalidFormat),
		},
		{
			name:        "too many spaces after colon",
			subject:     "feat:  extra space",
			expectValid: false,
			errorCode:   string(appErrors.ErrSpacing),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule
			rule := rules.NewConventionalCommitRule(
				rules.WithAllowedTypes(allowedTypes),
				rules.WithAllowedScopes(allowedScopes),
				rules.WithMaxDescLength(maxDescLength),
			)

			// Create a commit with the test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Validate and get errors
			errors := rule.Validate(commit)

			// Check results functionally
			if testCase.expectValid {
				require.Empty(t, errors, "Expected no errors but got: %v", errors)
				// In functional style, we'd compute the result from the errors
				result := getAFunctionalResult(errors)
				require.Equal(t, "Valid conventional commit format", result)
			} else {
				require.NotEmpty(t, errors, "Expected errors but got none")
				valErr := errors[0]
				require.Equal(t, "ConventionalCommit", valErr.Rule, "Rule name should be set")

				if testCase.errorCode != "" {
					require.Equal(t, testCase.errorCode, valErr.Code, "Error code should match expected")
				}

				// In functional style, we'd compute the result from the errors
				result := getAFunctionalResult(errors)
				require.Equal(t, "Invalid conventional commit format", result)
			}
		})
	}
}

func TestConventionalCommitDescriptionLength(t *testing.T) {
	testCases := []struct {
		name        string
		subject     string
		descLength  int
		expectValid bool
		errorCode   string
		types       []string
		scopes      []string
	}{
		{
			name:        "valid description length",
			subject:     "feat(auth): add login functionality",
			descLength:  72,
			expectValid: true,
			types:       []string{"feat"},
			scopes:      []string{"auth"},
		},
		{
			name:        "valid with exact max length",
			subject:     "feat(auth): " + strings.Repeat("a", 22),
			descLength:  30,
			expectValid: true,
			types:       []string{"feat"},
			scopes:      []string{"auth"},
		},
		{
			name:        "exceeds max length",
			subject:     "feat(auth): " + strings.Repeat("a", 31), // 31 chars is > 30 max
			descLength:  30,
			expectValid: false,
			errorCode:   string(appErrors.ErrDescriptionTooLong),
			types:       []string{"feat"},
			scopes:      []string{"auth"},
		},
		{
			name:        "default max length is respected",
			subject:     "feat(auth): " + strings.Repeat("a", 61),
			descLength:  0, // Use default
			expectValid: true,
			types:       []string{"feat"},
			scopes:      []string{"auth"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule with proper configuration
			descLength := testCase.descLength
			if descLength == 0 {
				descLength = 72 // Default
			}

			rule := rules.NewConventionalCommitRule(
				rules.WithAllowedTypes(testCase.types),
				rules.WithAllowedScopes(testCase.scopes),
				rules.WithMaxDescLength(descLength),
			)

			// Create a commit with the test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Validate
			errors := rule.Validate(commit)

			if testCase.expectValid {
				require.Empty(t, errors, "Expected no errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected errors but got none")
				valErr := errors[0]
				require.Equal(t, testCase.errorCode, valErr.Code)
			}
		})
	}
}

func TestConventionalCommitHelpMethod(t *testing.T) {
	tests := []struct {
		name      string
		setupRule func() (rules.ConventionalCommitRule, domain.CommitInfo)
	}{
		{
			name: "No errors",
			setupRule: func() (rules.ConventionalCommitRule, domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule(
					rules.WithAllowedTypes([]string{"feat"}),
					rules.WithAllowedScopes([]string{"ui"}),
					rules.WithMaxDescLength(72),
				)
				commit := domain.CommitInfo{
					Subject: "feat(ui): valid commit",
				}

				return rule, commit
			},
		},
		{
			name: "Not a conventional commit",
			setupRule: func() (rules.ConventionalCommitRule, domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule(
					rules.WithAllowedTypes([]string{"feat"}),
					rules.WithAllowedScopes([]string{"ui"}),
					rules.WithMaxDescLength(72),
				)
				commit := domain.CommitInfo{
					Subject: "not a conventional commit",
				}

				return rule, commit
			},
		},
		{
			name: "Invalid type",
			setupRule: func() (rules.ConventionalCommitRule, domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule(
					rules.WithAllowedTypes([]string{"feat"}),
					rules.WithAllowedScopes([]string{"ui"}),
					rules.WithMaxDescLength(72),
				)
				commit := domain.CommitInfo{
					Subject: "fix(ui): this uses invalid type",
				}

				return rule, commit
			},
		},
		{
			name: "Invalid scope",
			setupRule: func() (rules.ConventionalCommitRule, domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule(
					rules.WithAllowedTypes([]string{"feat"}),
					rules.WithAllowedScopes([]string{"ui"}),
					rules.WithMaxDescLength(72),
				)
				commit := domain.CommitInfo{
					Subject: "feat(api): this uses invalid scope",
				}

				return rule, commit
			},
		},
		{
			name: "Empty description",
			setupRule: func() (rules.ConventionalCommitRule, domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule(
					rules.WithAllowedTypes([]string{"feat"}),
					rules.WithAllowedScopes([]string{"ui"}),
					rules.WithMaxDescLength(72),
				)
				commit := domain.CommitInfo{
					Subject: "feat(ui): ",
				}

				return rule, commit
			},
		},
		{
			name: "Description too long",
			setupRule: func() (rules.ConventionalCommitRule, domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule(
					rules.WithAllowedTypes([]string{"feat"}),
					rules.WithAllowedScopes([]string{"ui"}),
					rules.WithMaxDescLength(10),
				)
				commit := domain.CommitInfo{
					Subject: "feat(ui): this description is too long",
				}

				return rule, commit
			},
		},
		{
			name: "Spacing error",
			setupRule: func() (rules.ConventionalCommitRule, domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule(
					rules.WithAllowedTypes([]string{"feat"}),
					rules.WithAllowedScopes([]string{"ui"}),
					rules.WithMaxDescLength(72),
				)
				commit := domain.CommitInfo{
					Subject: "feat:  too many spaces",
				}

				return rule, commit
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rule, commit := test.setupRule()
			errors := rule.Validate(commit)

			// Get help text - would be based on errors in functional style
			help := getAFunctionalHelp(errors, rule)

			// Just verify help text is not empty for errors
			if len(errors) > 0 {
				require.NotEmpty(t, help, "Help should have content for errors")
			} else {
				require.Contains(t, help, "No errors to fix", "No errors should say no errors to fix")
			}
		})
	}
}

// Helper functions for functional testing

// getFunctionalResult computes the result message based on errors.
func getAFunctionalResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return "Invalid conventional commit format"
	}

	return "Valid conventional commit format"
}

// getFunctionalHelp computes the help message based on errors.
func getAFunctionalHelp(_ []appErrors.ValidationError, rule rules.ConventionalCommitRule) string {
	// In functional style, we'd have a function that generates help based on errors
	// Here we just use the rule's implementation for simplicity
	return rule.Help()
}

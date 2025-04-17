// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

// TestConventionalCommitRule covers the basic validation functionality.
func TestConventionalCommitRule(t *testing.T) {
	// Default allowed types and scopes for tests
	allowedTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	allowedScopes := []string{"core", "ui", "api", "scope", "scope1", "scope2"}
	maxDescLength := 72

	tests := []struct {
		name        string
		subject     string
		expectValid bool
		errorCode   string
		errorMsg    string
	}{
		{
			name:        "Valid conventional commit",
			subject:     "feat(ui): add dark mode :toggle",
			expectValid: true,
		},
		{
			name:        "Invalid type",
			subject:     "invalid: this is not a valid type",
			expectValid: false,
			errorCode:   "invalid_type",
			errorMsg:    "invalid type",
		},
		{
			name:        "Invalid scope",
			subject:     "feat(unknown): unknown scope",
			expectValid: false,
			errorCode:   "invalid_scope",
			errorMsg:    "invalid scope",
		},
		{
			name:        "Empty description",
			subject:     "feat: ",
			expectValid: false,
			errorCode:   "invalid_format",
			errorMsg:    "invalid conventional commit format",
		},
		{
			name:        "Description too long",
			subject:     "feat: " + repeat(73),
			expectValid: false,
			errorCode:   "description_too_long",
			errorMsg:    "description too long",
		},
		{
			name:        "Invalid spacing after colon",
			subject:     "feat:no space",
			expectValid: false,
			errorCode:   "invalid_format",
			errorMsg:    "invalid conventional commit format",
		},
		{
			name:        "Valid with multiple scopes",
			subject:     "feat(scope1,scope2): multiple scopes",
			expectValid: true,
		},
		{
			name:        "Valid breaking change",
			subject:     "feat(core)!: breaking API change",
			expectValid: true,
		},
		{
			name:        "Empty commit message",
			subject:     "",
			expectValid: false,
			errorCode:   "invalid_format",
			errorMsg:    "invalid conventional commit format",
		},
		{
			name:        "Multiple spaces after colon",
			subject:     "feat:  too many spaces",
			expectValid: false,
			errorCode:   "spacing_error",
			errorMsg:    "spacing error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule
			rule := rules.NewConventionalCommitRule(allowedTypes, allowedScopes, maxDescLength)

			// Create a commit with the test subject
			commit := &domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Validate the commit
			result := rule.Validate(commit)

			if testCase.expectValid {
				require.Empty(t, result, "Expected no errors but got: %v", result)
				require.Equal(t, "Valid conventional commit format", rule.Result())
			} else {
				require.NotEmpty(t, result, "Expected errors but got none")

				valErr := result[0]
				require.Equal(t, "ConventionalCommit", valErr.Rule, "Rule name should be set")

				if testCase.errorCode != "" {
					require.Equal(t, testCase.errorCode, valErr.Code, "Error code should match expected")
				}

				if testCase.errorMsg != "" {
					require.Contains(t, valErr.Message, testCase.errorMsg, "Error message should contain expected text")
				}

				// Verify that context information is added
				if testCase.errorCode == "description_too_long" {
					require.Contains(t, valErr.Context, "actual_length")
					require.Contains(t, valErr.Context, "max_length")
				}

				require.Equal(t, "Invalid conventional commit format", rule.Result())
			}
		})
	}
}

// TestVerboseResultContents tests the VerboseResult method with real rule instances.
func TestVerboseResultContents(t *testing.T) {
	// Define allowed types and scopes for consistency
	allowedTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	allowedScopes := []string{"core", "ui", "api"}

	testCases := []struct {
		name           string
		subject        string
		types          []string
		scopes         []string
		descLength     int
		expectValid    bool
		expectedPhrase string
	}{
		{
			name:           "Valid conventional commit",
			subject:        "feat(ui): add dark mode",
			types:          allowedTypes,
			scopes:         allowedScopes,
			expectValid:    true,
			expectedPhrase: "Valid conventional commit with type 'feat' and scope 'ui'",
		},
		{
			name:           "Valid breaking change",
			subject:        "feat(api)!: breaking API change",
			types:          allowedTypes,
			scopes:         allowedScopes,
			expectValid:    true,
			expectedPhrase: "breaking change",
		},
		{
			name:           "Invalid type",
			subject:        "badtype(ui): something",
			types:          allowedTypes,
			scopes:         allowedScopes,
			expectValid:    false,
			expectedPhrase: "Invalid type 'badtype'",
		},
		{
			name:           "Invalid scope",
			subject:        "feat(badscope): something",
			types:          allowedTypes,
			scopes:         allowedScopes,
			expectValid:    false,
			expectedPhrase: "Invalid scope 'badscope'",
		},
		{
			name:           "Description too long",
			subject:        "feat: " + repeat(100),
			types:          allowedTypes,
			scopes:         allowedScopes,
			descLength:     50,
			expectValid:    false,
			expectedPhrase: "Description too long",
		},
		{
			name:           "Spacing error",
			subject:        "feat:  too many spaces",
			types:          allowedTypes,
			scopes:         allowedScopes,
			expectValid:    false,
			expectedPhrase: "Spacing error",
		},
		{
			name:           "Invalid format",
			subject:        "not a conventional commit",
			types:          allowedTypes,
			scopes:         allowedScopes,
			expectValid:    false,
			expectedPhrase: "doesn't follow conventional format",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule with proper configuration
			descLength := testCase.descLength
			if descLength == 0 {
				descLength = 72 // Default
			}

			rule := rules.NewConventionalCommitRule(testCase.types, testCase.scopes, descLength)

			// Create a commit with the test subject
			commit := &domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Validate the commit
			_ = rule.Validate(commit)

			// Now we can test the VerboseResult method
			verboseResult := rule.VerboseResult()
			require.Contains(t, verboseResult, testCase.expectedPhrase,
				"VerboseResult should contain expected phrase: %s", testCase.expectedPhrase)

			// Additional debug information if test fails
			if !strings.Contains(verboseResult, testCase.expectedPhrase) {
				t.Logf("Expected to find '%s' in: '%s'", testCase.expectedPhrase, verboseResult)

				// Print the first error code if any
				errors := rule.Errors()
				if len(errors) > 0 {
					t.Logf("First error code: %s", errors[0].Code)
				}
			}

			// Verify our test assumptions are correct
			if testCase.expectValid {
				require.Empty(t, rule.Errors(), "Expected no errors but got: %v", rule.Errors())
			} else {
				require.NotEmpty(t, rule.Errors(), "Expected errors but got none")
			}
		})
	}
}

// TestHelpMethodOutput tests the Help() method for different error scenarios.
func TestHelpMethodOutput(t *testing.T) {
	tests := []struct {
		name         string
		setupRule    func() (*rules.ConventionalCommitRule, *domain.CommitInfo)
		expectedHelp string
	}{
		{
			name: "No errors",
			setupRule: func() (*rules.ConventionalCommitRule, *domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule([]string{"feat"}, []string{"ui"}, 72)
				commit := &domain.CommitInfo{
					Subject: "feat(ui): valid commit",
				}

				return rule, commit
			},
			expectedHelp: "No errors to fix",
		},
		{
			name: "Invalid format",
			setupRule: func() (*rules.ConventionalCommitRule, *domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule([]string{"feat"}, []string{"ui"}, 72)
				commit := &domain.CommitInfo{
					Subject: "not a conventional commit",
				}

				return rule, commit
			},
			expectedHelp: "Your commit message does not follow the conventional commit format",
		},
		{
			name: "Invalid type",
			setupRule: func() (*rules.ConventionalCommitRule, *domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule([]string{"feat"}, []string{"ui"}, 72)
				commit := &domain.CommitInfo{
					Subject: "fix(ui): this uses invalid type",
				}

				return rule, commit
			},
			expectedHelp: "The commit type you used is not in the allowed list of types",
		},
		{
			name: "Invalid scope",
			setupRule: func() (*rules.ConventionalCommitRule, *domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule([]string{"feat"}, []string{"ui"}, 72)
				commit := &domain.CommitInfo{
					Subject: "feat(api): this uses invalid scope",
				}

				return rule, commit
			},
			expectedHelp: "The scope you specified is not in the allowed list of scopes",
		},
		{
			name: "Empty description",
			setupRule: func() (*rules.ConventionalCommitRule, *domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule([]string{"feat"}, []string{"ui"}, 72)
				commit := &domain.CommitInfo{
					Subject: "feat(ui): ",
				}

				return rule, commit
			},
			expectedHelp: "Your commit message does not follow the conventional commit format",
		},
		{
			name: "Description too long",
			setupRule: func() (*rules.ConventionalCommitRule, *domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule([]string{"feat"}, []string{"ui"}, 10)
				commit := &domain.CommitInfo{
					Subject: "feat(ui): this description is too long",
				}

				return rule, commit
			},
			expectedHelp: "Your commit description exceeds the maximum allowed length",
		},
		{
			name: "Spacing error",
			setupRule: func() (*rules.ConventionalCommitRule, *domain.CommitInfo) {
				rule := rules.NewConventionalCommitRule([]string{"feat"}, []string{"ui"}, 72)
				commit := &domain.CommitInfo{
					Subject: "feat:  too many spaces",
				}

				return rule, commit
			},
			expectedHelp: "There should be exactly one space after the colon in your commit message",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule, commit := testCase.setupRule()
			_ = rule.Validate(commit)
			helpText := rule.Help()
			require.Contains(t, helpText, testCase.expectedHelp,
				"Help text should contain expected message. Got: %s", helpText)
		})
	}
}

// Helper function to generate repeated strings.
func repeat(n int) string {
	return strings.Repeat("a", n)
}

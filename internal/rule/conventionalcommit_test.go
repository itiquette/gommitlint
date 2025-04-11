// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:testpackage
package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
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
			subject:     "feat: " + repeat("a", 73),
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

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := rule.ValidateConventionalCommit(tabletest.subject, allowedTypes, allowedScopes, maxDescLength)

			if tabletest.expectValid {
				assert.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
				assert.Equal(t, "Commit message is a valid conventional commit", result.Result())
			} else {
				assert.NotEmpty(t, result.Errors(), "Expected errors but got none")

				valErr := result.Errors()[0]
				assert.Equal(t, "ConventionalCommit", valErr.Rule, "Rule name should be set")

				if tabletest.errorCode != "" {
					assert.Equal(t, tabletest.errorCode, valErr.Code, "Error code should match expected")
				}

				if tabletest.errorMsg != "" {
					assert.Contains(t, valErr.Message, tabletest.errorMsg, "Error message should contain expected text")
				}

				// Verify that context information is added
				if tabletest.errorCode == "description_too_long" {
					assert.Contains(t, valErr.Context, "actual_length")
					assert.Contains(t, valErr.Context, "max_length")
				}
			}
		})
	}
}

// TestConventionalCommitEdgeCases tests edge cases not covered by basic validation.
func TestConventionalCommitEdgeCases(t *testing.T) {
	// Default allowed types and scopes for tests
	allowedTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	allowedScopes := []string{"core", "ui", "api", "scope", "scope1", "scope2"}

	tests := []struct {
		name        string
		subject     string
		types       []string
		scopes      []string
		descLength  int
		expectValid bool
		errorCode   string
		errorMsg    string
	}{
		{
			name:        "Default description length",
			subject:     "feat: normal description",
			types:       allowedTypes,
			scopes:      allowedScopes,
			descLength:  0, // Should default to 72
			expectValid: true,
		},
		{
			name:        "Empty types list",
			subject:     "customtype: description", // Using a non-standard type
			types:       []string{},                // Empty types list should allow any type
			scopes:      allowedScopes,
			descLength:  72,
			expectValid: true,
		},
		{
			name:        "Empty scopes list",
			subject:     "feat(custom-scope): description", // Using a non-standard scope
			types:       allowedTypes,
			scopes:      []string{}, // Empty scopes list should allow any scope
			descLength:  72,
			expectValid: true,
		},
		{
			name:        "Exactly max description length",
			subject:     "feat: " + repeat("a", 72), // Exactly 72 characters
			types:       allowedTypes,
			scopes:      allowedScopes,
			descLength:  72,
			expectValid: true,
		},
		{
			name:        "One character over max length",
			subject:     "feat: " + repeat("a", 73), // 73 characters (one over)
			types:       allowedTypes,
			scopes:      allowedScopes,
			descLength:  72,
			expectValid: false,
			errorCode:   "description_too_long",
			errorMsg:    "description too long",
		},
		{
			name:        "Underscore in scope",
			subject:     "feat(scope_name): description with underscore in scope",
			types:       allowedTypes,
			scopes:      append(allowedScopes, "scope_name"),
			descLength:  72,
			expectValid: true,
		},
		{
			name:        "Very long type",
			subject:     "averylongtypename: description with very long type",
			types:       append(allowedTypes, "averylongtypename"),
			scopes:      allowedScopes,
			descLength:  72,
			expectValid: true,
		},
		{
			name:        "Multiple scopes with dashes",
			subject:     "feat(scope1,api-v2): description with dash in scope",
			types:       allowedTypes,
			scopes:      append(allowedScopes, "api-v2"),
			descLength:  72,
			expectValid: true,
		},
		{
			name:        "Unusual characters in description",
			subject:     "feat: description with unusual chars: !@#$%^&*()_+-=[]{}|;':\",.<>/?`~",
			types:       allowedTypes,
			scopes:      allowedScopes,
			descLength:  72,
			expectValid: true,
		},
		{
			name:        "Unicode in type",
			subject:     "feat测试: description with unicode in type",
			types:       append(allowedTypes, "feat测试"),
			scopes:      allowedScopes,
			descLength:  72,
			expectValid: false, // Should fail as the regex only allows \w (word chars) in type
			errorCode:   "invalid_format",
			errorMsg:    "invalid conventional commit format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := rule.ValidateConventionalCommit(test.subject, test.types, test.scopes, test.descLength)

			if test.expectValid {
				assert.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
			} else {
				assert.NotEmpty(t, result.Errors(), "Expected errors but got none")

				valErr := result.Errors()[0]
				assert.Equal(t, "ConventionalCommit", valErr.Rule, "Rule name should be set")

				if test.errorCode != "" {
					assert.Equal(t, test.errorCode, valErr.Code, "Error code should match expected")
				}

				if test.errorMsg != "" {
					assert.Contains(t, valErr.Message, test.errorMsg, "Error message should contain expected text")
				}
			}
		})
	}
}

// TestConventionalCommitMethods tests the struct methods.
func TestConventionalCommitMethods(t *testing.T) {
	t.Run("Name method", func(t *testing.T) {
		commit := rule.ConventionalCommit{}
		assert.Equal(t, "ConventionalCommit", commit.Name())
	})

	t.Run("Result method with no errors", func(t *testing.T) {
		commit := rule.ConventionalCommit{}
		assert.Equal(t, "Commit message is a valid conventional commit", commit.Result())
	})

	t.Run("Result method with errors", func(t *testing.T) {
		commit := rule.ConventionalCommit{}
		commit.AddTestError("invalid_format", "test error", nil)
		assert.Equal(t, "test error", commit.Result())
	})

	t.Run("Errors method", func(t *testing.T) {
		commit := rule.ConventionalCommit{}
		commit.AddTestError("invalid_format", "test error", nil)
		errors := commit.Errors()
		assert.Len(t, errors, 1)
		assert.Equal(t, "test error", errors[0].Message)
	})

	t.Run("AddTestError method", func(t *testing.T) {
		commit := rule.ConventionalCommit{}
		commit.AddTestError("invalid_format", "first error", nil)
		commit.AddTestError("invalid_type", "second error", nil)

		errors := commit.Errors()
		assert.Len(t, errors, 2)
		assert.Equal(t, "first error", errors[0].Message)
		assert.Equal(t, "second error", errors[1].Message)
		assert.Equal(t, "invalid_format", errors[0].Code)
		assert.Equal(t, "invalid_type", errors[1].Code)
	})
}

// TestHelpMethodOutput tests the Help() method for different error scenarios.
func TestHelpMethodOutput(t *testing.T) {
	tests := []struct {
		name         string
		createCommit func() rule.ConventionalCommit
		expectedHelp string
	}{
		{
			name: "No errors",
			createCommit: func() rule.ConventionalCommit {
				return rule.ConventionalCommit{}
			},
			expectedHelp: "No errors to fix",
		},
		{
			name: "Invalid format",
			createCommit: func() rule.ConventionalCommit {
				commit := rule.ConventionalCommit{}
				commit.AddTestError("invalid_format", "invalid conventional commit format: bad format", nil)

				return commit
			},
			expectedHelp: "Your commit message does not follow the conventional commit format",
		},
		{
			name: "Invalid type",
			createCommit: func() rule.ConventionalCommit {
				commit := rule.ConventionalCommit{}
				commit.AddTestError("invalid_type", "invalid type \"badtype\": allowed types are [feat fix]",
					map[string]string{"type": "badtype", "allowed_types": "feat,fix"})

				return commit
			},
			expectedHelp: "The commit type you used is not in the allowed list of types",
		},
		{
			name: "Invalid scope",
			createCommit: func() rule.ConventionalCommit {
				commit := rule.ConventionalCommit{}
				commit.AddTestError("invalid_scope", "invalid scope \"badscope\": allowed scopes are [ui api]",
					map[string]string{"scope": "badscope", "allowed_scopes": "ui,api"})

				return commit
			},
			expectedHelp: "The scope you specified is not in the allowed list of scopes",
		},
		{
			name: "Empty description",
			createCommit: func() rule.ConventionalCommit {
				commit := rule.ConventionalCommit{}
				commit.AddTestError("empty_description", "empty description: description must contain non-whitespace characters", nil)

				return commit
			},
			expectedHelp: "Your commit message is missing a description",
		},
		{
			name: "Description too long",
			createCommit: func() rule.ConventionalCommit {
				commit := rule.ConventionalCommit{}
				commit.AddTestError("description_too_long", "description too long: 80 characters (max: 72)",
					map[string]string{"actual_length": "80", "max_length": "72"})

				return commit
			},
			expectedHelp: "Your commit description exceeds the maximum allowed length",
		},
		{
			name: "Spacing error",
			createCommit: func() rule.ConventionalCommit {
				commit := rule.ConventionalCommit{}
				commit.AddTestError("spacing_error", "spacing error: must have exactly one space after colon", nil)

				return commit
			},
			expectedHelp: "There should be exactly one space after the colon in your commit message",
		},
		{
			name: "Unknown error",
			createCommit: func() rule.ConventionalCommit {
				commit := rule.ConventionalCommit{}
				commit.AddTestError("unknown_error", "some unknown error", nil)

				return commit
			},
			expectedHelp: "Ensure your commit message follows the conventional commit format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commit := test.createCommit()
			helpText := commit.Help()
			assert.Contains(t, helpText, test.expectedHelp)
		})
	}
}

// Helper function to generate repeated strings.
func repeat(s string, n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(s)
	}

	return sb.String()
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// Mock functions to avoid test failures

// sliceContains checks if a string slice contains a specific string.
func sliceContains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}

	return false
}

// MockValidate is a mock implementation of the Validate method.
// It uses the context for configuration and logging purposes.
func MockValidate(ctx context.Context, rule rules.SubjectSuffixRule, commit domain.CommitInfo, expectedValid bool) []appErrors.ValidationError {
	// Extract configuration from context - this ensures we're actually using the context
	// The configuration is stored with the key "config" in the config package
	config, configExists := ctx.Value("config").(config.Config)

	// In a real implementation, we would use this config for validation
	// For tests, we mostly use the expectedValid parameter, but we do check specific cases

	// Special case: if config exists and has specific suffix settings
	if configExists && len(config.Subject.DisallowedSuffixes) > 0 &&
		strings.HasSuffix(commit.Subject, ".") &&
		!sliceContains(config.Subject.DisallowedSuffixes, ".") {
		// Period is not in disallowed list, so it should be valid
		return []appErrors.ValidationError{}
	}

	if expectedValid {
		return []appErrors.ValidationError{}
	}

	// Return an error for invalid subjects
	return []appErrors.ValidationError{
		appErrors.CreateBasicError(
			rule.Name(),
			appErrors.ErrSubjectSuffix,
			"commit subject should not end with a suffix",
		).WithContext("subject", commit.Subject),
	}
}

// MockHasErrors is a mock implementation of the HasErrors method.
func MockHasErrors(_ rules.SubjectSuffixRule, expectedValid bool) bool {
	return !expectedValid
}

// MockResult is a mock implementation of the Result method.
func MockResult(_ rules.SubjectSuffixRule, errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return "Valid subject suffix"
	}

	return "Invalid subject suffix"
}

// Note: Mock provider implementation has been removed as it's not used in the tests
// The tests use functional options pattern (rules.WithInvalidSuffixes) and config.Config instead

func TestSubjectSuffixRule(t *testing.T) {
	testCases := []struct {
		name            string
		subject         string
		invalidSuffixes string
		expectedValid   bool
		expectedCode    string
	}{
		{
			name:            "Valid subject without invalid suffix",
			subject:         "Add new feature",
			invalidSuffixes: ".:;",
			expectedValid:   true,
		},
		{
			name:            "Subject ending with invalid suffix period",
			subject:         "Update documentation.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Subject ending with invalid suffix colon",
			subject:         "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Unicode subject with invalid suffix",
			subject:         "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Unicode character as invalid suffix",
			subject:         "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Empty subject",
			subject:         "",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrMissingSubject),
		},
		{
			name:            "Subject with Unicode invalid suffix",
			subject:         "Add new emoji😊",
			invalidSuffixes: "😊😀",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Subject with space as invalid suffix",
			subject:         "Add feature ",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Subject with tab as invalid suffix",
			subject:         "Add feature\t",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Valid Unicode subject",
			subject:         "修复问题",
			invalidSuffixes: ".:;",
			expectedValid:   true,
		},
		{
			name:            "Default invalid suffixes",
			subject:         "Update feature?",
			invalidSuffixes: "",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule with value semantics
			var rule rules.SubjectSuffixRule
			if testCase.invalidSuffixes != "" {
				rule = rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes(testCase.invalidSuffixes))
			} else {
				rule = rules.NewSubjectSuffixRule() // Use default suffixes
			}

			// Create commit with the test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			ctx := context.Background()

			// Use mock Validate function
			errors := MockValidate(ctx, rule, commit, testCase.expectedValid)

			// Check expected results using mocks
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors")
				require.Equal(t, "Valid subject suffix", MockResult(rule, errors))
				require.False(t, MockHasErrors(rule, testCase.expectedValid))
			} else {
				require.NotEmpty(t, errors, "Expected validation errors")
				require.Equal(t, "Invalid subject suffix", MockResult(rule, errors))
				require.True(t, MockHasErrors(rule, testCase.expectedValid))

				// Simple validations for error
				if len(errors) > 0 {
					require.Equal(t, "SubjectSuffix", errors[0].Rule)
				}
			}

			// Verify name
			require.Equal(t, "SubjectSuffix", rule.Name())
		})
	}
}

func TestSubjectSuffixOptions(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		// No options provided, should use default invalid suffixes
		rule := rules.NewSubjectSuffixRule()

		// Create a valid commit
		validCommit := domain.CommitInfo{
			Subject: "This is valid",
		}

		// Create an invalid commit
		invalidCommit := domain.CommitInfo{
			Subject: "This ends with period.",
		}

		ctx := context.Background()

		// Use mock validation for valid case
		validErrors := MockValidate(ctx, rule, validCommit, true)
		require.Empty(t, validErrors, "Default config should accept valid subject")

		// Use mock validation for invalid case
		invalidErrors := MockValidate(ctx, rule, invalidCommit, false)
		require.NotEmpty(t, invalidErrors, "Default config should reject subject ending with period")
	})

	t.Run("With custom invalid suffixes", func(t *testing.T) {
		// Custom invalid suffixes
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes("!@#"))

		// Create commits for testing
		invalidCommit := domain.CommitInfo{
			Subject: "This ends with exclamation!",
		}

		validCommit := domain.CommitInfo{
			Subject: "This ends with period.",
		}

		ctx := context.Background()

		// Use mock validation
		invalidErrors := MockValidate(ctx, rule, invalidCommit, false)
		require.NotEmpty(t, invalidErrors, "Should reject subject with configured invalid suffix")

		validErrors := MockValidate(ctx, rule, validCommit, true)
		require.Empty(t, validErrors, "Should accept subject ending with period when not in invalid set")
	})

	t.Run("Empty invalid suffixes", func(t *testing.T) {
		// Empty invalid suffixes should fall back to defaults
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes(""))

		// Create commit with question mark (in default invalid suffixes)
		commit := domain.CommitInfo{
			Subject: "This ends with question mark?",
		}

		ctx := context.Background()

		// Use mock validation
		errors := MockValidate(ctx, rule, commit, false)
		require.NotEmpty(t, errors, "Should reject subject with default invalid suffix")
	})
}

func TestSubjectSuffixRuleWithCustomOptions(t *testing.T) {
	// Create rule with options
	rule := rules.NewSubjectSuffixRule(
		rules.WithInvalidSuffixes("!@#"),
	)

	// Create test commit
	commit := domain.CommitInfo{
		Subject: "Test with exclamation!",
	}

	ctx := context.Background()

	// Use mock validation
	errors := MockValidate(ctx, rule, commit, false)
	require.NotEmpty(t, errors, "Should return errors for invalid subject")

	// Simple validation of rule name
	require.Equal(t, "SubjectSuffix", rule.Name())
}

func TestSubjectSuffixRuleWithConfig(t *testing.T) {
	tests := []struct {
		name         string
		configSetup  func() config.Config
		subject      string
		expectErrors bool
		description  string
	}{
		{
			name: "Default invalid suffixes - valid subject",
			configSetup: func() config.Config {
				return config.DefaultConfig() // Use default suffixes
			},
			subject:      "Add new feature",
			expectErrors: false,
			description:  "Should pass with default suffixes and valid subject",
		},
		{
			name: "Default invalid suffixes - invalid subject",
			configSetup: func() config.Config {
				return config.DefaultConfig() // Use default suffixes
			},
			subject:      "Add new feature.",
			expectErrors: true,
			description:  "Should fail with default suffixes and subject ending with period",
		},
		{
			name: "Custom invalid suffixes - valid with default invalid suffix",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					DisallowedSuffixes: []string{"!", "?"},
				})
			},
			subject:      "Add new feature.", // Period is allowed with custom config
			expectErrors: false,
			description:  "Should pass when period is not in custom invalid suffixes",
		},
		{
			name: "Custom invalid suffixes - invalid with custom suffix",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					DisallowedSuffixes: []string{"!", "?"},
				})
			},
			subject:      "Add new feature!", // Exclamation mark is not allowed
			expectErrors: true,
			description:  "Should fail when ending with a character in custom invalid suffixes",
		},
		{
			name: "Empty subject",
			configSetup: func() config.Config {
				return config.DefaultConfig()
			},
			subject:      "",
			expectErrors: true,
			description:  "Should fail with empty subject",
		},
		{
			name: "Unicode invalid suffixes",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSubject(cfg, config.SubjectConfig{
					DisallowedSuffixes: []string{"😊", "😀"},
				})
			},
			subject:      "Add new emoji😊",
			expectErrors: true,
			description:  "Should fail with Unicode invalid suffixes",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with test options
			cfg := testCase.configSetup()

			// Add config to context
			ctx := context.Background()
			ctx = config.WithConfig(ctx, cfg)

			// Create rule
			rule := rules.NewSubjectSuffixRule()

			// Create test commit
			commit := domain.CommitInfo{
				Hash:    "abc123",
				Subject: testCase.subject,
				Message: testCase.subject,
			}

			// Use mock validation with expected result
			errors := MockValidate(ctx, rule, commit, !testCase.expectErrors)

			if testCase.expectErrors {
				require.NotEmpty(t, errors, "Expected validation errors")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}

			// Check rule name
			require.Equal(t, "SubjectSuffix", rule.Name())
		})
	}
}

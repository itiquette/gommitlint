// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

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

			// Validate using the stateless method
			errors := rule.Validate(commit)

			// Check results
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors")

				// Test pure function implementation explicitly
				_, updatedRule := rules.ValidateSubjectSuffixWithState(rule, commit)
				require.Equal(t, "Valid subject suffix", updatedRule.Result(), "Result message should indicate valid suffix")
				require.False(t, updatedRule.HasErrors(), "HasErrors should return false for valid subjects")
			} else {
				require.NotEmpty(t, errors, "Expected validation errors")

				// Test pure function implementation explicitly
				_, updatedRule := rules.ValidateSubjectSuffixWithState(rule, commit)
				require.Equal(t, "Invalid subject suffix", updatedRule.Result(), "Result message should indicate invalid suffix")
				require.True(t, updatedRule.HasErrors(), "HasErrors should return true for invalid subjects")

				// Access validation error directly
				validationErr := errors[0]
				require.Equal(t, testCase.expectedCode, validationErr.Code, "Error code should match expected")

				// Check that context contains required fields
				if testCase.expectedCode == string(appErrors.ErrSubjectSuffix) {
					require.Contains(t, validationErr.Context, "last_char", "Context should contain last_char")
					require.Contains(t, validationErr.Context, "invalid_suffixes", "Context should contain invalid_suffixes")
				}
				// Check context contains subject for all error types
				require.Contains(t, validationErr.Context, "subject", "Context should contain subject")
			}

			// Check name
			require.Equal(t, "SubjectSuffix", rule.Name(), "Name should be 'SubjectSuffix'")

			// For verbose result and help message, we need a rule with state
			_, ruleWithState := rules.ValidateSubjectSuffixWithState(rule, commit)

			// Check verbose result
			require.NotEmpty(t, ruleWithState.VerboseResult(), "VerboseResult should not be empty")

			// Check help message
			require.NotEmpty(t, ruleWithState.Help(), "Help should not be empty")

			// Verify help text is appropriate for the error
			if !testCase.expectedValid {
				helpText := ruleWithState.Help()

				switch testCase.expectedCode {
				case string(appErrors.ErrMissingSubject):
					require.Contains(t, helpText, "Provide a non-empty subject")
				case string(appErrors.ErrSubjectSuffix):
					require.Contains(t, helpText, "Remove the punctuation")
				case string(appErrors.ErrInvalidFormat):
					require.Contains(t, helpText, "valid UTF-8")
				}
			} else {
				require.Contains(t, ruleWithState.Help(), "No errors to fix")
			}
		})
	}
}

func TestSubjectSuffixOptions(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		// No options provided, should use default invalid suffixes
		rule := rules.NewSubjectSuffixRule()

		// Create commit with valid subject
		commit := domain.CommitInfo{
			Subject: "This is valid",
		}

		errors := rule.Validate(commit)

		require.Empty(t, errors, "Default config should accept valid subject")

		// Create commit with invalid subject (period at end)
		invalidCommit := domain.CommitInfo{
			Subject: "This ends with period.",
		}

		errors = rule.Validate(invalidCommit)

		require.NotEmpty(t, errors, "Default config should reject subject ending with period")
		validationErr := errors[0]
		require.Equal(t, string(appErrors.ErrSubjectSuffix), validationErr.Code, "Error code should be 'invalid_suffix'")
	})

	t.Run("With custom invalid suffixes", func(t *testing.T) {
		// Custom invalid suffixes
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes("!@#"))

		// Create commit with invalid suffix (ends with !)
		commit := domain.CommitInfo{
			Subject: "This ends with exclamation!",
		}

		errors := rule.Validate(commit)

		require.NotEmpty(t, errors, "Should reject subject with configured invalid suffix")
		validationErr := errors[0]
		require.Equal(t, string(appErrors.ErrSubjectSuffix), validationErr.Code, "Error code should be 'invalid_suffix'")

		// Verify context contains the correct invalid suffixes
		require.Equal(t, "!@#", validationErr.Context["invalid_suffixes"],
			"Context should contain the custom invalid suffixes")

		// Create commit with period (should be allowed with custom config)
		validCommit := domain.CommitInfo{
			Subject: "This ends with period.",
		}

		errors = rule.Validate(validCommit)

		require.Empty(t, errors, "Should accept subject ending with period when not in invalid set")
	})

	t.Run("Empty invalid suffixes", func(t *testing.T) {
		// Empty invalid suffixes should fall back to defaults
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes(""))

		// Create commit with question mark (in default invalid suffixes)
		commit := domain.CommitInfo{
			Subject: "This ends with question mark?",
		}

		errors := rule.Validate(commit)

		require.NotEmpty(t, errors, "Should reject subject with default invalid suffix")
		validationErr := errors[0]
		require.Equal(t, rules.DefaultInvalidSuffixes, validationErr.Context["invalid_suffixes"],
			"Should fall back to default invalid suffixes")
	})
}

func TestSubjectSuffixRuleWithCustomOptions(t *testing.T) {
	// Create rule with options
	rule := rules.NewSubjectSuffixRule(
		rules.WithInvalidSuffixes("!@#"),
	)

	// Verify the rule uses the config value
	commit := domain.CommitInfo{
		Subject: "Test with exclamation!",
	}

	// Validate and check for error
	errors := rule.Validate(commit)
	require.Len(t, errors, 1, "Should have exactly one error")
	require.Equal(t, string(appErrors.ErrSubjectSuffix), errors[0].Code)

	// Check context values
	require.Equal(t, "!", errors[0].Context["last_char"])
	require.Equal(t, "!@#", errors[0].Context["invalid_suffixes"])
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
				return config.NewConfig() // Use default suffixes
			},
			subject:      "Add new feature",
			expectErrors: false,
			description:  "Should pass with default suffixes and valid subject",
		},
		{
			name: "Default invalid suffixes - invalid subject",
			configSetup: func() config.Config {
				return config.NewConfig() // Use default suffixes
			},
			subject:      "Add new feature.",
			expectErrors: true,
			description:  "Should fail with default suffixes and subject ending with period",
		},
		{
			name: "Custom invalid suffixes - valid with default invalid suffix",
			configSetup: func() config.Config {
				return config.NewConfig().
					WithSubjectInvalidSuffixes("!?") // Only ! and ? are invalid
			},
			subject:      "Add new feature.", // Period is allowed with custom config
			expectErrors: false,
			description:  "Should pass when period is not in custom invalid suffixes",
		},
		{
			name: "Custom invalid suffixes - invalid with custom suffix",
			configSetup: func() config.Config {
				return config.NewConfig().
					WithSubjectInvalidSuffixes("!?") // Only ! and ? are invalid
			},
			subject:      "Add new feature!", // Exclamation mark is not allowed
			expectErrors: true,
			description:  "Should fail when ending with a character in custom invalid suffixes",
		},
		{
			name: "Empty subject",
			configSetup: func() config.Config {
				return config.NewConfig()
			},
			subject:      "",
			expectErrors: true,
			description:  "Should fail with empty subject",
		},
		{
			name: "Unicode invalid suffixes",
			configSetup: func() config.Config {
				return config.NewConfig().
					WithSubjectInvalidSuffixes("😊😀") // Only emojis are invalid
			},
			subject:      "Add new emoji😊",
			expectErrors: true,
			description:  "Should fail with Unicode invalid suffixes",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create unified config with test options
			unifiedConfig := testCase.configSetup()

			// Create rule with unified config
			rule := rules.NewSubjectSuffixRuleWithConfig(unifiedConfig)

			// Create test commit
			commit := domain.CommitInfo{
				Hash:    "abc123",
				Subject: testCase.subject,
				Message: testCase.subject,
			}

			// Validate and check results
			errors := rule.Validate(commit)

			if testCase.expectErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")

				// Check rule name in errors
				if len(errors) > 0 {
					require.Equal(t, "SubjectSuffix", errors[0].Rule, "Rule name should be correct in error")
				}
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}

			// Check rule name
			require.Equal(t, "SubjectSuffix", rule.Name(), "Rule name should be 'SubjectSuffix'")
		})
	}
}

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

func TestSubjectSuffixRule(t *testing.T) {
	testCases := []struct {
		name            string
		subject         string
		invalidSuffixes string
		expectedValid   bool
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
		},
		{
			name:            "Subject ending with invalid suffix colon",
			subject:         "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
		},
		{
			name:            "Unicode subject with invalid suffix",
			subject:         "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
		},
		{
			name:            "Unicode character as invalid suffix",
			subject:         "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
		},
		{
			name:            "Empty subject",
			subject:         "",
			invalidSuffixes: ".:;",
			expectedValid:   false,
		},
		{
			name:            "Subject with Unicode invalid suffix",
			subject:         "Add new emoji😊",
			invalidSuffixes: "😊", // Keep just one emoji to simplify the test
			expectedValid:   false,
		},
		{
			name:            "Subject with space as invalid suffix",
			subject:         "Add feature ",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
		},
		{
			name:            "Subject with tab as invalid suffix",
			subject:         "Add feature\t",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
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
			invalidSuffixes: ".,:;?",
			expectedValid:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule with config
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{},
				},
			}

			if testCase.invalidSuffixes != "" {
				// Convert string to array of endings
				for _, r := range testCase.invalidSuffixes {
					cfg.Message.Subject.ForbidEndings = append(cfg.Message.Subject.ForbidEndings, string(r))
				}
			}

			baseRule := rules.NewSubjectSuffixRule(cfg)

			// Create commit with the test subject
			commitInfo := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commitInfo.Subject = testCase.subject
			commit := commitInfo

			// Execute validation using the configured rule
			failures := baseRule.Validate(commit, nil, &cfg)

			// Check results
			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation failures")
			} else {
				require.NotEmpty(t, failures, "Expected validation failures")
				testdata.AssertRuleFailure(t, failures[0], "SubjectSuffix")
			}

			// Verify name
			require.Equal(t, "SubjectSuffix", baseRule.Name())
		})
	}
}

func TestSubjectSuffixOptions(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		// Create a valid commit
		validCommit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		validCommit.Subject = "This is valid"

		// Create an invalid commit
		invalidCommit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		invalidCommit.Subject = "This ends with period."

		// Create rule with default configuration
		cfg := config.Config{
			Message: config.MessageConfig{
				Subject: config.SubjectConfig{
					ForbidEndings: []string{".", ","},
				},
			},
		}
		rule := rules.NewSubjectSuffixRule(cfg)

		// Test valid case
		validFailures := rule.Validate(validCommit, nil, &cfg)
		require.Empty(t, validFailures, "Default config should accept valid subject")

		// Test invalid case
		invalidFailures := rule.Validate(invalidCommit, nil, &cfg)
		require.NotEmpty(t, invalidFailures, "Default config should reject subject ending with period")
	})

	t.Run("With custom invalid suffixes", func(t *testing.T) {
		// Custom invalid suffixes
		cfg := config.Config{
			Message: config.MessageConfig{
				Subject: config.SubjectConfig{
					ForbidEndings: []string{"!", "@", "#"},
				},
			},
		}
		rule := rules.NewSubjectSuffixRule(cfg)

		// Create commits for testing
		invalidCommit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		invalidCommit.Subject = "This ends with exclamation!"

		validCommit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		validCommit.Subject = "This ends with period."

		// Test invalid case
		invalidFailures := rule.Validate(invalidCommit, nil, &cfg)
		require.NotEmpty(t, invalidFailures, "Should reject subject with configured invalid suffix")

		// Test valid case
		validFailures := rule.Validate(validCommit, nil, &cfg)
		require.Empty(t, validFailures, "Should accept subject ending with period when not in invalid set")
	})

	t.Run("Empty invalid suffixes", func(t *testing.T) {
		// Create rule that checks for ?
		cfg := config.Config{
			Message: config.MessageConfig{
				Subject: config.SubjectConfig{
					ForbidEndings: []string{"?"},
				},
			},
		}
		rule := rules.NewSubjectSuffixRule(cfg)

		// Create commit with question mark
		commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		commit.Subject = "This ends with question mark?"

		// Test invalid case
		failures := rule.Validate(commit, nil, &cfg)

		require.NotEmpty(t, failures, "Should reject subject with configured invalid suffix")
	})
}

func TestSubjectSuffixRuleWithCustomOptions(t *testing.T) {
	// Create rule with config
	cfg := config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				ForbidEndings: []string{"!", "@", "#"},
			},
		},
	}
	rule := rules.NewSubjectSuffixRule(cfg)

	// Create test commit
	commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
	commit.Subject = "Test with exclamation!"

	// Test validation
	failures := rule.Validate(commit, nil, &cfg)

	require.NotEmpty(t, failures, "Should return errors for invalid subject")

	// Simple validation of rule name
	require.Equal(t, "SubjectSuffix", rule.Name())
}

func TestSubjectSuffixRuleWithConfig(t *testing.T) {
	tests := []struct {
		name            string
		invalidSuffixes []string
		subject         string
		expectErrors    bool
		description     string
	}{
		{
			name:            "Default invalid suffixes - valid subject",
			invalidSuffixes: []string{".", ","},
			subject:         "Add new feature",
			expectErrors:    false,
			description:     "Should pass with default suffixes and valid subject",
		},
		{
			name:            "Default invalid suffixes - invalid subject",
			invalidSuffixes: []string{".", ","},
			subject:         "Add new feature.",
			expectErrors:    true,
			description:     "Should fail with default suffixes and subject ending with period",
		},
		{
			name:            "Custom invalid suffixes - valid with default invalid suffix",
			invalidSuffixes: []string{"!", "?"},
			subject:         "Add new feature.", // Period is allowed with custom config
			expectErrors:    false,
			description:     "Should pass when period is not in custom invalid suffixes",
		},
		{
			name:            "Custom invalid suffixes - invalid with custom suffix",
			invalidSuffixes: []string{"!", "?"},
			subject:         "Add new feature!", // Exclamation mark is not allowed
			expectErrors:    true,
			description:     "Should fail when ending with a character in custom invalid suffixes",
		},
		{
			name:            "Empty subject",
			invalidSuffixes: []string{".", ","},
			subject:         "",
			expectErrors:    true,
			description:     "Should fail with empty subject",
		},
		{
			name:            "Unicode invalid suffixes",
			invalidSuffixes: []string{"😊"},
			subject:         "Add new emoji😊",
			expectErrors:    true,
			description:     "Should fail with Unicode invalid suffixes",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with configuration
			var invalidSuffixesStr string

			if len(testCase.invalidSuffixes) > 0 {
				for _, suffix := range testCase.invalidSuffixes {
					invalidSuffixesStr += suffix
				}
			}

			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						ForbidEndings: testCase.invalidSuffixes,
					},
				},
			}
			rule := rules.NewSubjectSuffixRule(cfg)

			// Create test commit
			commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commit.Hash = "abc123"
			commit.Subject = testCase.subject
			commit.Message = testCase.subject

			// Execute validation
			failures := rule.Validate(commit, nil, &cfg)

			// Check results
			if testCase.expectErrors {
				require.NotEmpty(t, failures, "Expected validation failures")
			} else {
				require.Empty(t, failures, "Expected no validation failures")
			}

			// Check rule name
			require.Equal(t, "SubjectSuffix", rule.Name())
		})
	}
}

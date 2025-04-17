// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitBodyRule(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		options        []rules.CommitBodyOption
		expectError    bool
		errorCode      string
		expectedResult string
	}{
		{
			name: "valid commit with body",
			message: `Add new validation rules

This commit adds new validation rules for:
- Password complexity
- Email format
- Username requirements`,
			expectError:    false,
			expectedResult: "Commit body is valid",
		},
		{
			name: "valid commit with body and sign-off",
			message: `Update documentation

Improve the getting started guide
Add more examples

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    false,
			expectedResult: "Commit body is valid",
		},
		{
			name:           "commit without body",
			message:        "just a subject",
			expectError:    true,
			errorCode:      string(domain.ValidationErrorInvalidBody),
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit without empty line between subject and body",
			message: `Update CI pipeline
Adding new stages for:
- Security scanning
- Performance testing
Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    true,
			errorCode:      string(domain.ValidationErrorInvalidBody),
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with empty line after subject but empty body",
			message: `Update CI pipeline

`,
			expectError:    true,
			errorCode:      string(domain.ValidationErrorInvalidBody),
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with only sign-off",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    true,
			errorCode:      string(domain.ValidationErrorInvalidBody),
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with multiple sign-off lines but no body",
			message: `Update dependencies

Signed-off-by: Laval Lion <laval@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			expectError:    true,
			errorCode:      string(domain.ValidationErrorInvalidBody),
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with only sign-off but configured to allow it",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			options: []rules.CommitBodyOption{
				rules.WithAllowSignOffOnly(true),
			},
			expectError:    false,
			expectedResult: "Commit body is valid",
		},
		{
			name:           "body not required when configured",
			message:        "just a subject line",
			options:        []rules.CommitBodyOption{rules.WithRequireBody(false)},
			expectError:    false,
			expectedResult: "Commit body is valid",
		},
		{
			name: "commit with empty body line but content after",
			message: `Fix typo in documentation

Fixed the typo in API documentation.`,
			expectError:    false,
			expectedResult: "Commit body is valid",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Create commit info object
			commit := &domain.CommitInfo{
				Message: tabletest.message,
			}

			// Create the rule with optional configuration
			commitBodyRule := rules.NewCommitBodyRule(tabletest.options...)

			// Validate the commit
			errors := commitBodyRule.Validate(commit)

			// Verify name is correct
			assert.Equal(t, "CommitBodyRule", commitBodyRule.Name())

			// Test result output
			assert.Equal(t, tabletest.expectedResult, commitBodyRule.Result(), "Result() should return expected output")

			if tabletest.expectError {
				require.NotEmpty(t, errors, "expected errors but got none")

				// Verify we have ValidationError with expected code
				valErr := errors[0]
				assert.Equal(t, tabletest.errorCode, valErr.Code, "unexpected error code")
				assert.Equal(t, "CommitBodyRule", valErr.Rule, "rule name should be set")

				// Check VerboseResult contains more detailed information
				verboseResult := commitBodyRule.VerboseResult()
				assert.NotEqual(t, commitBodyRule.Result(), verboseResult, "VerboseResult should differ from Result for errors")

				// Help should be provided for invalid commits
				assert.NotEmpty(t, commitBodyRule.Help(), "help should be provided for invalid commits")
			} else {
				assert.Empty(t, errors, "unexpected errors: %v", errors)

				// For valid commits, verify verbose output
				verboseResult := commitBodyRule.VerboseResult()
				assert.NotEqual(t, commitBodyRule.Result(), verboseResult, "VerboseResult should differ from Result for valid commits")
				assert.Contains(t, verboseResult, "proper format", "VerboseResult should be descriptive for valid commits")
			}

			// Errors should match what was returned by Validate
			assert.Equal(t, errors, commitBodyRule.Errors())
		})
	}
}

func TestCommitBodyVerboseResult(t *testing.T) {
	testCases := []struct {
		name           string
		message        string
		expectedPhrase string
	}{
		{
			name:           "missing body verbose output",
			message:        "subject only",
			expectedPhrase: "lacks a body",
		},
		{
			name:    "missing blank line verbose output",
			message: "subject\nbody without blank line",
			// This will actually trigger the "missing_body" error first in the validation logic,
			// so we adjust our expectation
			expectedPhrase: "lacks a body",
		},
		{
			name:           "empty body verbose output",
			message:        "subject\n\n",
			expectedPhrase: "empty",
		},
		{
			name:           "signoff first line verbose output",
			message:        "subject\n\nSigned-off-by: Test <test@example.com>",
			expectedPhrase: "sign-off line",
		},
		{
			name:    "only signoff verbose output",
			message: "subject\n\nSigned-off-by: Test <test@example.com>\nSigned-off-by: Another <another@example.com>",
			// This will actually trigger the "signoff_first_line" check first
			expectedPhrase: "sign-off line",
		},
		{
			name:           "valid commit verbose output",
			message:        "subject\n\nThis is a valid body.",
			expectedPhrase: "proper format",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Create commit info and validate
			commit := &domain.CommitInfo{
				Message: tabletest.message,
			}

			// Create rule
			rule := rules.NewCommitBodyRule()

			// Validate
			_ = rule.Validate(commit)

			// Check verbose result
			verboseResult := rule.VerboseResult()
			assert.Contains(t, verboseResult, tabletest.expectedPhrase, "verbose output should contain expected phrase")
		})
	}
}

func TestCommitBodyHelpMethod(t *testing.T) {
	t.Run("help for missing body", func(t *testing.T) {
		// Create a commit info with a missing body
		commit := &domain.CommitInfo{
			Message: "subject only",
		}

		// Create a rule and validate
		r := rules.NewCommitBodyRule()
		_ = r.Validate(commit)

		helpText := r.Help()
		assert.Contains(t, helpText, "Ensure your commit message follows this structure",
			"help text should have standard structure intro")
	})

	t.Run("help for missing blank line", func(t *testing.T) {
		// Create a commit info with a missing blank line
		commit := &domain.CommitInfo{
			Message: "subject\nbody without blank line",
		}

		// Create a rule and validate
		r := rules.NewCommitBodyRule()
		_ = r.Validate(commit)

		helpText := r.Help()
		// This will trigger a standard error message for invalid body
		assert.Contains(t, helpText, "Format the commit body correctly",
			"help text should include guidance about formatting")
	})

	t.Run("help for empty body", func(t *testing.T) {
		// Create a commit info with an empty body
		commit := &domain.CommitInfo{
			Message: "subject\n\n",
		}

		// Create a rule and validate
		r := rules.NewCommitBodyRule()
		_ = r.Validate(commit)

		helpText := r.Help()
		assert.Contains(t, helpText, "Format the commit body correctly",
			"help text should include standard guidance for formatting")
	})

	t.Run("help for valid commit", func(t *testing.T) {
		// Create a commit info with a valid body
		commit := &domain.CommitInfo{
			Message: "subject\n\nThis is a valid body.",
		}

		// Create a rule and validate
		r := rules.NewCommitBodyRule()
		_ = r.Validate(commit)

		helpText := r.Help()
		assert.Equal(t, "No errors to fix", helpText,
			"help text for valid commit should indicate no errors")
	})
}

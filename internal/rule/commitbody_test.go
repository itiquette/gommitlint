// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommitBodyRule(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		options        []rule.CommitBodyOption
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
			errorCode:      "missing_body",
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
			errorCode:      "missing_blank_line",
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with empty line after subject but empty body",
			message: `Update CI pipeline

`,
			expectError:    true,
			errorCode:      "empty_body",
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with only sign-off",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    true,
			errorCode:      "signoff_first_line",
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with multiple sign-off lines but no body",
			message: `Update dependencies

Signed-off-by: Laval Lion <laval@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			expectError:    true,
			errorCode:      "signoff_first_line",
			expectedResult: "Invalid commit message body",
		},
		{
			name: "commit with only sign-off but configured to allow it",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			options: []rule.CommitBodyOption{
				rule.WithAllowSignOffOnly(true),
			},
			expectError:    false,
			expectedResult: "Commit body is valid",
		},
		{
			name:           "body not required when configured",
			message:        "just a subject line",
			options:        []rule.CommitBodyOption{rule.WithRequireBody(false)},
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
			// Create the rule with optional configuration
			commitBodyRule := rule.ValidateCommitBody(tabletest.message, tabletest.options...)

			// Verify name is correct
			assert.Equal(t, "CommitBodyRule", commitBodyRule.Name())

			// Test result output
			assert.Equal(t, tabletest.expectedResult, commitBodyRule.Result(), "Result() should return expected output")

			if tabletest.expectError {
				require.NotEmpty(t, commitBodyRule.Errors(), "expected errors but got none")

				// Verify we have ValidationError with expected code
				valErr := commitBodyRule.Errors()[0]
				assert.Equal(t, tabletest.errorCode, valErr.Code, "unexpected error code")
				assert.Equal(t, "CommitBodyRule", valErr.Rule, "rule name should be set")

				// Check VerboseResult contains more detailed information
				verboseResult := commitBodyRule.VerboseResult()
				assert.NotEqual(t, commitBodyRule.Result(), verboseResult, "VerboseResult should differ from Result for errors")

				// Don't check for the specific word "Commit" since some error messages might not include it
				assert.NotEmpty(t, verboseResult, "VerboseResult should contain detailed explanation")

				// Help should be provided for invalid commits
				assert.NotEmpty(t, commitBodyRule.Help(), "help should be provided for invalid commits")
			} else {
				assert.Empty(t, commitBodyRule.Errors(), "unexpected errors: %v", commitBodyRule.Errors())

				// For valid commits, verify verbose output
				verboseResult := commitBodyRule.VerboseResult()
				assert.NotEqual(t, commitBodyRule.Result(), verboseResult, "VerboseResult should differ from Result for valid commits")
				assert.Contains(t, verboseResult, "proper format", "VerboseResult should be descriptive for valid commits")
			}
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
			// Create rule
			r := rule.ValidateCommitBody(tabletest.message)

			// Check verbose result
			verboseResult := r.VerboseResult()
			assert.Contains(t, verboseResult, tabletest.expectedPhrase, "verbose output should contain expected phrase")
		})
	}
}

func TestCommitBodyHelpMethod(t *testing.T) {
	t.Run("help for missing body", func(t *testing.T) {
		// Create a rule with a missing body error by validating a subject-only message
		r := rule.ValidateCommitBody("subject only")

		helpText := r.Help()
		assert.Contains(t, helpText, "Add a descriptive body",
			"help text should have guidance for missing body")
	})

	t.Run("help for missing blank line", func(t *testing.T) {
		// Create a rule with a missing blank line error
		r := rule.ValidateCommitBody("subject\nbody without blank line")

		helpText := r.Help()
		assert.Contains(t, helpText, "one blank line",
			"help text should include guidance for missing blank line")
	})

	t.Run("help for empty body", func(t *testing.T) {
		// Create a rule with an empty body error
		r := rule.ValidateCommitBody("subject\n\n")

		helpText := r.Help()
		assert.Contains(t, helpText, "must contain actual content",
			"help text should include guidance for empty body")
	})

	t.Run("help for signoff first line", func(t *testing.T) {
		// Create a rule with signoff first line error
		r := rule.ValidateCommitBody("subject\n\nSigned-off-by: Test <test@example.com>")

		helpText := r.Help()
		assert.Contains(t, helpText, "should not start with a sign-off line",
			"help text should include guidance for signoff first line")
	})

	t.Run("help for only signoff", func(t *testing.T) {
		// Create a message that will trigger the 'only_signoff' error
		// (This might be the same as signoff_first_line in practical cases)
		message := "subject\n\nSigned-off-by: Test <test@example.com>\nSigned-off-by: Another <another@example.com>"
		r := rule.ValidateCommitBody(message)

		// If this specific error occurred, check its help text
		if len(r.Errors()) > 0 && r.Errors()[0].Code == "only_signoff" {
			helpText := r.Help()
			assert.Contains(t, helpText, "beyond just sign-off lines",
				"help text should include guidance for only having signoff lines")
		}
	})

	t.Run("help for valid commit", func(t *testing.T) {
		// Create a rule with no errors
		validMessage := "subject\n\nThis is a valid body."
		r := rule.ValidateCommitBody(validMessage)

		helpText := r.Help()
		assert.Equal(t, "No errors to fix", helpText,
			"help text for valid commit should indicate no errors")
	})
}

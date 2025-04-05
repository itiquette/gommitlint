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
		name         string
		message      string
		options      []rule.CommitBodyOption
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid commit with body",
			message: `Add new validation rules

This commit adds new validation rules for:
- Password complexity
- Email format
- Username requirements`,
			expectError: false,
		},
		{
			name: "valid commit with body and sign-off",
			message: `Update documentation

Improve the getting started guide
Add more examples

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError: false,
		},
		{
			name:         "commit without body",
			message:      "just a subject",
			expectError:  true,
			errorMessage: "Commit message requires a body explaining the changes",
		},
		{
			name: "commit without empty line between subject and body",
			message: `Update CI pipeline
Adding new stages for:
- Security scanning
- Performance testing
Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:  true,
			errorMessage: "Commit message must have exactly one empty line between the subject and the body",
		},
		{
			name: "commit with empty line after subject but empty body",
			message: `Update CI pipeline

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:  true,
			errorMessage: "Commit message must have a non-empty body text",
		},
		{
			name: "commit with only sign-off",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:  true,
			errorMessage: "Commit message must have a non-empty body text",
		},
		{
			name: "commit with multiple sign-off lines but no body",
			message: `Update dependencies

Signed-off-by: Laval Lion <laval@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			expectError:  true,
			errorMessage: "Commit message must have a non-empty body text",
		},
		{
			name: "commit with only sign-off but configured to allow it",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			options: []rule.CommitBodyOption{
				rule.WithAllowSignOffOnly(true),
			},
			expectError: false,
		},
		{
			name:        "body not required when configured",
			message:     "just a subject line",
			options:     []rule.CommitBodyOption{rule.WithRequireBody(false)},
			expectError: false,
		},
		{
			name: "commit with empty body line but content after",
			message: `Fix typo in documentation

Fixed the typo in API documentation.`,
			expectError: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Create the rule with optional configuration
			commitBodyRule := rule.ValidateCommitBody(tabletest.message, tabletest.options...)

			// Verify name is correct
			assert.Equal(t, "CommitBodyRule", commitBodyRule.Name())

			// Test result output
			if tabletest.expectError {
				require.NotEmpty(t, commitBodyRule.Errors(), "expected errors but got none")
				assert.Contains(t, commitBodyRule.Result(), tabletest.errorMessage, "unexpected error message")
				assert.NotEmpty(t, commitBodyRule.Help(), "help should be provided for invalid commits")
			} else {
				assert.Empty(t, commitBodyRule.Errors(), "unexpected errors: %v", commitBodyRule.Errors())
				assert.Equal(t, "Commit body is valid", commitBodyRule.Result(), "unexpected message for valid commit")
			}
		})
	}
}

func TestCommitBodyHelpMethod(t *testing.T) {
	t.Run("help for missing body", func(t *testing.T) {
		r := &rule.CommitBodyRule{}
		r.AddError("Commit message requires a body explaining the changes")

		helpText := r.Help()
		assert.Contains(t, helpText, "Add a descriptive body",
			"help text should have guidance for missing body")
	})

	t.Run("help for missing blank line", func(t *testing.T) {
		r := &rule.CommitBodyRule{}
		r.AddError("Commit message must have exactly one empty line between the subject and the body")

		helpText := r.Help()
		assert.Contains(t, helpText, "one blank line",
			"help text should include guidance for missing blank line")
	})

	t.Run("help for empty body", func(t *testing.T) {
		r := &rule.CommitBodyRule{}
		r.AddError("Commit message must have a non-empty body text")

		helpText := r.Help()
		assert.Contains(t, helpText, "must contain actual content",
			"help text should include guidance for empty body")
	})
}

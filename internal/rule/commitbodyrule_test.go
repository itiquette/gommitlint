// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateCommitBodyRule(t *testing.T) {
	tests := []struct {
		name         string
		message      string
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
			errorMessage: "Commit message body is required",
		},
		{
			name: "commit with only sign-off",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:  true,
			errorMessage: "Commit message body is required",
		},
		{
			name: "commit with multiple sign-off lines but no body",
			message: `Update dependencies

Signed-off-by: Laval Lion <laval@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			expectError:  true,
			errorMessage: "Commit message body is required",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Use the new function name that matches the implementation
			rule := rule.ValidateCommitBodyRule(tabletest.message)

			if tabletest.expectError {
				require.NotEmpty(t, rule.Errors(), "expected errors but got none")
				require.Contains(t, rule.Result(), tabletest.errorMessage, "unexpected error message")

				return
			}

			require.Empty(t, rule.Errors(), "unexpected errors: %v", rule.Errors())
			require.Equal(t, "Commit body is valid", rule.Result(), "unexpected message for valid commit")
		})
	}
}

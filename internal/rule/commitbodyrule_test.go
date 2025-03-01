// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateCommitBody(t *testing.T) {
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
			name: "valid commit with body and DCO",
			message: `Update documentation

Improve the getting started guide
Add more examples

Signed-off-by: John Doe <john@example.com>`,
			expectError: false,
		},
		{
			name:         "commit without body",
			message:      "just a subject",
			expectError:  true,
			errorMessage: "Commit requires a descriptive body explaining the changes",
		},
		{
			name: "commit without empty line between subject and body",
			message: `Update CI pipeline
Adding new stages for:
- Security scanning
- Performance testing
Signed-off-by: John Doe <john@example.com>`,
			expectError:  true,
			errorMessage: "Commit message must have exactly one empty line between subject and body",
		},
		{
			name: "commit with multiple empty lines between subject and body",
			message: `Update CI pipeline



Adding new stages for:
- Security scanning
- Performance testing
Signed-off-by: John Doe <john@example.com>`,
			expectError:  true,
			errorMessage: "Commit message must have a non empty body text",
		},
		{
			name: "commit without empty line between subject and body",
			message: `Update CI pipeline
Adding new stages for:
- Security scanning
- Performance testing

Signed-off-by: John Doe <john@example.com>`,
			expectError:  true,
			errorMessage: "Commit message must have exactly one empty line between subject and body",
		},
		{
			name: "commit with only DCO",
			message: `Update config

Signed-off-by: John Doe <john@example.com>`,
			expectError:  true,
			errorMessage: "Commit body is required",
		},
		{
			name: "commit with empty lines and DCO",
			message: `Update tests

Signed-off-by: John Doe <john@example.com>`,
			expectError:  true,
			errorMessage: "Commit body is required",
		},
		{
			name: "commit with multiple DCO lines but no body",
			message: `Update dependencies

Signed-off-by: John Doe <john@example.com>
Signed-off-by: Jane Doe <jane@example.com>`,
			expectError:  true,
			errorMessage: "Commit body is required",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			check := rule.ValidateCommitBody(tabletest.message)

			if tabletest.expectError {
				require.NotEmpty(t, check.Errors(), "expected errors but got none")
				require.Contains(t, check.Result(), tabletest.errorMessage, "unexpected error message")

				return
			}

			require.Empty(t, check.Errors(), "unexpected errors: %v", check.Errors())
			require.Equal(t, "Commit body is valid", check.Result(), "unexpected message for valid commit")
		})
	}
}

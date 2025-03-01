// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateSignoff(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid sign-off",
			message: `Add new feature

Implement automatic logging system.

Signed-off-by: Laval Lion <laval.lion@cavora.org>`,
			expectError: false,
		},
		{
			name:        "valid sign-off with CRLF",
			message:     "Update docs\r\n\r\nImprove README\r\n\r\nSigned-off-by: Jane Smith <jane.smith@example.com>",
			expectError: false,
		},
		{
			name: "valid sign-off with multiple signers",
			message: `Fix bug

Update error handling.

Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Jane Smith <jane.smith@example.com>`,
			expectError: false,
		},
		{
			name: "missing sign-off signature",
			message: `Add feature

Implement new logging system.`,
			expectError:  true,
			errorMessage: "Commit must be signed-off",
		},
		{
			name: "malformed sign-off - wrong format",
			message: `Add test

Signed by: Laval Lion <laval.lion@cavora.org>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off",
		},
		{
			name: "malformed sign-off - invalid email",
			message: `Add test

Signed-off-by: John Doe <invalid-email>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off",
		},
		{
			name: "malformed sign-off - missing name",
			message: `Add test

Signed-off-by: <john.doe@example.com>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			check := rule.ValidateSignOff(tabletest.message)

			if tabletest.expectError {
				require.NotEmpty(t, check.Errors(), "expected errors but got none")
				require.Contains(t, check.Result(), tabletest.errorMessage, "unexpected error message")

				return
			}

			require.Empty(t, check.Errors(), "unexpected errors: %v", check.Errors())
			require.Equal(t, "Sign-off exists", check.Result(),
				"unexpected message for valid sign-off")
		})
	}
}

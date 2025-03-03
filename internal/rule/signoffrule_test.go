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

func TestValidateSignoffRule(t *testing.T) {
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
			message:     "Update docs\r\n\r\nImprove README\r\n\r\nSigned-off-by: Cragger Crocodile <cragger@svamp.org>",
			expectError: false,
		},
		{
			name: "valid sign-off with multiple signers",
			message: `Fix bug

Update error handling.

Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
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

Signed-off-by: Phoenix Fire <invalid-email>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off",
		},
		{
			name: "malformed sign-off - missing name",
			message: `Add test

Signed-off-by: <laval@cavora.org>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			rule := rule.ValidateSignOffRule(tabletest.message)

			if tabletest.expectError {
				require.NotEmpty(t, rule.Errors(), "expected errors but got none")
				require.Contains(t, rule.Result(), tabletest.errorMessage, "unexpected error message")

				return
			}

			require.Empty(t, rule.Errors(), "unexpected errors: %v", rule.Errors())
			require.Equal(t, "Sign-off exists", rule.Result(),
				"unexpected message for valid sign-off")
		})
	}
}

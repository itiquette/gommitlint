// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateDCO(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid DCO signature",
			message: `Add new feature

Implement automatic logging system.

Signed-off-by: John Doe <john.doe@example.com>`,
			expectError: false,
		},
		{
			name:        "valid DCO with CRLF",
			message:     "Update docs\r\n\r\nImprove README\r\n\r\nSigned-off-by: Jane Smith <jane.smith@example.com>",
			expectError: false,
		},
		{
			name: "valid DCO with multiple signers",
			message: `Fix bug

Update error handling.

Signed-off-by: John Doe <john.doe@example.com>
Signed-off-by: Jane Smith <jane.smith@example.com>`,
			expectError: false,
		},
		{
			name: "missing DCO signature",
			message: `Add feature

Implement new logging system.`,
			expectError:  true,
			errorMessage: "Commit must be signed-off with a Developer Certificate of Origin (DCO)",
		},
		{
			name: "malformed DCO - wrong format",
			message: `Add test

Signed by: John Doe <john.doe@example.com>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off with a Developer Certificate of Origin (DCO)",
		},
		{
			name: "malformed DCO - invalid email",
			message: `Add test

Signed-off-by: John Doe <invalid-email>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off with a Developer Certificate of Origin (DCO)",
		},
		{
			name: "malformed DCO - missing name",
			message: `Add test

Signed-off-by: <john.doe@example.com>`,
			expectError:  true,
			errorMessage: "Commit must be signed-off with a Developer Certificate of Origin (DCO)",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			check := rule.ValidateSignOff(tabletest.message)

			if tabletest.expectError {
				require.NotEmpty(t, check.Errors(), "expected errors but got none")
				require.Contains(t, check.Message(), tabletest.errorMessage, "unexpected error message")

				return
			}

			require.Empty(t, check.Errors(), "unexpected errors: %v", check.Errors())
			require.Equal(t, "Developer Certificate of Origin signature is valid", check.Message(),
				"unexpected message for valid DCO")
		})
	}
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSignOff(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		expectError  bool
		errorCode    string
		errorMessage string
	}{
		{
			name: "Valid sign-off",
			message: `Add new feature
Implement automatic logging system.
Signed-off-by: Laval Lion <laval.lion@cavora.org>`,
			expectError: false,
		},
		{
			name:        "Valid sign-off with crlf",
			message:     "Update docs\r\n\r\nImprove README\r\n\r\nSigned-off-by: Cragger Crocodile <cragger@svamp.org>",
			expectError: false,
		},
		{
			name: "Valid sign-off with multiple signers",
			message: `Fix bug
Update error handling.
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			expectError: false,
		},
		{
			name: "Missing sign-off signature",
			message: `Add feature
Implement new logging system.`,
			expectError:  true,
			errorCode:    "missing_signoff",
			errorMessage: "commit must be signed-off",
		},
		{
			name: "Malformed sign-off - wrong format",
			message: `Add test
Signed by: Laval Lion <laval.lion@cavora.org>`,
			expectError:  true,
			errorCode:    "invalid_format",
			errorMessage: "commit must be signed-off",
		},
		{
			name: "Malformed sign-off - invalid email",
			message: `Add test
Signed-off-by: Phoenix Fire <invalid-email>`,
			expectError:  true,
			errorCode:    "invalid_format",
			errorMessage: "commit must be signed-off",
		},
		{
			name: "Malformed sign-off - missing name",
			message: `Add test
Signed-off-by: <laval@cavora.org>`,
			expectError:  true,
			errorCode:    "invalid_format",
			errorMessage: "commit must be signed-off",
		},
		{
			name:         "Empty message",
			message:      "",
			expectError:  true,
			errorCode:    "empty_message",
			errorMessage: "commit message body is empty",
		},
		{
			name:         "Whitespace only message",
			message:      "   \n  \t  \n",
			expectError:  true,
			errorCode:    "empty_message",
			errorMessage: "commit message body is empty",
		},
	}

	for _, testtable := range tests {
		t.Run(testtable.name, func(t *testing.T) {
			// Call the rule
			rule := rule.ValidateSignOff(testtable.message)

			// Check errors as expected
			if testtable.expectError {
				require.NotEmpty(t, rule.Errors(), "expected errors but got none")

				// Check error code if specified
				if testtable.errorCode != "" {
					assert.Equal(t, testtable.errorCode, rule.Errors()[0].Code,
						"Error code should match expected")
				}

				// Check error message
				require.Contains(t, rule.Result(), testtable.errorMessage, "unexpected error message")

				// Verify rule name is set in error
				assert.Equal(t, "SignOff", rule.Errors()[0].Rule, "Rule name should be set in ValidationError")

				// Verify context data exists for appropriate errors
				if testtable.errorCode != "empty_message" {
					assert.Contains(t, rule.Errors()[0].Context, "message",
						"Context should contain the message")
				}

				// Verify Help() method provides guidance
				helpText := rule.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				if strings.Contains(rule.Result(), "empty") {
					assert.Contains(t, helpText, "Add a Developer Certificate of Origin sign-off",
						"Help should mention adding a sign-off")
				} else {
					assert.Contains(t, helpText, "Developer Certificate of Origin",
						"Help should mention DCO")
				}
			} else {
				require.Empty(t, rule.Errors(), "unexpected errors: %v", rule.Errors())
				require.Equal(t, "Sign-off exists", rule.Result(),
					"unexpected message for valid sign-off")

				// Test Help on valid case
				assert.Equal(t, "No errors to fix", rule.Help(),
					"Help for valid message should indicate nothing to fix")
			}

			// Verify Name() method
			assert.Equal(t, "SignOff", rule.Name(), "Rule name should be 'SignOff'")
		})
	}

	// Test for specific help messages based on error type
	t.Run("Specific help messages", func(t *testing.T) {
		// Empty message case
		emptyRule := rule.ValidateSignOff("")
		assert.Contains(t, emptyRule.Help(), "is currently empty",
			"Help for empty message should mention it's empty")

		// Missing signoff case
		missingRule := rule.ValidateSignOff("Add feature\nImplement new feature")
		assert.Contains(t, missingRule.Help(), "git commit -s",
			"Help for missing signoff should mention git commit -s")

		// Invalid format case
		invalidRule := rule.ValidateSignOff("Add feature\nSigned by: Invalid <format>")
		assert.Contains(t, invalidRule.Help(), "correctly formatted",
			"Help for invalid format should mention correct format")
	})
}

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

func TestValidateSignature(t *testing.T) {
	// Example of a valid GPG signature
	validGPGSignature := `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2

iQEcBAABCAAGBQJiJdwiAAoJEKbfr8MEmRYZlNQH/2f+/DjCj/3fVL2OFMElZIe6
r6te4sJFR5qSfm5Fg4zuvg8T3m1fUHSLQAddMjqrYtgZBOYwkqX2rnJA8QthI1n0
Wsh2QpuHqLLR1RP4rQQP0Vaz7ZCJa7dVXLjGSD9uFJFYl9NDZx04bxJQP8AUfTLk
e3ce/NzKGP8qTdBnRP3ca1Qw5CHLxAcTF79Z8v6VQu608QYVN9UnIYHLZEH7b9ov
q4XB9GFcGJiVbKQVGTJEIruvjRvvHcC2yY3kOUTGVUUI5+VAXkQvVxW10PqLlf13
LVlEr9iqRHQXX3ZocPjLcfRYVhyG1yw8PmGfRcYOKHdfDsUJSLYRXWKqTwXjRtc=
=zZJx
-----END PGP SIGNATURE-----`

	// Example of a valid SSH signature (properly base64-encoded for testing)
	validSSHSignature := `-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAABMAAAAHc3NoLXJzYQAAAQEAAAGBAJJU9e7iZTLCR6kiJz9jnj96
imcZvh/3zMfxLPPR3qZFKsJ5LQKAP8GzHY9AVP9SkMAO4Fj8eiLceGy3NKhSxW7ekpgw
IxOJXPScY8UlfvRZ9ftknNLOzQA7Mx+Kx+74QOciZ/Ctk7JAkOlCYKNbLKq8QwbQ+9RL
XY5MVe8iE9in3V4nM7dOtcNT5RWHjvmFtcVQCQUE9g4+xSQlg4RG0PvtQDU3DP3Tjjf+
R06m3OuBUUWX2wGaJSq9VKQlpv2aILXo5mhY7YVc82K8TJLK6pVsxgD73Hh4jCJ5J+8M
mzMiMUHVtZ8K1Z94mXzWMgvdBTvnOG1bLMBCJ1QpywAAASAAAAAIaGlzdG9yeQAAAnMA
AIJBh5q9m0CoHUo4gvdj4JEVzdj74FpMDKD+/s5MH6SXZTe6Ux4Pp7pOcvcF8cIVK2Hc
GDw7dXF9YzDM6TQGrZzHMLSKxJiEbKbJEwPBFahDUQP3o7MDfMXS9I23CLyd1JnmOGLP
a/K3XXxijJo6GNgpLhpkNzk2JnP5kSHLdp4LlLs/ByijLICOoBHQEAXJHhRp4GHuRpT4
a0G5Z3xhN014pKDLDI1/hEyemtSZlMkDJwoW9nc1q7AE8lNM1WBmeSZXqoZbgZn3V2Nl
FxlS+hzWnbOPMrRKuSfJ+H8mF6t1V3qUYtxHNQvHtcCvG0gx4auPSoxp7qVCVQ==
-----END SSH SIGNATURE-----`

	tests := []struct {
		name         string
		signature    string
		expectError  bool
		errorCode    string
		errorMessage string
	}{
		{
			name:        "Valid gpg signature",
			signature:   validGPGSignature,
			expectError: false,
		},
		{
			name:        "Valid ssh signature",
			signature:   validSSHSignature,
			expectError: false,
		},
		{
			name:         "Missing signature",
			signature:    "",
			expectError:  true,
			errorCode:    "missing_signature",
			errorMessage: "commit does not have a SSH/GPG signature",
		},
		{
			name:         "Empty signature with whitespace",
			signature:    "   \t\n",
			expectError:  true,
			errorCode:    "missing_signature",
			errorMessage: "commit does not have a SSH/GPG signature",
		},
		{
			name:         "Invalid signature format",
			signature:    "PARTIAL-SIGNATURE-EXAMPLE",
			expectError:  true,
			errorCode:    "unknown_signature_format",
			errorMessage: "invalid signature format",
		},
		{
			name:         "Incomplete gpg signature - missing end",
			signature:    "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\n\niQEcBAABCAAGBQJy...",
			expectError:  true,
			errorCode:    "incomplete_gpg_signature",
			errorMessage: "incomplete GPG signature",
		},
		{
			name:         "Incomplete ssh signature - missing end",
			signature:    "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg...",
			expectError:  true,
			errorCode:    "incomplete_ssh_signature",
			errorMessage: "incomplete SSH signature",
		},
		{
			name:         "Malformed gpg signature - missing version",
			signature:    "-----BEGIN PGP SIGNATURE-----\n-----END PGP SIGNATURE-----",
			expectError:  true,
			errorCode:    "invalid_gpg_format",
			errorMessage: "invalid GPG",
		},
		{
			name:         "Malformed ssh signature - invalid structure",
			signature:    "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHeHh4eA==\n-----END SSH SIGNATURE-----",
			expectError:  true,
			errorCode:    "invalid_ssh_format",
			errorMessage: "malformed SSH signature: too short",
		},
		{
			name:         "SSH signature with missing SSHSIG prefix",
			signature:    "-----BEGIN SSH SIGNATURE-----\nSW52YWxpZFByZWZpeA==\n-----END SSH SIGNATURE-----",
			expectError:  true,
			errorCode:    "invalid_ssh_format",
			errorMessage: "malformed SSH signature: missing SSHSIG prefix",
		},
		{
			name:         "SSH signature too short",
			signature:    "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHAA==\n-----END SSH SIGNATURE-----",
			expectError:  true,
			errorCode:    "invalid_ssh_format",
			errorMessage: "malformed SSH signature: too short",
		},
		{
			name:         "GPG signature with invalid content",
			signature:    "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\n\nInvalid Content\n-----END PGP SIGNATURE-----",
			expectError:  true,
			errorCode:    "invalid_gpg_format",
			errorMessage: "malformed GPG signature",
		},
		{
			name:         "SSH signature with wrong marker",
			signature:    "-----BEGIN SSH KEY-----\nU1NIU0lHAA==\n-----END SSH KEY-----",
			expectError:  true,
			errorCode:    "unknown_signature_format",
			errorMessage: "invalid signature format",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Call the rule
			rule := rule.ValidateSignature(tabletest.signature)

			// Check errors as expected
			if tabletest.expectError {
				require.NotEmpty(t, rule.Errors(), "expected errors but got none")

				// Check error code if specified
				if tabletest.errorCode != "" {
					assert.Equal(t, tabletest.errorCode, rule.Errors()[0].Code,
						"Error code should match expected")
				}

				// Check error message contains expected content
				errorFound := false

				for _, err := range rule.Errors() {
					if strings.Contains(err.Error(), tabletest.errorMessage) {
						errorFound = true

						break
					}
				}

				assert.True(t, errorFound, "Expected error containing '%s', but got: %v",
					tabletest.errorMessage, rule.Errors())

				// Verify rule name is set in ValidationError
				assert.Equal(t, "Signature", rule.Errors()[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify context information exists for specific error types
				if tabletest.errorCode == "invalid_gpg_format" ||
					tabletest.errorCode == "invalid_ssh_format" {
					assert.Contains(t, rule.Errors()[0].Context, "signature_type",
						"Context should contain signature type")
				}

				// Verify Help() method provides guidance
				helpText := rule.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")
				assert.Contains(t, helpText, "Sign your commit", "Help should explain how to sign commits")

				// Check for targeted help based on error code
				switch tabletest.errorCode {
				case "missing_signature":
					assert.Contains(t, helpText, "verify your identity",
						"Help should explain why signatures are important")
				case "incomplete_gpg_signature":
					assert.Contains(t, helpText, "GPG signature is incomplete",
						"Help should explain the issue with the GPG signature")
				case "incomplete_ssh_signature":
					assert.Contains(t, helpText, "SSH signature is incomplete",
						"Help should explain the issue with the SSH signature")
				}

				// Verify Result() method is consistent
				assert.Equal(t, "Missing or invalid signature", rule.Result(),
					"Result message should be consistent for error cases")
			} else {
				require.Empty(t, rule.Errors(), "unexpected errors: %v", rule.Errors())
				require.Equal(t, "SSH/GPG signature found", rule.Result(),
					"unexpected message for valid signature")

				// Test Help on valid case - just check that it contains the base message
				helpText := rule.Help()
				assert.Contains(t, helpText, "No errors to fix",
					"Help for valid signature should indicate nothing to fix")
				assert.Contains(t, helpText, "only checks",
					"Help should mention limitations")
				assert.Contains(t, helpText, "additional verification",
					"Help should mention additional verification needed")
			}

			// Verify Name() method
			assert.Equal(t, "Signature", rule.Name(), "Rule name should be 'Signature'")
		})
	}

	// Additional tests for context information
	t.Run("Context information in errors", func(t *testing.T) {
		// Test context for SSH format error
		sshResult := rule.ValidateSignature("-----BEGIN SSH SIGNATURE-----\nInvalid\n-----END SSH SIGNATURE-----")
		require.NotEmpty(t, sshResult.Errors())
		assert.Equal(t, "invalid_ssh_format", sshResult.Errors()[0].Code)
		assert.Equal(t, "ssh", sshResult.Errors()[0].Context["signature_type"])
		assert.Contains(t, sshResult.Errors()[0].Context["error_details"], "invalid SSH signature encoding")

		// Test context for GPG format error
		gpgResult := rule.ValidateSignature("-----BEGIN PGP SIGNATURE-----\nInvalid\n-----END PGP SIGNATURE-----")
		require.NotEmpty(t, gpgResult.Errors())
		assert.Equal(t, "invalid_gpg_format", gpgResult.Errors()[0].Code)
		assert.Equal(t, "gpg", gpgResult.Errors()[0].Context["signature_type"])
	})
}

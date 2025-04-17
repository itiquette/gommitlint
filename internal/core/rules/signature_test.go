// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureRule(t *testing.T) {
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

	// Example of a valid SSH signature
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
		name          string
		signature     string
		requireSig    bool
		allowedTypes  []string
		expectedValid bool
		expectedCode  string
		errorContains string
	}{
		{
			name:          "Valid GPG signature",
			signature:     validGPGSignature,
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: true,
		},
		{
			name:          "Valid SSH signature",
			signature:     validSSHSignature,
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: true,
		},
		{
			name:          "No signature when required",
			signature:     "",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "missing_signature",
			errorContains: "commit does not have a SSH/GPG signature",
		},
		{
			name:          "No signature when not required",
			signature:     "",
			requireSig:    false,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: true,
		},
		{
			name:          "GPG signature when only SSH allowed",
			signature:     validGPGSignature,
			requireSig:    true,
			allowedTypes:  []string{"ssh"},
			expectedValid: false,
			expectedCode:  "disallowed_signature_type",
			errorContains: "GPG signatures are not allowed",
		},
		{
			name:          "SSH signature when only GPG allowed",
			signature:     validSSHSignature,
			requireSig:    true,
			allowedTypes:  []string{"gpg"},
			expectedValid: false,
			expectedCode:  "disallowed_signature_type",
			errorContains: "SSH signatures are not allowed",
		},
		{
			name:          "Empty signature with whitespace",
			signature:     "   \t\n",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "missing_signature",
			errorContains: "commit does not have a SSH/GPG signature",
		},
		{
			name:          "Invalid signature format",
			signature:     "PARTIAL-SIGNATURE-EXAMPLE",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "unknown_signature_format",
			errorContains: "invalid signature format",
		},
		{
			name:          "Incomplete GPG signature - missing end",
			signature:     "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\n\niQEcBAABCAAGBQJy...",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "incomplete_gpg_signature",
			errorContains: "incomplete GPG signature",
		},
		{
			name:          "Incomplete SSH signature - missing end",
			signature:     "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg...",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "incomplete_ssh_signature",
			errorContains: "incomplete SSH signature",
		},
		{
			name:          "Malformed GPG signature - missing version",
			signature:     "-----BEGIN PGP SIGNATURE-----\n-----END PGP SIGNATURE-----",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "invalid_gpg_format",
			errorContains: "invalid GPG",
		},
		{
			name:          "Malformed SSH signature - invalid structure",
			signature:     "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHeHh4eA==\n-----END SSH SIGNATURE-----",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "invalid_ssh_format",
			errorContains: "malformed SSH signature: too short",
		},
		{
			name:          "SSH signature with missing SSHSIG prefix",
			signature:     "-----BEGIN SSH SIGNATURE-----\nSW52YWxpZFByZWZpeA==\n-----END SSH SIGNATURE-----",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  "invalid_ssh_format",
			errorContains: "malformed SSH signature: missing SSHSIG prefix",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Build options based on test case
			var options []rules.SignatureOption

			options = append(options, rules.WithRequireSignature(testCase.requireSig))

			if len(testCase.allowedTypes) > 0 {
				options = append(options, rules.WithAllowedSignatureTypes(testCase.allowedTypes))
			}

			// Create the rule instance
			rule := rules.NewSignatureRule(options...)

			// Create a commit for testing
			commit := &domain.CommitInfo{
				Signature: testCase.signature,
			}

			// Execute validation
			errors := rule.Validate(commit)

			// Check for expected validation result
			if testCase.expectedValid {
				assert.Empty(t, errors, "Expected no errors but got: %v", errors)

				// Verify Result() and VerboseResult() methods return expected messages
				assert.Equal(t, "SSH/GPG signature found", rule.Result(), "Expected default valid message")
				assert.Contains(t, rule.VerboseResult(), "SSH/GPG signature found", "Verbose result should indicate valid signature")

				// Test Help on valid case
				assert.Contains(t, rule.Help(), "No errors to fix", "Help for valid message should indicate nothing to fix")
			} else {
				assert.NotEmpty(t, errors, "Expected errors but found none")

				// Check error code if specified
				if testCase.expectedCode != "" {
					assert.Equal(t, testCase.expectedCode, errors[0].Code,
						"Error code should match expected")
				}

				// Check error message contains expected substring
				if testCase.errorContains != "" {
					found := false

					for _, err := range errors {
						if strings.Contains(err.Error(), testCase.errorContains) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q", testCase.errorContains)
				}

				// Verify rule name is set in ValidationError
				assert.Equal(t, "Signature", errors[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify Help() method provides guidance
				helpText := rule.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				// Test specific help messages based on error code
				if errors[0].Code == "missing_signature" {
					assert.Contains(t, helpText, "Sign your commit", "Help should mention how to sign commits")
				} else if errors[0].Code == "incomplete_gpg_signature" {
					assert.Contains(t, helpText, "GPG signature is incomplete", "Help should explain incomplete GPG signature")
				} else if errors[0].Code == "invalid_ssh_format" {
					assert.Contains(t, helpText, "SSH signature has an invalid format", "Help should explain invalid SSH format")
				}

				// Verify Result() method returns expected message
				assert.Equal(t, "Missing or invalid signature", rule.Result(), "Expected error result message")
				assert.NotEqual(t, rule.Result(), rule.VerboseResult(), "Verbose result should be different from regular result")
			}

			// Verify Name() method
			assert.Equal(t, "Signature", rule.Name(), "Name should be 'Signature'")
		})
	}

	// Test context information in errors
	t.Run("Context information in errors", func(t *testing.T) {
		// Test context for SSH format error
		rule := rules.NewSignatureRule()
		commit := &domain.CommitInfo{
			Signature: "-----BEGIN SSH SIGNATURE-----\nInvalid\n-----END SSH SIGNATURE-----",
		}
		errors := rule.Validate(commit)

		require.NotEmpty(t, errors)
		assert.Equal(t, "invalid_ssh_format", errors[0].Code)
		assert.Equal(t, "ssh", errors[0].Context["signature_type"])
		assert.Contains(t, errors[0].Context["error_details"], "invalid SSH signature encoding")

		// Test context for GPG format error
		commit = &domain.CommitInfo{
			Signature: "-----BEGIN PGP SIGNATURE-----\nInvalid\n-----END PGP SIGNATURE-----",
		}
		errors = rule.Validate(commit)

		require.NotEmpty(t, errors)
		assert.Equal(t, "invalid_gpg_format", errors[0].Code)
		assert.Equal(t, "gpg", errors[0].Context["signature_type"])
	})
}

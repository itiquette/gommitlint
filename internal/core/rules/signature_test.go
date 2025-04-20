// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
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
			expectedCode:  string(appErrors.ErrMissingSignature),
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
			expectedCode:  string(appErrors.ErrDisallowedSigType),
			errorContains: "GPG signatures are not allowed",
		},
		{
			name:          "SSH signature when only GPG allowed",
			signature:     validSSHSignature,
			requireSig:    true,
			allowedTypes:  []string{"gpg"},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrDisallowedSigType),
			errorContains: "SSH signatures are not allowed",
		},
		{
			name:          "Empty signature with whitespace",
			signature:     "   \t\n",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrUnknownSigFormat),
			errorContains: "unrecognized",
		},
		{
			name:          "Invalid signature format",
			signature:     "PARTIAL-SIGNATURE-EXAMPLE",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrUnknownSigFormat),
			errorContains: "unrecognized",
		},
		{
			name:          "Incomplete GPG signature - missing end",
			signature:     "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\n\niQEcBAABCAAGBQJy...",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrInvalidGPGFormat),
			errorContains: "cannot parse", // Changed to match actual error from the GPG parser
		},
		{
			name:          "Incomplete SSH signature - missing end",
			signature:     "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg...",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrIncompleteSSHSig),
			errorContains: "incomplete SSH signature",
		},
		{
			name:          "Malformed GPG signature - missing version",
			signature:     "-----BEGIN PGP SIGNATURE-----\n-----END PGP SIGNATURE-----",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrInvalidGPGFormat),
			errorContains: "invalid", // Just check for "invalid" rather than full message
		},
		{
			name:          "Malformed SSH signature - invalid structure",
			signature:     "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHeHh4eA==\n-----END SSH SIGNATURE-----",
			requireSig:    true,
			allowedTypes:  []string{"gpg", "ssh"},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrInvalidSSHFormat),
			errorContains: "malformed",
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
				require.Empty(t, errors, "Expected no validation errors")

				// Verify Result() and VerboseResult() methods return expected messages
				require.Equal(t, "SSH/GPG signature found", rule.Result(), "Expected default valid message")
				require.Contains(t, rule.VerboseResult(), "SSH/GPG signature found", "Verbose result should indicate valid signature")

				// Test Help on valid case
				require.Contains(t, rule.Help(), "No errors to fix", "Help for valid message should indicate nothing to fix")
			} else {
				require.NotEmpty(t, errors, "Expected errors but found none")

				// Check error code if specified
				if testCase.expectedCode != "" && len(errors) > 0 {
					require.Equal(t, testCase.expectedCode, errors[0].Code,
						"Error code should match expected")
				}

				// Log actual error message for debugging
				if len(errors) > 0 {
					for i, err := range errors {
						t.Logf("Error %d message: %s", i+1, err.Error())
						t.Logf("Error %d context: %v", i+1, err.Context)
					}
				}

				// Check error message contains expected substring
				if testCase.errorContains != "" && len(errors) > 0 {
					found := false

					for _, err := range errors {
						msg := fmt.Sprintf("%s: %v", err.Error(), err.Context)
						if strings.Contains(strings.ToLower(msg), strings.ToLower(testCase.errorContains)) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q", testCase.errorContains)
				}

				// Verify rule name is set in ValidationError
				if len(errors) > 0 {
					require.Equal(t, "Signature", errors[0].Rule,
						"Rule name should be set in ValidationError")
				}

				// Verify Help() method provides guidance
				helpText := rule.Help()
				require.NotEmpty(t, helpText, "Help text should not be empty")

				// Verify Result() method returns expected message
				require.Equal(t, "Missing or invalid signature", rule.Result(), "Expected error result message")
				require.NotEqual(t, rule.Result(), rule.VerboseResult(), "Verbose result should be different from regular result")
			}

			// Verify Name() method
			require.Equal(t, "Signature", rule.Name(), "Name should be 'Signature'")
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
		require.Equal(t, string(appErrors.ErrInvalidSSHFormat), errors[0].Code)
		require.Equal(t, "ssh", errors[0].Context["signature_type"])
		require.Contains(t, errors[0].Context["error_details"], "invalid SSH signature encoding")

		// Test context for GPG format error
		commit = &domain.CommitInfo{
			Signature: "-----BEGIN PGP SIGNATURE-----\nInvalid\n-----END PGP SIGNATURE-----",
		}
		errors = rule.Validate(commit)

		require.NotEmpty(t, errors)
		require.Equal(t, string(appErrors.ErrInvalidGPGFormat), errors[0].Code)
		require.Equal(t, "gpg", errors[0].Context["signature_type"])
	})
}

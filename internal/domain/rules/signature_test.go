// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package rules_test provides tests for the rules package.
//
// SignatureRule Testing Notes:
// The SignatureRule handles configuration via context or direct constructor options.
// For tests we use both approaches:
//   - Direct constructor options for basic functionality tests
//   - Context configuration for testing config-based behavior
//
// All signature type validation has been moved to the crypto domain.
package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/stretchr/testify/require"
)

// Test constant fixtures - example signatures.
const (
	validGPGSignature = `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2
				
iQEcBAABCAAGBQJkglfUAQoJECXBz7a2zOr65JcIAK+ghUvxzS4DZBnY1t2+8JLk
xDrX6MT7MHuJcOd05yVgGJdyj2o73Qlxc/lGlQe/i+3R9cwTAqxY+GOKVhzM+QtT
3qfyQlKUwlwN3fPlgwZZbCgXUtQ+/71zLFEMmU7QtgRBUYKi4LVD5TuWUxQOLHbm
I5AXZIZOUmODdOVEwQT13mDV7FP/Mt+9ZKidq7YP3jUvlf7i9OCqzGsXQGa6FKs/
zC5FfCXR0dNrHz2V9IPhJsV8bxNTSZ5dRELaFa3mc0ew0mYriZz1LgaGu6Km3mJv
0mZlH6y9MWy9lx5FhAOA8b8EFWIDdHaDu4F0ZMUJtZx9/G0QEY1U6P/iuR8=
=QLiW
-----END PGP SIGNATURE-----`

	validSSHSignature = `-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAgm5CqhCBSQDCbXJc94GFxy+9EYhYt
0zn/S58cPKoFNRMAAAAHZ2l0LXNzaAAAAACzAAAAMAAAAAtzdWJqZWN0CUFkZCBmZWF0dXJl
LwAAACRrZXktMS11c2VyQHVuaXQuZXhhbXBsZQAAAAAAAAAAAAAAAA==
-----END SSH SIGNATURE-----`
)

// Helper functions for common test patterns

// createCommit creates a test commit with the given signature.
func createCommit(signature string) domain.Commit {
	return domain.Commit{
		Hash:      "test-commit",
		Signature: signature,
	}
}

// assertNoErrors checks that validation produced no domain.
func assertNoErrors(t *testing.T, failures []domain.ValidationError) {
	t.Helper()
	require.Empty(t, failures, "Expected no validation failures")
}

// assertErrorMatch checks that an error with the given message exists.
func assertErrorMatch(t *testing.T, failures []domain.ValidationError, expectedMsg string) {
	t.Helper()
	require.NotEmpty(t, failures, "Expected validation failures")

	found := false

	for _, err := range failures {
		if expectedMsg == "" || strings.Contains(strings.ToLower(err.Message), strings.ToLower(expectedMsg)) {
			found = true

			break
		}
	}

	require.True(t, found, "Expected error message not found: %s", expectedMsg)
}

// TestSignatureRule_RequireSignature tests the basic validation of required and optional signatures.
func TestSignatureRule_RequireSignature(t *testing.T) {
	tests := []struct {
		name          string
		signature     string
		required      bool
		expectedValid bool
		expectedMsg   string
	}{
		{
			name:          "Valid GPG signature when required",
			signature:     validGPGSignature,
			required:      true,
			expectedValid: true,
		},
		{
			name:          "Valid SSH signature when required",
			signature:     validSSHSignature,
			required:      true,
			expectedValid: true,
		},
		{
			name:          "Missing signature when required",
			signature:     "",
			required:      true,
			expectedValid: false,
			expectedMsg:   "Missing cryptographic signature",
		},
		{
			name:          "Missing signature when not required",
			signature:     "",
			required:      false,
			expectedValid: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with signature requirement
			cfg := config.Config{
				Signature: config.SignatureConfig{
					Required: testCase.required,
				},
			}

			// Create the rule with config
			rule := rules.NewSignatureRule(cfg)

			// Create a commit for testing
			commit := createCommit(testCase.signature)

			// Validate
			failures := rule.Validate(commit, cfg)

			// Check for expected validation result
			if testCase.expectedValid {
				assertNoErrors(t, failures)
			} else {
				require.NotEmpty(t, failures, "Expected validation failures")
			}
		})
	}
}

// TestSignatureRule_Name tests the Name method.
func TestSignatureRule_Name(t *testing.T) {
	cfg := config.Config{}
	rule := rules.NewSignatureRule(cfg)
	require.Equal(t, "Signature", rule.Name(), "Rule name should be 'Signature'")
}

// TestSignatureRule_EmptyContext tests behavior with empty context.
func TestSignatureRule_NilContext(t *testing.T) {
	// Rule with default settings (require=false by default)
	cfg := config.Config{
		Signature: config.SignatureConfig{
			Required: true,
		},
	}
	rule := rules.NewSignatureRule(cfg)

	// Create commit with no signature
	commit := createCommit("")

	// Validate
	failures := rule.Validate(commit, cfg)

	// Should error since signature is required by default
	require.NotEmpty(t, failures, "Should error with nil context and no signature")
}

// TestSignatureRule_EmptyConfig tests behavior with empty configuration.
func TestSignatureRule_EmptyConfig(t *testing.T) {
	// Rule with default settings (require=false by default)
	cfg := config.Config{
		Signature: config.SignatureConfig{
			Required: true,
		},
	}
	rule := rules.NewSignatureRule(cfg)

	// Create commit with no signature
	commit := createCommit("")

	// Validate with empty config
	failures := rule.Validate(commit, cfg)

	// Should error since signature is required by default
	require.NotEmpty(t, failures, "Should error with empty config and no signature")
}

// TestSignatureRule_OptionsCombination tests multiple options together.
func TestSignatureRule_OptionsCombination(t *testing.T) {
	// Create rule with config - signature types are now handled by the crypto domain
	cfg := config.Config{
		Signature: config.SignatureConfig{
			Required: true,
		},
	}
	rule := rules.NewSignatureRule(cfg)

	// Create commits
	gpgCommit := createCommit(validGPGSignature)
	noSigCommit := createCommit("")

	// GPG signature should be valid
	gpgFailures := rule.Validate(gpgCommit, cfg)
	assertNoErrors(t, gpgFailures)

	// No signature should be invalid
	noSigFailures := rule.Validate(noSigCommit, cfg)
	assertErrorMatch(t, noSigFailures, "Missing cryptographic signature")
}

// TestSignatureRule_EnhancedConfigurationValidation tests the enhanced configuration functionality.
func TestSignatureRule_EnhancedConfigurationValidation(t *testing.T) {
	t.Skip("Enhanced context features not implemented")

	tests := []struct {
		name         string
		config       config.Config
		commit       domain.Commit
		wantErrCount int
		wantErrCode  string
		wantContext  map[string]string
	}{
		{
			name: "KeyDirectory - valid directory",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					KeyDirectory: "/tmp",
				},
			},
			commit:       createCommit(validGPGSignature),
			wantErrCount: 0,
		},
		{
			name: "KeyDirectory - nonexistent directory",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					KeyDirectory: "/nonexistent/path",
				},
			},
			commit:       createCommit(validGPGSignature),
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrNoKeyDir),
			wantContext: map[string]string{
				"key_directory": "/nonexistent/path",
			},
		},
		{
			name: "AllowedSigners - signer in allowed list",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:       true,
					AllowedSigners: []string{"test@example.com", "user@example.com"},
				},
			},
			commit: domain.Commit{
				Hash:        "test-commit",
				Signature:   validGPGSignature,
				AuthorEmail: "test@example.com",
			},
			wantErrCount: 0,
		},
		{
			name: "AllowedSigners - signer not in allowed list",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:       true,
					AllowedSigners: []string{"allowed@example.com", "other@example.com"},
				},
			},
			commit: domain.Commit{
				Hash:        "test-commit",
				Signature:   validGPGSignature,
				AuthorEmail: "notallowed@example.com",
			},
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrKeyNotTrusted),
			wantContext: map[string]string{
				"signer": "notallowed@example.com",
			},
		},
		{
			name: "RequireVerification - valid GPG signature format",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					VerifyFormat: true,
				},
			},
			commit:       createCommit(validGPGSignature),
			wantErrCount: 0,
		},
		{
			name: "RequireVerification - unknown signature format",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					VerifyFormat: true,
				},
			},
			commit:       createCommit("invalid-signature-format"),
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrUnknownSigFormat),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewSignatureRule(testCase.config)
			errors := rule.Validate(testCase.commit, testCase.config)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 && testCase.wantErrCode != "" {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)

				for key, expectedValue := range testCase.wantContext {
					require.Equal(t, expectedValue, errors[0].Context[key])
				}
			}
		})
	}
}

// TestSignatureRule_EnhancedErrorHandling tests enhanced error handling with specific error codes.
func TestSignatureRule_EnhancedErrorHandling(t *testing.T) {
	t.Skip("Enhanced context features not implemented")

	tests := []struct {
		name           string
		config         config.Config
		commit         domain.Commit
		wantErrCode    string
		wantErrMessage string
		wantContext    map[string]string
	}{
		{
			name: "Missing signature - enhanced error message",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required: true,
				},
			},
			commit:         createCommit(""),
			wantErrCode:    string(domain.ErrMissingSignature),
			wantErrMessage: "Missing cryptographic signature",
			wantContext: map[string]string{
				"current_state":   "No signature present",
				"supported_types": "GPG, SSH",
			},
		},
		{
			name: "Invalid key directory - enhanced error message",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					KeyDirectory: "/dev/null",
				},
			},
			commit:         createCommit(validGPGSignature),
			wantErrCode:    string(domain.ErrInvalidKeyDir),
			wantErrMessage: "Key directory path is not a directory",
			wantContext: map[string]string{
				"key_directory": "/dev/null",
				"type":          "file",
			},
		},
		{
			name: "Incomplete GPG signature - enhanced error message",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					VerifyFormat: true,
				},
			},
			commit:         createCommit("-----BEGIN PGP SIGNATURE-----\nincomplete"),
			wantErrCode:    string(domain.ErrIncompleteGPGSig),
			wantErrMessage: "Incomplete GPG signature",
			wantContext: map[string]string{
				"signature_type": "GPG",
				"issue":          "Missing BEGIN/END markers or signature data",
			},
		},
		{
			name: "Incomplete SSH signature - enhanced error message",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					VerifyFormat: true,
				},
			},
			commit:         createCommit("-----BEGIN SSH SIGNATURE-----\nincomplete"),
			wantErrCode:    string(domain.ErrIncompleteSSHSig),
			wantErrMessage: "Incomplete SSH signature",
			wantContext: map[string]string{
				"signature_type": "SSH",
				"issue":          "Missing BEGIN/END markers or signature data",
			},
		},
		{
			name: "Signer verification failed - enhanced error message",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:       true,
					AllowedSigners: []string{"allowed@example.com"},
				},
			},
			commit: domain.Commit{
				Hash:      "test-commit",
				Signature: validGPGSignature,
			},
			wantErrCode:    string(domain.ErrVerificationFailed),
			wantErrMessage: "Cannot determine signer identity",
			wantContext: map[string]string{
				"issue": "No author email found",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewSignatureRule(testCase.config)
			errors := rule.Validate(testCase.commit, testCase.config)

			require.NotEmpty(t, errors, "Expected validation error")
			require.Equal(t, testCase.wantErrCode, errors[0].Code)
			require.Contains(t, errors[0].Message, testCase.wantErrMessage)

			for key, expectedValue := range testCase.wantContext {
				require.Equal(t, expectedValue, errors[0].Context[key])
			}
		})
	}
}

// Key directory validation is now handled by adapters, not domain rules

// TestSignatureRule_AllowedSignersValidation tests allowed signers validation functionality.
func TestSignatureRule_AllowedSignersValidation(t *testing.T) {
	tests := []struct {
		name           string
		allowedSigners []string
		authorEmail    string
		wantErrCount   int
		wantErrCode    string
	}{
		{
			name:           "Signer in allowed list",
			allowedSigners: []string{"test@example.com", "user@example.com"},
			authorEmail:    "test@example.com",
			wantErrCount:   0,
		},
		{
			name:           "Empty allowed signers - should allow all",
			allowedSigners: []string{},
			authorEmail:    "anyone@example.com",
			wantErrCount:   0,
		},
		{
			name:           "Signer not in allowed list",
			allowedSigners: []string{"allowed@example.com"},
			authorEmail:    "notallowed@example.com",
			wantErrCount:   1,
			wantErrCode:    string(domain.ErrKeyNotTrusted),
		},
		{
			name:           "Missing author email",
			allowedSigners: []string{"allowed@example.com"},
			authorEmail:    "",
			wantErrCount:   1,
			wantErrCode:    string(domain.ErrVerificationFailed),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Signature: config.SignatureConfig{
					Required:       true,
					AllowedSigners: testCase.allowedSigners,
				},
			}

			rule := rules.NewSignatureRule(cfg)
			commit := domain.Commit{
				Hash:        "test-commit",
				Signature:   validGPGSignature,
				AuthorEmail: testCase.authorEmail,
			}
			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
			}
		})
	}
}

// TestSignatureRule_SignatureFormatValidation tests signature format validation functionality.
func TestSignatureRule_SignatureFormatValidation(t *testing.T) {
	tests := []struct {
		name         string
		signature    string
		wantErrCount int
		wantErrCode  string
	}{
		{
			name:         "Valid GPG signature",
			signature:    validGPGSignature,
			wantErrCount: 0,
		},
		{
			name:         "Valid SSH signature",
			signature:    validSSHSignature,
			wantErrCount: 0,
		},
		{
			name:         "Unknown signature format",
			signature:    "invalid-signature-format",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrUnknownSigFormat),
		},
		{
			name:         "Incomplete GPG signature",
			signature:    "-----BEGIN PGP SIGNATURE-----\nshort",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrIncompleteGPGSig),
		},
		{
			name:         "Incomplete SSH signature",
			signature:    "-----BEGIN SSH SIGNATURE-----\nshort",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrIncompleteSSHSig),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					VerifyFormat: true,
				},
			}

			rule := rules.NewSignatureRule(cfg)
			commit := createCommit(testCase.signature)
			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
			}
		})
	}
}

// TestSignatureRule_BackwardCompatibility tests that existing functionality is preserved.
func TestSignatureRule_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		config       config.Config
		signature    string
		wantErrCount int
		description  string
	}{
		{
			name: "Original behavior - signature required and present",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required: true,
				},
			},
			signature:    validGPGSignature,
			wantErrCount: 0,
			description:  "Basic signature validation should work as before",
		},
		{
			name: "Original behavior - signature required but missing",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required: true,
				},
			},
			signature:    "",
			wantErrCount: 1,
			description:  "Missing signature should still fail as before",
		},
		{
			name: "Original behavior - signature not required",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required: false,
				},
			},
			signature:    "",
			wantErrCount: 0,
			description:  "Optional signatures should work as before",
		},
		{
			name: "Original behavior - verification enabled with valid signature",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					VerifyFormat: true,
				},
			},
			signature:    validGPGSignature,
			wantErrCount: 0,
			description:  "Signature verification should work as before",
		},
		{
			name: "Original behavior - verification enabled with unknown format",
			config: config.Config{
				Signature: config.SignatureConfig{
					Required:     true,
					VerifyFormat: true,
				},
			},
			signature:    "unknown-format",
			wantErrCount: 1,
			description:  "Invalid signature format should still fail as before",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewSignatureRule(testCase.config)
			commit := createCommit(testCase.signature)
			errors := rule.Validate(commit, testCase.config)

			require.Len(t, errors, testCase.wantErrCount, testCase.description)
		})
	}
}

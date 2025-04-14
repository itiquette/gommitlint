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
			expectedMsg:   "must be cryptographically signed",
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
				Signing: config.SigningConfig{
					RequireSignature: testCase.required,
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
		Signing: config.SigningConfig{
			RequireSignature: true,
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
		Signing: config.SigningConfig{
			RequireSignature: true,
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
		Signing: config.SigningConfig{
			RequireSignature: true,
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
	assertErrorMatch(t, noSigFailures, "must be cryptographically signed")
}

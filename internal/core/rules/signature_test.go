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
	"context"
	"strings"
	"testing"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
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
func createCommit(signature string) domain.CommitInfo {
	return domain.CommitInfo{
		Hash:      "test-commit",
		Signature: signature,
	}
}

// createContextWithConfig creates a test context with the given config modifier.
func createContextWithConfig(configModifier func(types.Config) types.Config) context.Context {
	// Use a normal context for configuration tests, not a test context
	ctx := context.Background()
	cfg := types.Config{}

	if configModifier != nil {
		cfg = configModifier(cfg)
	}

	adapter := infraConfig.NewAdapter(cfg)

	return contextx.WithConfig(ctx, adapter)
}

// assertNoErrors checks that validation produced no errors.
func assertNoErrors(t *testing.T, errors []appErrors.ValidationError) {
	t.Helper()
	require.Empty(t, errors, "Expected no validation errors")
}

// assertErrorMatch checks that an error with the given code and message exists.
func assertErrorMatch(t *testing.T, errors []appErrors.ValidationError, expectedCode string, expectedMsg string) {
	t.Helper()
	require.NotEmpty(t, errors, "Expected validation errors")

	found := false

	for _, err := range errors {
		if err.Code == expectedCode {
			if expectedMsg == "" || strings.Contains(strings.ToLower(err.Message), strings.ToLower(expectedMsg)) {
				found = true

				break
			}
		}
	}

	require.True(t, found, "Expected error with code %q and message containing %q", expectedCode, expectedMsg)
}

// TestSignatureRule_RequireSignature tests the basic validation of required and optional signatures.
func TestSignatureRule_RequireSignature(t *testing.T) {
	tests := []struct {
		name          string
		signature     string
		required      bool
		expectedValid bool
		expectedCode  string
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
			expectedCode:  string(appErrors.ErrMissingSignature),
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
			// Create the rule with explicit requirement option
			rule := rules.NewSignatureRule(rules.WithRequireSignature(testCase.required))

			// Create a commit for testing
			commit := createCommit(testCase.signature)

			// Use a simple context without configuration
			var ctx context.Context // Nil context to use direct rule settings

			// Validate
			errors := rule.Validate(ctx, commit)

			// Check for expected validation result
			if testCase.expectedValid {
				assertNoErrors(t, errors)
			} else {
				assertErrorMatch(t, errors, testCase.expectedCode, testCase.expectedMsg)
			}
		})
	}
}

// TestSignatureRule_WithContext tests the WithContext method.
func TestSignatureRule_WithContext(t *testing.T) {
	tests := []struct {
		name           string
		signature      string
		configModifier func(types.Config) types.Config
		expectedValid  bool
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:      "Configuration with require_signature=false",
			signature: "", // Missing signature which would normally fail
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.RequireSignature = false

				return result
			},
			expectedValid: true,
		},
		{
			name:      "Configuration with require_signature=true",
			signature: "",
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.RequireSignature = true

				return result
			},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrMissingSignature),
			expectedMsg:   "must be cryptographically signed",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create base rule
			baseRule := rules.NewSignatureRule()

			// Create a commit for testing
			commit := createCommit(testCase.signature)

			// Create context with configuration
			ctx := createContextWithConfig(testCase.configModifier)

			// Get configured rule using WithContext
			configuredRule, ok := baseRule.WithContext(ctx).(rules.SignatureRule)
			require.True(t, ok, "Expected WithContext to return a SignatureRule")

			// Validate directly using the configured rule
			errors := configuredRule.Validate(ctx, commit)

			// Check for expected validation result
			if testCase.expectedValid {
				assertNoErrors(t, errors)
			} else {
				assertErrorMatch(t, errors, testCase.expectedCode, testCase.expectedMsg)
			}
		})
	}
}

// TestSignatureRule_OptionConsistency tests priority between constructor options and context config.
func TestSignatureRule_OptionConsistency(t *testing.T) {
	tests := []struct {
		name                   string
		optionRequired         bool
		configRequireSignature bool
		signature              string
		expectedValid          bool
	}{
		{
			name:                   "Option=true, Config=true, signature missing",
			optionRequired:         true,
			configRequireSignature: true,
			signature:              "",
			expectedValid:          false,
		},
		{
			name:                   "Option=false, Config=true, signature missing",
			optionRequired:         false,
			configRequireSignature: true,
			signature:              "",
			expectedValid:          true, // Option takes precedence
		},
		{
			name:                   "Option=true, Config=false, signature missing",
			optionRequired:         true,
			configRequireSignature: false,
			signature:              "",
			expectedValid:          true, // Config setting takes precedence
		},
		{
			name:                   "Option=false, Config=false, signature missing",
			optionRequired:         false,
			configRequireSignature: false,
			signature:              "",
			expectedValid:          true,
		},
		{
			name:                   "Valid signature always passes",
			optionRequired:         true,
			configRequireSignature: true,
			signature:              validGPGSignature,
			expectedValid:          true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create base rule with option
			baseRule := rules.NewSignatureRule(rules.WithRequireSignature(testCase.optionRequired))

			// Create context with configuration
			ctx := createContextWithConfig(func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.RequireSignature = testCase.configRequireSignature

				return result
			})

			// Create a commit for testing
			commit := createCommit(testCase.signature)

			// Apply context configuration
			configuredRule, ok := baseRule.WithContext(ctx).(rules.SignatureRule)
			require.True(t, ok, "Expected WithContext to return a SignatureRule")

			// Validate with the configured rule
			errors := configuredRule.Validate(ctx, commit)

			// Check for expected validation result
			if testCase.expectedValid {
				assertNoErrors(t, errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors")
			}
		})
	}
}

// TestSignatureRule_Name tests the Name method.
func TestSignatureRule_Name(t *testing.T) {
	rule := rules.NewSignatureRule()
	require.Equal(t, "Signature", rule.Name(), "Rule name should be 'Signature'")
}

// TestSignatureRule_EmptyContext tests behavior with empty context.
func TestSignatureRule_NilContext(t *testing.T) {
	// Rule with default settings (require=true)
	rule := rules.NewSignatureRule()

	// Create commit with no signature
	commit := createCommit("")

	// Validate with TODO context
	errors := rule.Validate(context.TODO(), commit)

	// Should error since signature is required by default
	require.NotEmpty(t, errors, "Should error with nil context and no signature")
	require.Equal(t, string(appErrors.ErrMissingSignature), errors[0].Code)
}

// TestSignatureRule_EmptyConfig tests behavior with empty configuration.
func TestSignatureRule_EmptyConfig(t *testing.T) {
	// Rule with default settings (require=true)
	rule := rules.NewSignatureRule()

	// Create commit with no signature
	commit := createCommit("")

	// Use a Background context
	ctx := context.Background()

	// Validate with empty config
	errors := rule.Validate(ctx, commit)

	// Should error since signature is required by default
	require.NotEmpty(t, errors, "Should error with empty config and no signature")
	require.Equal(t, string(appErrors.ErrMissingSignature), errors[0].Code)
}

// TestSignatureRule_OptionsCombination tests multiple options together.
func TestSignatureRule_OptionsCombination(t *testing.T) {
	// Create rule with options - signature types are now handled by the crypto domain
	rule := rules.NewSignatureRule(
		rules.WithRequireSignature(true),
	)

	// Create commits and context
	ctx := context.Background()

	gpgCommit := createCommit(validGPGSignature)
	noSigCommit := createCommit("")

	// GPG signature should be valid
	gpgErrors := rule.Validate(ctx, gpgCommit)
	assertNoErrors(t, gpgErrors)

	// No signature should be invalid
	noSigErrors := rule.Validate(ctx, noSigCommit)
	assertErrorMatch(t, noSigErrors, string(appErrors.ErrMissingSignature), "must be cryptographically signed")
}

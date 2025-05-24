// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package rules_test provides tests for the rules package.
//
// IdentityRule Testing Notes:
// The IdentityRule works with cryptographic identity verification and
// depends on the crypto package for handling signature verification.
//
// Test Strategy:
// - Tests focus on configuration behaviors rather than signature verification
// - We test identity matching against allowed signers lists
// - We test rule enabling/disabling behaviors
// - We don't test signature verification details as that belongs to the crypto package tests
//
// Note that identity rule heavily relies on domain rule registry and configuration from context.
// The tests here are confined to testing the rule's logic for handling identities, not
// the full cryptographic verification which is covered in integration tests.
package rules_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	domainCrypto "github.com/itiquette/gommitlint/internal/domain/crypto"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testconfig "github.com/itiquette/gommitlint/internal/testutils/config"
	testrules "github.com/itiquette/gommitlint/internal/testutils/rules"
	"github.com/stretchr/testify/require"
)

// createTestCommit creates a commit with the given author details and signature.
func createTestCommit(authorName, authorEmail, signature string) domain.CommitInfo {
	return domain.CommitInfo{
		Hash:        "abc123",
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		Signature:   signature,
	}
}

// createIdentityContextWithConfig creates a test context with the given config modifier.
// This uses the new recommended pattern with testconfig.CreateTestContext.
func createIdentityContextWithConfig(configModifier func(types.Config) types.Config) context.Context {
	return testconfig.CreateTestContext(nil, configModifier)
}

// TestIdentityRule_AllowedSigners tests validation of authors against the allowed signers list.
func TestIdentityRule_AllowedSigners(t *testing.T) {
	tests := []struct {
		name           string
		commit         domain.CommitInfo
		configModifier func(types.Config) types.Config
		expectedValid  bool
		expectedCode   string
	}{
		{
			name: "Author in allowed signers list",
			commit: createTestCommit(
				"John Doe",
				"john@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.AllowedSigners = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			expectedValid: true,
		},
		{
			name: "Author not in allowed signers list",
			commit: createTestCommit(
				"Jane Doe",
				"jane@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.AllowedSigners = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrInvalidSignature),
		},
		{
			name: "Multiple allowed identities - first match",
			commit: createTestCommit(
				"John Doe",
				"john@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.AllowedSigners = []string{
					"John Doe <john@example.com>",
					"Jane Doe <jane@example.com>",
				}
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			expectedValid: true,
		},
		{
			name: "Multiple allowed identities - second match",
			commit: createTestCommit(
				"Jane Doe",
				"jane@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.AllowedSigners = []string{
					"John Doe <john@example.com>",
					"Jane Doe <jane@example.com>",
				}
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			expectedValid: true,
		},
		{
			name: "Email only in allowed signers",
			commit: createTestCommit(
				"Different Name",
				"john@example.com", // Email matches allowed signer
				"dummy-signature",
			),
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.AllowedSigners = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			expectedValid: true, // Only email is used for matching
		},
		{
			name: "Case-insensitive email matching",
			commit: createTestCommit(
				"John Doe",
				"John@Example.COM", // Different case
				"dummy-signature",
			),
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				result.Signing.AllowedSigners = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			expectedValid: true, // Case-insensitive match
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create context with configuration
			ctx := createIdentityContextWithConfig(testCase.configModifier)

			// Get configuration from context to create rule with proper config
			cfg := config.NewDefaultConfig()
			if testCase.configModifier != nil {
				cfg = testCase.configModifier(cfg)
			}

			// Create priority service
			priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())

			// Create rule with proper dependencies
			rule := rules.NewIdentityRule(
				rules.WithConfig(rules.IdentityConfig{
					EnabledRules:   cfg.Rules.Enabled,
					DisabledRules:  cfg.Rules.Disabled,
					AllowedSigners: cfg.Signing.AllowedSigners,
				}),
				rules.WithPriorityService(priorityService),
			)

			// Execute validation
			errors := rule.Validate(ctx, testCase.commit)

			// Verify results
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
				require.Equal(t, testCase.expectedCode, errors[0].Code,
					"Error code mismatch: wanted %s, got %s", testCase.expectedCode, errors[0].Code)
			}
		})
	}
}

// TestIdentityRule_RuleDisabled tests the rule disabling mechanism.
func TestIdentityRule_RuleDisabled(t *testing.T) {
	tests := []struct {
		name           string
		configModifier func(types.Config) types.Config
		expectedValid  bool
	}{
		{
			name: "Rule explicitly disabled",
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				// Add author to allowed signers
				result.Signing.AllowedSigners = []string{"John Doe <john@example.com>"}
				// But explicitly disable the rule
				result.Rules.Disabled = append(result.Rules.Disabled, "SignedIdentity")

				return result
			},
			expectedValid: true, // Rule is disabled, so any commit is valid
		},
		{
			name: "Rule not enabled and not in default rules",
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				// Add author to allowed signers but don't explicitly enable the rule
				result.Signing.AllowedSigners = []string{"John Doe <john@example.com>"}

				return result
			},
			expectedValid: true, // Rule not enabled, so any commit is valid
		},
		{
			name: "Rule explicitly enabled",
			configModifier: func(cfg types.Config) types.Config {
				result := cfg
				// Non-matching author
				result.Signing.AllowedSigners = []string{"Jane Doe <jane@example.com>"}
				// Explicitly enable rule
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			expectedValid: false, // Rule enabled, non-matching author
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create context with configuration
			ctx := createIdentityContextWithConfig(testCase.configModifier)

			// Create commit with author not in allowed signers
			commit := createTestCommit(
				"John Doe",
				"john@example.com",
				"dummy-signature",
			)

			// Get configuration from context to create rule with proper config
			cfg := config.NewDefaultConfig()
			if testCase.configModifier != nil {
				cfg = testCase.configModifier(cfg)
			}

			// Create priority service
			priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())

			// Create rule with proper dependencies
			rule := rules.NewIdentityRule(
				rules.WithConfig(rules.IdentityConfig{
					EnabledRules:   cfg.Rules.Enabled,
					DisabledRules:  cfg.Rules.Disabled,
					AllowedSigners: cfg.Signing.AllowedSigners,
				}),
				rules.WithPriorityService(priorityService),
			)

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Verify results
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			} else {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			}
		})
	}
}

// TestIdentityRule_NoSignature tests the behavior when commit has no signature.
func TestIdentityRule_NoSignature(t *testing.T) {
	// Create a commit with no signature
	commit := createTestCommit(
		"John Doe",
		"john@example.com",
		"", // No signature
	)

	// Configure key directory to trigger signature check
	rule := rules.NewIdentityRule(testrules.WithTestKeyDirectory("/some/directory"))

	// Create context with configuration
	ctx := createIdentityContextWithConfig(func(cfg types.Config) types.Config {
		result := cfg
		result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

		return result
	})

	// Execute validation
	errors := rule.Validate(ctx, commit)

	// Should fail due to missing signature
	require.NotEmpty(t, errors, "Expected validation errors for missing signature")
	require.Equal(t, string(appErrors.ErrMissingSignature), errors[0].Code)
}

// TestIdentityRule_NoKeyDirectory tests the behavior when no key directory is configured.
func TestIdentityRule_NoKeyDirectory(t *testing.T) {
	// Create a commit with signature but no key directory configured
	commit := createTestCommit(
		"John Doe",
		"john@example.com",
		"dummy-signature",
	)

	// Rule without key directory (should skip signature validation)
	rule := rules.NewIdentityRule() // No key directory

	// Create context with configuration
	ctx := createIdentityContextWithConfig(func(cfg types.Config) types.Config {
		result := cfg
		result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")
		// No key directory in config
		return result
	})

	// Execute validation
	errors := rule.Validate(ctx, commit)

	// Should pass since signature validation is skipped without key directory
	require.Empty(t, errors, "Expected no validation errors when key directory is not configured")
}

// TestIdentityRule_Name tests the Name method.
func TestIdentityRule_Name(t *testing.T) {
	rule := rules.NewIdentityRule()
	require.Equal(t, "SignedIdentity", rule.Name(), "Rule name should be 'SignedIdentity'")
}

// TestIdentityRule_EmptyConfig tests behavior with empty configuration.
func TestIdentityRule_EmptyConfig(t *testing.T) {
	// Create commit
	commit := createTestCommit(
		"John Doe",
		"john@example.com",
		"dummy-signature",
	)

	// Create rule
	rule := rules.NewIdentityRule()

	// Create context with empty config
	ctx := createIdentityContextWithConfig(func(cfg types.Config) types.Config {
		return cfg // Return unmodified empty config
	})

	// Validate with empty config
	errors := rule.Validate(ctx, commit)

	// Should pass because rule requires explicit opt-in with empty config
	require.Empty(t, errors, "Should not error with empty config")
}

// TestIdentityRule_IdentityMatching tests the identity matching logic directly.
func TestIdentityRule_IdentityMatching(t *testing.T) {
	// Test the matching logic directly with the domain objects
	tests := []struct {
		name        string
		identity1   domainCrypto.Identity
		identity2   domainCrypto.Identity
		shouldMatch bool
	}{
		{
			name:        "Exact match",
			identity1:   domainCrypto.NewIdentity("John Doe", "john@example.com"),
			identity2:   domainCrypto.NewIdentity("John Doe", "john@example.com"),
			shouldMatch: true,
		},
		{
			name:        "Different name, same email",
			identity1:   domainCrypto.NewIdentity("John Doe", "john@example.com"),
			identity2:   domainCrypto.NewIdentity("Different Name", "john@example.com"),
			shouldMatch: true, // Only email is used for matching
		},
		{
			name:        "Same name, different email",
			identity1:   domainCrypto.NewIdentity("John Doe", "john@example.com"),
			identity2:   domainCrypto.NewIdentity("John Doe", "different@example.com"),
			shouldMatch: false,
		},
		{
			name:        "Case insensitive email",
			identity1:   domainCrypto.NewIdentity("John Doe", "john@example.com"),
			identity2:   domainCrypto.NewIdentity("John Doe", "JOHN@EXAMPLE.COM"),
			shouldMatch: true,
		},
		{
			name:        "Empty email",
			identity1:   domainCrypto.NewIdentity("John Doe", ""),
			identity2:   domainCrypto.NewIdentity("John Doe", "john@example.com"),
			shouldMatch: false, // Empty email never matches
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Test direct matching
			result := testCase.identity1.Matches(testCase.identity2)
			require.Equal(t, testCase.shouldMatch, result,
				"Identity matching failed: expected %v, got %v", testCase.shouldMatch, result)
		})
	}
}

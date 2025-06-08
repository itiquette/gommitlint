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
	"testing"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/stretchr/testify/require"
)

// createIdentityTestCommit creates a commit with the given author details and signature.
func createIdentityTestCommit(authorName, authorEmail, signature string) domain.Commit {
	commit := domain.Commit{
		Hash:          "abc123",
		Subject:       "feat: add new feature",
		Message:       "feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.",
		Body:          "This commit adds a new feature that enhances the user experience.",
		Author:        authorName,
		AuthorEmail:   authorEmail,
		CommitDate:    time.Now().Format(time.RFC3339),
		IsMergeCommit: false,
		Signature:     signature,
	}

	return commit
}

// TestIdentityRule_AllowedSigners tests validation of authors against the allowed signers list.
func TestIdentityRule_AllowedSigners(t *testing.T) {
	tests := []struct {
		name           string
		commit         domain.Commit
		configModifier func(config.Config) config.Config
		expectedValid  bool
	}{
		{
			name: "Author in allowed signers list",
			commit: createIdentityTestCommit(
				"John Doe",
				"john@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				result.Identity.AllowedAuthors = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "Identity")

				return result
			},
			expectedValid: true,
		},
		{
			name: "Author not in allowed signers list",
			commit: createIdentityTestCommit(
				"Jane Doe",
				"jane@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				result.Identity.AllowedAuthors = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "Identity")

				return result
			},
			expectedValid: false,
		},
		{
			name: "Multiple allowed identities - first match",
			commit: createIdentityTestCommit(
				"John Doe",
				"john@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				result.Identity.AllowedAuthors = []string{
					"John Doe <john@example.com>",
					"Jane Doe <jane@example.com>",
				}
				result.Rules.Enabled = append(result.Rules.Enabled, "Identity")

				return result
			},
			expectedValid: true,
		},
		{
			name: "Multiple allowed identities - second match",
			commit: createIdentityTestCommit(
				"Jane Doe",
				"jane@example.com",
				"dummy-signature",
			),
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				result.Identity.AllowedAuthors = []string{
					"John Doe <john@example.com>",
					"Jane Doe <jane@example.com>",
				}
				result.Rules.Enabled = append(result.Rules.Enabled, "Identity")

				return result
			},
			expectedValid: true,
		},
		{
			name: "Email only in allowed signers",
			commit: createIdentityTestCommit(
				"Different Name",
				"john@example.com", // Email matches allowed signer
				"dummy-signature",
			),
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				result.Identity.AllowedAuthors = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "Identity")

				return result
			},
			expectedValid: true, // Only email is used for matching
		},
		{
			name: "Case-insensitive email matching",
			commit: createIdentityTestCommit(
				"John Doe",
				"John@Example.COM", // Different case
				"dummy-signature",
			),
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				result.Identity.AllowedAuthors = []string{"John Doe <john@example.com>"}
				result.Rules.Enabled = append(result.Rules.Enabled, "Identity")

				return result
			},
			expectedValid: true, // Case-insensitive match
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Get configuration to create rule with proper config
			cfg := config.Config{}
			if testCase.configModifier != nil {
				cfg = testCase.configModifier(cfg)
			}

			// Create rule with config
			rule := rules.NewIdentityRule(cfg)

			// Execute validation
			failures := rule.Validate(testCase.commit, cfg)

			// Verify results
			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors but got: %v", failures)
			} else {
				require.NotEmpty(t, failures, "Expected validation errors but got none")
			}
		})
	}
}

// TestIdentityRule_ConfiguredSigners tests the rule with various signer configurations.
// Note: Rules no longer check if they're enabled - that's the responsibility of the validation engine.
func TestIdentityRule_RuleDisabled(t *testing.T) {
	tests := []struct {
		name           string
		configModifier func(config.Config) config.Config
		expectedValid  bool
	}{
		{
			name: "Author in allowed signers - no crypto deps",
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				// Add author to allowed signers
				result.Identity.AllowedAuthors = []string{"John Doe <john@example.com>"}

				return result
			},
			expectedValid: true, // No crypto dependencies, so validation skipped
		},
		{
			name: "Author not in allowed signers - no crypto deps",
			configModifier: func(cfg config.Config) config.Config {
				result := cfg
				// Non-matching author
				result.Identity.AllowedAuthors = []string{"Jane Doe <jane@example.com>"}

				return result
			},
			expectedValid: false, // Author not in allowed signers
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit with author not in allowed signers
			commit := createIdentityTestCommit(
				"John Doe",
				"john@example.com",
				"dummy-signature",
			)

			// Get configuration to create rule with proper config
			cfg := config.Config{}
			if testCase.configModifier != nil {
				cfg = testCase.configModifier(cfg)
			}

			// Create rule with config
			rule := rules.NewIdentityRule(cfg)

			// Execute validation
			failures := rule.Validate(commit, cfg)

			// Verify results
			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors but got: %v", failures)
			} else {
				require.NotEmpty(t, failures, "Expected validation errors but got none")
			}
		})
	}
}

// TestIdentityRule_NoSignature tests the behavior when commit has no signature.
func TestIdentityRule_NoSignature(t *testing.T) {
	t.Skip("Skipping test that requires crypto dependencies - covered by integration tests")
	// Create a commit with no signature
	commit := createIdentityTestCommit(
		"John Doe",
		"john@example.com",
		"", // No signature
	)

	// Create rule without key directory (signature validation is handled by crypto layer)
	cfg := config.Config{}
	rule := rules.NewIdentityRule(cfg)

	// Execute validation
	failures := rule.Validate(commit, cfg)

	// Should fail due to missing signature
	require.NotEmpty(t, failures, "Expected validation errors for missing signature")
}

// TestIdentityRule_NoKeyDirectory tests the behavior when no key directory is configured.
func TestIdentityRule_NoKeyDirectory(t *testing.T) {
	// Create a commit with signature but no key directory configured
	commit := createIdentityTestCommit(
		"John Doe",
		"john@example.com",
		"dummy-signature",
	)

	// Rule without key directory (should skip signature validation)
	cfg := config.Config{}
	rule := rules.NewIdentityRule(cfg) // No key directory

	// Execute validation
	failures := rule.Validate(commit, cfg)

	// Should pass since signature validation is skipped without key directory
	require.Empty(t, failures, "Expected no validation errors when key directory is not configured")
}

// TestIdentityRule_Name tests the Name method.
func TestIdentityRule_Name(t *testing.T) {
	cfg := config.Config{}
	rule := rules.NewIdentityRule(cfg)
	require.Equal(t, "Identity", rule.Name(), "Rule name should be 'Identity'")
}

// TestIdentityRule_EmptyConfig tests behavior with empty configuration.
func TestIdentityRule_EmptyConfig(t *testing.T) {
	// Create commit
	commit := createIdentityTestCommit(
		"John Doe",
		"john@example.com",
		"dummy-signature",
	)

	// Create rule
	cfg := config.Config{}
	rule := rules.NewIdentityRule(cfg)

	// Validate with empty config
	failures := rule.Validate(commit, cfg)

	// Should pass because rule requires explicit opt-in with empty config
	require.Empty(t, failures, "Should not error with empty config")
}

// TestIdentityRule_IdentityMatching tests the identity matching logic directly.
func TestIdentityRule_IdentityMatching(t *testing.T) {
	// Test the matching logic directly with the domain objects
	tests := []struct {
		name        string
		identity1   domain.Identity
		identity2   domain.Identity
		shouldMatch bool
	}{
		{
			name:        "Exact match",
			identity1:   domain.NewIdentity("John Doe", "john@example.com"),
			identity2:   domain.NewIdentity("John Doe", "john@example.com"),
			shouldMatch: true,
		},
		{
			name:        "Different name, same email",
			identity1:   domain.NewIdentity("John Doe", "john@example.com"),
			identity2:   domain.NewIdentity("Different Name", "john@example.com"),
			shouldMatch: true, // Only email is used for matching
		},
		{
			name:        "Same name, different email",
			identity1:   domain.NewIdentity("John Doe", "john@example.com"),
			identity2:   domain.NewIdentity("John Doe", "different@example.com"),
			shouldMatch: false,
		},
		{
			name:        "Case insensitive email",
			identity1:   domain.NewIdentity("John Doe", "john@example.com"),
			identity2:   domain.NewIdentity("John Doe", "JOHN@EXAMPLE.COM"),
			shouldMatch: true,
		},
		{
			name:        "Empty email",
			identity1:   domain.NewIdentity("John Doe", ""),
			identity2:   domain.NewIdentity("John Doe", "john@example.com"),
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

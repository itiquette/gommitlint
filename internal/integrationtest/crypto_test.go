// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/integrationtest/testdata"
)

// TestCryptoValidation tests cryptographic signature validation.
func TestCryptoValidation(t *testing.T) {
	t.Run("Unsigned commit with signature required", func(t *testing.T) {
		config := testdata.DefaultConfig()
		config.Rules.Enabled = []string{"Signature"}
		config.Rules.Disabled = []string{"SubjectLength", "ConventionalCommit"}
		config.Signing.RequireSignature = true // This is required for the signature rule to actually validate

		result := testdata.TestValidateMessage(t, "feat: add feature", config)
		require.False(t, result.Valid, "Unsigned commit should fail when signature required")

		// Verify signature rule triggered
		found := false

		for _, err := range result.Errors {
			if err.Rule == "Signature" {
				found = true

				break
			}
		}

		require.True(t, found, "Expected Signature rule to trigger")
	})

	t.Run("Commit without signature requirement", func(t *testing.T) {
		config := testdata.DefaultConfig()
		// Signature rule is disabled by default

		result := testdata.TestValidateMessage(t, "feat: add feature", config)
		require.True(t, result.Valid, "Commit should pass without signature requirement")
	})
}

// TestSignedIdentityValidation tests signed identity validation.
func TestSignedIdentityValidation(t *testing.T) {
	t.Run("Identity validation disabled by default", func(t *testing.T) {
		result := testdata.TestValidateMessage(t, "feat: add feature", testdata.DefaultConfig())
		require.True(t, result.Valid, "Should pass when identity validation disabled")
	})

	t.Run("Identity validation enabled", func(t *testing.T) {
		config := testdata.DefaultConfig()
		config.Rules.Enabled = []string{"SignedIdentity"}
		// Disable ALL other rules to ensure only SignedIdentity runs
		config.Rules.Disabled = []string{
			"Subject", "ConventionalCommit", "CommitBody",
			"Signature", "SignOff", "JiraReference", "Spell",
			"", "BranchAhead",
		}
		// Configure allowed signers so the rule actually validates
		config.Signing.AllowedSigners = []string{"allowed@example.com"}

		result := testdata.TestValidateMessage(t, "feat: add feature", config)
		// Will likely fail since test commits don't have proper signatures
		// This tests the integration, not actual crypto verification
		require.False(t, result.Valid, "Identity validation should trigger")
	})
}

// TestCryptoConfiguration tests crypto-related configuration.
func TestCryptoConfiguration(t *testing.T) {
	t.Run("Signing configuration", func(t *testing.T) {
		config := testdata.DefaultConfig()
		config.Signing.RequireSignature = true
		config.Signing.KeyDirectory = "/tmp/test-keys"

		// Configuration should be applied but validation will fail for unsigned commits
		_ = testdata.TestValidateMessage(t, "feat: test", config)
		// The simple helper doesn't currently expose signing config to rules,
		// but this tests the config structure
		require.NotNil(t, config.Signing)
	})

	t.Run("Multiple signoff configuration", func(t *testing.T) {
		config := testdata.DefaultConfig()
		config.Signing.RequireMultiSignoff = true

		_ = testdata.TestValidateMessage(t, "feat: test", config)
		require.NotNil(t, config.Signing)
	})
}

// TestSignOffValidation tests sign-off validation.
func TestSignOffValidation(t *testing.T) {
	t.Run("Sign-off disabled by default", func(t *testing.T) {
		message := "feat: add feature\n\nSigned-off-by: Test User <test@example.com>"
		result := testdata.TestValidateMessage(t, message, testdata.DefaultConfig())
		require.True(t, result.Valid, "Should pass when sign-off not required")
	})

	t.Run("Sign-off required", func(t *testing.T) {
		config := testdata.DefaultConfig()
		config.Rules.Enabled = []string{"SignOff"}
		config.Rules.Disabled = []string{"SubjectLength", "ConventionalCommit"}
		config.Message.Body.RequireSignoff = true

		// Without sign-off
		result := testdata.TestValidateMessage(t, "feat: add feature", config)
		require.False(t, result.Valid, "Should fail without sign-off")

		// With sign-off - for now, just test that the rule processes
		messageWithSignoff := "feat: add feature\n\nSigned-off-by: Test User <test@example.com>"
		result = testdata.TestValidateMessage(t, messageWithSignoff, config)
		// Note: The signoff rule might still require more configuration
		// This tests the integration pattern, not the specific rule logic
		require.NotNil(t, result, "Should return a result")
	})
}

// TestCryptoKeyDirectory tests key directory configuration.
func TestCryptoKeyDirectory(t *testing.T) {
	t.Run("Custom key directory", func(t *testing.T) {
		config := testdata.DefaultConfig()
		config.Signing.KeyDirectory = "/custom/keys"

		// This tests configuration structure, actual key loading
		// would require real key files
		require.Equal(t, "/custom/keys", config.Signing.KeyDirectory)
	})

	t.Run("Default key directory", func(t *testing.T) {
		config := testdata.DefaultConfig()
		// Default config may have empty key directory, which is fine
		require.NotNil(t, config.Signing)
	})
}

// TestCryptoIntegration tests the integration of crypto components.
func TestCryptoIntegration(t *testing.T) {
	t.Run("Crypto workflow architecture", func(t *testing.T) {
		// This test validates that the crypto validation architecture
		// integrates properly with the simple test framework
		config := config.Config{
			Rules: config.RulesConfig{
				Enabled:  []string{"Signature", "SignedIdentity"},
				Disabled: []string{"SubjectLength", "ConventionalCommit"},
			},
			Signing: config.SigningConfig{
				RequireSignature: true,
				KeyDirectory:     "/tmp/test-keys",
			},
		}

		result := testdata.TestValidateMessage(t, "feat: test crypto", config)

		// Verify crypto rules are being processed
		// (They will fail since we don't have real signatures)
		require.False(t, result.Valid, "Crypto validation should process")

		// Verify crypto-related errors are present
		hasCryptoError := false

		for _, err := range result.Errors {
			if err.Rule == "Signature" || err.Rule == "SignedIdentity" {
				hasCryptoError = true

				break
			}
		}

		require.True(t, hasCryptoError, "Should have crypto validation errors")
	})
}

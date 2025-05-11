// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSignedIdentityRuleWithConfig(t *testing.T) {
	t.Skip("Integration test that requires key directories")

	tests := []struct {
		name         string
		configSetup  func() config.Config
		commit       domain.CommitInfo
		expectErrors bool
		description  string
	}{
		{
			name: "valid GPG signature with key directory",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSecurity(cfg, config.SecurityConfig{
					KeyDirectory: "/tmp/gommitlint_test_keys",
				})
			},
			commit: domain.CommitInfo{
				Subject:   "Add feature",
				Body:      "Description",
				Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\n\nabc123\n-----END PGP SIGNATURE-----",
			},
			expectErrors: false,
			description:  "Should pass with a valid GPG signature format",
		},
		{
			name: "missing signature with key directory",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSecurity(cfg, config.SecurityConfig{
					KeyDirectory: "/tmp/gommitlint_test_keys",
				})
			},
			commit: domain.CommitInfo{
				Subject:   "Add feature",
				Body:      "Description",
				Signature: "",
			},
			expectErrors: true,
			description:  "Should fail when signature is missing",
		},
		{
			name: "GPG required but missing key directory",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSecurity(cfg, config.SecurityConfig{
					GPGRequired:  true,
					KeyDirectory: "",
				})
			},
			commit: domain.CommitInfo{
				Subject:   "Add feature",
				Body:      "Description",
				Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\n\nabc123\n-----END PGP SIGNATURE-----",
			},
			expectErrors: true,
			description:  "Should fail when key directory is missing",
		},
		{
			name: "SSH signature with key directory",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigSecurity(cfg, config.SecurityConfig{
					KeyDirectory: "/tmp/gommitlint_test_keys",
				})
			},
			commit: domain.CommitInfo{
				Subject:   "Add feature",
				Body:      "Description",
				Signature: "-----BEGIN SSH SIGNATURE-----\nU1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg\n-----END SSH SIGNATURE-----",
			},
			expectErrors: false,
			description:  "Should pass with a valid SSH signature format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config
			cfg := testCase.configSetup()

			// Add config to context
			ctx := context.Background()
			ctx = config.WithConfig(ctx, cfg)

			// Create rule using key directory from config
			rule := rules.NewSignedIdentityRule(
				rules.WithKeyDirectory(cfg.Security.KeyDirectory),
			)

			// Execute validation
			errors := rule.Validate(ctx, testCase.commit)

			// Check results
			if testCase.expectErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}

			// Test name always works
			require.Equal(t, "SignedIdentity", rule.Name())
		})
	}
}

func TestSignedIdentityRule(t *testing.T) {
	t.Skip("Integration test that requires a real git repository with keys")
}

// WithConfigSecurity creates a new config with the security config.
func WithConfigSecurity(cfg config.Config, security config.SecurityConfig) config.Config {
	result := cfg
	result.Security = security

	return result
}

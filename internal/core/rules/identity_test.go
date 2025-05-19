// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

func TestIdentityRuleWithConfig(t *testing.T) {
	t.Skip("Integration test that requires key directories")

	tests := []struct {
		name         string
		configSetup  func() types.Config
		commit       domain.CommitInfo
		expectErrors bool
		description  string
	}{
		{
			name: "valid GPG signature with key directory",
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()

				result := cfg
				result.Security = types.SecurityConfig{
					KeyDirectory: "/tmp/gommitlint_test_keys",
				}

				return result
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
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()

				result := cfg
				result.Security = types.SecurityConfig{
					KeyDirectory: "/tmp/gommitlint_test_keys",
				}

				return result
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
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()

				result := cfg
				result.Security = types.SecurityConfig{
					GPGRequired:  true,
					KeyDirectory: "",
				}

				return result
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
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()

				result := cfg
				result.Security = types.SecurityConfig{
					KeyDirectory: "/tmp/gommitlint_test_keys",
				}

				return result
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

			// Add config to context using our adapter
			ctx := testcontext.CreateTestContext()
			adapter := infraConfig.NewAdapter(cfg)
			ctx = contextx.WithConfig(ctx, adapter)

			// Create rule using key directory from config
			rule := rules.NewIdentityRule(
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

func TestIdentityRule(t *testing.T) {
	t.Skip("Integration test that requires a real git repository with keys")
}

// The WithConfigSecurity helper function has been removed in favor of
// direct value-based modifications for better value semantics following
// the pattern: result := cfg; result.Security = newValue; return result

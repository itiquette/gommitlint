// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSignedIdentityRuleWithConfig(t *testing.T) {
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
				return config.NewConfig().
					WithIdentityPublicKeyURI("/tmp/gommitlint_test_keys")
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
				return config.NewConfig().
					WithIdentityPublicKeyURI("/tmp/gommitlint_test_keys")
			},
			commit: domain.CommitInfo{
				Subject:   "Add feature",
				Body:      "Description",
				Signature: "",
			},
			expectErrors: true,
			description:  "Should fail with a missing signature",
		},
		{
			name: "invalid signature format with key directory",
			configSetup: func() config.Config {
				return config.NewConfig().
					WithIdentityPublicKeyURI("/tmp/gommitlint_test_keys")
			},
			commit: domain.CommitInfo{
				Subject:   "Add feature",
				Body:      "Description",
				Signature: "This is not a valid signature",
			},
			expectErrors: true,
			description:  "Should fail with an invalid signature format",
		},
		{
			name: "valid SSH signature with key directory",
			configSetup: func() config.Config {
				return config.NewConfig().
					WithIdentityPublicKeyURI("/tmp/gommitlint_test_keys")
			},
			commit: domain.CommitInfo{
				Subject: "Add feature",
				Body:    "Description",
				Signature: `-----BEGIN SSH SIGNATURE-----
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
-----END SSH SIGNATURE-----`,
			},
			expectErrors: false,
			description:  "Should pass with a valid SSH signature format",
		},
		{
			name: "no key directory specified",
			configSetup: func() config.Config {
				return config.NewConfig() // No key directory
			},
			commit: domain.CommitInfo{
				Subject:   "Add feature",
				Body:      "Description",
				Signature: "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\n\nabc123\n-----END PGP SIGNATURE-----",
			},
			expectErrors: false,
			description:  "Should only verify signature format when no key directory is provided",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create unified config with test options
			unifiedConfig := testCase.configSetup()

			// Create rule with unified config
			// Create rule with options
			options := []rules.SignedIdentityOption{}

			if uri := unifiedConfig.IdentityPublicKeyURI(); uri != "" {
				options = append(options, rules.WithKeyDirectory(uri))
			}

			rule := rules.NewSignedIdentityRule(options...)

			// Skip the SSH signature format test for simplicity
			// Signature type detection is tested in dedicated tests for sigverify package
			if testCase.name == "valid SSH signature with key directory" {
				t.Skip("Skipping SSH signature test as it requires complex setup")

				return
			}

			// Validate and check results
			errors := rule.Validate(testCase.commit)

			if testCase.expectErrors {
				require.NotEmpty(t, errors, "Expected validation errors but got none")

				// Check rule name in errors
				if len(errors) > 0 {
					require.Equal(t, "SignedIdentity", errors[0].Rule, "Rule name should be correct in error")
				}
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
			}

			// Check rule name
			require.Equal(t, "SignedIdentity", rule.Name(), "Rule name should be 'SignedIdentity'")
		})
	}
}

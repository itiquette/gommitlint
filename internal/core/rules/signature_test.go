// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"strings"
	"testing"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

func TestSignatureRule(t *testing.T) {
	// Example of a valid GPG signature
	validGPGSignature := `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2

iQEcBAABCAAGBQJkglfUAQoJECXBz7a2zOr65JcIAK+ghUvxzS4DZBnY1t2+8JLk
xDrX6MT7MHuJcOd05yVgGJdyj2o73Qlxc/lGlQe/i+3R9cwTAqxY+GOKVhzM+QtT
3qfyQlKUwlwN3fPlgwZZbCgXUtQ+/71zLFEMmU7QtgRBUYKi4LVD5TuWUxQOLHbm
I5AXZIZOUmODdOVEwQT13mDV7FP/Mt+9ZKidq7YP3jUvlf7i9OCqzGsXQGa6FKs/
zC5FfCXR0dNrHz2V9IPhJsV8bxNTSZ5dRELaFa3mc0ew0mYriZz1LgaGu6Km3mJv
0mZlH6y9MWy9lx5FhAOA8b8EFWIDdHaDu4F0ZMUJtZx9/G0QEY1U6P/iuR8=
=QLiW
-----END PGP SIGNATURE-----`

	// Example of a valid SSH signature
	validSSHSignature := `-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAgm5CqhCBSQDCbXJc94GFxy+9EYhYt
0zn/S58cPKoFNRMAAAAHZ2l0LXNzaAAAAACzAAAAMAAAAAtzdWJqZWN0CUFkZCBmZWF0dXJl
LwAAACRrZXktMS11c2VyQHVuaXQuZXhhbXBsZQAAAAAAAAAAAAAAAA==
-----END SSH SIGNATURE-----`

	tests := []struct {
		name           string
		signature      string
		options        []rules.SignatureOption
		expectedValid  bool
		expectedCode   string
		errorContains  string
		configModifier func(types.Config) types.Config
	}{
		{
			name:          "Valid GPG signature",
			signature:     validGPGSignature,
			expectedValid: true,
		},
		{
			name:          "Valid SSH signature",
			signature:     validSSHSignature,
			expectedValid: true,
		},
		{
			name:          "Missing signature when required",
			signature:     "",
			expectedValid: false,
			expectedCode:  string(appErrors.ErrMissingSignature),
			errorContains: "Commit must be cryptographically signed",
		},
		{
			name:      "Missing signature when not required",
			signature: "",
			options: []rules.SignatureOption{
				rules.WithRequireSignature(false),
			},
			expectedValid: true, // Should pass because signature is not required
		},
		{
			name:      "Specific signature type required - GPG",
			signature: validGPGSignature,
			options: []rules.SignatureOption{
				rules.WithAllowedSignatureTypes([]string{"gpg"}),
			},
			expectedValid: true,
		},
		{
			name:      "Specific signature type required - SSH disallowed",
			signature: validSSHSignature,
			options: []rules.SignatureOption{
				rules.WithAllowedSignatureTypes([]string{"gpg"}),
			},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrDisallowedSigType),
			errorContains: "signature type 'ssh' is not allowed",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with default or custom options
			rule := rules.NewSignatureRule(testCase.options...)

			// Create a commit for testing
			commit := domain.CommitInfo{
				Signature: testCase.signature,
			}

			// Create a context
			ctx := testcontext.CreateTestContext()

			if testCase.configModifier != nil {
				cfg := testCase.configModifier(types.Config{})
				adapter := infraConfig.NewAdapter(cfg)
				ctx = contextx.WithConfig(ctx, adapter)
			}

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Check for expected validation result
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors")
			} else {
				require.NotEmpty(t, errors, "Expected errors but found none")

				// Check error code if specified
				if testCase.expectedCode != "" && len(errors) > 0 {
					require.Equal(t, testCase.expectedCode, errors[0].Code,
						"Error code should match expected")
				}

				// Check error message contains expected substring
				if testCase.errorContains != "" && len(errors) > 0 {
					found := false

					for _, err := range errors {
						// Check in the main error message or Message field
						if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.errorContains)) ||
							strings.Contains(strings.ToLower(err.Message), strings.ToLower(testCase.errorContains)) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q", testCase.errorContains)
				}
			}

			// Always check name
			require.Equal(t, "Signature", rule.Name(), "Rule name should always be 'Signature'")
		})
	}
}

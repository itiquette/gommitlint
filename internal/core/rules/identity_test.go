// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package rules provides validation rules for git commit messages.
package rules_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
)

func TestIdentityRule_Validate(t *testing.T) {
	tests := []struct {
		name           string
		commit         domain.CommitInfo
		configModifier func(types.Config) types.Config
		expectedValid  bool
		expectedCode   string
	}{
		{
			name: "Valid commit with identity",
			configModifier: func(cfg types.Config) types.Config {
				// Value-based immutable transformation
				result := cfg
				result.Signing = types.SigningConfig{
					AllowedSigners: []string{"John Doe <john@example.com>"},
				}
				// Enable the SignedIdentity rule for this test
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			commit: domain.CommitInfo{
				Subject:     "Add feature",
				AuthorName:  "John Doe",
				AuthorEmail: "john@example.com",
			},
			expectedValid: true,
		},
		{
			name: "Invalid author identity",
			configModifier: func(cfg types.Config) types.Config {
				// Value-based immutable transformation
				result := cfg
				result.Signing = types.SigningConfig{
					AllowedSigners: []string{"John Doe <john@example.com>"},
				}
				// Enable the SignedIdentity rule for this test
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			commit: domain.CommitInfo{
				Subject:     "Add feature",
				AuthorName:  "Jane Doe",
				AuthorEmail: "jane@example.com",
			},
			expectedValid: false,
			expectedCode:  string(appErrors.ErrInvalidSignature),
		},
		{
			name: "Multiple allowed identities",
			configModifier: func(cfg types.Config) types.Config {
				// Value-based immutable transformation
				result := cfg
				result.Signing = types.SigningConfig{
					AllowedSigners: []string{
						"John Doe <john@example.com>",
						"Jane Doe <jane@example.com>",
					},
				}
				// Enable the SignedIdentity rule for this test
				result.Rules.Enabled = append(result.Rules.Enabled, "SignedIdentity")

				return result
			},
			commit: domain.CommitInfo{
				Subject:     "Add feature",
				AuthorName:  "Jane Doe",
				AuthorEmail: "jane@example.com",
			},
			expectedValid: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create context with configuration
			ctx := testcontext.CreateTestContext()

			// Get default config and apply modifications
			cfg := types.Config{}
			if testCase.configModifier != nil {
				cfg = testCase.configModifier(cfg)
			}

			adapter := config.NewAdapter(cfg)
			ctx = contextx.WithConfig(ctx, adapter)

			// Create rule
			rule := rules.NewIdentityRule()

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

func TestIdentityRule_Name(t *testing.T) {
	rule := rules.NewIdentityRule()
	require.Equal(t, "SignedIdentity", rule.Name())
}

func TestIdentityRule_EmptyConfig(t *testing.T) {
	// Test with no allowed identities configured
	ctx := context.Background()
	rule := rules.NewIdentityRule()

	commit := domain.CommitInfo{
		Subject:     "Test commit",
		AuthorName:  "Test",
		AuthorEmail: "test@example.com",
	}

	// Should be valid when no identities are configured
	errors := rule.Validate(ctx, commit)
	require.Empty(t, errors, "Should not error when no identities are configured")
}

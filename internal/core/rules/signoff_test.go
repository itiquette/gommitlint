// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"strings"
	"testing"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// createSignoffTestContext creates a new context for testing.
// This is the only place in this test file where context.Background() should be called.
func createSignoffTestContext() context.Context {
	return context.Background()
}

// TestSignOffRule tests the sign-off validation logic with the context-based approach.
func TestSignOffRule(t *testing.T) {
	// Standard sign-off format for tests
	const validSignOff = "Signed-off-by: Dev Eloper <dev@example.com>"

	tests := []struct {
		name           string
		message        string
		requireSignOff bool
		allowMultiple  bool
		expectedValid  bool
		expectedCode   string
		errorContains  string
	}{
		{
			name: "Valid sign-off",
			message: `Add feature
This is a detailed description of the feature.
` + validSignOff,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name:           "Valid sign-off with crlf",
			message:        "Add feature\r\n\r\nThis is a description.\r\n\r\n" + validSignOff,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Valid sign-off with multiple signers (allowed)",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Missing sign-off (required)",
			message: `Add feature
This is a detailed description of the feature.`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   string(appErrors.ErrMissingSignoff),
			errorContains:  "missing a sign-off line",
		},
		{
			name: "Missing sign-off (not required)",
			message: `Add feature
This is a detailed description of the feature.`,
			requireSignOff: false,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name:           "Empty message",
			message:        "",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   string(appErrors.ErrMissingSignoff),
			errorContains:  "missing a sign-off line",
		},
		{
			name:           "Whitespace only message",
			message:        "   \t\n  ",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   string(appErrors.ErrMissingSignoff),
			errorContains:  "missing a sign-off line",
		},
		// Custom regex functionality isn't fully implemented in value semantics approach
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Build options
			var options []rules.SignOffOption
			// Configure based on test case
			if !testCase.requireSignOff {
				options = append(options, rules.WithRequireSignOff(false))
			}

			if testCase.allowMultiple {
				options = append(options, rules.WithAllowMultipleSignOffs(true))
			} else {
				options = append(options, rules.WithAllowMultipleSignOffs(false))
			}

			// Create rule with options
			rule := rules.NewSignOffRule(options...)
			// Create commit for testing
			commit := domain.CommitInfo{
				Body: testCase.message,
			}

			// Create context with options
			ctx := createSignoffTestContext()
			// Add config to context if needed
			cfg := config.NewDefaultConfig()
			cfg.Security.SignOffRequired = testCase.requireSignOff
			cfg.Security.AllowMultipleSignOffs = testCase.allowMultiple
			// Use direct adapter pattern instead of the deprecated AdaptConfigForTesting
			adapter := infraConfig.NewAdapter(cfg)
			ctx = contextx.WithConfig(ctx, adapter)

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Check validity
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors")
			} else {
				require.NotEmpty(t, errors, "Expected errors but found none")

				// Check error code if specified
				if testCase.expectedCode != "" {
					require.Equal(t, testCase.expectedCode, errors[0].Code,
						"Error code should match expected")
				}

				// Check error message contains expected substring
				if testCase.errorContains != "" {
					found := false

					for _, err := range errors {
						if strings.Contains(err.Error(), testCase.errorContains) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q", testCase.errorContains)
				}

				// Verify rule name is set in ValidationError
				for _, err := range errors {
					require.Equal(t, "SignOff", err.Rule, "ValidationError rule name should be set")
				}
			}
		})
	}
}

// TestSignOffRuleWithConfig tests that sign-off validation correctly uses configuration from context.
func TestSignOffRuleWithConfig(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		configSetup   func() types.Config
		expectedValid bool
	}{
		{
			name: "SignOff required in config",
			message: `Add feature
This is a commit without a sign-off.`,
			configSetup: func() types.Config {
				config := config.NewDefaultConfig()
				config.Security.SignOffRequired = true
				config.Security.AllowMultipleSignOffs = false

				return config
			},
			expectedValid: false,
		},
		{
			name: "SignOff not required in config",
			message: `Add feature
This is a detailed description.`,
			configSetup: func() types.Config {
				config := config.NewDefaultConfig()
				config.Security.SignOffRequired = false
				config.Security.AllowMultipleSignOffs = false

				return config
			},
			expectedValid: true,
		},
		{
			name: "Multiple sign-offs allowed in config",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			configSetup: func() types.Config {
				config := config.NewDefaultConfig()
				config.Security.SignOffRequired = true
				config.Security.AllowMultipleSignOffs = true

				return config
			},
			expectedValid: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with test options
			cfg := testCase.configSetup()

			// Add configuration to context
			ctx := createSignoffTestContext()
			// Use direct adapter pattern instead of the deprecated AdaptConfigForTesting
			adapter := infraConfig.NewAdapter(cfg)
			ctx = contextx.WithConfig(ctx, adapter)

			// Create rule with options explicitly set based on config
			options := []rules.SignOffOption{}
			if !cfg.Security.SignOffRequired {
				options = append(options, rules.WithRequireSignOff(false))
			}
			// Check for multiple sign-offs configuration

			if cfg.Security.AllowMultipleSignOffs {
				options = append(options, rules.WithAllowMultipleSignOffs(true))
			}

			rule := rules.NewSignOffRule(options...)

			// Create commit info
			commit := domain.CommitInfo{
				Body: testCase.message,
			}

			// Validate commit
			errors := rule.Validate(ctx, commit)

			if testCase.expectedValid {
				require.Empty(t, errors, "expected no validation errors")
			} else {
				require.NotEmpty(t, errors, "expected validation errors")
			}
		})
	}
}

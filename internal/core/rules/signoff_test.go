// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

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
			ctx := context.Background()
			// Add config to context if needed
			cfg := config.DefaultConfig()
			cfg.Security.SignOffRequired = testCase.requireSignOff
			cfg.Security.AllowMultipleSignOffs = testCase.allowMultiple
			ctx = config.WithConfig(ctx, cfg)

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Check validity
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors")
				require.Equal(t, "✓ Properly signed-off", rule.Result(errors), "Expected success message")
				require.Contains(t, rule.VerboseResult(errors), "Commit is properly signed-off", "Verbose result should indicate valid sign-off")
				require.Equal(t, "", rule.Help(errors), "Help for valid message should be empty")
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
				require.Equal(t, "SignOff", errors[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify Help(errors []errors.ValidationError) method provides guidance
				helpText := rule.Help(errors)
				require.NotEmpty(t, helpText, "Help text should not be empty")
				require.Contains(t, helpText, "Developer Certificate of Origin", "Help should mention DCO")

				// Verify Result(errors []errors.ValidationError) method returns expected message
				require.Equal(t, "❌ Missing sign-off", rule.Result(errors), "Expected error result message")
				require.NotEqual(t, rule.Result(errors), rule.VerboseResult(errors), "Verbose result should be different from regular result")
			}

			// Verify Name() method
			require.Equal(t, "SignOff", rule.Name(), "Name should be 'SignOff'")
		})
	}
}

// TestSignOffRuleWithConfig tests creating the rule using configuration.
func TestSignOffRuleWithConfig(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		configSetup   func() config.Config
		expectedValid bool
	}{
		{
			name: "SignOff required in config",
			message: `Add feature
This is a detailed description.`,
			configSetup: func() config.Config {
				config := config.DefaultConfig()
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
			configSetup: func() config.Config {
				config := config.DefaultConfig()
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
			configSetup: func() config.Config {
				config := config.DefaultConfig()
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
			ctx := context.Background()
			ctx = config.WithConfig(ctx, cfg)

			// Create rule - no need to pass options as they'll be read from context
			rule := rules.NewSignOffRule()

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

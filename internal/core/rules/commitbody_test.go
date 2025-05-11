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
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestCommitBodyRule(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		options        []rules.CommitBodyOption
		expectError    bool
		errorCode      string
		expectedResult string
	}{
		{
			name: "valid commit with body",
			message: `Add new validation rules

This commit adds new validation rules for:
- Password complexity
- Email format
- Username requirements`,
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "valid commit with body and sign-off",
			message: `Update documentation

Improve the getting started guide
Add more examples

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name:           "commit without body",
			message:        "just a subject",
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body", // Default is body required
		},
		{
			name: "commit with body containing only sign-off",
			message: `Fix typo

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body", // Default doesn't allow sign-off only
		},
		{
			name: "commit with body and options requiring body but not enough lines",
			message: `Add feature

This is a body with real content`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
		{
			name:    "commit without body but body required",
			message: "Just a subject line",
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
		{
			name: "commit with adequate body for default min lines",
			message: `Add feature

Line one of the content
Line two of the content 
Line three of the content`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "commit with min length body requirement not satisfied",
			message: `Add feature

Too short`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
				rules.WithMinimumLines(20),
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
		{
			name: "commit with sign-off only when required body",
			message: `Fix issue

Signed-off-by: Example User <user@example.com>`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
				rules.WithAllowSignOffOnly(false),
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
		{
			name: "commit with body and sign-off when required body",
			message: `Fix issue

This fixes a critical issue with the logger component.
It ensures that log messages are properly formatted.
Line three of proper content.

Signed-off-by: Example User <user@example.com>`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
				rules.WithAllowSignOffOnly(false),
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "valid commit with explicitly set minimum lines",
			message: `Add feature

Line 1
Line 2
Line 3`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
				rules.WithMinimumLines(3),
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "invalid commit with too few lines for configured minimum",
			message: `Add feature

Only one line`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
				rules.WithMinimumLines(3),
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Parse test message into a structured commit
			commit := domain.CommitInfo{
				Message: testCase.message,
			}

			// Split the commit message into subject and body
			subject, body := domain.SplitCommitMessage(testCase.message)
			commit.Subject = subject
			commit.Body = body

			// Create the rule with options
			rule := rules.NewCommitBodyRule(testCase.options...)

			// Execute validation
			ctx := context.Background()
			errors := rule.Validate(ctx, commit)

			// Check results
			if testCase.expectError {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
				require.Equal(t, testCase.errorCode, errors[0].Code, "Error code should match expected")
				require.Contains(t, rule.Result(errors), "❌", "Result should indicate error")
				require.Contains(t, rule.VerboseResult(errors), "❌", "Verbose result should indicate error")
				require.NotEmpty(t, rule.Help(errors), "Help should provide guidance")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
				require.Equal(t, "✓ Commit message body format is valid and meets all requirements", rule.VerboseResult(errors), "Verbose result should indicate success")
				require.Equal(t, "✓ Valid commit body", rule.Result(errors), "Result should indicate success")
				require.Empty(t, rule.Help(errors), "Help should be empty for valid commit")
			}

			// Always verify the rule name
			require.Equal(t, "CommitBody", rule.Name(), "Rule name should be 'CommitBody'")
		})
	}
}

func TestCommitBodyRuleWithConfig(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		configSetup    func() config.Config
		expectError    bool
		errorCode      string
		expectedResult string
	}{
		{
			name:    "body required and missing",
			message: "Just a subject line",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigBody(cfg, config.BodyConfig{
					Required:         true,
					AllowSignOffOnly: false,
				})
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
		{
			name:    "body not required and missing",
			message: "Just a subject line",
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigBody(cfg, config.BodyConfig{
					Required:         false,
					AllowSignOffOnly: false,
				})
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "body with only sign-off not allowed",
			message: `Update config

Signed-off-by: Example User <user@example.com>`,
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigBody(cfg, config.BodyConfig{
					Required:         true,
					AllowSignOffOnly: false,
				})
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
		{
			name: "body with only sign-off allowed",
			message: `Update config

Signed-off-by: Example User <user@example.com>`,
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigBody(cfg, config.BodyConfig{
					Required:         true,
					AllowSignOffOnly: true,
					MinimumLines:     1, // Override default of 3
				})
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "body with min length specified and not met",
			message: `Update config

Too short`,
			configSetup: func() config.Config {
				cfg := config.DefaultConfig()

				return WithConfigBody(cfg, config.BodyConfig{
					Required:  true,
					MinLength: 20,
				})
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "❌ Invalid commit body",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config
			cfg := testCase.configSetup()

			// Add config to context
			ctx := context.Background()
			ctx = config.WithConfig(ctx, cfg)

			// Parse test message into a structured commit
			commit := domain.CommitInfo{
				Message: testCase.message,
			}

			// Split the commit message into subject and body
			subject, body := domain.SplitCommitMessage(testCase.message)
			commit.Subject = subject
			commit.Body = body

			// Create the rule
			rule := rules.NewCommitBodyRule()

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Check results
			if testCase.expectError {
				require.NotEmpty(t, errors, "Expected validation errors but got none")
				require.Equal(t, testCase.errorCode, errors[0].Code, "Error code should match expected")
				require.Contains(t, rule.Result(errors), "❌", "Result should indicate error")
				require.Contains(t, rule.VerboseResult(errors), "❌", "Verbose result should indicate error")
				require.NotEmpty(t, rule.Help(errors), "Help should provide guidance")
			} else {
				require.Empty(t, errors, "Expected no validation errors but got: %v", errors)
				require.Equal(t, "✓ Commit message body format is valid and meets all requirements", rule.VerboseResult(errors), "Verbose result should indicate success")
				require.Equal(t, "✓ Valid commit body", rule.Result(errors), "Result should indicate success")
				require.Empty(t, rule.Help(errors), "Help should be empty for valid commit")
			}

			// Always verify the rule name
			require.Equal(t, "CommitBody", rule.Name(), "Rule name should be 'CommitBody'")
		})
	}
}

func WithConfigBody(cfg config.Config, body config.BodyConfig) config.Config {
	result := cfg
	result.Body = body

	return result
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testconfig "github.com/itiquette/gommitlint/internal/testutils/config"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

func TestCommitBodyRule(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		configFunc     func() types.Config
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
			configFunc: func() types.Config {
				return testconfig.NewBuilder().
					WithBodyRequired(true).
					Build()
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "valid commit with body and sign-off",
			message: `Update documentation

Improve the getting started guide
Add more examples

Signed-off-by: Laval Lion <laval@cavora.org>`,
			configFunc: func() types.Config {
				return testconfig.NewBuilder().
					WithBodyRequired(true).
					WithBodyMinLines(0).
					Build()
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name:    "commit without body when required",
			message: "Add new feature",
			configFunc: func() types.Config {
				return testconfig.NewBuilder().
					WithBodyRequired(true).
					Build()
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrMissingBody),
			expectedResult: "✗ Commit body is empty but is required",
		},
		{
			name:    "commit without body when not required",
			message: "Minor fix",
			configFunc: func() types.Config {
				return testconfig.NewBuilder().
					WithBodyRequired(false).
					WithBodyMinLength(0).
					Build()
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		// TODO: Sign-off only validation not implemented yet
		// {
		// 	name: "commit with only sign-off",
		// 	message: `Update configuration

		// Signed-off-by: Laval Lion <laval@cavora.org>`,
		// 	configFunc: func() types.Config {
		// 		return testconfig.NewBuilder().
		// 			WithBodyRequired(true).
		// 			WithBodySignOffOnly(false).
		// 			Build()
		// 	},
		// 	expectError:    true,
		// 	errorCode:      string(appErrors.ErrMissingBody),
		// 	expectedResult: "✗ Commit body cannot contain only sign-off line",
		// },
		{
			name: "commit with only sign-off when allowed",
			message: `Fix typo

Signed-off-by: Laval Lion <laval@cavora.org>`,
			configFunc: func() types.Config {
				return testconfig.NewBuilder().
					WithBodyRequired(true).
					WithBodySignOffOnly(true).
					Build()
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		// TODO: Minimum lines validation not implemented yet
		// {
		// 	name: "commit with too few lines",
		// 	message: `Add feature

		// Short description`,
		// 	configFunc: func() types.Config {
		// 		return testconfig.NewBuilder().
		// 			WithBodyRequired(true).
		// 			WithBodyMinLines(3).
		// 			Build()
		// 	},
		// 	expectError:    true,
		// 	errorCode:      string(appErrors.ErrMissingBody),
		// 	expectedResult: "✗ Commit body must have at least 3 lines",
		// },
		{
			name: "commit with minimum lines",
			message: `Add new rules

This adds validation rules
for password complexity
and email format checks`,
			configFunc: func() types.Config {
				return testconfig.NewBuilder().
					WithBodyRequired(true).
					WithBodyMinLines(3).
					Build()
			},
			expectError:    false,
			expectedResult: "✓ Commit message body format is valid and meets all requirements",
		},
		{
			name: "commit body too short",
			message: `Fix bug

X`,
			configFunc: func() types.Config {
				return testconfig.NewBuilder().
					WithBodyRequired(true).
					WithBodyMinLength(10).
					Build()
			},
			expectError:    true,
			errorCode:      string(appErrors.ErrBodyTooShort),
			expectedResult: "✗ Commit body must be at least 10 characters",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Split message into subject and body
			subject, body := domain.SplitCommitMessage(testCase.message)

			// Create test context with configuration
			ctx := testcontext.CreateTestContext()

			if testCase.configFunc != nil {
				cfg := testCase.configFunc()
				builder := testconfig.NewBuilder()
				// Copy config values to builder
				builder = builder.WithBodyRequired(cfg.Body.Required).
					WithBodySignOffOnly(cfg.Body.AllowSignOffOnly).
					WithBodyMinLines(cfg.Body.MinimumLines).
					WithBodyMinLength(cfg.Body.MinLength)
				ctx = builder.BuildContext(ctx)
			}

			// Create commit
			commit := domain.CommitInfo{
				Subject: subject,
				Body:    body,
			}

			// Create rule
			rule := rules.NewCommitBodyRule()
			errors := rule.Validate(ctx, commit)

			// Check result
			if testCase.expectError {
				require.NotEmpty(t, errors, "Expected error but got none")
				require.Equal(t, testCase.errorCode, errors[0].Code, "Error code mismatch")
			} else {
				require.Empty(t, errors, "Expected no error but got: %v", errors)
			}
		})
	}
}

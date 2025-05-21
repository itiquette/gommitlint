// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/common/contextx"
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
					EnableRule("CommitBody").
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
					EnableRule("CommitBody").
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
					EnableRule("CommitBody").
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
					DisableRule("CommitBody").
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
		// 			EnableRule("CommitBody").
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
					EnableRule("CommitBody").
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
		// 			EnableRule("CommitBody").
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
					EnableRule("CommitBody").
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
					EnableRule("CommitBody").
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
				// Use the config directly with the adapter
				adapter := testconfig.NewAdapter(cfg)
				ctx = contextx.WithConfig(ctx, adapter)
			}

			// Create commit
			commit := domain.CommitInfo{
				Subject: subject,
				Body:    body,
			}

			// Create and configure rule with context
			baseRule := rules.NewCommitBodyRule()
			rule, ok := baseRule.WithContext(ctx).(rules.CommitBodyRule)
			require.True(t, ok, "Expected WithContext to return a CommitBodyRule")

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

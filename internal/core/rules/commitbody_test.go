// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

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
			expectedResult: "Valid commit body",
		},
		{
			name: "valid commit with body and sign-off",
			message: `Update documentation

Improve the getting started guide
Add more examples

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    false,
			expectedResult: "Valid commit body",
		},
		{
			name:           "commit without body",
			message:        "just a subject",
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "Invalid commit body",
		},
		{
			name: "commit without empty line between subject and body",
			message: `Update CI pipeline
Adding new stages for:
- Security scanning
- Performance testing
Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "Invalid commit body",
		},
		{
			name: "commit with empty line after subject but empty body",
			message: `Update CI pipeline

`,
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "Invalid commit body",
		},
		{
			name: "commit with only sign-off",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "Invalid commit body",
		},
		{
			name: "commit with multiple sign-off lines but no body",
			message: `Update dependencies

Signed-off-by: Laval Lion <laval@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			expectError:    true,
			errorCode:      string(appErrors.ErrInvalidBody),
			expectedResult: "Invalid commit body",
		},
		{
			name: "commit with only sign-off but configured to allow it",
			message: `Update config

Signed-off-by: Laval Lion <laval@cavora.org>`,
			options: []rules.CommitBodyOption{
				rules.WithRequireBody(true),
				rules.WithAllowSignOffOnly(true),
			},
			expectError:    false,
			expectedResult: "Valid commit body",
		},
		{
			name:           "body not required when configured",
			message:        "just a subject line",
			options:        []rules.CommitBodyOption{rules.WithRequireBody(false)},
			expectError:    false,
			expectedResult: "Valid commit body",
		},
		{
			name: "commit with empty body line but content after",
			message: `Fix typo in documentation

Fixed the typo in API documentation.`,
			expectError:    false,
			expectedResult: "Valid commit body",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit info object
			commit := domain.CommitInfo{
				Message: testCase.message,
			}

			// Always require body for test cases unless specified otherwise in options
			options := append([]rules.CommitBodyOption{rules.WithRequireBody(true)}, testCase.options...)

			// Test using value semantics
			rule := rules.NewCommitBodyRule(options...)
			errors := rule.Validate(commit)

			// Verify name is correct
			require.Equal(t, "CommitBody", rule.Name(), "Rule name should be 'CommitBody'")

			if testCase.expectError {
				require.NotEmpty(t, errors, "expected errors but got none")

				// Access ValidationError directly (no type assertion needed)
				require.GreaterOrEqual(t, len(errors), 1, "should have at least one error")
				valErr := errors[0]
				require.Equal(t, testCase.errorCode, valErr.Code, "unexpected error code")
				require.Equal(t, "CommitBody", valErr.Rule, "rule name should be set")
			} else {
				require.Empty(t, errors, "unexpected errors: %v", errors)
			}
		})
	}
}

func TestCommitBodyVerboseResult(t *testing.T) {
	testCases := []struct {
		name        string
		message     string
		expectError bool
	}{
		{
			name:        "missing body",
			message:     "subject only",
			expectError: true,
		},
		{
			name:        "missing blank line",
			message:     "subject\nbody without blank line",
			expectError: true,
		},
		{
			name:        "empty body",
			message:     "subject\n\n",
			expectError: true,
		},
		{
			name:        "signoff first line",
			message:     "subject\n\nSigned-off-by: Test <test@example.com>",
			expectError: true,
		},
		{
			name:        "only signoff lines",
			message:     "subject\n\nSigned-off-by: Test <test@example.com>\nSigned-off-by: Another <another@example.com>",
			expectError: true,
		},
		{
			name:        "valid commit",
			message:     "subject\n\nThis is a valid body.",
			expectError: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit info
			commit := domain.CommitInfo{
				Message: testCase.message,
			}

			// Use value semantics
			rule := rules.NewCommitBodyRule(rules.WithRequireBody(true))
			errors := rule.Validate(commit)

			if testCase.expectError {
				require.NotEmpty(t, errors, "expected validation errors for invalid commit")
			} else {
				require.Empty(t, errors, "expected no validation errors for valid commit")
			}
		})
	}
}

func TestCommitBodyHelpMethod(t *testing.T) {
	t.Run("valid and invalid commits", func(t *testing.T) {
		testCases := []struct {
			name        string
			message     string
			expectError bool
		}{
			{
				name:        "missing body",
				message:     "subject only",
				expectError: true,
			},
			{
				name:        "missing blank line",
				message:     "subject\nbody without blank line",
				expectError: true,
			},
			{
				name:        "empty body",
				message:     "subject\n\n",
				expectError: true,
			},
			{
				name:        "valid commit",
				message:     "subject\n\nThis is a valid body.",
				expectError: false,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// Create commit info
				commit := domain.CommitInfo{
					Message: testCase.message,
				}

				// Create rule with value semantics and validate
				rule := rules.NewCommitBodyRule(rules.WithRequireBody(true))
				errors := rule.Validate(commit)

				if testCase.expectError {
					require.NotEmpty(t, errors, "expected validation errors")
				} else {
					require.Empty(t, errors, "expected no validation errors")
				}
			})
		}
	})
}

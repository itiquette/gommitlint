// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
	"github.com/stretchr/testify/require"
)

func TestCommitBodyRule(t *testing.T) {
	tests := []struct {
		name             string
		message          string
		minLines         int
		minLength        int
		allowSignOffOnly bool
		expectError      bool
		errorCode        string
		description      string
	}{
		{
			name: "valid commit with body",
			message: `Add new validation rules

This commit adds new validation rules for:
- Password complexity
- Email format
- Username requirements`,
			expectError: false,
			description: "Valid commit with proper body content",
		},
		{
			name: "valid commit with body and sign-off",
			message: `Update documentation

Improve the getting started guide
Add more examples

Signed-off-by: Laval Lion <laval@cavora.org>`,
			minLines:    0,
			expectError: false,
			description: "Valid commit with body and sign-off",
		},
		{
			name:        "commit without body is allowed by default",
			message:     "Add new feature",
			expectError: false,
			description: "Body is not required by default",
		},
		{
			name:        "commit without body when not required",
			message:     "Minor fix",
			minLength:   0,
			expectError: false,
			description: "Should pass when body is not required",
		},
		{
			name: "commit with only sign-off is allowed by default",
			message: `Update configuration

Signed-off-by: Laval Lion <laval@cavora.org>`,
			allowSignOffOnly: false,
			expectError:      true,
			errorCode:        string(domain.ErrMissingBody),
			description:      "Sign-off only is NOT allowed when allowSignOffOnly is false",
		},
		{
			name: "commit with only sign-off when allowed",
			message: `Fix typo

Signed-off-by: Laval Lion <laval@cavora.org>`,
			allowSignOffOnly: true,
			expectError:      false,
			description:      "Should pass when only sign-off is present and allowed",
		},
		{
			name: "commit with too few lines",
			message: `Add feature

Short description`,
			minLines:    3,
			expectError: true,
			errorCode:   string(domain.ErrBodyTooShort),
			description: "Should fail when body has fewer lines than required",
		},
		{
			name: "commit with minimum lines",
			message: `Add new rules

This adds validation rules
for password complexity
and email format checks`,
			minLines:    3,
			expectError: false,
			description: "Should pass when body meets minimum line requirement",
		},
		{
			name: "commit body too short in length",
			message: `Fix bug

X`,
			minLength:   10,
			expectError: true,
			errorCode:   string(domain.ErrBodyTooShort),
			description: "Should fail when body is shorter than minimum length",
		},
		{
			name: "commit body meets minimum length",
			message: `Fix bug

This is a longer description that meets the minimum character requirement`,
			minLength:   10,
			expectError: false,
			description: "Should pass when body meets minimum length requirement",
		},
		// Additional test cases for edge scenarios
		{
			name: "empty body with blank lines is allowed",
			message: `Fix issue


`,
			expectError: false,
			description: "Empty body with blank lines is allowed by default",
		},
		{
			name: "body with trailing whitespace",
			message: `Update feature

This description has trailing spaces   
And some content here`,
			expectError: false,
			description: "Should pass with trailing whitespace",
		},
		{
			name: "body with multiple sign-offs",
			message: `Implement new feature

Add comprehensive validation logic

Signed-off-by: Developer One <dev1@example.com>
Signed-off-by: Developer Two <dev2@example.com>`,
			expectError: false,
			description: "Should pass with multiple sign-offs",
		},
		{
			name: "body exactly at minimum lines",
			message: `Add feature

Line one
Line two
Line three`,
			minLines:    3,
			expectError: false,
			description: "Should pass when exactly at minimum lines",
		},
		{
			name: "body exactly at minimum length",
			message: `Fix bug

1234567890`,
			minLength:   10,
			expectError: false,
			description: "Should pass when exactly at minimum length",
		},
		// Test combination of constraints
		{
			name: "body meets both line and length requirements",
			message: `Major update

This is line one with sufficient length
This is line two with sufficient length  
This is line three with sufficient length`,
			minLines:    3,
			minLength:   20,
			expectError: false,
			description: "Should pass when meeting both line and length requirements",
		},
		{
			name: "body meets lines but not length",
			message: `Update

X
Y
Z`,
			minLines:    3,
			minLength:   20,
			expectError: true,
			errorCode:   string(domain.ErrBodyTooShort),
			description: "Should fail when meeting lines but not length requirement",
		},
		{
			name: "body meets length but not lines",
			message: `Update

This is a very long single line that exceeds the minimum character requirement`,
			minLines:    3,
			minLength:   20,
			expectError: true,
			errorCode:   string(domain.ErrBodyTooShort),
			description: "Should fail when meeting length but not lines requirement",
		},
		// Test complex body content
		{
			name:        "body with code blocks and formatting",
			message:     "Add new API endpoint\n\nThis commit adds a new REST API endpoint:\n\n```go\nfunc (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {\n    // Implementation here\n}\n```\n\nThe endpoint supports:\n- User creation\n- Input validation\n- Error handling",
			minLines:    5,
			expectError: false,
			description: "Should handle body with code blocks and formatting",
		},
		{
			name: "body with bullets and lists",
			message: `Refactor authentication system

Changes include:
• Updated password hashing
• Improved session management  
• Added multi-factor authentication

Technical details:
1. Migrated from MD5 to bcrypt
2. Extended session timeout
3. Integrated TOTP support`,
			minLines:    7,
			expectError: false,
			description: "Should handle body with bullets and numbered lists",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using factory
			commit := testdata.Commit(testCase.message)

			// Create config with options
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinLines:         testCase.minLines,
						MinLength:        testCase.minLength,
						AllowSignoffOnly: testCase.allowSignOffOnly,
					},
				},
			}

			rule := rules.NewCommitBodyRule(cfg)
			failures := rule.Validate(commit, cfg)

			// Check result
			if testCase.expectError {
				require.NotEmpty(t, failures, "Expected error but got none for case: %s", testCase.description)

				if testCase.errorCode != "" {
					testdata.AssertRuleFailure(t, failures[0], "CommitBody")
				}
			} else {
				require.Empty(t, failures, "Expected no error but got: %v for case: %s", failures, testCase.description)
			}

			// Verify rule name
			require.Equal(t, "CommitBody", rule.Name(), "Rule name should be 'CommitBody'")
		})
	}
}

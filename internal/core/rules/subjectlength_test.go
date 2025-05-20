// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// createTestSubjectLengthContext creates a context with the given max length setting.
func createTestSubjectLengthContext(maxLength int) context.Context {
	// Create a base config to adapt
	cfg := config.NewDefaultConfig()
	cfg.Message.Subject.MaxLength = maxLength
	cfg.Message.Subject.Case = "sentence"
	cfg.Message.Subject.RequireImperative = true
	cfg.Conventional.Types = []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}

	// Add to context using direct adapter pattern
	ctx := testcontext.CreateTestContext()
	adapter := infraConfig.NewAdapter(cfg)

	return contextx.WithConfig(ctx, adapter)
}

func TestSubjectLengthRule(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		maxLength   int
		expectError bool
	}{
		{
			name:        "Within default length",
			subject:     "Add new feature",
			maxLength:   72,
			expectError: false,
		},
		{
			name:        "Exactly at max length",
			subject:     strings.Repeat("a", 72),
			maxLength:   72,
			expectError: false,
		},
		{
			name:        "Exceeds max length",
			subject:     strings.Repeat("a", 73),
			maxLength:   72,
			expectError: true,
		},
		{
			name:        "Custom max length",
			subject:     "This is a commit message with exactly 50 character",
			maxLength:   50,
			expectError: false,
		},
		{
			name:        "Exceeds custom max length",
			subject:     "This is a commit message that exceeds the custom limit",
			maxLength:   50,
			expectError: true,
		},
		{
			name:        "Empty subject is allowed",
			subject:     "",
			maxLength:   50,
			expectError: false,
		},
		{
			name:        "UTF-8 characters",
			subject:     "Add emoji support: 🎉🎊😃🔥✨",
			maxLength:   50,
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a rule without options
			rule := rules.NewSubjectLengthRule()

			// Create context with configuration
			ctx := createTestSubjectLengthContext(testCase.maxLength)
			commit := domain.CommitInfo{Subject: testCase.subject}

			errors := rule.Validate(ctx, commit)

			if testCase.expectError {
				require.NotEmpty(t, errors, "Expected validation error")
				require.Len(t, errors, 1, "Expected exactly one error")
				require.Equal(t, "SubjectLength", errors[0].Rule, "Rule name should be SubjectLength")
				require.Equal(t, string(appErrors.ErrMaxLengthExceeded), errors[0].Code, "Error code should be ErrMaxLengthExceeded")
				require.Contains(t, errors[0].Error(), "too long", "Error message should contain 'too long'")

				// Check context values
				actualLengthStr, hasActualLength := errors[0].Context["actual"]
				require.True(t, hasActualLength, "Error should have actual context")

				actualLength, err := strconv.Atoi(actualLengthStr)
				require.NoError(t, err, "actual should be a valid integer")
				require.Equal(t, len(testCase.subject), actualLength, "actual should match subject length")

				maxLengthStr, hasMaxLength := errors[0].Context["max"]
				require.True(t, hasMaxLength, "Error should have max context")

				maxLength, err := strconv.Atoi(maxLengthStr)
				require.NoError(t, err, "max should be a valid integer")
				require.Equal(t, testCase.maxLength, maxLength, "max should match configured max length")

				subject, hasSubject := errors[0].Context["subject"]
				require.True(t, hasSubject, "Error should have subject context")
				require.Equal(t, testCase.subject, subject, "subject context should match original subject")

				// Help is now accessible through GetHelp() method
				help := errors[0].GetHelp()
				require.NotEmpty(t, help, "Error should have help text")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

func TestSubjectLengthRuleVerbosity(t *testing.T) {
	// Create a rule
	rule := rules.NewSubjectLengthRule()

	// Create a rule with verbosity
	verboseRule := rule.WithVerbosity("verbose")
	require.Equal(t, "SubjectLength", verboseRule.Name(), "Rule name should be unchanged")

	// Test with a long subject
	subject := strings.Repeat("a", 80)
	ctx := createTestSubjectLengthContext(50)
	commit := domain.CommitInfo{Subject: subject}

	errors := verboseRule.Validate(ctx, commit)
	require.NotEmpty(t, errors, "Should have validation errors")
}

func TestSubjectLengthRuleName(t *testing.T) {
	rule := rules.NewSubjectLengthRule()
	require.Equal(t, "SubjectLength", rule.Name())
}

func TestSubjectLengthRuleWithoutContext(t *testing.T) {
	rule := rules.NewSubjectLengthRule()

	// Create a commit with a long subject
	commit := domain.CommitInfo{
		Subject: strings.Repeat("a", 100),
	}

	// Create a context without configuration
	ctx := testcontext.CreateTestContext()

	// When no config is available, the rule should use default max length of 72
	errors := rule.Validate(ctx, commit)
	require.NotEmpty(t, errors, "Should return error when subject exceeds default max length")
	require.Len(t, errors, 1, "Should return exactly one error")
	require.Equal(t, string(appErrors.ErrMaxLengthExceeded), errors[0].Code, "Error code should be ErrMaxLengthExceeded")
	require.Contains(t, errors[0].Error(), "28 characters too long", "Error should indicate how many characters over the limit")
}

// Test that UTF-8 characters are counted correctly (by bytes, not runes).
func TestSubjectLengthRuleUTF8(t *testing.T) {
	rule := rules.NewSubjectLengthRule()

	// Create a subject with multi-byte UTF-8 characters
	// Each emoji is typically 4 bytes
	subject := "Test " + strings.Repeat("🎉", 20) // 5 + 20*4 = 85 bytes

	ctx := createTestSubjectLengthContext(72)
	commit := domain.CommitInfo{Subject: subject}

	errors := rule.Validate(ctx, commit)

	// Should fail because byte length exceeds 72
	require.NotEmpty(t, errors, "Expected validation error for UTF-8 string")

	// Verify the actual byte length
	actualLength := len(subject)
	expectedLength := 5 + 20*utf8.RuneLen('🎉')
	require.Equal(t, expectedLength, actualLength, "UTF-8 byte count should be correct")
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
	"github.com/stretchr/testify/require"
)

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
			// Create rule with configuration
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: testCase.maxLength,
					},
				},
			}
			if testCase.maxLength == 0 {
				cfg.Message.Subject.MaxLength = 72 // default
			}

			rule := rules.NewSubjectLengthRule(cfg)

			// Create commit using builder
			commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commit.Subject = testCase.subject

			failures := rule.Validate(commit, nil, &cfg)

			if testCase.expectError {
				require.NotEmpty(t, failures, "Expected validation failure")
				require.Len(t, failures, 1, "Expected exactly one failure")

				// Use assertion helper
				testdata.AssertRuleFailure(t, failures[0], "SubjectLength")
				require.Contains(t, failures[0].Message, "exceeds", "Message should contain 'exceeds'")

				// Help is accessible through Help field
				help := failures[0].Help
				require.NotEmpty(t, help, "Failure should have help text")
			} else {
				require.Empty(t, failures, "Expected no validation failures")
			}
		})
	}
}

func TestSubjectLengthRuleName(t *testing.T) {
	cfg := config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				MaxLength: 72,
			},
		},
	}
	rule := rules.NewSubjectLengthRule(cfg)
	require.Equal(t, "SubjectLength", rule.Name())
}

func TestSubjectLengthRuleWithoutContext(t *testing.T) {
	cfg := config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				MaxLength: 72, // default
			},
		},
	}
	rule := rules.NewSubjectLengthRule(cfg)

	// Create a commit with a long subject using builder
	commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
	commit.Subject = strings.Repeat("a", 100)

	// When no config is available, the rule should use default max length of 72
	failures := rule.Validate(commit, nil, &cfg)
	require.NotEmpty(t, failures, "Should return failure when subject exceeds default max length")
	require.Len(t, failures, 1, "Should return exactly one failure")
	testdata.AssertRuleFailure(t, failures[0], "SubjectLength")
	require.Contains(t, failures[0].Message, "exceeds 72 characters", "Message should indicate it exceeds the limit")
}

// Test that UTF-8 characters are counted correctly (by bytes, not runes).
func TestSubjectLengthRuleUTF8(t *testing.T) {
	// Create a subject with multi-byte UTF-8 characters
	// Each emoji is typically 4 bytes
	subject := "Test " + strings.Repeat("🎉", 20) // 5 + 20*4 = 85 bytes

	// Create rule with max length of 72
	cfg := config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				MaxLength: 72,
			},
		},
	}
	rule := rules.NewSubjectLengthRule(cfg)

	// Create commit using builder
	commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
	commit.Subject = subject

	// Validate
	failures := rule.Validate(commit, nil, &cfg)

	// Should fail because byte length exceeds 72
	require.NotEmpty(t, failures, "Expected validation failure for UTF-8 string")

	// Verify the actual byte length
	actualLength := len(subject)
	expectedLength := 5 + 20*utf8.RuneLen('🎉')
	require.Equal(t, expectedLength, actualLength, "UTF-8 byte count should be correct")
}

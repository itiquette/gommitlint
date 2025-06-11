// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
)

func TestText_ProgressiveDisclosure(t *testing.T) {
	// Create test report with multiple errors for one rule
	commit := domain.Commit{
		Hash:    "abc1234",
		Subject: "not conventional commit",
		Message: "not conventional commit",
	}

	errors := []domain.ValidationError{
		{
			Code:    "invalid_conventional_format",
			Message: "Commit message doesn't follow conventional format",
			Help:    "Use format: type(scope): description (e.g., 'feat: add login')",
			Context: map[string]string{
				"expected": "type(scope): description",
				"found":    "not conventional commit",
			},
		},
		{
			Code:    "invalid_case",
			Message: "First word 'not' should be in sentence case",
			Help:    "Change 'not' to sentence case",
			Context: map[string]string{
				"current_word":  "not",
				"expected_word": "Not",
			},
		},
	}

	ruleResults := []domain.RuleReport{
		{
			Name:    "TestRule",
			Status:  domain.StatusFailed,
			Message: "Rule failed",
			Errors:  errors,
		},
	}

	commitReport := domain.CommitReport{
		Commit:      commit,
		RuleResults: ruleResults,
	}

	report := domain.Report{
		Commits: []domain.CommitReport{commitReport},
		Summary: domain.ReportSummary{
			TotalCommits:  1,
			PassedCommits: 0,
			AllPassed:     false,
		},
	}

	tests := []struct {
		name            string
		options         TextOptions
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "Basic level - concise summary on same line",
			options: TextOptions{
				Verbose:  false,
				ShowHelp: false,
			},
			wantContains: []string{
				"✗ TestRule:",
				"wrong format",
				"wrong case",
				"FAIL: 0 of 1 rules passed",
			},
			wantNotContains: []string{
				"[invalid_conventional_format]",
				"Help:",
				"    ", // No indented details
			},
		},
		{
			name: "Verbose level - detailed breakdown without codes",
			options: TextOptions{
				Verbose:  true,
				ShowHelp: false,
			},
			wantContains: []string{
				"✗ TestRule:",
				"    Commit message doesn't follow conventional format",
				"    Expected: type(scope): description",
				"    Found: not conventional commit",
				"    First word 'not' should be in sentence case",
				"    Current Word: not",
				"    Expected Word: Not",
				"FAIL: 0 of 1 rules passed",
			},
			wantNotContains: []string{
				"[invalid_conventional_format]",
				"Help:",
			},
		},
		{
			name: "Very verbose level - codes and help text",
			options: TextOptions{
				Verbose:  true,
				ShowHelp: true,
			},
			wantContains: []string{
				"✗ TestRule:",
				"    [invalid_conventional_format] Commit message doesn't follow conventional format",
				"    Help: Use format: type(scope): description (e.g., 'feat: add login')",
				"    [invalid_case] First word 'not' should be in sentence case",
				"    Help: Change 'not' to sentence case",
				"FAIL: 0 of 1 rules passed",
			},
			wantNotContains: []string{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Text(report, testCase.options)

			for _, want := range testCase.wantContains {
				require.Contains(t, result, want, "Expected output to contain: %s", want)
			}

			for _, notWant := range testCase.wantNotContains {
				require.NotContains(t, result, notWant, "Expected output to NOT contain: %s", notWant)
			}
		})
	}
}

func TestText_SingleError(t *testing.T) {
	// Test single error formatting
	commit := domain.Commit{
		Hash:    "abc1234",
		Subject: "test commit",
		Message: "test commit",
	}

	singleError := []domain.ValidationError{
		{
			Code:    "missing_signoff",
			Message: "Missing sign-off",
			Help:    "Add 'Signed-off-by: Your Name <email@example.com>'",
		},
	}

	ruleResults := []domain.RuleReport{
		{
			Name:    "SignOff",
			Status:  domain.StatusFailed,
			Message: "Rule failed",
			Errors:  singleError,
		},
	}

	commitReport := domain.CommitReport{
		Commit:      commit,
		RuleResults: ruleResults,
	}

	report := domain.Report{
		Commits: []domain.CommitReport{commitReport},
		Summary: domain.ReportSummary{
			TotalCommits:  1,
			PassedCommits: 0,
			AllPassed:     false,
		},
	}

	// Basic level should show single error message directly
	basicOptions := TextOptions{Verbose: false, ShowHelp: false}
	result := Text(report, basicOptions)

	require.Contains(t, result, "✗ SignOff: Missing sign-off")
	require.NotContains(t, result, "[missing_signoff]")
	require.NotContains(t, result, "Help:")
}

func TestCreateErrorSummary(t *testing.T) {
	tests := []struct {
		name   string
		errors []domain.ValidationError
		want   string
	}{
		{
			name:   "Empty errors",
			errors: []domain.ValidationError{},
			want:   "Unknown error",
		},
		{
			name: "Single error",
			errors: []domain.ValidationError{
				{Code: "subject_too_long", Message: "Subject too long"},
			},
			want: "Subject too long",
		},
		{
			name: "Multiple different errors",
			errors: []domain.ValidationError{
				{Code: "subject_too_long", Message: "Subject too long"},
				{Code: "invalid_case", Message: "Wrong case"},
			},
			want: "too long, wrong case",
		},
		{
			name: "Multiple same type errors",
			errors: []domain.ValidationError{
				{Code: "misspelled_word", Message: "Misspelled: word1"},
				{Code: "misspelled_word", Message: "Misspelled: word2"},
			},
			want: "misspelled (2 issues)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := createErrorSummary(testCase.errors)

			// Check that all expected parts are present (order may vary for maps)
			if strings.Contains(testCase.want, ",") {
				parts := strings.Split(testCase.want, ", ")
				for _, part := range parts {
					require.Contains(t, result, strings.TrimSpace(part))
				}
			} else {
				require.Equal(t, testCase.want, result)
			}
		})
	}
}

func TestExtractErrorType(t *testing.T) {
	tests := []struct {
		name  string
		error domain.ValidationError
		want  string
	}{
		{
			name:  "Known error code",
			error: domain.ValidationError{Code: "subject_too_long", Message: "Subject is too long"},
			want:  "too long",
		},
		{
			name:  "Unknown error code - fallback to message",
			error: domain.ValidationError{Code: "unknown_code", Message: "Some custom error message"},
			want:  "Some custom",
		},
		{
			name:  "Short message fallback",
			error: domain.ValidationError{Code: "unknown_code", Message: "Short"},
			want:  "Short",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := extractErrorType(testCase.error)
			require.Equal(t, testCase.want, result)
		})
	}
}

func TestFormatContextKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "snake_case",
			input:    "current_word",
			expected: "Current Word",
		},
		{
			name:     "single word",
			input:    "expected",
			expected: "Expected",
		},
		{
			name:     "multiple underscores",
			input:    "expected_case_format",
			expected: "Expected Case Format",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := formatContextKey(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

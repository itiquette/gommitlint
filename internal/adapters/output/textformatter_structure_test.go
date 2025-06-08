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

// - Proper spacing and visual hierarchy throughout.
func TestTextFormatter_StructuralIntegrity(t *testing.T) {
	testReport := createTestReport()

	tests := []struct {
		name            string
		options         TextOptions
		expectHelp      bool
		expectDividers  int
		expectColorCode bool
	}{
		{
			name:            "verbose mode structure",
			options:         TextOptions{Verbose: true, VerboseLevel: 1, UseColor: false},
			expectHelp:      false,
			expectDividers:  3,
			expectColorCode: false,
		},
		{
			name:            "extra verbose mode structure",
			options:         TextOptions{Verbose: true, VerboseLevel: 2, UseColor: false},
			expectHelp:      true,
			expectDividers:  3,
			expectColorCode: false,
		},
		{
			name:            "colored dividers",
			options:         TextOptions{Verbose: true, VerboseLevel: 1, UseColor: true},
			expectHelp:      false,
			expectDividers:  3,
			expectColorCode: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Text(testReport, testCase.options)

			// Verify core structural requirements
			assertStructuralIntegrity(t, result, testCase.expectHelp, testCase.expectDividers, testCase.expectColorCode)
		})
	}
}

// TestTextFormatter_EdgeCases validates structure in edge case scenarios.
func TestTextFormatter_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		report          domain.Report
		expectDividers  int
		expectRuleCount int
	}{
		{
			name:            "single error",
			report:          createSingleErrorReport(),
			expectDividers:  1,
			expectRuleCount: 1,
		},
		{
			name:            "multiple rules",
			report:          createMultipleRulesReport(),
			expectDividers:  3, // 2 errors in first rule + 1 error in second rule
			expectRuleCount: 2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			options := TextOptions{Verbose: true, VerboseLevel: 1, UseColor: false}
			result := Text(testCase.report, options)

			// Count dividers and rules
			dividerCount := strings.Count(result, dividerPattern)
			require.Equal(t, testCase.expectDividers, dividerCount,
				"Must have exactly %d dividers", testCase.expectDividers)

			ruleCount := strings.Count(result, "✗ ")
			require.Equal(t, testCase.expectRuleCount, ruleCount,
				"Must have exactly %d failed rules", testCase.expectRuleCount)
		})
	}
}

// assertStructuralIntegrity validates the core formatting structure.
func assertStructuralIntegrity(t *testing.T, result string, expectHelp bool, expectDividers int, expectColorCode bool) {
	t.Helper()

	// Rule structure: empty row after rule name before first error
	require.Contains(t, result, "✗ TestRule:\n\n    Error Code: first_error",
		"CRITICAL: Empty row must exist after rule name")

	// Error structure: structured format for all errors
	require.Contains(t, result, "    Error Code: first_error", "Must use structured error format")
	require.Contains(t, result, "    Error Message: First error message", "Must include error message")
	require.Contains(t, result, "    Expected: expected1", "Must include context fields")
	require.Contains(t, result, "    Actual: value1", "Must include actual values")

	// Help text behavior
	if expectHelp {
		require.Contains(t, result, "    Help:", "Help section must be present in extra verbose mode")
		require.Contains(t, result, "First error help text", "Help content must be visible")
	} else {
		require.NotContains(t, result, "    Help:", "Help section must NOT be present in verbose mode")
	}

	// Divider validation
	if expectColorCode {
		coloredDividers := strings.Count(result, lightOrangeANSI214+dividerPattern+resetANSI)
		require.Equal(t, expectDividers, coloredDividers, "Must have colored dividers when colors enabled")
	} else {
		plainDividers := strings.Count(result, dividerPattern)
		colorCodes := strings.Count(result, lightOrangeANSI214)

		require.Equal(t, expectDividers, plainDividers, "Must have plain dividers when colors disabled")
		require.Equal(t, 0, colorCodes, "Must have no color codes when colors disabled")
	}

	// Spacing validation: empty rows between error sections
	dividerToCheck := dividerPattern
	if expectColorCode {
		dividerToCheck = lightOrangeANSI214 + dividerPattern + resetANSI
	}

	require.Contains(t, result, dividerToCheck+"\n\n    Error Code: second_error",
		"CRITICAL: Empty row must exist between error sections")
}

// Test data factories

func createTestReport() domain.Report {
	return domain.Report{
		Commits: []domain.CommitReport{{
			Commit: domain.Commit{
				Hash:    "abc1234",
				Subject: "Test commit",
				Message: "Test commit",
			},
			RuleResults: []domain.RuleReport{{
				Name:   "TestRule",
				Status: domain.StatusFailed,
				Errors: []domain.ValidationError{
					{
						Code:    "first_error",
						Message: "First error message",
						Help:    "First error help text",
						Context: map[string]string{"actual": "value1", "expected": "expected1"},
					},
					{
						Code:    "second_error",
						Message: "Second error message",
						Help:    "Second error help text",
						Context: map[string]string{"first_word": "actual2", "expected_word": "expected2"},
					},
					{
						Code:    "third_error",
						Message: "Third error message",
						Help:    "Third error help text",
						Context: map[string]string{"actual": "value3", "expected": "expected3"},
					},
				},
			}},
		}},
		Summary: domain.ReportSummary{TotalCommits: 1, AllPassed: false},
	}
}

func createSingleErrorReport() domain.Report {
	return domain.Report{
		Commits: []domain.CommitReport{{
			Commit: domain.Commit{Hash: "abc1234", Subject: "Test commit"},
			RuleResults: []domain.RuleReport{{
				Name:   "SingleRule",
				Status: domain.StatusFailed,
				Errors: []domain.ValidationError{{
					Code:    "single_error",
					Message: "Single error message",
					Help:    "Single error help",
					Context: map[string]string{"actual": "wrong", "expected": "right"},
				}},
			}},
		}},
		Summary: domain.ReportSummary{TotalCommits: 1, AllPassed: false},
	}
}

func createMultipleRulesReport() domain.Report {
	return domain.Report{
		Commits: []domain.CommitReport{{
			Commit: domain.Commit{Hash: "abc1234", Subject: "Test commit"},
			RuleResults: []domain.RuleReport{
				{
					Name:   "FirstRule",
					Status: domain.StatusFailed,
					Errors: []domain.ValidationError{
						{Code: "error_a", Message: "Error A", Help: "Help A"},
						{Code: "error_b", Message: "Error B", Help: "Help B"},
					},
				},
				{
					Name:   "SecondRule",
					Status: domain.StatusFailed,
					Errors: []domain.ValidationError{
						{Code: "error_c", Message: "Error C", Help: "Help C"},
					},
				},
			},
		}},
		Summary: domain.ReportSummary{TotalCommits: 1, AllPassed: false},
	}
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package format

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
)

// SimpleFormatter formats validation results.
// Much simpler than the current UnifiedFormatter.
type SimpleFormatter struct {
	format string // "text", "json", "github", "gitlab"
}

// NewSimpleFormatter creates a formatter for the given format.
func NewSimpleFormatter(format string) SimpleFormatter {
	return SimpleFormatter{format: format}
}

// Format converts validation results to string output.
func (f SimpleFormatter) Format(results []domain.ValidationResult) string {
	switch f.format {
	case "json":
		return f.formatJSON(results)
	case "github":
		return f.formatGitHub(results)
	case "gitlab":
		return f.formatGitLab(results)
	default:
		return f.formatText(results)
	}
}

// FormatSingle formats a single validation result.
func (f SimpleFormatter) FormatSingle(result domain.ValidationResult) string {
	return f.Format([]domain.ValidationResult{result})
}

func (f SimpleFormatter) formatText(results []domain.ValidationResult) string {
	var builder strings.Builder

	// Summary
	total := len(results)
	passed := 0

	for _, r := range results {
		if r.Passed() {
			passed++
		}
	}

	if total > 1 {
		fmt.Fprintf(&builder, "Validated %d commits: %d passed, %d failed\n\n", total, passed, total-passed)
	}

	// Individual results
	for _, result := range results {
		if result.Passed() {
			fmt.Fprintf(&builder, "✓ %s\n", result.Commit.Subject)
		} else {
			fmt.Fprintf(&builder, "✗ %s\n", result.Commit.Subject)

			for _, failure := range result.Failures {
				fmt.Fprintf(&builder, "  %s: %s\n", failure.Rule, failure.Message)

				if failure.Help != "" {
					fmt.Fprintf(&builder, "    → %s\n", failure.Help)
				}
			}
		}

		fmt.Fprintln(&builder)
	}

	return builder.String()
}

func (f SimpleFormatter) formatJSON(results []domain.ValidationResult) string {
	return ConvertValidationResultsToJSON(results)
}

func (f SimpleFormatter) formatGitHub(results []domain.ValidationResult) string {
	var builder strings.Builder

	for _, result := range results {
		if !result.Passed() {
			for _, failure := range result.Failures {
				fmt.Fprintf(&builder, "::error file=%s,title=%s::%s\n",
					result.Commit.Hash, failure.Rule, failure.Message)
			}
		}
	}

	return builder.String()
}

func (f SimpleFormatter) formatGitLab(results []domain.ValidationResult) string {
	var builder strings.Builder

	for _, result := range results {
		if !result.Passed() {
			for _, failure := range result.Failures {
				fmt.Fprintf(&builder, "ERROR: %s - %s: %s\n",
					result.Commit.Hash[:7], failure.Rule, failure.Message)
			}
		}
	}

	return builder.String()
}

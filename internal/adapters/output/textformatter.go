// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/domain"
)

// TextOptions contains configuration for text formatting.
type TextOptions struct {
	Verbose   bool
	ShowHelp  bool
	LightMode bool
}

// Text formats a domain report as plain text with colors (pure function).
func Text(report domain.Report, options TextOptions) string {
	var builder strings.Builder

	colors := getColorScheme(options.LightMode)

	// Summary for multiple commits
	if len(report.Commits) > 1 {
		if report.Summary.AllPassed {
			builder.WriteString(colors.Success(fmt.Sprintf("SUCCESS: All %d commits passed validation\n\n", report.Summary.TotalCommits)))
		} else {
			builder.WriteString(colors.Warning(fmt.Sprintf("SUMMARY: %d of %d commits passed validation\n", report.Summary.PassedCommits, report.Summary.TotalCommits)))
			writeFailedRulesSummary(&builder, report.Summary, colors)
			builder.WriteString("\n")
		}
	}

	// Format each commit
	for i, commitReport := range report.Commits {
		writeCommitHeader(&builder, commitReport, i, len(report.Commits), colors)
		writeCommitRules(&builder, commitReport, colors, options.Verbose)
	}

	// Format repository-level results at the end
	if len(report.Repository.RuleResults) > 0 {
		writeRepositoryResults(&builder, report.Repository.RuleResults, colors, options.ShowHelp)
	}

	return builder.String()
}

// getColorScheme returns the appropriate color scheme based on light mode setting.
func getColorScheme(lightMode bool) colorScheme {
	if lightMode {
		return colorScheme{
			Success: color.New(color.FgGreen).SprintFunc(),
			Warning: color.New(color.FgRed).SprintFunc(),
			Error:   color.New(color.FgRed, color.Bold).SprintFunc(),
			Header:  color.New(color.FgBlue, color.Bold).SprintFunc(),
			Bold:    color.New(color.Bold).SprintFunc(),
			Muted:   color.New(color.FgBlack).SprintFunc(),
		}
	}

	return colorScheme{
		Success: color.New(color.FgGreen).SprintFunc(),
		Warning: color.New(color.FgYellow).SprintFunc(),
		Error:   color.New(color.FgRed, color.Bold).SprintFunc(),
		Header:  color.New(color.FgCyan, color.Bold).SprintFunc(),
		Bold:    color.New(color.Bold).SprintFunc(),
		Muted:   color.New(color.FgHiBlack).SprintFunc(),
	}
}

func writeFailedRulesSummary(builder *strings.Builder, summary domain.ReportSummary, colors colorScheme) {
	if len(summary.FailedRules) == 0 {
		return
	}

	builder.WriteString(colors.Bold("Rule failures summary:\n"))

	// Sort rule names
	ruleNames := make([]string, 0, len(summary.FailedRules))
	for ruleName := range summary.FailedRules {
		ruleNames = append(ruleNames, ruleName)
	}

	sort.Strings(ruleNames)

	for _, ruleName := range ruleNames {
		count := summary.FailedRules[ruleName]
		builder.WriteString(fmt.Sprintf("  - %s: %d failure(s)\n", colors.Bold(ruleName), count))
	}
}

func writeCommitHeader(builder *strings.Builder, commitReport domain.CommitReport, index, totalCommits int, colors colorScheme) {
	if commitReport.Commit.Hash == "" {
		return
	}

	if totalCommits > 1 {
		builder.WriteString(colors.Header(fmt.Sprintf("COMMIT #%d:\n", index+1)))
	}

	divider := strings.Repeat("=", 80)
	builder.WriteString(colors.Header(divider) + "\n")

	shortSHA := commitReport.Commit.Hash
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}

	builder.WriteString(fmt.Sprintf("%s %s\n", colors.Header("COMMIT-SHA:"), colors.Bold(shortSHA)))
	builder.WriteString(fmt.Sprintf("%s %s\n", colors.Header("SUBJECT:"), commitReport.Commit.Subject))

	if commitReport.Commit.Message != "" {
		parts := strings.SplitN(commitReport.Commit.Message, "\n", 2)
		if len(parts) > 1 && parts[1] != "" {
			builder.WriteString(fmt.Sprintf("%s\n%s\n", colors.Header("MESSAGE:"), parts[1]))
		}
	}

	builder.WriteString(colors.Header(divider) + "\n\n")
}

func writeCommitRules(builder *strings.Builder, commitReport domain.CommitReport, colors colorScheme, verbose bool) {
	// Sort rules alphabetically
	sortedRules := make([]domain.RuleReport, len(commitReport.RuleResults))
	copy(sortedRules, commitReport.RuleResults)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Name < sortedRules[j].Name
	})

	passedCount := 0

	for _, ruleReport := range sortedRules {
		if ruleReport.Status == domain.StatusPassed {
			passedCount++
			symbol := colors.Success("✓")
			builder.WriteString(fmt.Sprintf("%s %s: %s\n", symbol, colors.Bold(ruleReport.Name), ruleReport.Message))
		} else if ruleReport.Status == domain.StatusFailed {
			symbol := colors.Error("✗")
			builder.WriteString(fmt.Sprintf("%s %s: \n", symbol, colors.Bold(ruleReport.Name)))

			for _, err := range ruleReport.Errors {
				message := err.Message
				if verbose && err.Code != "" {
					message = fmt.Sprintf("[%s] %s", err.Code, message)
				}

				builder.WriteString(fmt.Sprintf("    %s\n", message))
			}

			builder.WriteString("\n")
		}
	}

	// Summary line
	totalRules := len(sortedRules)
	if passedCount == totalRules {
		builder.WriteString(colors.Success(fmt.Sprintf("\nPASS: All %d rules passed\n\n", totalRules)))
	} else {
		builder.WriteString(colors.Warning(fmt.Sprintf("\nFAIL: %d of %d rules passed\n\n", passedCount, totalRules)))
	}
}

func writeRepositoryResults(builder *strings.Builder, repoResults []domain.RuleReport, colors colorScheme, showHelp bool) {
	builder.WriteString("\n")
	builder.WriteString("================================================================================\n")
	builder.WriteString("REPOSITORY VALIDATION\n")
	builder.WriteString("================================================================================\n")
	builder.WriteString("\n")

	for _, ruleReport := range repoResults {
		var icon string

		if ruleReport.Status == domain.StatusFailed {
			icon = colors.Error("✗")
		} else {
			icon = colors.Success("✓")
		}

		builder.WriteString(fmt.Sprintf("%s %s: ", icon, ruleReport.Name))

		if ruleReport.Status == domain.StatusFailed {
			builder.WriteString("\n")

			for _, err := range ruleReport.Errors {
				builder.WriteString(fmt.Sprintf("    %s\n", err.Message))
			}
		} else {
			builder.WriteString(ruleReport.Message + "\n")
		}

		if showHelp && len(ruleReport.Errors) > 0 && ruleReport.Errors[0].Help != "" {
			builder.WriteString(fmt.Sprintf("    Help: %s\n", ruleReport.Errors[0].Help))
		}

		builder.WriteString("\n")
	}
}

// colorScheme provides color functions for different output elements.
type colorScheme struct {
	Success func(a ...interface{}) string
	Warning func(a ...interface{}) string
	Error   func(a ...interface{}) string
	Header  func(a ...interface{}) string
	Bold    func(a ...interface{}) string
	Muted   func(a ...interface{}) string
}

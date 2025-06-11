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
	Verbose      bool
	ShowHelp     bool
	ShowRuleHelp bool
	RuleHelpName string
	UseColor     bool
}

// Text formats a domain report as plain text with colors (pure function).
func Text(report domain.Report, options TextOptions) string {
	var builder strings.Builder

	colors := getColorScheme(options.UseColor)

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
		writeCommitRules(&builder, commitReport, colors, options)
	}

	// Format repository-level results at the end
	if len(report.Repository.RuleResults) > 0 {
		writeRepositoryResults(&builder, report.Repository.RuleResults, colors, options)
	}

	return builder.String()
}

// createErrorSummary creates a concise summary for multiple errors (pure function).
func createErrorSummary(errors []domain.ValidationError) string {
	if len(errors) == 0 {
		return "Unknown error"
	}

	if len(errors) == 1 {
		return errors[0].Message
	}

	// For multiple errors, create a brief summary
	errorTypes := make(map[string]int)

	for _, err := range errors {
		// Extract error type from code or message
		errorType := extractErrorType(err)
		errorTypes[errorType]++
	}

	// Build summary
	var parts []string

	for errorType, count := range errorTypes {
		if count == 1 {
			parts = append(parts, errorType)
		} else {
			parts = append(parts, fmt.Sprintf("%s (%d issues)", errorType, count))
		}
	}

	return strings.Join(parts, ", ")
}

// extractErrorType extracts a human-readable error type from a validation error.
func extractErrorType(err domain.ValidationError) string {
	switch err.Code {
	case "subject_too_long":
		return "too long"
	case "invalid_case":
		return "wrong case"
	case "invalid_suffix":
		return "invalid ending"
	case "missing_subject":
		return "missing subject"
	case "non_imperative":
		return "not imperative"
	case "invalid_conventional_format":
		return "wrong format"
	case "invalid_conventional_type":
		return "invalid type"
	case "missing_conventional_scope":
		return "missing scope"
	case "invalid_conventional_scope":
		return "invalid scope"
	case "conventional_desc_too_long":
		return "description too long"
	case "missing_body":
		return "missing body"
	case "body_too_short":
		return "body too short"
	case "invalid_body":
		return "invalid body"
	case "missing_jira":
		return "missing JIRA"
	case "invalid_project":
		return "invalid project"
	case "misplaced_jira":
		return "misplaced JIRA"
	case "missing_signoff":
		return "missing sign-off"
	case "missing_signature":
		return "missing signature"
	case "key_not_trusted":
		return "unauthorized author"
	case "misspelled_word":
		return "misspelled"
	case "too_many_commits":
		return "too many commits"
	default:
		// Fallback: use first few words of message
		words := strings.Fields(err.Message)
		if len(words) > 2 {
			return strings.Join(words[:2], " ")
		}

		return err.Message
	}
}

// getColorScheme returns the appropriate color scheme based on color usage setting.
func getColorScheme(useColor bool) colorScheme {
	if !useColor {
		// No color - return functions that return plain text
		noOp := func(a ...interface{}) string {
			return fmt.Sprint(a...)
		}

		return colorScheme{
			Success: noOp,
			Warning: noOp,
			Error:   noOp,
			Header:  noOp,
			Bold:    noOp,
			Muted:   noOp,
		}
	}

	// Universal bright colors that work on both dark and light backgrounds
	return colorScheme{
		Success: color.New(color.FgHiGreen).SprintFunc(),
		Warning: color.New(color.FgHiYellow).SprintFunc(),
		Error:   color.New(color.FgHiRed, color.Bold).SprintFunc(),
		Header:  color.New(color.FgHiCyan, color.Bold).SprintFunc(),
		Bold:    color.New(color.Bold).SprintFunc(),
		Muted:   color.New(color.Reset).SprintFunc(),
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

func writeCommitRules(builder *strings.Builder, commitReport domain.CommitReport, colors colorScheme, options TextOptions) {
	// Filter rules if specific rule help is requested
	var rulesToShow []domain.RuleReport

	if options.ShowRuleHelp && options.RuleHelpName != "" {
		// Show only the requested rule (exact match)
		for _, ruleReport := range commitReport.RuleResults {
			if ruleReport.Name == options.RuleHelpName {
				rulesToShow = append(rulesToShow, ruleReport)

				break
			}
		}
	} else {
		// Show all rules
		rulesToShow = commitReport.RuleResults
	}

	// Sort rules alphabetically
	sort.Slice(rulesToShow, func(i, j int) bool {
		return rulesToShow[i].Name < rulesToShow[j].Name
	})

	passedCount := 0

	for _, ruleReport := range rulesToShow {
		if ruleReport.Status == domain.StatusPassed {
			passedCount++
			symbol := colors.Success("✓")
			builder.WriteString(fmt.Sprintf("%s %s: %s\n", symbol, colors.Bold(ruleReport.Name), ruleReport.Message))
		} else if ruleReport.Status == domain.StatusFailed {
			symbol := colors.Error("✗")

			// Determine if we should show help for this specific case
			shouldShowHelp := options.ShowHelp || (options.ShowRuleHelp && options.RuleHelpName != "")

			if !options.Verbose && !shouldShowHelp {
				// Basic level: show concise summary on same line
				if len(ruleReport.Errors) == 1 {
					builder.WriteString(fmt.Sprintf("%s %s: %s\n", symbol, colors.Bold(ruleReport.Name), ruleReport.Errors[0].Message))
				} else {
					// Multiple errors: create summary
					errorSummary := createErrorSummary(ruleReport.Errors)
					builder.WriteString(fmt.Sprintf("%s %s: %s\n", symbol, colors.Bold(ruleReport.Name), errorSummary))
				}
			} else {
				// Verbose and very verbose: show details
				builder.WriteString(fmt.Sprintf("%s %s: \n", symbol, colors.Bold(ruleReport.Name)))

				for _, err := range ruleReport.Errors {
					message := err.Message

					// Show error codes only with -vv (showHelp) or specific rule help
					if shouldShowHelp && err.Code != "" {
						message = fmt.Sprintf("[%s] %s", err.Code, message)
					}

					builder.WriteString(fmt.Sprintf("    %s\n", message))

					// Show context with -v (verbose but not showHelp) unless specific rule help
					if options.Verbose && !shouldShowHelp && len(err.Context) > 0 {
						// Sort context keys for consistent output
						var keys []string
						for k := range err.Context {
							keys = append(keys, k)
						}

						sort.Strings(keys)

						for _, key := range keys {
							// Format context key nicely
							displayKey := formatContextKey(key)
							builder.WriteString(fmt.Sprintf("    %s: %s\n", displayKey, err.Context[key]))
						}
					}

					// Show help with -vv or specific rule help
					if shouldShowHelp && err.Help != "" {
						builder.WriteString(fmt.Sprintf("    Help: %s\n", err.Help))
					}
				}

				builder.WriteString("\n")
			}
		}
	}

	// Summary line - only show if we're showing all rules or if we found the specific rule
	if !options.ShowRuleHelp || len(rulesToShow) > 0 {
		totalRules := len(rulesToShow)
		if passedCount == totalRules {
			builder.WriteString(colors.Success(fmt.Sprintf("\nPASS: All %d rules passed\n\n", totalRules)))
		} else {
			builder.WriteString(colors.Warning(fmt.Sprintf("\nFAIL: %d of %d rules passed\n\n", passedCount, totalRules)))
		}
	} else if options.ShowRuleHelp {
		// Rule not found
		builder.WriteString(colors.Warning(fmt.Sprintf("\nRule '%s' not found in validation results\n\n", options.RuleHelpName)))
	}
}

func writeRepositoryResults(builder *strings.Builder, repoResults []domain.RuleReport, colors colorScheme, options TextOptions) {
	// Filter rules if specific rule help is requested
	var rulesToShow []domain.RuleReport

	if options.ShowRuleHelp && options.RuleHelpName != "" {
		// Show only the requested rule (exact match)
		for _, ruleReport := range repoResults {
			if ruleReport.Name == options.RuleHelpName {
				rulesToShow = append(rulesToShow, ruleReport)

				break
			}
		}
	} else {
		// Show all rules
		rulesToShow = repoResults
	}

	// Only show repository section if we have rules to display
	if len(rulesToShow) == 0 {
		return
	}

	builder.WriteString("\n")
	builder.WriteString("================================================================================\n")
	builder.WriteString("REPOSITORY VALIDATION\n")
	builder.WriteString("================================================================================\n")
	builder.WriteString("\n")

	for _, ruleReport := range rulesToShow {
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

		shouldShowHelp := options.ShowHelp || (options.ShowRuleHelp && options.RuleHelpName != "")
		if shouldShowHelp && len(ruleReport.Errors) > 0 && ruleReport.Errors[0].Help != "" {
			builder.WriteString(fmt.Sprintf("    Help: %s\n", ruleReport.Errors[0].Help))
		}

		builder.WriteString("\n")
	}
}

// formatContextKey formats a context key for display.
func formatContextKey(key string) string {
	// Convert snake_case to Title Case
	words := strings.Split(key, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[0:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
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

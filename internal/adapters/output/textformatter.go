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

// Constants for formatting consistency and maintainability.
const (
	dividerPattern     = "··················································································"
	lightOrangeANSI214 = "\033[38;5;214m"
	resetANSI          = "\033[0m"
	baseIndent         = "    "
	maxLineWidth       = 80
)

// TextOptions contains configuration for text formatting.
type TextOptions struct {
	Verbose      bool
	VerboseLevel int // 0=quiet, 1=verbose (-v), 2=extra verbose (-vv)
	ShowHelp     bool
	ShowRuleHelp bool
	RuleHelpName string
	UseColor     bool
}

// Text formats a domain report as plain text with colors (pure function).
func Text(report domain.Report, options TextOptions) string {
	var builder strings.Builder

	colors := getColorScheme(options.UseColor)

	// Format each commit
	for i, commitReport := range report.Commits {
		writeCommitHeader(&builder, commitReport, i, len(report.Commits), colors)
		writeCommitRules(&builder, commitReport, colors, options, nil) // Don't show repo rules per commit
	}

	// Repository rules after all commits
	if len(report.Repository.RuleResults) > 0 {
		writeRepositoryRules(&builder, report.Repository.RuleResults, colors, options)
	}

	// Summary for multiple commits - show at the end
	if len(report.Commits) > 1 {
		if report.Summary.AllPassed {
			builder.WriteString(colors.Success(fmt.Sprintf("SUCCESS: All %d commits passed validation\n\n", report.Summary.TotalCommits)))
		} else {
			builder.WriteString(colors.Warning(fmt.Sprintf("SUMMARY: %d of %d commits passed validation\n\n", report.Summary.PassedCommits, report.Summary.TotalCommits)))
			writeFailedRulesSummary(&builder, report.Summary, colors)
		}
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
// Uses the error message directly, taking the first few words for brevity.
func extractErrorType(err domain.ValidationError) string {
	words := strings.Fields(err.Message)
	if len(words) > 2 {
		return strings.Join(words[:2], " ")
	}

	return err.Message
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
	// Using high intensity colors for better readability
	return colorScheme{
		Success: color.New(color.FgHiGreen).SprintFunc(),
		Warning: color.New(color.FgHiYellow).SprintFunc(),
		Error:   color.New(color.FgHiRed, color.Bold).SprintFunc(),
		Header:  color.New(color.FgHiCyan, color.Bold).SprintFunc(),
		Bold:    color.New(color.Bold).SprintFunc(),
		Muted:   color.New(color.Reset).SprintFunc(), // Default terminal color
	}
}

func writeFailedRulesSummary(builder *strings.Builder, summary domain.ReportSummary, colors colorScheme) {
	if len(summary.FailedRules) == 0 {
		return
	}

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
	builder.WriteString(fmt.Sprintf("%s %s\n", colors.Header("DATE:"), commitReport.Commit.CommitDate))

	if commitReport.Commit.Message != "" {
		parts := strings.SplitN(commitReport.Commit.Message, "\n", 2)
		if len(parts) > 1 && parts[1] != "" {
			builder.WriteString(fmt.Sprintf("%s\n%s\n", colors.Header("MESSAGE:"), parts[1]))
		}
	}

	builder.WriteString(colors.Header(divider) + "\n\n")
}

func writeCommitRules(builder *strings.Builder, commitReport domain.CommitReport, colors colorScheme, options TextOptions, _ []domain.RuleReport) {
	// Only use commit rules
	commitRules := commitReport.RuleResults

	// Filter rules if specific rule help is requested
	var rulesToShow []domain.RuleReport

	if options.ShowRuleHelp && options.RuleHelpName != "" {
		// Show only the requested rule (exact match)
		for _, ruleReport := range commitRules {
			if ruleReport.Name == options.RuleHelpName {
				rulesToShow = append(rulesToShow, ruleReport)

				break
			}
		}
	} else {
		// Show all commit rules
		rulesToShow = commitRules
	}

	// Sort rules alphabetically
	sort.Slice(rulesToShow, func(i, j int) bool {
		return rulesToShow[i].Name < rulesToShow[j].Name
	})

	passedCount := writeRuleReports(builder, rulesToShow, colors, options)

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

// writeRepositoryRules writes repository-level validation results.
func writeRepositoryRules(builder *strings.Builder, repoResults []domain.RuleReport, colors colorScheme, options TextOptions) {
	divider := strings.Repeat("=", 80)
	builder.WriteString(colors.Header(divider) + "\n")
	builder.WriteString(colors.Header("REPOSITORY VALIDATION:") + "\n")
	builder.WriteString(colors.Header(divider) + "\n\n")

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
		// Show all repository rules
		rulesToShow = repoResults
	}

	// Sort rules alphabetically
	sort.Slice(rulesToShow, func(i, j int) bool {
		return rulesToShow[i].Name < rulesToShow[j].Name
	})

	passedCount := writeRuleReports(builder, rulesToShow, colors, options)

	// Repository summary line
	totalRules := len(rulesToShow)
	if passedCount == totalRules {
		builder.WriteString(colors.Success(fmt.Sprintf("\nPASS: All %d repository rules passed\n\n", totalRules)))
	} else {
		builder.WriteString(colors.Warning(fmt.Sprintf("\nFAIL: %d of %d repository rules passed\n\n", passedCount, totalRules)))
	}
}

// writeFormattedError writes a single error with appropriate formatting based on verbosity level.
func writeFormattedError(builder *strings.Builder, err domain.ValidationError, options TextOptions, showHelpText bool, colors colorScheme) {
	if options.VerboseLevel == 0 {
		// Default: just the message
		builder.WriteString(fmt.Sprintf("%s%s\n", baseIndent, err.Message))

		return
	}

	// Use structured format for verbose modes
	if err.Code != "" {
		builder.WriteString(fmt.Sprintf("%s%s %s\n",
			baseIndent,
			colors.Bold("Error Code:"),
			colors.Warning(err.Code)))
	}

	builder.WriteString(fmt.Sprintf("%s%s %s\n",
		baseIndent,
		colors.Bold("Error Message:"),
		err.Message))

	// Show context in structured format
	if len(err.Context) > 0 {
		// Define the order for key fields
		orderedKeys := []string{
			"first_word", "expected_word", // Case validation - first word pair
			"found", "expected", // Alternative Found/Expected pair
			"actual", "actual_length", "expected_length", // Generic actual/expected pair
		}

		// Write ordered fields first
		for _, key := range orderedKeys {
			if value, ok := err.Context[key]; ok {
				displayKey := formatTechnicalContextKey(key)
				formattedValue := formatContextValue(key, value, colors)
				builder.WriteString(fmt.Sprintf("%s%s %s\n", baseIndent, colors.Muted(displayKey+":"), formattedValue))
			}
		}

		// Write remaining context fields (excluding subject and ordered keys)
		for key, value := range err.Context {
			if !contains(orderedKeys, key) && key != "subject" {
				displayKey := formatContextKey(key)
				formattedValue := formatContextValue(key, value, colors)
				builder.WriteString(fmt.Sprintf("%s%s %s\n", baseIndent, colors.Muted(displayKey+":"), formattedValue))
			}
		}
	}

	// Show help with -vv or specific rule help
	if showHelpText && err.Help != "" {
		builder.WriteString(fmt.Sprintf("\n%s%s\n", baseIndent, colors.Bold("Help:")))
		writeHelpSection(builder, err.Help, colors)
	}
}

// formatTechnicalContextKey formats specific technical context keys with better names.
func formatTechnicalContextKey(key string) string {
	switch key {
	case "first_word":
		return "Actual"
	case "expected_word":
		return "Expected"
	case "actual_length":
		return "Actual Length"
	case "expected_length":
		return "Expected Length"
	case "commit_format":
		return "Commit Format"
	default:
		return formatContextKey(key)
	}
}

// formatContextValue applies appropriate colors to context values based on their key and content.
func formatContextValue(key, value string, colors colorScheme) string {
	switch key {
	case "first_word", "actual_case", "found", "actual", "actual_length":
		return colors.Warning(value)
	case "expected_word", "expected", "expected_length":
		return colors.Success(value)
	case "commit_format":
		if value == "conventional" {
			return colors.Success(value)
		}

		return colors.Warning(value)
	default:
		return value
	}
}

// writeHelpSection writes the help text with proper formatting and consistent indentation.
func writeHelpSection(builder *strings.Builder, helpText string, colors colorScheme) {
	lines := strings.Split(helpText, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			builder.WriteString("\n")
		} else if strings.HasPrefix(line, "✓") {
			// Success examples - highlight in green, wrap with consistent indentation
			wrappedLines := wrapTextWithIndent(line)
			for _, wrappedLine := range wrappedLines {
				builder.WriteString(fmt.Sprintf("%s%s\n", baseIndent, colors.Success(wrappedLine)))
			}
		} else if strings.HasPrefix(line, "•") {
			// Bullet points - use muted color, wrap with hanging indent
			wrappedLines := wrapTextWithIndent(line)
			for lineIndex, wrappedLine := range wrappedLines {
				if lineIndex == 0 {
					builder.WriteString(fmt.Sprintf("%s%s\n", baseIndent, colors.Muted(wrappedLine)))
				} else {
					// Continuation lines get extra indent to align with bullet content
					builder.WriteString(fmt.Sprintf("%s  %s\n", baseIndent, colors.Muted(wrappedLine)))
				}
			}
		} else if strings.HasSuffix(strings.TrimSpace(line), ":") && !strings.HasPrefix(line, " ") {
			// Section headers - make them bold, wrap with consistent indentation
			wrappedLines := wrapTextWithIndent(line)
			for _, wrappedLine := range wrappedLines {
				builder.WriteString(fmt.Sprintf("%s%s\n", baseIndent, colors.Bold(wrappedLine)))
			}
		} else {
			// Regular text - wrap with consistent indentation
			wrappedLines := wrapTextWithIndent(line)
			for _, wrappedLine := range wrappedLines {
				builder.WriteString(fmt.Sprintf("%s%s\n", baseIndent, wrappedLine))
			}
		}
	}
}

// wrapTextWithIndent wraps text to 80 characters while maintaining proper indentation.
func wrapTextWithIndent(text string) []string {
	const maxWidth = 80
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string

	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		// Check if adding this word would exceed the width
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}

		testLine += word

		if len(testLine) <= maxWidth {
			currentLine = testLine
		} else {
			// Current line is full, start a new one
			if currentLine != "" {
				lines = append(lines, currentLine)
			}

			currentLine = word
		}
	}

	// Add the last line if there's content
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// contains checks if a string slice contains a value.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
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

// writeRuleReports writes rule reports and returns the count of passed rules.
func writeRuleReports(builder *strings.Builder, rulesToShow []domain.RuleReport, colors colorScheme, options TextOptions) int {
	passedCount := 0

	for _, ruleReport := range rulesToShow {
		symbol := "✓"
		statusColor := colors.Success

		if len(ruleReport.Errors) > 0 {
			symbol = "✗"
			statusColor = colors.Error
		} else {
			passedCount++
		}

		// Rule name and status with colon format
		if len(ruleReport.Errors) == 0 {
			// Passed rule - just show checkmark and name
			builder.WriteString(fmt.Sprintf("%s %s\n", statusColor(symbol), colors.Bold(ruleReport.Name)))
		} else if len(ruleReport.Errors) == 1 && options.VerboseLevel == 0 {
			// Single error at basic level - show inline
			builder.WriteString(fmt.Sprintf("%s %s: %s\n", statusColor(symbol), colors.Bold(ruleReport.Name), ruleReport.Errors[0].Message))
		} else {
			// Multiple errors or verbose mode
			if options.VerboseLevel == 0 {
				// Basic level with multiple errors - show rule name with colon and errors without indentation
				builder.WriteString(fmt.Sprintf("%s %s:\n", statusColor(symbol), colors.Bold(ruleReport.Name)))

				for _, err := range ruleReport.Errors {
					// Basic level formatting - just the message without indentation
					builder.WriteString(err.Message + "\n")
				}
			} else {
				// Verbose and very verbose: show details
				builder.WriteString(fmt.Sprintf("%s %s:\n", statusColor(symbol), colors.Bold(ruleReport.Name)))

				for _, err := range ruleReport.Errors {
					// Add empty row before each error section (including the first one)
					builder.WriteString("\n")

					showHelpText := options.VerboseLevel >= 2 || options.ShowHelp || (options.ShowRuleHelp && options.RuleHelpName != "")
					writeFormattedError(builder, err, options, showHelpText, colors)

					// Add light dim orange divider after every error section in verbose modes
					if options.VerboseLevel >= 1 {
						writeDivider(builder, options.UseColor)
					}
				}

				builder.WriteString("\n")
			}
		}

		if len(ruleReport.Errors) > 0 {
			builder.WriteString("\n")
		}
	}

	return passedCount
}

// writeDivider writes a formatted divider with optional color support.
func writeDivider(builder *strings.Builder, useColor bool) {
	dividerText := dividerPattern
	if useColor {
		dividerText = lightOrangeANSI214 + dividerPattern + resetANSI
	}

	builder.WriteString(fmt.Sprintf("\n%s%s\n", baseIndent, dividerText))
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

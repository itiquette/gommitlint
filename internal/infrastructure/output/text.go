// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package output provides output formatting for validation results.
package output

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/domain"
)

// TextFormatter formats validation results as text.
type TextFormatter struct {
	verbose   bool
	showHelp  bool
	lightMode bool
}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter(verbose, showHelp bool, options ...bool) *TextFormatter {
	lightMode := false
	if len(options) > 0 {
		lightMode = options[0]
	}

	return &TextFormatter{
		verbose:   verbose,
		showHelp:  showHelp,
		lightMode: lightMode,
	}
}

// ColorScheme defines colors for different UI elements.
type ColorScheme struct {
	Success     func(format string, a ...interface{}) string
	Error       func(format string, a ...interface{}) string
	Warning     func(format string, a ...interface{}) string
	Info        func(format string, a ...interface{}) string
	VerboseInfo func(format string, a ...interface{}) string
	ErrorDetail func(format string, a ...interface{}) string
	HelpText    func(format string, a ...interface{}) string
	Bold        func(format string, a ...interface{}) string
	Italic      func(format string, a ...interface{}) string
	Header      func(format string, a ...interface{}) string
}

// Format formats validation results as text.
func (f *TextFormatter) Format(results *domain.ValidationResults) string {
	var builder strings.Builder

	// Check for NO_COLOR environment variable
	noColor := false
	if val, exists := os.LookupEnv("NO_COLOR"); exists && val != "" {
		noColor = true
		color.NoColor = true
	}

	// Set up color scheme based on mode preference and accessibility
	colorScheme := getColorScheme(f.lightMode, noColor)

	// Use Unicode symbols based on terminal capabilities
	passSymbol := colorScheme.Success("PASS")
	failSymbol := colorScheme.Error("FAIL")

	if canHandleUnicode() {
		passSymbol = colorScheme.Success("✓")
		failSymbol = colorScheme.Error("✗")
	}

	// For multiple commits, show a summary header first
	if len(results.CommitResults) > 1 {
		// Print overall summary
		if results.AllPassed() {
			builder.WriteString(colorScheme.Success("SUCCESS: "))
			builder.WriteString(fmt.Sprintf("All %d commits passed validation\n\n", results.TotalCommits))
		} else {
			builder.WriteString(colorScheme.Warning("SUMMARY: "))
			builder.WriteString(fmt.Sprintf("%d of %d commits passed validation\n",
				results.PassedCommits, results.TotalCommits))

			// Add rule summary if there are failures
			if len(results.RuleSummary) > 0 {
				builder.WriteString(colorScheme.Bold("Rule failures summary:\n"))

				// Sort rules by name for consistent output
				var ruleNames []string
				for ruleName := range results.RuleSummary {
					ruleNames = append(ruleNames, ruleName)
				}

				sort.Strings(ruleNames)

				for _, ruleName := range ruleNames {
					count := results.RuleSummary[ruleName]
					builder.WriteString(fmt.Sprintf("  - %s: %d failure(s)\n",
						colorScheme.Bold(ruleName), count))
				}
			}

			builder.WriteString("\n")
		}
	}

	// Process each commit result
	for i, commitResult := range results.CommitResults {
		// Print commit header with details
		if commitResult.CommitInfo != nil {
			// When we have multiple commits, show numbering
			if len(results.CommitResults) > 1 {
				builder.WriteString(colorScheme.Header(fmt.Sprintf("COMMIT #%d:\n", i+1)))
			}

			// Print commit divider
			divider := strings.Repeat("=", 80)
			builder.WriteString(colorScheme.Header(divider) + "\n")

			// Get the short SHA (first 7 characters)
			shortSHA := commitResult.CommitInfo.Hash
			if len(shortSHA) > 7 {
				shortSHA = shortSHA[:7]
			}

			// Print commit info header
			builder.WriteString(fmt.Sprintf("%s %s\n",
				colorScheme.Header("COMMIT-SHA:"),
				colorScheme.Bold(shortSHA)))

			// Print subject line
			builder.WriteString(fmt.Sprintf("%s %s\n",
				colorScheme.Header("SUBJECT:"),
				commitResult.CommitInfo.Subject))

			// Print full message if body exists and is not empty
			if commitResult.CommitInfo.Message != "" {
				// Extract body (everything after the first line)
				parts := strings.SplitN(commitResult.CommitInfo.Message, "\n", 2)
				if len(parts) > 1 && parts[1] != "" {
					builder.WriteString(fmt.Sprintf("%s\n%s\n",
						colorScheme.Header("MESSAGE:"),
						parts[1]))
				}
			}

			builder.WriteString(colorScheme.Header(divider) + "\n\n")
		}

		// Sort rule results alphabetically by name
		sortedRules := make([]domain.RuleResult, len(commitResult.RuleResults))
		copy(sortedRules, commitResult.RuleResults)
		sort.Slice(sortedRules, func(i, j int) bool {
			return strings.ToLower(sortedRules[i].RuleName) < strings.ToLower(sortedRules[j].RuleName)
		})

		// Print validation results for each rule
		passedRules := 0
		totalRules := 0

		for _, ruleResult := range sortedRules {
			// Skip rules with StatusSkipped status - these are disabled rules
			if ruleResult.Status == domain.StatusSkipped {
				continue
			}

			totalRules++
			ruleName := colorScheme.Bold(ruleResult.RuleName)

			if ruleResult.Status == domain.StatusPassed {
				// Success
				passedRules++

				builder.WriteString(fmt.Sprintf("%s %s: ", passSymbol, ruleName))

				// Print basic result message
				builder.WriteString(ruleResult.Message + "\n")

				// In verbose mode, add detailed information
				if f.verbose {
					builder.WriteString(fmt.Sprintf("    %s\n",
						colorScheme.VerboseInfo(ruleResult.VerboseMessage)))
				}
			} else if ruleResult.Status == domain.StatusFailed {
				// Error
				builder.WriteString(fmt.Sprintf("%s %s: ", failSymbol, ruleName))

				// Print basic error message in red
				builder.WriteString(colorScheme.Error(ruleResult.Message) + "\n")

				// In verbose mode or show help mode, add detailed error information
				if f.verbose {
					builder.WriteString(fmt.Sprintf("  %s\n",
						colorScheme.ErrorDetail(ruleResult.VerboseMessage)))
				}

				// Display validation errors
				if len(ruleResult.Errors) > 0 {
					for _, err := range ruleResult.Errors {
						builder.WriteString(fmt.Sprintf("    - %s\n",
							colorScheme.Error(err.Message)))

						// Show context if verbose
						if f.verbose && len(err.Context) > 0 {
							for key, value := range err.Context {
								builder.WriteString(fmt.Sprintf("      %s: %s\n",
									key, value))
							}
						}
					}
				}

				// If help mode, show full help text
				if f.showHelp && ruleResult.HelpMessage != "" {
					builder.WriteString("\n")

					// Properly indent and style all help text lines
					helpLines := strings.Split(ruleResult.HelpMessage, "\n")
					for _, line := range helpLines {
						builder.WriteString(fmt.Sprintf("  %s\n",
							colorScheme.HelpText(line)))
					}

					builder.WriteString("\n")
				} else if f.verbose {
					// In verbose mode, show a compact tip if we have a rule ID
					if ruleResult.RuleID != "" {
						builder.WriteString(fmt.Sprintf("    %s Run with '--rulehelp=%s' for detailed instructions\n\n",
							colorScheme.Warning("TIP:"), strings.ToLower(ruleResult.RuleID)))
					}
				}
			}
		}

		// Print summary line for this commit
		if totalRules > 0 {
			if passedRules == totalRules {
				builder.WriteString(fmt.Sprintf("\n%s All rules passed (%d/%d)\n",
					colorScheme.Success("SUCCESS:"), passedRules, totalRules))
			} else {
				builder.WriteString(fmt.Sprintf("\n%s %d of %d rules passed\n",
					colorScheme.Warning("FAIL:"), passedRules, totalRules))

				// Add a help hint if we're not in verbose mode
				if !f.verbose {
					builder.WriteString(colorScheme.Warning("TIP:") + " Run with '--verbose' for more details or '--rulehelp=<rule>' for specific rule help\n")
				}
			}
		} else {
			builder.WriteString(fmt.Sprintf("\n%s No active rules to evaluate\n",
				colorScheme.Warning("INFO:")))
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

// FormatRuleHelp formats help for a specific rule.
func (f *TextFormatter) FormatRuleHelp(ruleName string, results *domain.ValidationResults) string {
	var builder strings.Builder

	// Check for NO_COLOR environment variable
	noColor := false
	if val, exists := os.LookupEnv("NO_COLOR"); exists && val != "" {
		noColor = true
		color.NoColor = true
	}

	// Set up color scheme
	colorScheme := getColorScheme(f.lightMode, noColor)

	builder.WriteString(fmt.Sprintf("\n%s Rule Help:\n", colorScheme.Bold(ruleName)))

	foundHelp := false

	// Look through all rule results to find matching rule
	for _, commitResult := range results.CommitResults {
		for _, ruleResult := range commitResult.RuleResults {
			if strings.EqualFold(ruleResult.RuleID, ruleName) ||
				strings.EqualFold(ruleResult.RuleName, ruleName) {
				if ruleResult.HelpMessage != "" {
					// Properly indent and style all help text lines
					helpLines := strings.Split(ruleResult.HelpMessage, "\n")
					for _, line := range helpLines {
						builder.WriteString(fmt.Sprintf("  %s\n",
							colorScheme.HelpText(line)))
					}

					builder.WriteString("\n")

					foundHelp = true

					break
				}
			}
		}

		if foundHelp {
			break
		}
	}

	if !foundHelp {
		builder.WriteString(fmt.Sprintf("No help available for rule '%s'\n", ruleName))

		// Show available rules
		builder.WriteString("Available rules:\n")

		// Collect unique rule names
		ruleNames := make(map[string]bool)

		for _, commitResult := range results.CommitResults {
			for _, ruleResult := range commitResult.RuleResults {
				ruleNames[ruleResult.RuleName] = true
			}
		}

		// Sort rule names alphabetically
		var sortedNames []string
		for name := range ruleNames {
			sortedNames = append(sortedNames, name)
		}

		sort.Strings(sortedNames)

		// List available rules
		for _, name := range sortedNames {
			builder.WriteString(fmt.Sprintf("  - %s\n", colorScheme.Bold(name)))
		}
	}

	return builder.String()
}

// getColorScheme returns appropriate color functions based on mode and accessibility.
func getColorScheme(lightMode, noColor bool) ColorScheme {
	if noColor {
		// Return a scheme with no colors
		return ColorScheme{
			Success:     fmt.Sprintf,
			Error:       fmt.Sprintf,
			Warning:     fmt.Sprintf,
			Info:        fmt.Sprintf,
			VerboseInfo: fmt.Sprintf,
			ErrorDetail: fmt.Sprintf,
			HelpText:    fmt.Sprintf,
			Bold:        fmt.Sprintf,
			Italic:      fmt.Sprintf,
			Header:      fmt.Sprintf,
		}
	}

	// Keep standard green/red for pass/fail indicators
	successColor := color.New(color.FgGreen, color.Bold).SprintfFunc()
	errorColor := color.New(color.FgRed, color.Bold).SprintfFunc()
	regularErrorColor := color.New(color.FgRed).SprintfFunc()

	if lightMode {
		// Light mode colors (higher contrast on light backgrounds)
		return ColorScheme{
			Success:     successColor,                                        // Keep green for success
			Error:       errorColor,                                          // Keep red for errors
			Warning:     color.New(color.FgBlack, color.Bold).SprintfFunc(),  // Black for warnings
			Info:        color.New(color.FgBlue).SprintfFunc(),               // Blue for info
			VerboseInfo: color.New(color.FgBlue).SprintfFunc(),               // Blue for verbose success info
			ErrorDetail: regularErrorColor,                                   // Regular red for error details
			HelpText:    color.New(color.FgBlue, color.Italic).SprintfFunc(), // Blue italic for help text
			Bold:        color.New(color.Bold).SprintfFunc(),
			Italic:      color.New(color.Italic).SprintfFunc(),
			Header:      color.New(color.FgBlue, color.Bold).SprintfFunc(), // Blue bold for header
		}
	}

	// Dark mode colors (default - better for most terminal backgrounds)
	return ColorScheme{
		Success:     successColor,                                        // Keep green for success
		Error:       errorColor,                                          // Keep red for errors
		Warning:     color.New(color.FgYellow, color.Bold).SprintfFunc(), // Yellow for warnings
		Info:        color.New(color.FgCyan).SprintfFunc(),               // Cyan for info
		VerboseInfo: color.New(color.FgHiCyan).SprintfFunc(),             // Bright cyan for verbose success
		ErrorDetail: regularErrorColor,                                   // Regular red for error details
		HelpText:    color.New(color.FgCyan, color.Italic).SprintfFunc(), // Cyan italic for help text
		Bold:        color.New(color.Bold).SprintfFunc(),
		Italic:      color.New(color.Italic).SprintfFunc(),
		Header:      color.New(color.FgMagenta, color.Bold).SprintfFunc(), // Magenta bold for header
	}
}

// canHandleUnicode checks if the terminal supports Unicode characters.
func canHandleUnicode() bool {
	// Check common environment variables for UTF-8 support
	for _, envVar := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		value := strings.ToLower(os.Getenv(envVar))
		if strings.Contains(value, "utf-8") || strings.Contains(value, "utf8") {
			return true
		}
	}

	return false
}

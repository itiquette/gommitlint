// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package output provides output formatting for validation results.
package output

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
)

// TextFormatter formats validation results as text.
// It implements the domain.ResultFormatter interface.
type TextFormatter struct {
	verbose   bool
	showHelp  bool
	lightMode bool
	colors    ColorScheme
	symbols   struct {
		pass string
		fail string
	}
}

// Ensure TextFormatter implements domain.ResultFormatter.
var _ domain.ResultFormatter = TextFormatter{}

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

// NewTextFormatter creates a new text formatter.
// It implements domain.ResultFormatter interface.
// This implementation returns a value rather than a pointer.
func NewTextFormatter() TextFormatter {
	formatter := TextFormatter{
		verbose:   false,
		showHelp:  false,
		lightMode: false,
	}

	// Initialize the color scheme and symbols
	formatter = formatter.withInitializedColorScheme()
	formatter = formatter.withInitializedSymbols()

	return formatter
}

// withInitializedColorScheme creates a new formatter with initialized color scheme.
// This is a pure function that returns a new TextFormatter without modifying the original.
func (f TextFormatter) withInitializedColorScheme() TextFormatter {
	result := f

	// Check for NO_COLOR environment variable
	noColor := os.Getenv("NO_COLOR") != ""
	if noColor {
		color.NoColor = true
	}

	result.colors = getColorScheme(result.lightMode, noColor)

	return result
}

// withInitializedSymbols creates a new formatter with initialized symbols.
// This is a pure function that returns a new TextFormatter without modifying the original.
func (f TextFormatter) withInitializedSymbols() TextFormatter {
	result := f

	result.symbols.pass = result.colors.Success("PASS")
	result.symbols.fail = result.colors.Error("FAIL")

	if canHandleUnicode() {
		result.symbols.pass = result.colors.Success("✓")
		result.symbols.fail = result.colors.Error("✗")
	}

	return result
}

// WithVerbose sets the verbose flag for the formatter.
// This implementation returns a value rather than a pointer.
func (f TextFormatter) WithVerbose(verbose bool) TextFormatter {
	result := f
	result.verbose = verbose

	// Re-initialize the formatter with the new settings
	result = result.withInitializedColorScheme()
	result = result.withInitializedSymbols()

	return result
}

// WithShowHelp sets the showHelp flag for the formatter.
// This implementation returns a value rather than a pointer.
func (f TextFormatter) WithShowHelp(showHelp bool) TextFormatter {
	result := f
	result.showHelp = showHelp

	return result
}

// WithLightMode sets the lightMode flag for the formatter.
// This implementation returns a value rather than a pointer.
func (f TextFormatter) WithLightMode(lightMode bool) TextFormatter {
	result := f
	result.lightMode = lightMode

	// Re-initialize the formatter with the new settings
	result = result.withInitializedColorScheme()
	result = result.withInitializedSymbols()

	return result
}

// These methods are now replaced by withInitializedColorScheme and withInitializedSymbols
// which follow functional programming patterns by returning new instances
// instead of modifying state.

// Format formats validation results as text.
func (f TextFormatter) Format(results domain.ValidationResults) string {
	var builder strings.Builder

	// For multiple commits, show a summary header first
	if len(results.CommitResults) > 1 {
		f.formatOverallSummary(&builder, results)
	}

	// Process each commit result
	for i, commitResult := range results.CommitResults {
		f.formatCommitHeader(&builder, commitResult, i, len(results.CommitResults))
		f.formatRuleResults(&builder, commitResult)
	}

	return builder.String()
}

// formatOverallSummary formats the summary for multiple commits.
func (f TextFormatter) formatOverallSummary(builder *strings.Builder, results domain.ValidationResults) {
	if results.AllPassed() {
		builder.WriteString(f.colors.Success("SUCCESS: "))
		builder.WriteString(fmt.Sprintf("All %d commits passed validation\n\n", results.TotalCommits))

		return
	}

	builder.WriteString(f.colors.Warning("SUMMARY: "))
	builder.WriteString(fmt.Sprintf("%d of %d commits passed validation\n",
		results.PassedCommits, results.TotalCommits))

	f.formatRuleSummary(builder, results)
	builder.WriteString("\n")
}

// formatRuleSummary formats the summary of rule failures.
func (f TextFormatter) formatRuleSummary(builder *strings.Builder, results domain.ValidationResults) {
	if len(results.RuleSummary) == 0 {
		return
	}

	builder.WriteString(f.colors.Bold("Rule failures summary:\n"))

	// Sort rules by name for consistent output
	ruleNames := getSortedRuleNames(results.RuleSummary)

	for _, ruleName := range ruleNames {
		count := results.RuleSummary[ruleName]
		builder.WriteString(fmt.Sprintf("  - %s: %d failure(s)\n",
			f.colors.Bold(ruleName), count))
	}
}

// formatCommitHeader formats the commit header information.
func (f TextFormatter) formatCommitHeader(builder *strings.Builder,
	commitResult domain.CommitResult, index, totalCommits int) {
	if commitResult.CommitInfo.Hash == "" {
		return
	}

	// When we have multiple commits, show numbering
	if totalCommits > 1 {
		builder.WriteString(f.colors.Header(fmt.Sprintf("COMMIT #%d:\n", index+1)))
	}

	// Print commit divider
	divider := strings.Repeat("=", 80)
	builder.WriteString(f.colors.Header(divider) + "\n")

	// Get the short SHA (first 7 characters)
	shortSHA := commitResult.CommitInfo.Hash
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}

	// Print commit info
	builder.WriteString(fmt.Sprintf("%s %s\n",
		f.colors.Header("COMMIT-SHA:"),
		f.colors.Bold(shortSHA)))
	builder.WriteString(fmt.Sprintf("%s %s\n",
		f.colors.Header("SUBJECT:"),
		commitResult.CommitInfo.Subject))

	f.formatCommitMessage(builder, commitResult.CommitInfo.Message)
	builder.WriteString(f.colors.Header(divider) + "\n\n")
}

// formatCommitMessage formats the commit message body if it exists.
func (f TextFormatter) formatCommitMessage(builder *strings.Builder, message string) {
	if message == "" {
		return
	}

	// Extract body (everything after the first line)
	parts := strings.SplitN(message, "\n", 2)
	if len(parts) > 1 && parts[1] != "" {
		builder.WriteString(fmt.Sprintf("%s\n%s\n",
			f.colors.Header("MESSAGE:"),
			parts[1]))
	}
}

// formatRuleResults formats the rule validation results.
func (f TextFormatter) formatRuleResults(builder *strings.Builder, commitResult domain.CommitResult) {
	// Sort rule results alphabetically by name
	sortedRules := getSortedRuleResults(commitResult.RuleResults)

	// Print validation results for each rule
	passedRules := 0
	totalRules := 0

	for _, ruleResult := range sortedRules {
		// Skip rules with StatusSkipped status - these are disabled rules
		if ruleResult.Status == domain.StatusSkipped {
			continue
		}

		totalRules++
		ruleName := f.colors.Bold(ruleResult.RuleName)

		// Fix inconsistent message status for JiraReference and CommitsAhead rules
		if ruleResult.Status == domain.StatusFailed {
			// Handle specific cases where default message is misleading
			if ruleResult.RuleName == "JiraReference" && strings.Contains(ruleResult.Message, "Valid Jira reference") {
				// Override default message for failed JiraReference rule
				if len(ruleResult.Errors) > 0 && ruleResult.Errors[0].Message != "" {
					ruleResult.Message = "Missing Jira issue key"
				} else {
					ruleResult.Message = "Invalid Jira reference"
				}
			} else if ruleResult.RuleName == "CommitsAhead" && strings.Contains(ruleResult.Message, "HEAD is 0 commit") {
				// Extract correct ahead count from error if possible
				if len(ruleResult.Errors) > 0 {
					for _, err := range ruleResult.Errors {
						if ahead, exists := err.Context["commits_ahead"]; exists {
							if aheadCount, err := strconv.Atoi(ahead); err == nil {
								ruleResult.Message = fmt.Sprintf("Too many commits ahead of main (%d)", aheadCount)

								break
							}
						}
					}
				}
			}
		}

		if ruleResult.Status == domain.StatusPassed {
			passedRules++

			f.formatPassedRule(builder, ruleName, ruleResult)
		} else if ruleResult.Status == domain.StatusFailed {
			f.formatFailedRule(builder, ruleName, ruleResult)
		}
	}

	f.formatRuleSummaryLine(builder, passedRules, totalRules)
}

// formatPassedRule formats a passed rule result.
func (f TextFormatter) formatPassedRule(builder *strings.Builder, ruleName string, ruleResult domain.RuleResult) {
	builder.WriteString(fmt.Sprintf("%s %s: ", f.symbols.pass, ruleName))
	builder.WriteString(ruleResult.Message + "\n")

	// In verbose mode, add detailed information
	if f.verbose {
		builder.WriteString(fmt.Sprintf("    %s\n",
			f.colors.VerboseInfo(ruleResult.VerboseMessage)))
	}
}

// formatFailedRule formats a failed rule result.
func (f TextFormatter) formatFailedRule(builder *strings.Builder, ruleName string, ruleResult domain.RuleResult) {
	builder.WriteString(fmt.Sprintf("%s %s: ", f.symbols.fail, ruleName))
	builder.WriteString(f.colors.Error(ruleResult.Message) + "\n")

	// In verbose mode, add detailed error information
	if f.verbose {
		builder.WriteString(fmt.Sprintf("  %s\n",
			f.colors.ErrorDetail(ruleResult.VerboseMessage)))
	}

	f.formatRuleErrors(builder, ruleResult.Errors)

	// When showHelp is enabled, show the actual rule.Help() output instead of our hardcoded text
	if !f.showHelp {
		// Just let the code below handle formatting the help message
		// This allows the actual rule.Help() content to be used
		// ONLY use our hardcoded help when showHelp is NOT enabled (for backward compatibility)
		// Handle special rules differently - direct fix to show proper helpful messages
		if ruleName == "JiraReference" {
			builder.WriteString("\n")
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("Include a valid Jira issue key (e.g., PROJECT-123) in your commit subject.")))
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("For conventional commits, place the key at the end of the first line:")))
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("- feat(auth): add login feature PROJ-123")))
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("- fix: resolve timeout issue [PROJ-123]")))
			builder.WriteString("\n")

			return
		}

		if ruleName == "CommitsAhead" {
			builder.WriteString("\n")
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("Your branch is too far ahead of the reference branch.")))
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("To fix this, either:")))
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("1. Merge the reference branch into your branch")))
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("2. Rebase your branch onto the latest reference")))
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("3. Squash some commits to reduce the total count")))
			builder.WriteString("\n")

			return
		}
	}

	// This is the MAIN place where help text gets displayed with --extra-verbose flag
	if f.showHelp {
		// Always show the rule's Help() text when showHelp is true
		builder.WriteString("\n")

		// Always use the full HelpMessage content from the rule
		if ruleResult.HelpMessage != "" {
			// Properly indent and style all help text lines
			helpLines := strings.Split(ruleResult.HelpMessage, "\n")
			for _, line := range helpLines {
				builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText(line)))
			}
		} else {
			// Fallback if HelpMessage is empty (shouldn't happen)
			builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("Run with '--rulehelp="+ruleName+"' for specific guidance on this rule.")))
		}

		builder.WriteString("\n")

		return
	}

	// ONLY for non-showHelp mode (regular output without --extra-verbose)
	// For rules with "No errors to fix" in help message or empty help messages, show a suggestion
	if strings.Contains(ruleResult.HelpMessage, "No errors to fix") || ruleResult.HelpMessage == "" {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("Run with '--rulehelp="+ruleName+"' for specific guidance on this rule.")))
		builder.WriteString("\n")

		return
	}

	// Use standard help text formatting for other cases
	f.formatHelpText(builder, ruleResult)
}

// formatRuleErrors formats the validation errors for a rule.
func (f TextFormatter) formatRuleErrors(builder *strings.Builder, validationErrors []errors.ValidationError) {
	if len(validationErrors) == 0 {
		return
	}

	// Create a text formatter for validation errors
	errorFormatter := errors.NewTextFormatter(!f.verbose) // invert verbose setting: we want compact when not verbose

	for _, err := range validationErrors {
		// Use the enhanced error formatting if a help message is available
		// Use enhanced error formatting based on verbosity
		if !f.verbose {
			// Simple format for non-verbose mode
			builder.WriteString(fmt.Sprintf("    - %s\n",
				f.colors.Error(err.Message)))
		} else {
			// In verbose mode, use the enhanced formatter to get rich error display
			// Indent the text
			errorText := errorFormatter.FormatError(err)

			// Add indentation to each line
			lines := strings.Split(strings.TrimSpace(errorText), "\n")
			for _, line := range lines {
				builder.WriteString(fmt.Sprintf("    %s\n", line))
			}
		}
	}
}

// formatHelpText formats the help text for a failed rule.
func (f TextFormatter) formatHelpText(builder *strings.Builder, ruleResult domain.RuleResult) {
	// Standard behavior
	if f.showHelp {
		builder.WriteString("\n")

		// First, check for enhanced errors with help messages
		if len(ruleResult.Errors) > 0 {
			helpMessage := ""

			// Try to get help from the first error with a help message
			for _, err := range ruleResult.Errors {
				if help := err.GetHelp(); help != "" {
					helpMessage = help

					break
				}
			}

			// If we found a help message, use it
			if helpMessage != "" {
				// Properly indent and style all help text lines
				helpLines := strings.Split(helpMessage, "\n")
				for _, line := range helpLines {
					builder.WriteString(fmt.Sprintf("  %s\n",
						f.colors.HelpText(line)))
				}

				builder.WriteString("\n")

				return
			}
		}

		// Fallback to rule's help message if enhanced error doesn't have one
		if ruleResult.HelpMessage != "" {
			// Properly indent and style all help text lines
			helpLines := strings.Split(ruleResult.HelpMessage, "\n")
			for _, line := range helpLines {
				builder.WriteString(fmt.Sprintf("  %s\n",
					f.colors.HelpText(line)))
			}

			builder.WriteString("\n")
		}
	} else if f.verbose && ruleResult.RuleID != "" {
		// In verbose mode, show a compact tip if we have a rule ID
		builder.WriteString(fmt.Sprintf("    %s Run with '--rulehelp=%s' for detailed instructions\n\n",
			f.colors.Warning("TIP:"), strings.ToLower(ruleResult.RuleID)))
	}
}

// formatRuleSummaryLine formats the summary line for a commit.
func (f TextFormatter) formatRuleSummaryLine(builder *strings.Builder, passedRules, totalRules int) {
	if totalRules == 0 {
		builder.WriteString(fmt.Sprintf("\n%s No active rules to evaluate\n\n",
			f.colors.Warning("INFO:")))

		return
	}

	if passedRules == totalRules {
		builder.WriteString(fmt.Sprintf("\n%s All rules passed (%d/%d)\n\n",
			f.colors.Success("SUCCESS:"), passedRules, totalRules))

		return
	}

	builder.WriteString(fmt.Sprintf("\n%s %d of %d rules passed\n",
		f.colors.Warning("FAIL:"), passedRules, totalRules))

	// Add a help hint if we're not in verbose mode
	if !f.verbose {
		builder.WriteString(f.colors.Warning("TIP:") +
			" Run with '--verbose' for more details or '--rulehelp=<rule>' for specific rule help\n\n")
	} else {
		builder.WriteString("\n")
	}
}

// FormatRuleHelp formats help for a specific rule.
// This implementation ensures proper initialization following functional patterns.
func (f TextFormatter) FormatRuleHelp(ruleName string, results domain.ValidationResults) string {
	var builder strings.Builder

	// Ensure colors are initialized using functional pattern
	initializedFormatter := f.withInitializedColorScheme()

	builder.WriteString(fmt.Sprintf("\n%s Rule Help:\n", initializedFormatter.colors.Bold(ruleName)))

	if initializedFormatter.formatSpecificRuleHelp(&builder, ruleName, results) {
		return builder.String()
	}

	// If no help was found, show available rules
	builder.WriteString(fmt.Sprintf("No help available for rule '%s'\n", ruleName))
	initializedFormatter.formatAvailableRules(&builder, results)

	return builder.String()
}

// formatSpecificRuleHelp formats help for a specific rule if available.
func (f TextFormatter) formatSpecificRuleHelp(builder *strings.Builder, ruleName string, results domain.ValidationResults) bool {
	// Look through all rule results to find matching rule
	for _, commitResult := range results.CommitResults {
		for _, ruleResult := range commitResult.RuleResults {
			if !strings.EqualFold(ruleResult.RuleID, ruleName) &&
				!strings.EqualFold(ruleResult.RuleName, ruleName) {
				continue
			}

			if ruleResult.HelpMessage == "" {
				continue
			}

			// Properly indent and style all help text lines
			helpLines := strings.Split(ruleResult.HelpMessage, "\n")
			for _, line := range helpLines {
				builder.WriteString(fmt.Sprintf("  %s\n",
					f.colors.HelpText(line)))
			}

			builder.WriteString("\n")

			return true
		}
	}

	return false
}

// formatAvailableRules formats the list of available rules.
func (f TextFormatter) formatAvailableRules(builder *strings.Builder, results domain.ValidationResults) {
	builder.WriteString("Available rules:\n")

	// Collect unique rule names
	ruleNames := make(map[string]bool)

	for _, commitResult := range results.CommitResults {
		for _, ruleResult := range commitResult.RuleResults {
			ruleNames[ruleResult.RuleName] = true
		}
	}

	// Sort and list available rules
	sortedNames := make([]string, 0)
	for name := range ruleNames {
		sortedNames = append(sortedNames, name)
	}

	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		builder.WriteString(fmt.Sprintf("  - %s\n", f.colors.Bold(name)))
	}
}

// getSortedRuleNames returns a sorted slice of rule names from a map.
func getSortedRuleNames(ruleMap map[string]int) []string {
	ruleNames := make([]string, 0)
	for ruleName := range ruleMap {
		ruleNames = append(ruleNames, ruleName)
	}

	sort.Strings(ruleNames)

	return ruleNames
}

// getSortedRuleResults returns a sorted slice of rule results.
func getSortedRuleResults(ruleResults []domain.RuleResult) []domain.RuleResult {
	sortedRules := make([]domain.RuleResult, len(ruleResults))
	copy(sortedRules, ruleResults)
	sort.Slice(sortedRules, func(i, j int) bool {
		return strings.ToLower(sortedRules[i].RuleName) < strings.ToLower(sortedRules[j].RuleName)
	})

	return sortedRules
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

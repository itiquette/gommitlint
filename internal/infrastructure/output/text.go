// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package output provides output formatters for validation results.
package output

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// TextFormatter formats validation results as text.
type TextFormatter struct {
	verbose    bool
	showHelp   bool
	lightMode  bool
	colors     ColorScheme
	symbols    SymbolSet
	lineWidth  int
	maxDetails int
}

// NewTextFormatter creates a new TextFormatter.
func NewTextFormatter() TextFormatter {
	// Create a new formatter with default values
	formatter := TextFormatter{
		verbose:    false,
		showHelp:   false,
		lightMode:  false,
		lineWidth:  80,
		maxDetails: 3,
	}

	// Initialize the formatter with default color scheme and symbols
	formatter = formatter.withInitializedColorScheme()
	formatter = formatter.withInitializedSymbols()

	return formatter
}

// withInitializedColorScheme initializes the color scheme for the formatter.
// Returns a new formatter with the initialized color scheme.
func (f TextFormatter) withInitializedColorScheme() TextFormatter {
	result := f

	// Create a color scheme based on the terminal's capabilities and light/dark mode
	if f.lightMode {
		// Light mode uses different colors that work better on light backgrounds
		result.colors = ColorScheme{
			Success:  color.New(color.FgGreen).SprintFunc(),
			Warning:  color.New(color.FgRed).SprintFunc(),
			Error:    color.New(color.FgRed, color.Bold).SprintFunc(),
			Info:     color.New(color.FgBlue).SprintFunc(),
			Bold:     color.New(color.Bold).SprintFunc(),
			Header:   color.New(color.FgBlue, color.Bold).SprintFunc(),
			Muted:    color.New(color.FgBlack).SprintFunc(),
			HelpText: color.New(color.Italic, color.FgBlue).SprintFunc(),
		}
	} else {
		// Dark mode (default)
		result.colors = ColorScheme{
			Success:  color.New(color.FgGreen).SprintFunc(),
			Warning:  color.New(color.FgYellow).SprintFunc(),
			Error:    color.New(color.FgRed, color.Bold).SprintFunc(),
			Info:     color.New(color.FgCyan).SprintFunc(),
			Bold:     color.New(color.Bold).SprintFunc(),
			Header:   color.New(color.FgCyan, color.Bold).SprintFunc(),
			Muted:    color.New(color.FgHiBlack).SprintFunc(),
			HelpText: color.New(color.Italic, color.FgCyan).SprintFunc(),
		}
	}

	return result
}

// withInitializedSymbols initializes the symbols for the formatter.
// Returns a new formatter with the initialized symbols.
func (f TextFormatter) withInitializedSymbols() TextFormatter {
	result := f

	// Use color functions from the color scheme
	result.symbols = SymbolSet{
		pass:  f.colors.Success("✓"),
		fail:  f.colors.Error("✗"),
		skip:  f.colors.Muted("-"),
		arrow: f.colors.Muted("→"),
	}

	return result
}

// WithVerbose sets the verbose flag for the formatter.
func (f TextFormatter) WithVerbose(verbose bool) TextFormatter {
	result := f
	result.verbose = verbose

	// Re-initialize the formatter with the new settings
	result = result.withInitializedColorScheme()
	result = result.withInitializedSymbols()

	return result
}

// WithShowHelp sets the showHelp flag for the formatter.
func (f TextFormatter) WithShowHelp(showHelp bool) TextFormatter {
	result := f
	result.showHelp = showHelp

	return result
}

// WithLightMode sets the lightMode flag for the formatter.
func (f TextFormatter) WithLightMode(lightMode bool) TextFormatter {
	result := f
	result.lightMode = lightMode

	// Re-initialize the formatter with the new settings
	result = result.withInitializedColorScheme()
	result = result.withInitializedSymbols()

	return result
}

// Format formats validation results as text.
func (f TextFormatter) Format(ctx context.Context, results domain.ValidationResults) string {
	logger := log.Logger(ctx)
	logger.Trace().Bool("verbose", f.verbose).Bool("light_mode", f.lightMode).Bool("show_help", f.showHelp).Int("total_commits", results.TotalCommits).Msg("Entering TextFormatter.Format")

	var builder strings.Builder

	// Get the configuration from context for use in formatting
	cfg := config.GetConfig(ctx)

	// For multiple commits, show a summary header first
	if len(results.CommitResults) > 1 {
		f.formatOverallSummary(&builder, results)
	}

	// Process each commit result
	for i, commitResult := range results.CommitResults {
		f.formatCommitHeader(&builder, commitResult, i, len(results.CommitResults))
		f.formatRuleResults(ctx, &builder, commitResult, cfg)
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

	// Format each rule failure summary using Map
	formattedSummaries := contextx.Map(ruleNames, func(ruleName string) string {
		count := results.RuleSummary[ruleName]

		return fmt.Sprintf("  - %s: %d failure(s)\n", f.colors.Bold(ruleName), count)
	})

	// Join all formatted summaries and write to builder
	builder.WriteString(strings.Join(formattedSummaries, ""))
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
func (f TextFormatter) formatRuleResults(ctx context.Context, builder *strings.Builder, commitResult domain.CommitResult, cfg config.Config) {
	// Debug info about actual config values
	debugFile, _ := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if debugFile != nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "\n==== TEXT FORMATTER ====\n")
		fmt.Fprintf(debugFile, "CONFIG: enabled=%v disabled=%v\n", cfg.Rules.EnabledRules, cfg.Rules.DisabledRules)

		// Show specific rule status for our problematic rules
		fmt.Fprintf(debugFile, "CommitBody explicitly enabled: %v\n",
			config.IsRuleEnabled("CommitBody", cfg.Rules.EnabledRules, cfg.Rules.DisabledRules))
		fmt.Fprintf(debugFile, "JiraReference explicitly enabled: %v\n",
			config.IsRuleEnabled("JiraReference", cfg.Rules.EnabledRules, cfg.Rules.DisabledRules))

		// Log all rules before filtering
		fmt.Fprintf(debugFile, "\nInitial rule results before filtering:\n")

		for _, rule := range commitResult.RuleResults {
			fmt.Fprintf(debugFile, "INITIAL RULE: %s (status=%s)\n", rule.RuleName, rule.Status)
		}
	}

	// Sort rule results alphabetically by name
	sortedRules := getSortedRuleResults(commitResult.RuleResults)

	// Use the application-level filter to properly filter rules based on configuration
	filteredRules := validate.GetFilteredRuleResults(ctx, sortedRules)

	// Use Map to create RuleCount records for each filtered rule
	type RuleCount struct {
		rule   domain.RuleResult
		passed bool
	}

	ruleCounts := contextx.Map(filteredRules, func(rule domain.RuleResult) RuleCount {
		return RuleCount{
			rule:   rule,
			passed: rule.Status == domain.StatusPassed,
		}
	})

	// Extract active rules and count passed rules
	activeRules := contextx.Map(ruleCounts, func(rc RuleCount) domain.RuleResult {
		return rc.rule
	})

	// Count total rules (after all filtering)
	totalRules := len(activeRules)

	// Count passed rules by filtering the ruleCounts
	passedRules := len(contextx.Filter(ruleCounts, func(rc RuleCount) bool {
		return rc.passed
	}))

	if debugFile != nil {
		fmt.Fprintf(debugFile, "FILTERED RULES COUNT: %d (after removing disabled rules)\n", totalRules)

		// Log which rules are counted
		activeRuleNames := make([]string, 0, len(activeRules))
		for _, rule := range activeRules {
			activeRuleNames = append(activeRuleNames, rule.RuleName)
		}

		fmt.Fprintf(debugFile, "COUNTED RULES: %v\n", activeRuleNames)
	}

	// Process each active rule
	contextx.ForEach(activeRules, func(ruleResult domain.RuleResult) {
		ruleName := f.colors.Bold(ruleResult.RuleName)

		if ruleResult.Status == domain.StatusPassed {
			f.formatPassedRule(builder, ruleName, ruleResult)
		} else if ruleResult.Status == domain.StatusFailed {
			f.formatFailedRule(builder, ruleName, ruleResult)
		}
	})

	f.formatRuleSummaryLine(builder, passedRules, totalRules)
}

// formatPassedRule formats a passed rule result.
func (f TextFormatter) formatPassedRule(builder *strings.Builder, ruleName string, ruleResult domain.RuleResult) {
	// Always show basic output
	builder.WriteString(fmt.Sprintf("%s %s: ", f.symbols.pass, ruleName))
	builder.WriteString(ruleResult.Message + "\n")

	// Only show detailed output in verbose mode
	if !f.verbose {
		return
	}

	builder.WriteString(fmt.Sprintf("    %s\n", ruleResult.VerboseMessage))
}

// formatFailedRule formats a failed rule result.
func (f TextFormatter) formatFailedRule(builder *strings.Builder, ruleName string, ruleResult domain.RuleResult) {
	// Basic output
	builder.WriteString(fmt.Sprintf("%s %s: ", f.symbols.fail, ruleName))
	builder.WriteString("\n  \n")

	// Format errors
	for _, err := range ruleResult.Errors {
		builder.WriteString(fmt.Sprintf("    %s\n", f.colors.Error(err.Error())))
	}

	// Add a blank line for spacing
	builder.WriteString("\n")

	// Show help if available
	if f.showHelp && ruleResult.HelpMessage != "" {
		builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText(ruleResult.HelpMessage)))
		builder.WriteString("\n")
	} else {
		// Otherwise just show a hint on how to get help
		builder.WriteString(fmt.Sprintf("  %s\n\n", f.colors.Muted(fmt.Sprintf("Run with '--rulehelp=%s' for specific guidance on this rule.", ruleResult.RuleName))))
	}
}

// formatRuleSummaryLine formats the summary line.
func (f TextFormatter) formatRuleSummaryLine(builder *strings.Builder, passedRules, totalRules int) {
	if passedRules == totalRules {
		builder.WriteString(f.colors.Success(fmt.Sprintf("\nPASS: All %d rules passed\n\n", totalRules)))

		return
	}

	builder.WriteString(f.colors.Warning(fmt.Sprintf("\nFAIL: %d of %d rules passed\n\n", passedRules, totalRules)))
}

// getSortedRuleNames returns sorted rule names from the rule summary map.
func getSortedRuleNames(summary map[string]int) []string {
	ruleNames := make([]string, 0, len(summary))
	for ruleName := range summary {
		ruleNames = append(ruleNames, ruleName)
	}

	sort.Strings(ruleNames)

	return ruleNames
}

// getSortedRuleResults returns rule results sorted by rule name.
func getSortedRuleResults(results []domain.RuleResult) []domain.RuleResult {
	// Create a copy to avoid mutating the original
	sortedResults := make([]domain.RuleResult, len(results))
	copy(sortedResults, results)

	// Sort by rule name
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].RuleName < sortedResults[j].RuleName
	})

	return sortedResults
}

// Note: We keep the wrapText function signature to avoid unused function warnings,
// but we don't use it in our current formatter implementation.

// ColorScheme provides color functions for different parts of the output.
type ColorScheme struct {
	Success  func(a ...interface{}) string
	Warning  func(a ...interface{}) string
	Error    func(a ...interface{}) string
	Info     func(a ...interface{}) string
	Bold     func(a ...interface{}) string
	Header   func(a ...interface{}) string
	Muted    func(a ...interface{}) string
	HelpText func(a ...interface{}) string
}

// SymbolSet provides symbols for different statuses.
type SymbolSet struct {
	pass  string
	fail  string
	skip  string
	arrow string
}

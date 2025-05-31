// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package output provides output formatters for validation results.
package format

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
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
func (f TextFormatter) Format(ctx context.Context, results interface{}) string {
	validationResults, ok := results.(domain.ValidationResults)
	if !ok {
		return "Error: invalid results type"
	}
	// Estimate initial capacity based on number of commits and results
	initialCapacity := 1000                                 // Base size
	initialCapacity += 200 * len(validationResults.Results) // Headers

	// Add space for rule results (roughly 100 bytes per rule per commit)
	for _, commit := range validationResults.Results {
		initialCapacity += 100 * len(commit.RuleResults)
	}

	var builder strings.Builder

	builder.Grow(initialCapacity)

	// Formatters should receive pre-processed data from application layer
	// No need to access context configuration directly
	cfg := config.Config{
		Rules: config.RulesConfig{
			Enabled:  []string{},
			Disabled: []string{},
		},
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				MaxLength:         50,
				Case:              "sentence",
				RequireImperative: true,
			},
		},
	}

	// For multiple commits, show a summary header first
	if len(validationResults.Results) > 1 {
		f.formatOverallSummary(&builder, validationResults)
	}

	// Process each commit result
	for i, commitResult := range validationResults.Results {
		f.formatCommitHeader(&builder, commitResult, i, len(validationResults.Results))
		f.formatRuleResults(ctx, &builder, commitResult, cfg)
	}

	return builder.String()
}

// ContentType returns the MIME type for text output.
func (f TextFormatter) ContentType() string {
	return "text/plain"
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
	// Get failed rules using the new method
	failedRules := results.GetFailedRules()
	if len(failedRules) == 0 {
		return
	}

	builder.WriteString(f.colors.Bold("Rule failures summary:\n"))

	// Sort rules by name for consistent output
	ruleNames := getSortedRuleNames(failedRules)

	// Format each rule failure summary using Map
	formattedSummaries := domain.MapSliceCompat(ruleNames, func(ruleName string) string {
		count := failedRules[ruleName]

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
func (f TextFormatter) formatRuleResults(_ context.Context, builder *strings.Builder, commitResult domain.CommitResult, _ config.Config) {
	// Sort rule results alphabetically by name
	sortedRules := getSortedRuleResults(commitResult.RuleResults)

	// Formatters should not filter rules - that's application layer responsibility
	// Accept pre-filtered rules from application layer
	filteredRules := sortedRules

	// Use Map to create RuleCount records for each filtered rule
	type RuleCount struct {
		rule   domain.RuleResult
		passed bool
	}

	ruleCounts := domain.MapSliceCompat(filteredRules, func(rule domain.RuleResult) RuleCount {
		return RuleCount{
			rule:   rule,
			passed: rule.Status == domain.StatusPassed,
		}
	})

	// Extract active rules and count passed rules
	activeRules := domain.MapSliceCompat(ruleCounts, func(rc RuleCount) domain.RuleResult {
		return rc.rule
	})

	// Count total rules (after all filtering)
	totalRules := len(activeRules)

	// Count passed rules by filtering the ruleCounts
	passedRules := len(domain.FilterSliceCompat(ruleCounts, func(rc RuleCount) bool {
		return rc.passed
	}))

	// Process each active rule
	for _, ruleResult := range activeRules {
		ruleName := f.colors.Bold(ruleResult.RuleName)

		if ruleResult.Status == domain.StatusPassed {
			f.formatPassedRule(builder, ruleName, ruleResult)
		} else if ruleResult.Status == domain.StatusFailed {
			f.formatFailedRule(builder, ruleName, ruleResult)
		}
	}

	f.formatRuleSummaryLine(builder, passedRules, totalRules)
}

// formatPassedRule formats a passed rule result.
func (f TextFormatter) formatPassedRule(builder *strings.Builder, ruleName string, ruleResult domain.RuleResult) {
	// Always show basic output
	builder.WriteString(fmt.Sprintf("%s %s: ", f.symbols.pass, ruleName))
	// Generate message on-the-fly
	message := Result(ruleResult.RuleName, ruleResult.Errors)
	builder.WriteString(message + "\n")

	// Only show detailed output in verbose mode
	if !f.verbose {
		return
	}

	// Generate verbose message on-the-fly
	verboseMessage := VerboseResult(ruleResult.RuleName, ruleResult.Errors)
	builder.WriteString(fmt.Sprintf("    %s\n", verboseMessage))
}

// formatFailedRule formats a failed rule result.
func (f TextFormatter) formatFailedRule(builder *strings.Builder, ruleName string, ruleResult domain.RuleResult) {
	// Basic output
	builder.WriteString(fmt.Sprintf("%s %s: ", f.symbols.fail, ruleName))
	builder.WriteString("\n")

	// Format errors with progressive detail based on formatter settings
	for _, err := range ruleResult.Errors {
		// Determine verbosity level from formatter flags
		verbosityLevel := 0
		if f.verbose {
			verbosityLevel = 1
		}

		if f.showHelp {
			verbosityLevel = 2
		}

		// Use progressive formatting
		// Since err is already of type ValidationError, we don't need type assertion
		{
			// Format based on verbosity level
			var formatted string

			switch verbosityLevel {
			case 0:
				// Basic: just the message
				formatted = err.Message
			case 1:
				// Medium: message with code
				if err.Code != "" {
					formatted = fmt.Sprintf("[%s] %s", err.Code, err.Message)
				} else {
					formatted = err.Message
				}
			default:
				// High: message with code and context
				if err.Code != "" {
					formatted = fmt.Sprintf("[%s] %s", err.Code, err.Message)
				} else {
					formatted = err.Message
				}

				if len(err.Context) > 0 {
					var contextParts []string
					for k, v := range err.Context {
						contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
					}

					formatted = fmt.Sprintf("%s (%s)", formatted, strings.Join(contextParts, ", "))
				}
			}
			// Split by newlines and indent each line
			lines := strings.Split(formatted, "\n")
			for _, line := range lines {
				if line != "" {
					builder.WriteString(fmt.Sprintf("    %s\n", line))
				}
			}
		}
	}

	// Add a blank line for spacing
	builder.WriteString("\n")
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

// Ensure TextFormatter implements Formatter.
var _ Formatter = (*TextFormatter)(nil)

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package output provides output formatting for validation results.
package output

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
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

// These methods are now replaced by withInitializedColorScheme and withInitializedSymbols
// which follow functional programming patterns by returning new instances
// instead of modifying state.

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
		f.formatRuleResults(&builder, commitResult, cfg)
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
func (f TextFormatter) formatRuleResults(builder *strings.Builder, commitResult domain.CommitResult, cfg config.Config) {
	// Debug info about actual config values
	debugFile, _ := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if debugFile != nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "CONFIG: enabled=%v disabled=%v\n", cfg.Rules.EnabledRules, cfg.Rules.DisabledRules)
		
		// Log all rules before filtering
		for _, rule := range commitResult.RuleResults {
			fmt.Fprintf(debugFile, "INITIAL RULE: %s (status=%s)\n", rule.RuleName, rule.Status)
		}
	}
	
	// Sort rule results alphabetically by name
	sortedRules := getSortedRuleResults(commitResult.RuleResults)

	// Use FilterMap to both filter out skipped rules and count passed rules in one pass
	type RuleCount struct {
		rule   domain.RuleResult
		passed bool
	}

	// Filter out skipped rules and map to a structure that tracks if the rule passed
	ruleCounts := contextx.FilterMap(sortedRules,
		// Filter predicate: include only non-skipped rules AND not disabled by config
		func(rule domain.RuleResult) bool {
			// cfg is now passed as a parameter to formatRuleResults
			
			// Add more detailed debugging
			debugFile, _ := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY, 0644)
			if debugFile != nil {
				defer debugFile.Close()
				fmt.Fprintf(debugFile, "CHECKING RULE: %s (enabled_rules: %v, disabled_rules: %v)\n", 
				         rule.RuleName, cfg.Rules.EnabledRules, cfg.Rules.DisabledRules)
			}
			
			// Generic solution: Always check the YAML config file to determine if a rule
			// should be shown, regardless of what the runtime config says
			yamlConfig := ".gommitlint.yaml"
			if _, err := os.Stat(yamlConfig); err == nil {
				content, err := os.ReadFile(yamlConfig)
				if err == nil {
					configContent := string(content)
					// Simple check to see if the rule is explicitly enabled in the config file
					if strings.Contains(configContent, "enabled_rules:") && 
					   strings.Contains(configContent, "- "+rule.RuleName) {
						if debugFile != nil {
							fmt.Fprintf(debugFile, "CONFIG FILE OVERRIDE: Explicitly showing rule %s from config file\n", rule.RuleName)
						}
						return true
					}
				}
			}
			
			// Check if the rule is explicitly enabled (highest priority)
			// This ensures that rules specified in enabled_rules will always be shown,
			// regardless of other conflicting settings
			isExplicitlyEnabled := false
			for _, enabledRule := range cfg.Rules.EnabledRules {
				if enabledRule == rule.RuleName {
					if debugFile != nil {
						fmt.Fprintf(debugFile, "OUTPUT: Including explicitly enabled rule: %s\n", rule.RuleName)
					}
					isExplicitlyEnabled = true
					break
				}
			}
			
			// If explicitly enabled, always show it
			if isExplicitlyEnabled {
				return true
			}
			
			// Check if rule is explicitly disabled (second priority)
			// Rules in disabled_rules should be hidden unless explicitly enabled
			for _, disabledRule := range cfg.Rules.DisabledRules {
				if disabledRule == rule.RuleName {
					if debugFile != nil {
						fmt.Fprintf(debugFile, "OUTPUT: Filtering out explicitly disabled rule: %s\n", rule.RuleName)
					}
					return false
				}
			}
			
			// For rules neither explicitly enabled nor disabled:
			// Use default behavior - hide rules in DefaultDisabledRules
			if config.DefaultDisabledRules[rule.RuleName] {
				// Log which rules we're filtering out
				if debugFile != nil {
					fmt.Fprintf(debugFile, "OUTPUT: Filtering out disabled-by-default rule: %s\n", rule.RuleName)
				}
				return false
			}

			// Only include rules that haven't been skipped
			return rule.Status != domain.StatusSkipped
		},
		// Map function: convert to RuleCount with passed flag
		func(rule domain.RuleResult) RuleCount {
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
	
	if debugFile, _ := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY, 0644); debugFile != nil {
		defer debugFile.Close()
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

	// Only show verbose output when verbose flag is given
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

	// Special rules have standardized help messages for consistency
	if ruleName == "JiraReference" {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("Include a valid Jira issue key (e.g., PROJECT-123) in your commit subject.")))
		builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("For conventional commits, place the key at the end of the first line:")))
		builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("- feat(auth): add login feature PROJ-123")))
		builder.WriteString(fmt.Sprintf("  %s\n", f.colors.HelpText("- fix: resolve timeout issue [PROJ-123]")))
		builder.WriteString("\n")

		return
	}

	// Show help text for rules with --extra-verbose flag
	if f.showHelp {
		// Always show the rule's Help(errors []errors.ValidationError) text when showHelp is true
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

	// Process each error and format it
	contextx.ForEach(validationErrors, func(err errors.ValidationError) {
		// Check for suggested corrections
		suggestedForm, hasSuggestion := err.Context["suggested_form"]
		suggestionText, hasText := err.Context["suggestion_text"]

		// Show the suggestion if available (always show, regardless of verbose mode)
		if hasSuggestion {
			builder.WriteString(fmt.Sprintf("    %s\n",
				f.colors.Success("✓ SUGGESTED CORRECTION:")))

			if hasText {
				builder.WriteString(fmt.Sprintf("      %s\n",
					f.colors.Bold(suggestionText)))
			}

			builder.WriteString(fmt.Sprintf("      %s\n",
				f.colors.Success(suggestedForm)))
			builder.WriteString("\n")
		}

		// Format error message based on verbosity
		if !f.verbose {
			// Simple format for non-verbose mode
			builder.WriteString(fmt.Sprintf("    - %s\n",
				f.colors.Error(err.Message)))
		} else {
			// In verbose mode, use the enhanced formatter to get rich error display
			errorText := errorFormatter.FormatError(err)

			// Format and indent each line
			indentedLines := contextx.Map(
				strings.Split(strings.TrimSpace(errorText), "\n"),
				func(line string) string {
					return fmt.Sprintf("    %s\n", line)
				},
			)

			// Join all lines and write to builder
			builder.WriteString(strings.Join(indentedLines, ""))
		}
	})
}

// formatHelpText formats the help text for a failed rule.
func (f TextFormatter) formatHelpText(builder *strings.Builder, ruleResult domain.RuleResult) {
	// Standard behavior
	if f.showHelp {
		builder.WriteString("\n")

		// First, check for enhanced errors with help messages
		if len(ruleResult.Errors) > 0 {
			// Use Some to find the first error with a help message
			helpMessage := ""

			// Find the first error with a non-empty help message
			foundError := contextx.Some(ruleResult.Errors, func(err errors.ValidationError) bool {
				if help := err.GetHelp(); help != "" {
					helpMessage = help

					return true
				}

				return false
			})

			// If we found a help message, use it
			if foundError && helpMessage != "" {
				// Format helpLines using Map to transform each line
				helpLines := strings.Split(helpMessage, "\n")
				formattedLines := contextx.Map(helpLines, func(line string) string {
					return "  " + f.colors.HelpText(line)
				})

				// Join all formatted lines with newlines
				builder.WriteString(strings.Join(formattedLines, "\n"))
				builder.WriteString("\n\n")

				return
			}
		}

		// Fallback to rule's help message if enhanced error doesn't have one
		if ruleResult.HelpMessage != "" {
			// Format helpLines using Map to transform each line
			helpLines := strings.Split(ruleResult.HelpMessage, "\n")
			formattedLines := contextx.Map(helpLines, func(line string) string {
				return "  " + f.colors.HelpText(line)
			})

			// Join all formatted lines with newlines
			builder.WriteString(strings.Join(formattedLines, "\n"))
			builder.WriteString("\n\n")
		}
	} else if f.verbose && ruleResult.RuleID != "" {
		// In verbose mode, show a compact tip if we have a rule ID
		builder.WriteString(fmt.Sprintf("    %s Run with '--rulehelp=%s' for detailed instructions\n\n",
			f.colors.Warning("TIP:"), strings.ToLower(ruleResult.RuleID)))
	}
}

// formatRuleSummaryLine formats the summary line for a commit.
func (f TextFormatter) formatRuleSummaryLine(builder *strings.Builder, passedRules, totalRules int) {
	// Log for debugging
	debugFile, _ := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if debugFile != nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "SUMMARY: %d of %d rules passed\n", passedRules, totalRules)
	}

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
	// Create a flattened list of all rule results from all commits
	// Using a nested loop isn't very functional, let's refactor this
	// Function to check if a rule matches the requested name
	ruleMatches := func(ruleResult domain.RuleResult) bool {
		return strings.EqualFold(ruleResult.RuleID, ruleName) ||
			strings.EqualFold(ruleResult.RuleName, ruleName)
	}

	// Function to format a rule's help message
	formatRuleHelp := func(ruleResult domain.RuleResult) bool {
		if ruleResult.HelpMessage == "" {
			return false
		}

		// Format help lines using functional Map
		helpLines := strings.Split(ruleResult.HelpMessage, "\n")
		formattedLines := contextx.Map(helpLines, func(line string) string {
			return "  " + f.colors.HelpText(line)
		})

		// Join all formatted lines with newlines
		builder.WriteString(strings.Join(formattedLines, "\n"))
		builder.WriteString("\n\n")

		return true
	}

	// Search through all commit results
	for _, commitResult := range results.CommitResults {
		// Filter for matching rules
		matchingRules := contextx.Filter(commitResult.RuleResults, ruleMatches)

		// Format the first matching rule with a help message (if any)
		for _, rule := range matchingRules {
			if formatRuleHelp(rule) {
				return true
			}
		}
	}

	return false
}

// formatAvailableRules formats the list of available rules.
func (f TextFormatter) formatAvailableRules(builder *strings.Builder, results domain.ValidationResults) {
	builder.WriteString("Available rules:\n")

	// Collect unique rule names using a map
	ruleNames := make(map[string]bool)

	// Extract all rule names from the results
	for _, commitResult := range results.CommitResults {
		for _, ruleResult := range commitResult.RuleResults {
			ruleNames[ruleResult.RuleName] = true
		}
	}

	// Convert the map keys to a slice
	uniqueRuleNames := make([]string, 0, len(ruleNames))
	for name := range ruleNames {
		uniqueRuleNames = append(uniqueRuleNames, name)
	}

	// Sort the names using Map to transform them
	sortedNames := contextx.DeepCopy(uniqueRuleNames)
	sort.Strings(sortedNames)

	// Use Map to transform each rule name into a formatted string
	formattedRules := contextx.Map(sortedNames, func(name string) string {
		return "  - " + f.colors.Bold(name)
	})

	// Join all formatted rules with newlines
	builder.WriteString(strings.Join(formattedRules, "\n"))
	builder.WriteString("\n")
}

// getSortedRuleNames returns a sorted slice of rule names from a map.
func getSortedRuleNames(ruleMap map[string]int) []string {
	// Extract keys from map
	keys := make([]string, 0, len(ruleMap))
	for k := range ruleMap {
		keys = append(keys, k)
	}

	// Create a deep copy of the keys and sort it (following functional patterns)
	sorted := contextx.DeepCopy(keys)
	sort.Strings(sorted)

	return sorted
}

// getSortedRuleResults returns a sorted slice of rule results.
func getSortedRuleResults(ruleResults []domain.RuleResult) []domain.RuleResult {
	// Create a deep copy using our utility
	sortedRules := contextx.DeepCopy(ruleResults)

	// Sort the copy by rule name
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

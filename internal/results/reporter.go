// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package results

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal"
)

// OutputFormat defines the format for validation results.
type OutputFormat string

const (
	// FormatText outputs results in human-readable text format.
	FormatText OutputFormat = "text"

	// FormatJSON outputs results in machine-readable JSON format.
	FormatJSON OutputFormat = "json"
)

// ReporterOptions configures the behavior of result reporting.
type ReporterOptions struct {
	// Format specifies the output format (text, json, etc.)
	Format OutputFormat

	// Verbose enables detailed output mode
	Verbose bool

	// ShowHelp enables rule help text display
	ShowHelp bool

	// RuleToShowHelp specifies a specific rule to show help for
	RuleToShowHelp string

	// LightMode enables color scheme for light backgrounds
	LightMode bool

	// Writer specifies where output should be written (defaults to stdout)
	Writer io.Writer
}

// Reporter handles the formatting and output of validation results.
type Reporter struct {
	aggregator *Aggregator
	options    ReporterOptions
}

// NewReporter creates a new result reporter.
func NewReporter(aggregator *Aggregator, options ReporterOptions) *Reporter {
	// Set defaults if not specified
	if options.Format == "" {
		options.Format = FormatText
	}

	if options.Writer == nil {
		options.Writer = os.Stdout
	}

	return &Reporter{
		aggregator: aggregator,
		options:    options,
	}
}

// GenerateReport produces a complete validation report.
func (r *Reporter) GenerateReport() error {
	switch r.options.Format {
	case FormatText:
		return r.generateTextReport()
	case FormatJSON:
		return r.generateJSONReport()
	default:
		return fmt.Errorf("unsupported output format: %s", r.options.Format)
	}
}

// generateTextReport produces a human-readable text report.
func (r *Reporter) generateTextReport() error {
	if !r.aggregator.HasAnyResults() {
		fmt.Fprintln(r.options.Writer, "No commits were validated")

		return nil
	}

	// If specific rule help requested, show that and exit
	if r.options.ShowHelp && r.options.RuleToShowHelp != "" {
		return r.showRuleHelp()
	}

	// Create print options from reporter options
	printOpts := &internal.PrintOptions{
		Verbose:        r.options.Verbose,
		ShowHelp:       r.options.ShowHelp,
		RuleToShowHelp: r.options.RuleToShowHelp,
		LightMode:      r.options.LightMode,
	}

	// Print each commit result
	for _, result := range r.aggregator.GetCommitResults() {
		err := internal.PrintReport(result.Rules, &result.CommitInfo, printOpts)
		if err != nil {
			return err
		}
	}

	// Print overall summary if multiple commits
	if r.aggregator.GetSummary().TotalCommits > 1 {
		r.printOverallSummary()
	}

	return nil
}

// generateJSONReport produces a machine-readable JSON report.
func (r *Reporter) generateJSONReport() error {
	// Create a simplified data structure for JSON output
	type jsonRule struct {
		Name    string   `json:"name"`
		Passed  bool     `json:"passed"`
		Message string   `json:"message"`
		Errors  []string `json:"errors,omitempty"`
	}

	type jsonCommit struct {
		SHA     string     `json:"sha"`
		Subject string     `json:"subject"`
		Passed  bool       `json:"passed"`
		Rules   []jsonRule `json:"rules"`
	}

	type jsonSummary struct {
		TotalCommits  int          `json:"totalCommits"`
		PassedCommits int          `json:"passedCommits"`
		FailedCommits int          `json:"failedCommits"`
		PassRate      float64      `json:"passRate"`
		Commits       []jsonCommit `json:"commits"`
		Status        string       `json:"status"`
	}

	// Build the JSON structure
	commits := make([]jsonCommit, 0, len(r.aggregator.GetCommitResults()))

	for _, result := range r.aggregator.GetCommitResults() {
		// Extract SHA from commit info
		sha := "unknown"
		if result.CommitInfo.RawCommit != nil {
			sha = result.CommitInfo.RawCommit.Hash.String()
		}

		// Convert rules to JSON format
		rules := make([]jsonRule, 0, len(result.Rules))

		for _, rule := range result.Rules {
			// Convert validation errors to strings
			errorMessages := make([]string, 0, len(rule.Errors()))
			for _, err := range rule.Errors() {
				errorMessages = append(errorMessages, err.Error())
			}

			rules = append(rules, jsonRule{
				Name:    rule.Name(),
				Passed:  len(rule.Errors()) == 0,
				Message: rule.Result(),
				Errors:  errorMessages,
			})
		}

		commits = append(commits, jsonCommit{
			SHA:     sha,
			Subject: result.CommitInfo.Subject,
			Passed:  result.Passed,
			Rules:   rules,
		})
	}

	// Calculate pass rate
	passRate := 0.0
	summary := r.aggregator.GetSummary()

	if summary.TotalCommits > 0 {
		passRate = float64(summary.PassedCommits) / float64(summary.TotalCommits)
	}

	// Determine overall status
	status := "success"
	if !r.aggregator.AllRulesPassed() {
		status = "failure"
	}

	// Build complete summary
	jsonOutput := jsonSummary{
		TotalCommits:  summary.TotalCommits,
		PassedCommits: summary.PassedCommits,
		FailedCommits: summary.TotalCommits - summary.PassedCommits,
		PassRate:      passRate,
		Commits:       commits,
		Status:        status,
	}

	// Marshal to JSON and write to output
	encoder := json.NewEncoder(r.options.Writer)
	encoder.SetIndent("", "  ")

	return encoder.Encode(jsonOutput)
}

// showRuleHelp displays help for a specific rule.
func (r *Reporter) showRuleHelp() error {
	// To avoid redundancy, find the first occurrence of the requested rule
	// and display its help information
	ruleToShowHelp := strings.ToLower(r.options.RuleToShowHelp)

	for _, result := range r.aggregator.GetCommitResults() {
		for _, rule := range result.Rules {
			if strings.EqualFold(rule.Name(), ruleToShowHelp) {
				fmt.Fprintf(r.options.Writer, "\n%s Rule Help:\n", rule.Name())

				// Print help text with proper indentation
				helpLines := strings.Split(rule.Help(), "\n")
				for _, line := range helpLines {
					fmt.Fprintf(r.options.Writer, "  %s\n", line)
				}

				fmt.Fprintln(r.options.Writer, "")

				return nil
			}
		}
	}

	// Rule not found, show available rules
	fmt.Fprintf(r.options.Writer, "No help found for rule: %s\n", r.options.RuleToShowHelp)
	fmt.Fprintln(r.options.Writer, "Available rules:")

	// Get a unique list of rule names
	ruleNames := make(map[string]bool)

	for _, result := range r.aggregator.GetCommitResults() {
		for _, rule := range result.Rules {
			ruleNames[rule.Name()] = true
		}
	}

	// Print sorted rule names
	sortedNames := make([]string, 0, len(ruleNames))
	for name := range ruleNames {
		sortedNames = append(sortedNames, name)
	}

	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		fmt.Fprintf(r.options.Writer, "  - %s\n", name)
	}

	return nil
}

// printOverallSummary prints a summary of all validation results.
func (r *Reporter) printOverallSummary() {
	summary := r.aggregator.GetSummary()

	// Create a divider line
	divider := strings.Repeat("=", 80)

	// Define colors based on settings
	var summaryColor func(format string, a ...interface{}) string

	// Check for NO_COLOR environment variable
	noColor := false
	if val, exists := os.LookupEnv("NO_COLOR"); exists && val != "" {
		noColor = true
		color.NoColor = true
	}

	if noColor {
		// No color mode - just use plain text
		summaryColor = fmt.Sprintf
	} else if r.options.LightMode {
		// Light mode - use blue bold (works well on light backgrounds)
		summaryColor = color.New(color.FgBlue, color.Bold).SprintfFunc()
	} else {
		// Dark mode - use cyan bold (works well on dark backgrounds)
		summaryColor = color.New(color.FgHiBlue, color.Bold).SprintfFunc()
	}

	// Calculate failed commits
	failedCommits := summary.TotalCommits - summary.PassedCommits

	// Print the summary
	fmt.Fprintln(r.options.Writer, summaryColor(divider))
	fmt.Fprintln(r.options.Writer, summaryColor("OVERALL SUMMARY"))
	fmt.Fprintln(r.options.Writer, summaryColor(divider))
	fmt.Fprintf(r.options.Writer, "%s Validated %d commits\n", summaryColor("Result:"), summary.TotalCommits)
	fmt.Fprintf(r.options.Writer, "  %s %d commits passed\n", summaryColor("Passed:"), summary.PassedCommits)
	fmt.Fprintf(r.options.Writer, "  %s %d commits failed\n", summaryColor("Failed:"), failedCommits)

	// Show most common failures if any
	if failedCommits > 0 && len(summary.FailedRuleTypes) > 0 {
		fmt.Fprintln(r.options.Writer, "")
		fmt.Fprintf(r.options.Writer, "%s\n", summaryColor("Most Common Issues:"))

		topFailures := r.aggregator.GetMostFrequentFailures(3) // Show top 3
		for i, ruleName := range topFailures {
			count := summary.FailedRuleTypes[ruleName]
			fmt.Fprintf(r.options.Writer, "  %d. %s: %d occurrences\n", i+1, ruleName, count)
		}
	}

	fmt.Fprintln(r.options.Writer, "")
}

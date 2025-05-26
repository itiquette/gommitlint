// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// GitHubActionsFormatter formats validation results for GitHub Actions.
// It implements the domain.ResultFormatter interface.
type GitHubActionsFormatter struct {
	verbose  bool
	showHelp bool
}

// Ensure GitHubActionsFormatter implements outgoing.ResultFormatter.
var _ outgoing.ResultFormatter = GitHubActionsFormatter{}

// NewGitHubFormatter creates a new GitHub Actions formatter.
// It implements domain.ResultFormatter interface.
func NewGitHubFormatter() GitHubActionsFormatter {
	return GitHubActionsFormatter{
		verbose:  false,
		showHelp: false,
	}
}

// WithVerbose sets the verbose flag for the formatter.
func (f GitHubActionsFormatter) WithVerbose(verbose bool) GitHubActionsFormatter {
	result := f
	result.verbose = verbose

	return result
}

// WithShowHelp sets the showHelp flag for the formatter.
func (f GitHubActionsFormatter) WithShowHelp(showHelp bool) GitHubActionsFormatter {
	result := f
	result.showHelp = showHelp

	return result
}

// Format formats validation results for GitHub Actions output.
func (f GitHubActionsFormatter) Format(_ context.Context, results interface{}) string {
	validationResults, ok := results.(domain.ValidationResults)
	if !ok {
		return "Error: invalid results type"
	}

	var builder strings.Builder

	// Print summary
	builder.WriteString("::group::Summary\n")
	fmt.Fprintf(&builder, "Validated %d commits\n", validationResults.TotalCommits)
	fmt.Fprintf(&builder, "Passed: %d, Failed: %d\n", validationResults.PassedCommits, validationResults.TotalCommits-validationResults.PassedCommits)
	builder.WriteString("::endgroup::\n")

	// Print details for each commit
	for i, commitResult := range validationResults.Results {
		fmt.Fprintf(&builder, "::group::Commit #%d: %s\n", i+1, commitResult.CommitInfo.Hash)
		fmt.Fprintf(&builder, "Subject: %s\n", commitResult.CommitInfo.Subject)

		if commitResult.CommitInfo.IsMergeCommit {
			fmt.Fprintf(&builder, "Type: Merge Commit\n")
		}

		// Count failed rules manually instead of using FailedRulesCount()
		failedRules := 0

		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.Status == domain.StatusFailed {
				failedRules++
			}
		}

		if commitResult.Passed {
			fmt.Fprintf(&builder, "✅ All rules passed\n")
		} else {
			fmt.Fprintf(&builder, "❌ %d rules failed\n", failedRules)
		}

		// Print rule results
		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.Status == domain.StatusFailed {
				for _, err := range ruleResult.Errors {
					// Print error in GitHub format
					fmt.Fprintf(&builder, "::error file=%s,line=1,title=%s::%s\n",
						commitResult.CommitInfo.Hash,
						ruleResult.RuleName,
						err.Error())
				}
			}

			// Print rule status
			status := "✅"
			if ruleResult.Status == domain.StatusFailed {
				status = "❌"
			}

			message := ruleResult.Message
			if f.verbose {
				message = ruleResult.VerboseMessage
			}

			fmt.Fprintf(&builder, "%s %s: %s\n", status, ruleResult.RuleName, message)

			// Show rule help if requested - using HelpMessage instead of HelpText
			if f.showHelp && ruleResult.HelpMessage != "" {
				fmt.Fprintf(&builder, "   Help: %s\n", ruleResult.HelpMessage)
			}
		}

		builder.WriteString("::endgroup::\n")
	}

	// Set output for GitHub Actions
	if !validationResults.AllPassed() {
		builder.WriteString("::set-output name=passed::false\n")
	} else {
		builder.WriteString("::set-output name=passed::true\n")
	}

	return builder.String()
}

// FormatRuleHelp formats help text for a specific rule.
func (f GitHubActionsFormatter) FormatRuleHelp(ruleName string, results domain.ValidationResults) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "::group::Help for rule: %s\n", ruleName)

	// Find the rule in results
	found := false

	for _, commitResult := range results.Results {
		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.RuleName == ruleName {
				found = true
				// Using HelpMessage instead of HelpText
				fmt.Fprintf(&builder, "%s\n", ruleResult.HelpMessage)

				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		fmt.Fprintf(&builder, "No help available for rule: %s\n", ruleName)
	}

	builder.WriteString("::endgroup::\n")

	return builder.String()
}

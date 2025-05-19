// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package output

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/domain"
)

// GitLabCIFormatter formats validation results for GitLab CI.
// It implements the domain.ResultFormatter interface.
type GitLabCIFormatter struct {
	verbose  bool
	showHelp bool
}

// Ensure GitLabCIFormatter implements domain.ResultFormatter.
var _ domain.ResultFormatter = GitLabCIFormatter{}

// NewGitLabFormatter creates a new GitLab CI formatter.
// It implements domain.ResultFormatter interface.
func NewGitLabFormatter() GitLabCIFormatter {
	return GitLabCIFormatter{
		verbose:  false,
		showHelp: false,
	}
}

// WithVerbose sets the verbose flag for the formatter.
func (f GitLabCIFormatter) WithVerbose(verbose bool) GitLabCIFormatter {
	result := f
	result.verbose = verbose

	return result
}

// WithShowHelp sets the showHelp flag for the formatter.
func (f GitLabCIFormatter) WithShowHelp(showHelp bool) GitLabCIFormatter {
	result := f
	result.showHelp = showHelp

	return result
}

// Format formats validation results for GitLab CI output.
func (f GitLabCIFormatter) Format(ctx context.Context, results domain.ValidationResults) string {
	logger := log.Logger(ctx)
	logger.Trace().Bool("verbose", f.verbose).Bool("show_help", f.showHelp).Int("total_commits", results.TotalCommits).Msg("Entering GitLabCIFormatter.Format")

	var builder strings.Builder

	// Print summary
	builder.WriteString("section_start:$(date +%s):summary[collapsed=true]\n")
	fmt.Fprintf(&builder, "Validated %d commits\n", results.TotalCommits)
	fmt.Fprintf(&builder, "Passed: %d, Failed: %d\n", results.PassedCommits, results.TotalCommits-results.PassedCommits)
	builder.WriteString("section_end:$(date +%s):summary\n")

	// Print details for each commit
	for i, commitResult := range results.Results {
		sectionName := fmt.Sprintf("commit_%d", i+1)
		fmt.Fprintf(&builder, "section_start:$(date ):%s[collapsed=true]\n", sectionName)
		fmt.Fprintf(&builder, "Commit #%d: %s\n", i+1, commitResult.CommitInfo.Hash)
		fmt.Fprintf(&builder, "Subject: %s\n", commitResult.CommitInfo.Subject)

		if commitResult.CommitInfo.IsMergeCommit {
			fmt.Fprintf(&builder, "Type: Merge Commit\n")
		}

		// Count failed rules manually
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
					// Print error in GitLab format
					fmt.Fprintf(&builder, "ERROR: %s:1:%s - %s\n",
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

			// Show rule help if requested
			if f.showHelp && ruleResult.HelpMessage != "" {
				fmt.Fprintf(&builder, "   Help: %s\n", ruleResult.HelpMessage)
			}
		}

		fmt.Fprintf(&builder, "section_end:$(date ):%s\n", sectionName)
	}

	return builder.String()
}

// FormatRuleHelp formats help text for a specific rule.
func (f GitLabCIFormatter) FormatRuleHelp(ruleName string, results domain.ValidationResults) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "section_start:$(date ):rulehelp[collapsed=true]\n")
	fmt.Fprintf(&builder, "Help for rule: %s\n", ruleName)

	// Find the rule in results
	found := false

	for _, commitResult := range results.Results {
		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.RuleName == ruleName {
				found = true

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

	builder.WriteString("section_end:$(date +%s):rulehelp\n")

	return builder.String()
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package output provides output formatting for validation results.
package output

import (
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
)

// TextFormatter formats validation results as text.
type TextFormatter struct {
	verbose  bool
	showHelp bool
}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter(verbose, showHelp bool, _ ...bool) *TextFormatter {
	return &TextFormatter{
		verbose:  verbose,
		showHelp: showHelp,
	}
}

// Format formats validation results as text.
func (f *TextFormatter) Format(results *domain.ValidationResults) string {
	var builder strings.Builder

	// Summary of results
	builder.WriteString("Validation Results\n")
	builder.WriteString("-----------------\n")

	// Status
	if results.AllPassed() {
		builder.WriteString("Status: ✅ PASSED\n")
	} else {
		builder.WriteString("Status: ❌ FAILED\n")
	}

	// Counts
	builder.WriteString("Total commits: ")
	builder.WriteString(strconv.Itoa(results.TotalCommits))
	builder.WriteString("\n")
	builder.WriteString("Passed commits: ")
	builder.WriteString(strconv.Itoa(results.PassedCommits))
	builder.WriteString("\n")
	builder.WriteString("Failed commits: ")
	builder.WriteString(strconv.Itoa(results.TotalCommits - results.PassedCommits))
	builder.WriteString("\n\n")

	// Commit details
	if len(results.CommitResults) > 0 {
		builder.WriteString("Commit Details\n")
		builder.WriteString("--------------\n")

		for _, commitResult := range results.CommitResults {
			// Commit header
			builder.WriteString("Commit: ")
			builder.WriteString(commitResult.CommitInfo.Hash)

			if len(commitResult.CommitInfo.Hash) > 8 {
				builder.WriteString(" (")
				builder.WriteString(commitResult.CommitInfo.Hash[:8])
				builder.WriteString(")")
			}

			builder.WriteString("\n")

			// Commit message
			builder.WriteString("Subject: ")
			builder.WriteString(commitResult.CommitInfo.Subject)
			builder.WriteString("\n")

			// Status
			if commitResult.Passed {
				builder.WriteString("Status: ✅ PASSED\n")
			} else {
				builder.WriteString("Status: ❌ FAILED\n")
			}

			// Rule results
			if len(commitResult.RuleResults) > 0 {
				builder.WriteString("Rules:\n")

				for _, ruleResult := range commitResult.RuleResults {
					// Rule status
					if ruleResult.Status == domain.StatusPassed {
						builder.WriteString("  ✅ ")
					} else {
						builder.WriteString("  ❌ ")
					}

					// Rule name
					builder.WriteString(ruleResult.RuleName)
					builder.WriteString(": ")

					// Rule message
					if f.verbose {
						builder.WriteString(ruleResult.VerboseMessage)
					} else {
						builder.WriteString(ruleResult.Message)
					}

					builder.WriteString("\n")

					// Rule errors
					if ruleResult.Status == domain.StatusFailed && len(ruleResult.Errors) > 0 {
						for _, err := range ruleResult.Errors {
							builder.WriteString("    - ")
							builder.WriteString(err.Message)
							builder.WriteString("\n")

							// Show context if verbose
							if f.verbose && len(err.Context) > 0 {
								for key, value := range err.Context {
									builder.WriteString("      ")
									builder.WriteString(key)
									builder.WriteString(": ")
									builder.WriteString(value)
									builder.WriteString("\n")
								}
							}
						}

						// Show help if requested
						if f.showHelp {
							builder.WriteString("    Help: ")
							builder.WriteString(ruleResult.HelpMessage)
							builder.WriteString("\n")
						}
					}
				}
			}

			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// FormatRuleHelp formats help information for a specific rule.
func (f *TextFormatter) FormatRuleHelp(ruleName string, results *domain.ValidationResults) string {
	var builder strings.Builder

	builder.WriteString("Help for rule: ")
	builder.WriteString(ruleName)
	builder.WriteString("\n")
	builder.WriteString("-----------------\n\n")

	// Try to find help information in the results
	found := false

	for _, commitResult := range results.CommitResults {
		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.RuleName == ruleName {
				builder.WriteString("Description: ")
				builder.WriteString(ruleResult.Message)
				builder.WriteString("\n\n")

				builder.WriteString("Help: ")
				builder.WriteString(ruleResult.HelpMessage)
				builder.WriteString("\n\n")

				// Show examples for common rules
				switch ruleName {
				case "ConventionalCommitRule":
					builder.WriteString("Examples:\n")
					builder.WriteString("  Good: feat(api): add new endpoint for user authentication\n")
					builder.WriteString("  Good: fix: resolve null pointer exception in login flow\n")
					builder.WriteString("  Bad: Added new feature\n")
					builder.WriteString("  Bad: fix stuff\n")
				case "SubjectLengthRule":
					builder.WriteString("Examples:\n")
					builder.WriteString("  Good: Add user authentication feature\n")
					builder.WriteString("  Bad: Add extremely long and unnecessarily verbose description of a simple change that could be described much more concisely\n")
				case "ImperativeVerbRule":
					builder.WriteString("Examples:\n")
					builder.WriteString("  Good: Add feature X\n")
					builder.WriteString("  Good: Fix bug Y\n")
					builder.WriteString("  Bad: Added feature X\n")
					builder.WriteString("  Bad: Fixing bug Y\n")
				}

				found = true

				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		builder.WriteString("No detailed help available for this rule.\n")

		// Generic help based on rule name
		switch ruleName {
		case "ConventionalCommitRule":
			builder.WriteString("\nConventional Commit Rule enforces the conventional commit format:\n")
			builder.WriteString("  <type>[optional scope]: <description>\n\n")
			builder.WriteString("Common types: feat, fix, docs, style, refactor, test, chore\n")
		case "SubjectLengthRule":
			builder.WriteString("\nSubject Length Rule ensures that commit subject lines are of appropriate length.\n")
			builder.WriteString("Usually the maximum length is 50-100 characters.\n")
		case "CommitBodyRule":
			builder.WriteString("\nCommit Body Rule checks that commits have a proper body section.\n")
			builder.WriteString("A good commit should explain the change in detail in the body.\n")
		case "ImperativeVerbRule":
			builder.WriteString("\nImperative Verb Rule ensures that commit subjects use the imperative mood.\n")
			builder.WriteString("Use 'Add feature' instead of 'Added feature' or 'Adds feature'.\n")
		}
	}

	return builder.String()
}

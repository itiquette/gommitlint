// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package formatting provides pure functions for formatting rule validation results.
package formatting

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/errors"
)

// FormatResult creates a concise result message for a rule.
func FormatResult(ruleName string, errs []errors.ValidationError) string {
	if len(errs) == 0 {
		return ruleName + ": Passed"
	}

	return fmt.Sprintf("%s: Failed with %d error(s)", ruleName, len(errs))
}

// FormatVerboseResult creates a detailed result message.
func FormatVerboseResult(ruleName string, errs []errors.ValidationError) string {
	if len(errs) == 0 {
		return ruleName + ": All checks passed"
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s: Found %d error(s):\n", ruleName, len(errs)))

	for i, err := range errs {
		builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Message))

		help := err.GetHelp()
		if help != "" {
			builder.WriteString(fmt.Sprintf("     Help: %s\n", help))
		}
	}

	return builder.String()
}

// FormatHelp creates help text based on errors or provides default help.
func FormatHelp(ruleName string, errs []errors.ValidationError) string {
	// Collect unique help messages from errors
	helpTexts := make(map[string]bool)

	for _, err := range errs {
		if help := err.GetHelp(); help != "" {
			helpTexts[help] = true
		}
	}

	// If we have specific help from errors, use that
	if len(helpTexts) > 0 {
		var helps []string
		for help := range helpTexts {
			helps = append(helps, help)
		}

		return strings.Join(helps, "\n")
	}

	// Otherwise return default help for the rule
	return GetDefaultHelp(ruleName)
}

// GetDefaultHelp returns default help text for a rule.
func GetDefaultHelp(ruleName string) string {
	switch ruleName {
	case "SubjectLength":
		return "Keep commit subjects under the configured maximum length (default: 50 characters)"
	case "ImperativeVerb":
		return "Use imperative mood in commit subjects (e.g., 'Add feature' not 'Added feature')"
	case "CommitBody":
		return "Include a commit body to provide context for your changes"
	case "Conventional":
		return "Follow conventional commit format: type(scope): description"
	case "JiraReference":
		return "Include JIRA ticket reference in your commit message"
	case "Signature":
		return "Sign your commits using GPG or SSH"
	case "SignedIdentity":
		return "Ensure commit signature matches the configured identity"
	case "SignOff":
		return "Add sign-off to your commit (Signed-off-by: name <email>)"
	case "SubjectCase":
		return "Use the configured case style for commit subjects"
	case "SubjectSuffix":
		return "Avoid disallowed suffixes in commit subjects (e.g., periods)"
	case "CommitsAhead":
		return "Too many commits ahead - consider squashing or rebasing"
	case "SpellCheck":
		return "Fix spelling errors in your commit message"
	default:
		return fmt.Sprintf("Follow the %s rule guidelines", ruleName)
	}
}

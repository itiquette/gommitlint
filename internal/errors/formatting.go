// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// WithHelp adds help text to an existing ValidationError.
func WithHelp(err ValidationError, help string) ValidationError {
	// Create a copy of the existing error
	result := err

	// Set the help field directly
	result.Help = help

	return result
}

// GetHelp retrieves the help message from the error.
func (e ValidationError) GetHelp() string {
	return e.Help
}

// NewFormatValidationError creates a rich error for format validation failures.
// This is the only helper actually used in the codebase.
func NewFormatValidationError(ruleName string, message string, helpText string, subject string) ValidationError {
	err := New(ruleName, ErrInvalidFormat, message)
	err = err.WithHelp(helpText)
	err = err.WithContext("subject", subject)

	return err
}

// NewMissingJiraError creates a standardized error for missing JIRA references.
func NewMissingJiraError(ruleName string, commitMsg string, expectedPattern string) ValidationError {
	err := New(ruleName, ErrMissingJira, "Missing JIRA reference")
	err = err.WithUserMessage("Commit message must contain a JIRA issue reference")
	err = err.WithHelp(fmt.Sprintf("Include a JIRA reference like %s in your commit message", expectedPattern))
	err = err.WithContextMap(map[string]string{
		"commit_message": commitMsg,
		"pattern":        expectedPattern,
	})

	return err
}

// NewNonImperativeError creates a standardized error for non-imperative verbs.
func NewNonImperativeError(ruleName string, word string, suggestions []string) ValidationError {
	err := New(ruleName, ErrNonImperative, fmt.Sprintf("Word '%s' is not in imperative mood", word))
	err = err.WithUserMessage("Commit subject should start with an imperative verb")

	help := fmt.Sprintf("Use the imperative form of '%s'", word)
	if len(suggestions) > 0 {
		help = "Try: " + strings.Join(suggestions, ", ")
	}

	err = err.WithHelp(help)

	err = err.WithContextMap(map[string]string{
		"word":        word,
		"suggestions": strings.Join(suggestions, ", "),
	})

	return err
}

// NewBodyError creates a standardized error for commit body issues.
func NewBodyError(code ValidationErrorCode, ruleName string, issue string, actual string, requirement string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(fmt.Sprintf("Your commit body %s. %s", actual, requirement))
	err = err.WithContextMap(map[string]string{
		"issue":       issue,
		"actual":      actual,
		"requirement": requirement,
	})

	return err
}

// NewConventionalCommitError creates a standardized error for conventional commit format issues.
func NewConventionalCommitError(code ValidationErrorCode, ruleName string, issue string, format string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp("Use the format: " + format)
	err = err.WithContext("expected_format", format)

	return err
}

// NewSpellingError creates a standardized error for spelling mistakes.
func NewSpellingError(ruleName string, word string, position int) ValidationError {
	err := New(ruleName, ErrSpellCheckFailed, fmt.Sprintf("Spelling error: '%s'", word))
	err = err.WithUserMessage("Misspelled word: '%s'", word)
	err = err.WithHelp(fmt.Sprintf("Check the spelling of '%s' or add it to the ignore list if it's a technical term", word))
	err = err.WithContextMap(map[string]string{
		"word":     word,
		"position": strconv.Itoa(position),
	})

	return err
}

// NewSignatureError creates a standardized error for signature issues.
func NewSignatureError(code ValidationErrorCode, ruleName string, issue string, requirement string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(requirement)

	return err
}

// NewIdentityError creates a standardized error for identity verification issues.
func NewIdentityError(code ValidationErrorCode, ruleName string, issue string, help string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(help)

	return err
}

// NewLengthError creates a standardized error for length violations.
func NewLengthError(code ValidationErrorCode, ruleName string, issue string, actual int, maximum int) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(fmt.Sprintf("Keep it under %d characters", maximum))
	err = err.WithContextMap(map[string]string{
		"actual": strconv.Itoa(actual),
		"max":    strconv.Itoa(maximum),
	})

	return err
}

// NewCaseError creates a standardized error for case violations.
func NewCaseError(code ValidationErrorCode, ruleName string, issue string, expected string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp("Use " + expected)

	return err
}

// NewSuffixError creates a standardized error for suffix violations.
func NewSuffixError(code ValidationErrorCode, ruleName string, issue string, help string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(help)

	return err
}

// NewBranchError creates a standardized error for branch issues.
func NewBranchError(code ValidationErrorCode, ruleName string, issue string, help string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(help)

	return err
}

// NewSignOffError creates a standardized error for sign-off issues.
func NewSignOffError(code ValidationErrorCode, ruleName string, issue string, help string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(help)

	return err
}

// NewJiraError creates a standardized error for JIRA-related issues.
func NewJiraError(code ValidationErrorCode, ruleName string, issue string, help string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(help)

	return err
}

// NewValidationError creates a generic validation error.
func NewValidationError(code ValidationErrorCode, ruleName string, issue string, help string) ValidationError {
	err := New(ruleName, code, issue)
	err = err.WithUserMessage(issue)
	err = err.WithHelp(help)

	return err
}

// FormatAtLevel formats the error with progressive detail based on verbosity level.
// Level 0: Basic message and help.
// Level 1: Add rule and key details.
// Level 2+: Show all context.
func (e ValidationError) FormatAtLevel(level int) string {
	switch level {
	case 0:
		return e.formatBasic()
	case 1:
		return e.formatWithRule()
	default:
		return e.formatDetailed()
	}
}

// formatBasic shows just the error message and help text.
func (e ValidationError) formatBasic() string {
	// Just the error and help
	out := e.Message
	if e.Help != "" {
		out += "\n  " + e.Help
	}

	return out
}

// formatWithRule adds rule information and key context details.
func (e ValidationError) formatWithRule() string {
	// Start with basic format
	out := e.formatBasic()
	out += "\n  Rule: " + e.Rule

	// Show only important context keys
	important := []string{"actual", "expected", "max", "min", "detected", "suggested"}
	hasDetails := false

	for _, key := range important {
		if val, exists := e.Context[key]; exists {
			if !hasDetails {
				out += "\n  Details:"
				hasDetails = true
			}

			out += fmt.Sprintf("\n    %s: %s", key, val)
		}
	}

	return out
}

// formatDetailed shows all available information.
func (e ValidationError) formatDetailed() string {
	// Start with rule format
	out := e.formatWithRule()

	// Add all remaining context not already shown
	shown := map[string]bool{
		"actual": true, "expected": true,
		"max": true, "min": true,
		"detected": true, "suggested": true,
	}

	// Check if we have additional context to show
	hasAdditional := false

	for k := range e.Context {
		if !shown[k] {
			hasAdditional = true

			break
		}
	}

	if hasAdditional {
		// Sort keys for consistent output
		var keys []string

		for k := range e.Context {
			if !shown[k] {
				keys = append(keys, k)
			}
		}

		sort.Strings(keys)

		// Add remaining context
		for _, k := range keys {
			out += fmt.Sprintf("\n    %s: %s", k, e.Context[k])
		}
	}

	// Add error code at the highest verbosity
	out += "\n  Error Code: " + e.Code

	return out
}

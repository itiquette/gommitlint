// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"encoding/json"
	"fmt"
	"sort"
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

// FormatAsText formats the error as text for CLI output.
func (e ValidationError) FormatAsText(verbose bool) string {
	if !verbose {
		return fmt.Sprintf("[%s] %s", e.Rule, e.Message)
	}

	// More detailed output for verbose mode
	var builder strings.Builder

	fmt.Fprintf(&builder, "Rule:    %s\n", e.Rule)
	fmt.Fprintf(&builder, "Code:    %s\n", e.Code)
	fmt.Fprintf(&builder, "Message: %s\n", e.Message)

	if e.Help != "" {
		fmt.Fprintf(&builder, "Help:    %s\n", e.Help)
	}

	// Print context if available
	if len(e.Context) > 0 {
		fmt.Fprintf(&builder, "Context:\n")

		// Get keys and sort them for consistent output
		contextKeys := make([]string, 0, len(e.Context))
		for k := range e.Context {
			contextKeys = append(contextKeys, k)
		}

		sort.Strings(contextKeys)

		// Format and write each context entry
		for _, k := range contextKeys {
			fmt.Fprintf(&builder, "  %s: %s\n", k, e.Context[k])
		}
	}

	return builder.String()
}

// FormatAsJSON formats the error as JSON.
func (e ValidationError) FormatAsJSON() ([]byte, error) {
	// Create a representation with help as a first-class field
	representation := struct {
		Rule    string            `json:"rule"`
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Help    string            `json:"help,omitempty"`
		Context map[string]string `json:"context,omitempty"`
	}{
		Rule:    e.Rule,
		Code:    e.Code,
		Message: e.Message,
		Help:    e.Help,
		Context: e.Context,
	}

	return json.Marshal(representation)
}

// FormatAsMarkdown formats the error as Markdown for documentation.
func (e ValidationError) FormatAsMarkdown() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "### %s: %s\n\n", e.Rule, e.Message)
	fmt.Fprintf(&builder, "**Code:** `%s`\n\n", e.Code)

	if e.Help != "" {
		fmt.Fprintf(&builder, "**Help:** %s\n\n", e.Help)
	}

	// Print all context entries
	if len(e.Context) > 0 {
		fmt.Fprintf(&builder, "**Context:**\n\n")

		// Format each context entry as a Markdown list item
		contextEntries := make([]string, 0, len(e.Context))
		for k, v := range e.Context {
			contextEntries = append(contextEntries, fmt.Sprintf("- `%s`: %s", k, v))
		}

		// Sort for consistent output
		sort.Strings(contextEntries)

		// Join all formatted entries with newlines
		builder.WriteString(strings.Join(contextEntries, "\n"))
		builder.WriteString("\n\n")
	}

	return builder.String()
}

// NewFormatValidationError creates a rich error for format validation failures.
// This is the only helper actually used in the codebase.
func NewFormatValidationError(ruleName string, message string, helpText string, subject string) ValidationError {
	err := New(ruleName, ErrInvalidFormat, message)
	err = err.WithHelp(helpText)
	err = err.WithContext("subject", subject)

	return err
}

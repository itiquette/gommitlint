// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EnhanceValidationError adds help text to an existing ValidationError.
func EnhanceValidationError(err ValidationError, help string) ValidationError {
	// Create a copy of the existing error
	result := err

	// Add help to the context
	result = result.WithContext("help", help)

	return result
}

// WithCommitSHA adds a commit SHA to the error context.
func (e ValidationError) WithCommitSHA(commitSHA string) ValidationError {
	return e.WithContext("commit_sha", commitSHA)
}

// GetHelp retrieves the help message from the error context.
func (e ValidationError) GetHelp() string {
	if help, ok := e.Context["help"]; ok {
		return help
	}

	return ""
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

	help := e.GetHelp()
	if help != "" {
		fmt.Fprintf(&builder, "Help:    %s\n", help)
	}

	// Skip help when printing context
	if len(e.Context) > 0 {
		fmt.Fprintf(&builder, "Context:\n")

		for k, v := range e.Context {
			if k != "help" {
				fmt.Fprintf(&builder, "  %s: %s\n", k, v)
			}
		}
	}

	return builder.String()
}

// FormatAsJSON formats the error as JSON.
func (e ValidationError) FormatAsJSON() ([]byte, error) {
	// Create a representation with help extracted from context
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
		Help:    e.GetHelp(),
		Context: make(map[string]string),
	}

	// Copy context excluding help
	for k, v := range e.Context {
		if k != "help" {
			representation.Context[k] = v
		}
	}

	return json.Marshal(representation)
}

// FormatAsMarkdown formats the error as Markdown for documentation.
func (e ValidationError) FormatAsMarkdown() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "### %s: %s\n\n", e.Rule, e.Message)
	fmt.Fprintf(&builder, "**Code:** `%s`\n\n", e.Code)

	help := e.GetHelp()
	if help != "" {
		fmt.Fprintf(&builder, "**Help:** %s\n\n", help)
	}

	// Add context section if needed (excluding help)
	hasContext := false

	for k := range e.Context {
		if k != "help" {
			hasContext = true

			break
		}
	}

	if hasContext {
		fmt.Fprintf(&builder, "**Context:**\n\n")

		for k, v := range e.Context {
			if k != "help" {
				fmt.Fprintf(&builder, "- `%s`: %s\n", k, v)
			}
		}

		fmt.Fprintf(&builder, "\n")
	}

	return builder.String()
}

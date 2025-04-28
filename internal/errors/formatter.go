// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorFormatter defines an interface for formatting validation errors.
type ErrorFormatter interface {
	FormatError(err ValidationError) string
	FormatErrors(errs []ValidationError) string
}

// TextFormatter formats errors as text.
type TextFormatter struct {
	Verbose bool
}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter(verbose bool) TextFormatter {
	return TextFormatter{
		Verbose: verbose,
	}
}

// FormatError formats a single error as text.
func (f TextFormatter) FormatError(err ValidationError) string {
	return err.FormatAsText(f.Verbose)
}

// FormatErrors formats multiple errors as text.
func (f TextFormatter) FormatErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "No validation errors found."
	}

	var result string

	for index, err := range errs {
		if index > 0 {
			result += "\n\n"
		}

		result += f.FormatError(err)
	}

	return result
}

// JSONFormatter formats errors as JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() JSONFormatter {
	return JSONFormatter{}
}

// FormatError formats a single error as JSON.
func (f JSONFormatter) FormatError(err ValidationError) string {
	data, jsonErr := err.FormatAsJSON()
	if jsonErr != nil {
		return fmt.Sprintf(`{"error":"Failed to format error as JSON: %s"}`, jsonErr)
	}

	return string(data)
}

// FormatErrors formats multiple errors as JSON.
func (f JSONFormatter) FormatErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return `{"errors":[],"count":0}`
	}

	type errorRepresentation struct {
		Rule    string            `json:"rule"`
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Help    string            `json:"help,omitempty"`
		Context map[string]string `json:"context,omitempty"`
	}

	representations := make([]errorRepresentation, len(errs))

	for index, err := range errs {
		helpText := ""
		if help, ok := err.Context["help"]; ok {
			helpText = help
		}

		// Copy context excluding help
		contextMap := make(map[string]string)

		for k, v := range err.Context {
			if k != "help" {
				contextMap[k] = v
			}
		}

		representations[index] = errorRepresentation{
			Rule:    err.Rule,
			Code:    err.Code,
			Message: err.Message,
			Help:    helpText,
			Context: contextMap,
		}
	}

	data, err := json.Marshal(map[string]interface{}{
		"errors": representations,
		"count":  len(errs),
	})

	if err != nil {
		return fmt.Sprintf(`{"error":"Failed to format errors as JSON: %s"}`, err)
	}

	return string(data)
}

// MarkdownFormatter formats errors as Markdown.
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new Markdown formatter.
func NewMarkdownFormatter() MarkdownFormatter {
	return MarkdownFormatter{}
}

// FormatError formats a single error as Markdown.
func (f MarkdownFormatter) FormatError(err ValidationError) string {
	return err.FormatAsMarkdown()
}

// FormatErrors formats multiple errors as Markdown.
func (f MarkdownFormatter) FormatErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "# No validation errors found."
	}

	var result string

	result = fmt.Sprintf("# Validation Errors (%d)\n\n", len(errs))

	for _, err := range errs {
		result += err.FormatAsMarkdown() + "\n"
	}

	return result
}

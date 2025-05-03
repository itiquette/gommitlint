// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"encoding/json"
	"fmt"

	"github.com/itiquette/gommitlint/internal/contextx"
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

	formatted := contextx.Map(errs, func(err ValidationError) string {
		return f.FormatError(err)
	})

	return contextx.Reduce(formatted, "", func(acc string, errStr string) string {
		if acc == "" {
			return errStr
		}

		return acc + "\n\n" + errStr
	})
}

// JSONFormatter formats errors as JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() JSONFormatter {
	return JSONFormatter{}
}

// FormatError formats a single error as JSON.
func (JSONFormatter) FormatError(err ValidationError) string {
	data, jsonErr := err.FormatAsJSON()
	if jsonErr != nil {
		return fmt.Sprintf(`{"error":"Failed to format error as JSON: %s"}`, jsonErr)
	}

	return string(data)
}

// FormatErrors formats multiple errors as JSON.
func (JSONFormatter) FormatErrors(errs []ValidationError) string {
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

	// Transform ValidationError to errorRepresentation using Map
	representations := contextx.Map(errs, func(err ValidationError) errorRepresentation {
		helpText := ""
		if help, ok := err.Context["help"]; ok {
			helpText = help
		}

		// Filter context excluding help
		contextMap := contextx.Reduce(
			contextx.Map(
				contextx.Filter(
					contextx.Keys(err.Context),
					func(keyName string) bool { return keyName != "help" },
				),
				func(keyName string) struct {
					key   string
					value string
				} {
					return struct {
						key   string
						value string
					}{
						key:   keyName,
						value: err.Context[keyName],
					}
				},
			),
			make(map[string]string),
			func(acc map[string]string, pair struct {
				key   string
				value string
			}) map[string]string {
				result := contextx.DeepCopyMap(acc)
				result[pair.key] = pair.value

				return result
			},
		)

		return errorRepresentation{
			Rule:    err.Rule,
			Code:    err.Code,
			Message: err.Message,
			Help:    helpText,
			Context: contextMap,
		}
	})

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
func (MarkdownFormatter) FormatError(err ValidationError) string {
	return err.FormatAsMarkdown()
}

// FormatErrors formats multiple errors as Markdown.
func (MarkdownFormatter) FormatErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "# No validation errors found."
	}

	header := fmt.Sprintf("# Validation Errors (%d)\n\n", len(errs))

	formattedErrors := contextx.Map(errs, func(err ValidationError) string {
		return err.FormatAsMarkdown()
	})

	errorContent := contextx.Reduce(formattedErrors, "", func(acc string, errStr string) string {
		return acc + errStr + "\n"
	})

	return header + errorContent
}

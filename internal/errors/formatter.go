// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/slices"
)

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
	// Map verbose flag to level
	level := 0
	if f.Verbose {
		level = 2 // Full details
	}

	return err.FormatAtLevel(level)
}

// FormatErrors formats multiple errors as text.
func (f TextFormatter) FormatErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "No validation errors found."
	}

	formatted := slices.Map(errs, func(err ValidationError) string {
		return f.FormatError(err)
	})

	return slices.Reduce(formatted, "", func(acc string, errStr string) string {
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
	// Create a representation with help as a first-class field
	representation := struct {
		Rule    string            `json:"rule"`
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Help    string            `json:"help,omitempty"`
		Context map[string]string `json:"context,omitempty"`
	}{
		Rule:    err.Rule,
		Code:    err.Code,
		Message: err.Message,
		Help:    err.Help,
		Context: err.Context,
	}

	data, jsonErr := json.Marshal(representation)
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
	representations := slices.Map(errs, func(err ValidationError) errorRepresentation {
		helpText := ""
		if help, ok := err.Context["help"]; ok {
			helpText = help
		}

		// Filter context excluding help using FilterMapKeys
		contextMap := slices.FilterMapKeys(err.Context, []string{"help"})

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
	var builder strings.Builder

	fmt.Fprintf(&builder, "### %s: %s\n\n", err.Rule, err.Message)
	fmt.Fprintf(&builder, "**Code:** `%s`\n\n", err.Code)

	if err.Help != "" {
		fmt.Fprintf(&builder, "**Help:** %s\n\n", err.Help)
	}

	// Print all context entries
	if len(err.Context) > 0 {
		fmt.Fprintf(&builder, "**Context:**\n\n")

		// Format each context entry as a Markdown list item
		contextEntries := make([]string, 0, len(err.Context))
		for k, v := range err.Context {
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

// FormatErrors formats multiple errors as Markdown.
func (MarkdownFormatter) FormatErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "# No validation errors found."
	}

	header := fmt.Sprintf("# Validation Errors (%d)\n\n", len(errs))

	formattedErrors := slices.Map(errs, func(err ValidationError) string {
		return MarkdownFormatter{}.FormatError(err)
	})

	errorContent := slices.Reduce(formattedErrors, "", func(acc string, errStr string) string {
		return acc + errStr + "\n"
	})

	return header + errorContent
}

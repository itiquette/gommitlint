// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// FormatterAdapter adapts various output formatters to the domain.ResultFormatter interface.
// It serves as a factory for creating appropriate formatters based on format type.
type FormatterAdapter struct {
	format     string
	verbose    bool
	lightMode  bool
	showHelp   bool
	targetRule string
}

// NewFormatterAdapter creates a new formatter adapter.
func NewFormatterAdapter(format string, options ...FormatterOption) *FormatterAdapter {
	adapter := &FormatterAdapter{
		format:     format,
		verbose:    false,
		lightMode:  false,
		showHelp:   false,
		targetRule: "",
	}

	// Apply options
	for _, option := range options {
		option(adapter)
	}

	return adapter
}

// FormatterOption is a function that configures a FormatterAdapter.
type FormatterOption func(*FormatterAdapter)

// WithVerbose sets the verbose flag.
func WithVerbose(verbose bool) FormatterOption {
	return func(a *FormatterAdapter) {
		a.verbose = verbose
	}
}

// WithLightMode sets the light mode flag.
func WithLightMode(lightMode bool) FormatterOption {
	return func(a *FormatterAdapter) {
		a.lightMode = lightMode
	}
}

// WithShowHelp sets the show help flag.
func WithShowHelp(showHelp bool) FormatterOption {
	return func(a *FormatterAdapter) {
		a.showHelp = showHelp
	}
}

// WithTargetRule sets the specific rule to show help for.
func WithTargetRule(ruleName string) FormatterOption {
	return func(a *FormatterAdapter) {
		a.targetRule = ruleName
	}
}

// Format formats the validation results according to the configured format.
func (a *FormatterAdapter) Format(results domain.ValidationResults) string {
	var formatter domain.ResultFormatter

	switch a.format {
	case "json":
		formatter = NewJSONFormatter()
	case "github":
		formatter = NewGitHubFormatter()
	case "gitlab":
		formatter = NewGitLabFormatter()
	default:
		// Default to text format
		formatter = NewTextFormatter(a.verbose, a.showHelp, a.lightMode)
	}

	// Delegate formatting to the appropriate formatter
	return formatter.Format(results)
}

// Ensure FormatterAdapter implements domain.ResultFormatter.
var _ domain.ResultFormatter = (*FormatterAdapter)(nil)

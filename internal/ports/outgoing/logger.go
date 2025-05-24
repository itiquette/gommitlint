// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package outgoing

import "context"

// Logger defines the minimal interface for error logging operations.
// This is an outgoing port that will be implemented by adapters.
type Logger interface {
	Debug(msg string, kvs ...interface{})
	Info(msg string, kvs ...interface{})
	Warn(msg string, kvs ...interface{})
	Error(msg string, kvs ...interface{})
}

// ResultFormatter defines an interface for formatting validation results.
// This interface decouples the domain from specific output formats.
type ResultFormatter interface {
	// Format converts validation results to a formatted string.
	Format(ctx context.Context, results interface{}) string
}

// ReportGenerator defines an interface for generating reports from validation results.
// This decouples the application layer from specific report implementations.
//
// Follows functional programming principles with value semantics.
type ReportGenerator interface {
	// GenerateReport creates a report from validation results.
	GenerateReport(ctx context.Context, results interface{}) error

	// GenerateSummary creates a brief summary report.
	GenerateSummary(ctx context.Context, results interface{}) error

	// WithVerbose returns a new generator with verbose setting updated.
	WithVerbose(verbose bool) ReportGenerator

	// WithShowHelp returns a new generator with showHelp setting updated.
	WithShowHelp(showHelp bool) ReportGenerator

	// WithRuleToShowHelp returns a new generator with ruleToShowHelp setting updated.
	WithRuleToShowHelp(ruleName string) ReportGenerator
}

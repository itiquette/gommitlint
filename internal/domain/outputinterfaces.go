// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import "context"

// ResultFormatter defines an interface for formatting validation results.
// This interface decouples the domain from specific output formats.
type ResultFormatter interface {
	// Format converts validation results to a formatted string.
	Format(ctx context.Context, results ValidationResults) string
}

// ReportGenerator defines an interface for generating reports from validation results.
// This decouples the application layer from specific report implementations.
//
// Follows functional programming principles with value semantics.
type ReportGenerator interface {
	// GenerateReport creates a report from validation results.
	GenerateReport(ctx context.Context, results ValidationResults) error

	// GenerateSummary creates a brief summary report.
	GenerateSummary(ctx context.Context, results ValidationResults) error

	// WithVerbose returns a new generator with verbose setting updated.
	WithVerbose(verbose bool) ReportGenerator

	// WithShowHelp returns a new generator with showHelp setting updated.
	WithShowHelp(showHelp bool) ReportGenerator

	// WithRuleToShowHelp returns a new generator with ruleToShowHelp setting updated.
	WithRuleToShowHelp(ruleName string) ReportGenerator
}

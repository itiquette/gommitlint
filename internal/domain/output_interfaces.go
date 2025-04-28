// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// ResultFormatter defines an interface for formatting validation results.
// This interface decouples the domain from specific output formats.
type ResultFormatter interface {
	// Format converts validation results to a formatted string.
	Format(results ValidationResults) string
}

// ReportGenerator defines an interface for generating reports from validation results.
// This decouples the application layer from specific report implementations.
//
// Note: This interface maintains both imperative methods (for backward compatibility)
// and functional methods (for new code). New implementations should focus on the
// functional methods and provide dummy implementations of the imperative methods.
type ReportGenerator interface {
	// GenerateReport creates a report from validation results.
	GenerateReport(results ValidationResults) error

	// SetVerbose enables or disables verbose output in reports.
	// Deprecated: Use WithVerbose for functional style instead.
	SetVerbose(verbose bool)

	// SetShowHelp enables or disables showing help messages in reports.
	// Deprecated: Use WithShowHelp for functional style instead.
	SetShowHelp(showHelp bool)

	// SetRuleToShowHelp sets a specific rule to show help for.
	// Deprecated: Use WithRuleToShowHelp for functional style instead.
	SetRuleToShowHelp(ruleName string)
}

// FunctionalReportGenerator extends ReportGenerator with functional methods.
// This is a supplementary interface to be used alongside ReportGenerator for
// implementations that support functional programming patterns.
//
// Implementations should use value semantics and return new instances for all
// configuration changes, following functional programming principles.
type FunctionalReportGenerator interface {
	ReportGenerator

	// WithVerbose returns a new generator with verbose setting updated.
	WithVerbose(verbose bool) FunctionalReportGenerator

	// WithShowHelp returns a new generator with showHelp setting updated.
	WithShowHelp(showHelp bool) FunctionalReportGenerator

	// WithRuleToShowHelp returns a new generator with ruleToShowHelp setting updated.
	WithRuleToShowHelp(ruleName string) FunctionalReportGenerator
}

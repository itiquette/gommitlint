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
type ReportGenerator interface {
	// GenerateReport creates a report from validation results.
	GenerateReport(results ValidationResults) error

	// SetVerbose enables or disables verbose output in reports.
	SetVerbose(verbose bool)

	// SetShowHelp enables or disables showing help messages in reports.
	SetShowHelp(showHelp bool)

	// SetRuleToShowHelp sets a specific rule to show help for.
	SetRuleToShowHelp(ruleName string)
}

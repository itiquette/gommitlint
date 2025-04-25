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

// RepositoryFactory defines an interface for creating repository-related services.
// This interface follows the Abstract Factory pattern to decouple the application layer
// from specific repository implementations.
//
// The factory allows the application to obtain access to different repository services
// without knowing how they are implemented, which supports the Dependency Inversion Principle.
type RepositoryFactory interface {
	// CreateGitCommitService creates a commit service for accessing git commits.
	CreateGitCommitService() GitCommitService

	// CreateInfoProvider creates an info provider for repository information.
	CreateInfoProvider() RepositoryInfoProvider

	// CreateCommitAnalyzer creates a commit analyzer for repository analysis.
	CreateCommitAnalyzer() CommitAnalyzer

	// CreateFullService creates a complete repository service that combines
	// all the capabilities of the specialized services.
	CreateFullService() GitRepositoryService
}

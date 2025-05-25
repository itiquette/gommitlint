// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package orchestration provides high-level coordination of application use cases.
package orchestration

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// ValidationOrchestrator orchestrates the validation workflow.
// It coordinates between validation and report generation.
type ValidationOrchestrator struct {
	validationService incoming.ValidationService
	formatter         outgoing.ResultFormatter
	logger            outgoing.Logger
}

// NewValidationOrchestrator creates a new validation orchestrator.
func NewValidationOrchestrator(
	validationService incoming.ValidationService,
	formatter outgoing.ResultFormatter,
	logger outgoing.Logger,
) ValidationOrchestrator {
	return ValidationOrchestrator{
		validationService: validationService,
		formatter:         formatter,
		logger:            logger,
	}
}

// ValidateAndReport validates commits and generates a report.
// Returns 0 for success, 2 for validation failure, 1 for errors.
func (o ValidationOrchestrator) ValidateAndReport(
	ctx context.Context,
	ref string,
	skipMergeCommits bool,
	options ports.ReportOptions,
) (int, error) {
	// Validate the commit
	commitResult, err := o.validationService.ValidateCommit(ctx, ref, skipMergeCommits)
	if err != nil {
		return 1, fmt.Errorf("failed to validate commit: %w", err)
	}

	// Create validation results
	results := domain.NewValidationResults().WithResult(commitResult)

	// Create report generator with provided options
	reportOpts := convertReportOptions(options)
	reportGenerator := report.NewGenerator(reportOpts, o.formatter, o.logger)

	// Generate report
	err = reportGenerator.GenerateReport(ctx, results)
	if err != nil {
		return 1, fmt.Errorf("failed to generate report: %w", err)
	}

	// Return appropriate exit code
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// ValidateRangeAndReport validates a range of commits and generates a report.
func (o ValidationOrchestrator) ValidateRangeAndReport(
	ctx context.Context,
	fromHash, toHash string,
	skipMergeCommits bool,
	options ports.ReportOptions,
) (int, error) {
	// Validate the commit range
	results, err := o.validationService.ValidateCommitRange(ctx, fromHash, toHash, skipMergeCommits)
	if err != nil {
		return 1, fmt.Errorf("failed to validate commit range: %w", err)
	}

	// Create report generator with provided options
	reportOpts := convertReportOptions(options)
	reportGenerator := report.NewGenerator(reportOpts, o.formatter, o.logger)

	// Generate report
	err = reportGenerator.GenerateReport(ctx, results)
	if err != nil {
		return 1, fmt.Errorf("failed to generate report: %w", err)
	}

	// Return appropriate exit code
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// ValidateMessageAndReport validates a message and generates a report.
func (o ValidationOrchestrator) ValidateMessageAndReport(
	ctx context.Context,
	message string,
	options ports.ReportOptions,
) (int, error) {
	// Validate the message
	results, err := o.validationService.ValidateMessage(ctx, message)
	if err != nil {
		return 1, fmt.Errorf("failed to validate message: %w", err)
	}

	// Create report generator with provided options
	reportOpts := convertReportOptions(options)
	reportGenerator := report.NewGenerator(reportOpts, o.formatter, o.logger)

	// Generate report
	err = reportGenerator.GenerateReport(ctx, results)
	if err != nil {
		return 1, fmt.Errorf("failed to generate report: %w", err)
	}

	// Return appropriate exit code
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// convertReportOptions converts port-level options to application-level options.
func convertReportOptions(options ports.ReportOptions) report.Options {
	// Convert format string to report.Format
	var format report.Format

	switch options.Format {
	case "json":
		format = report.FormatJSON
	case "github":
		format = report.FormatGitHubActions
	case "gitlab":
		format = report.FormatGitLabCI
	default:
		format = report.FormatText
	}

	return report.Options{
		Format:         format,
		Verbose:        options.Verbose,
		ShowHelp:       options.ShowHelp || options.ExtraVerbose,
		RuleToShowHelp: options.RuleToShowHelp,
		LightMode:      options.LightMode,
		Writer:         options.Writer,
	}
}

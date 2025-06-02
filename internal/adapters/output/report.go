// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package report provides functionality for generating validation reports.
package format

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
)

// Logger provides structured logging capabilities.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// Format represents the output format for reports.
type Format string

const (
	// FormatText is the plain text output format.
	FormatText Format = "text"
	// FormatJSON is the JSON output format.
	FormatJSON Format = "json"
	// FormatGitHubActions is the GitHub Actions output format.
	FormatGitHubActions Format = "github"
	// FormatGitLabCI is the GitLab CI output format.
	FormatGitLabCI Format = "gitlab"
)

// Options contains options for generating reports.
type Options struct {
	// Format is the output format.
	Format Format
	// Verbose indicates whether to include verbose details.
	Verbose bool
	// ShowHelp indicates whether to include help messages.
	ShowHelp bool
	// RuleToShowHelp specifies a specific rule to show help for.
	RuleToShowHelp string
	// LightMode indicates whether to use light mode colors.
	LightMode bool
	// Writer is the output writer.
	Writer io.Writer
}

// ReportService generates reports from validation results.
// It generates reports using the provided formatter and logger.
type ReportService struct {
	options   Options
	formatter Formatter
	logger    Logger
}

// Options returns a copy of the generator's options.
// This is useful for inspection and testing.
func (g ReportService) Options() Options {
	return copyOptions(g.options)
}

// ReportService creates validation reports with immutable configuration.

// NewReportService creates a new report generator.
// The writer must be provided in options - no default will be set.
func NewReportService(options Options, formatter Formatter, logger Logger) ReportService {
	return ReportService{
		options:   options,
		formatter: formatter,
		logger:    logger,
	}
}

// GenerateReport generates a validation report using the configured formatter.
func (g ReportService) GenerateReport(ctx context.Context, results interface{}) error {
	validationResults, ok := results.(domain.ValidationResults)
	if !ok {
		return errors.New("invalid results type: expected domain.ValidationResults")
	}

	g.logger.Info("Generating validation report",
		"format", string(g.options.Format),
		"verbose", g.options.Verbose,
		"show_help", g.options.ShowHelp,
		"total_commits", validationResults.TotalCommits,
		"passed_commits", validationResults.PassedCommits)

	// Use the injected formatter with context
	report := g.formatter.Format(ctx, validationResults)

	// Use a pure function to write the report
	if err := writeReport(g.options.Writer, report); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	// Note: The caller is responsible for flushing the writer if needed

	// Use a pure function to handle failure cases
	return handleFailure(validationResults, g.options.Writer)
}

// writeReport is a pure function to write content to a writer.
func writeReport(writer io.Writer, content string) error {
	// If there's nothing to write, consider it success (empty reports are valid)
	if len(content) == 0 {
		return nil
	}

	// Write the content to the writer
	bytesWritten, err := writer.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to write %d bytes: %w", len(content), err)
	}

	if bytesWritten != len(content) {
		return fmt.Errorf("short write: wrote %d of %d bytes", bytesWritten, len(content))
	}

	return nil
}

// handleFailure is a pure function that returns appropriate error for failed validations.
func handleFailure(_ domain.ValidationResults, _ io.Writer) error {
	// The application layer doesn't decide about exit codes or stdout handling
	// That's the responsibility of the CLI adapter
	return nil
}

//

// copyOptions creates a deep copy of Options to ensure immutability.
func copyOptions(opts Options) Options {
	return Options{
		Format:         opts.Format,
		Verbose:        opts.Verbose,
		ShowHelp:       opts.ShowHelp,
		RuleToShowHelp: opts.RuleToShowHelp,
		LightMode:      opts.LightMode,
		Writer:         opts.Writer,
	}
}

// WithVerbose returns a new ReportService with verbose setting updated.
// Uses explicit field copying for clarity and to maintain immutability.
func (g ReportService) WithVerbose(verbose bool) ReportService {
	return ReportService{
		options: Options{
			Format:         g.options.Format,
			Verbose:        verbose, // Only this field changes
			ShowHelp:       g.options.ShowHelp,
			RuleToShowHelp: g.options.RuleToShowHelp,
			LightMode:      g.options.LightMode,
			Writer:         g.options.Writer,
		},
		formatter: g.formatter,
		logger:    g.logger,
	}
}

// WithShowHelp returns a new ReportService with showHelp setting updated.
// Uses explicit field copying for clarity and to maintain immutability.
func (g ReportService) WithShowHelp(showHelp bool) ReportService {
	return ReportService{
		options: Options{
			Format:         g.options.Format,
			Verbose:        g.options.Verbose,
			ShowHelp:       showHelp, // Only this field changes
			RuleToShowHelp: g.options.RuleToShowHelp,
			LightMode:      g.options.LightMode,
			Writer:         g.options.Writer,
		},
		formatter: g.formatter,
		logger:    g.logger,
	}
}

// WithRuleToShowHelp returns a new ReportService with ruleToShowHelp setting updated.
// Uses explicit field copying for clarity and to maintain immutability.
func (g ReportService) WithRuleToShowHelp(ruleName string) ReportService {
	return ReportService{
		options: Options{
			Format:         g.options.Format,
			Verbose:        g.options.Verbose,
			ShowHelp:       g.options.ShowHelp,
			RuleToShowHelp: ruleName, // Only this field changes
			LightMode:      g.options.LightMode,
			Writer:         g.options.Writer,
		},
		formatter: g.formatter,
		logger:    g.logger,
	}
}

// GenerateSummary generates a brief summary report.
// This is used for quick command-line feedback.
func (g ReportService) GenerateSummary(_ context.Context, results interface{}) error {
	validationResults, ok := results.(domain.ValidationResults)
	if !ok {
		return errors.New("invalid results type: expected domain.ValidationResults")
	}

	g.logger.Info("Generating validation summary",
		"total_commits", validationResults.TotalCommits,
		"passed_commits", validationResults.PassedCommits)

	// Generate the summary content using a pure function
	summary := buildSummary(validationResults)

	// Write the summary using the pure function
	if err := writeReport(g.options.Writer, summary); err != nil {
		return err
	}

	// Handle failure cases using a pure function
	return handleFailure(validationResults, g.options.Writer)
}

// buildSummary is a pure function that creates a summary string from validation results.
func buildSummary(results domain.ValidationResults) string {
	var builder strings.Builder
	if results.AllPassed() {
		builder.WriteString("✅ All commits passed validation\n")
	} else {
		builder.WriteString("❌ Some commits failed validation\n")
		builder.WriteString(fmt.Sprintf("   Failed: %d/%d\n",
			results.TotalCommits-results.PassedCommits, results.TotalCommits))
	}

	return builder.String()
}

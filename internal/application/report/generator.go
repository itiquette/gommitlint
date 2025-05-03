// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package report provides functionality for generating validation reports.
package report

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

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

// Generator generates reports from validation results.
// It implements the domain.ReportGenerator interface.
type Generator struct {
	options   Options
	formatter domain.ResultFormatter
}

// Options returns a copy of the generator's options.
// This is useful for inspection and testing.
func (g Generator) Options() Options {
	return copyOptions(g.options)
}

// Ensure Generator implements domain.ReportGenerator.
var _ domain.ReportGenerator = Generator{}

// NewGenerator creates a new report generator.
func NewGenerator(options Options, formatter domain.ResultFormatter) Generator {
	// Ensure writer is initialized with a default if not provided
	if options.Writer == nil {
		options.Writer = os.Stdout
	}

	return Generator{
		options:   options,
		formatter: formatter,
	}
}

// GenerateReport implements the domain.ReportGenerator interface.
func (g Generator) GenerateReport(ctx context.Context, results domain.ValidationResults) error {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("format", string(g.options.Format)).
		Bool("verbose", g.options.Verbose).
		Bool("show_help", g.options.ShowHelp).
		Int("total_commits", results.TotalCommits).
		Int("passed_commits", results.PassedCommits).
		Msg("Entering Generator.GenerateReport")

	// Use the injected formatter with context
	report := g.formatter.Format(ctx, results)

	// Use a pure function to write the report
	if err := writeReport(g.options.Writer, report); err != nil {
		return err
	}

	// Use a pure function to handle failure cases
	return handleFailure(results, g.options.Writer)
}

// writeReport is a pure function to write content to a writer.
func writeReport(writer io.Writer, content string) error {
	_, err := writer.Write([]byte(content))

	return err
}

// handleFailure is a pure function that returns appropriate error for failed validations.
func handleFailure(results domain.ValidationResults, writer io.Writer) error {
	// If the validation failed and this is the final output, write an error code
	if !results.AllPassed() && writer == os.Stdout {
		// This is a convention to indicate failure to the shell
		return fmt.Errorf("validation failed: %d of %d commits failed",
			results.TotalCommits-results.PassedCommits, results.TotalCommits)
	}

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

// WithVerbose returns a new Generator with verbose setting updated.
func (g Generator) WithVerbose(verbose bool) domain.ReportGenerator {
	newOptions := copyOptions(g.options)
	newOptions.Verbose = verbose

	return Generator{
		options:   newOptions,
		formatter: g.formatter,
	}
}

// WithShowHelp returns a new Generator with showHelp setting updated.
func (g Generator) WithShowHelp(showHelp bool) domain.ReportGenerator {
	newOptions := copyOptions(g.options)
	newOptions.ShowHelp = showHelp

	return Generator{
		options:   newOptions,
		formatter: g.formatter,
	}
}

// WithRuleToShowHelp returns a new Generator with ruleToShowHelp setting updated.
func (g Generator) WithRuleToShowHelp(ruleName string) domain.ReportGenerator {
	newOptions := copyOptions(g.options)
	newOptions.RuleToShowHelp = ruleName

	return Generator{
		options:   newOptions,
		formatter: g.formatter,
	}
}

// GenerateSummary generates a brief summary report.
// This is used for quick command-line feedback.
func (g Generator) GenerateSummary(ctx context.Context, results domain.ValidationResults) error {
	logger := log.Logger(ctx)
	logger.Trace().
		Int("total_commits", results.TotalCommits).
		Int("passed_commits", results.PassedCommits).
		Msg("Entering Generator.GenerateSummary")

	// Generate the summary content using a pure function
	summary := buildSummary(results)

	// Write the summary using the pure function
	if err := writeReport(g.options.Writer, summary); err != nil {
		return err
	}

	// Handle failure cases using a pure function
	return handleFailure(results, g.options.Writer)
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

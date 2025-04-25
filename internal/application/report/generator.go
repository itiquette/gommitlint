// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package report provides functionality for generating validation reports.
package report

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
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

// Ensure Generator implements domain.ReportGenerator.
var _ domain.ReportGenerator = (*Generator)(nil)

// NewGenerator creates a new report generator.
func NewGenerator(options Options, formatter domain.ResultFormatter) *Generator {
	return &Generator{
		options:   options,
		formatter: formatter,
	}
}

// GenerateReport implements the domain.ReportGenerator interface.
func (g *Generator) GenerateReport(results domain.ValidationResults) error {
	// Use the injected formatter
	report := g.formatter.Format(results)

	_, err := g.options.Writer.Write([]byte(report))
	if err != nil {
		return err
	}

	// If the validation failed and this is the final output, write an error code
	if !results.AllPassed() && g.options.Writer == os.Stdout {
		// This is a convention to indicate failure to the shell
		return fmt.Errorf("validation failed: %d of %d commits failed", results.TotalCommits-results.PassedCommits, results.TotalCommits)
	}

	return nil
}

// Generate is kept for backward compatibility.
func (g *Generator) Generate(results domain.ValidationResults) error {
	return g.GenerateReport(results)
}

// SetVerbose enables or disables verbose output in reports.
// This implements the domain.ReportGenerator interface.
func (g *Generator) SetVerbose(verbose bool) {
	g.options.Verbose = verbose
}

// SetShowHelp enables or disables showing help messages in reports.
// This implements the domain.ReportGenerator interface.
func (g *Generator) SetShowHelp(showHelp bool) {
	g.options.ShowHelp = showHelp
}

// SetRuleToShowHelp sets a specific rule to show help for.
// This implements the domain.ReportGenerator interface.
func (g *Generator) SetRuleToShowHelp(ruleName string) {
	g.options.RuleToShowHelp = ruleName
}

// GenerateSummary generates a brief summary report.
// This is used for quick command-line feedback.
func (g *Generator) GenerateSummary(results domain.ValidationResults) error {
	var builder strings.Builder
	if results.AllPassed() {
		builder.WriteString("✅ All commits passed validation\n")
	} else {
		builder.WriteString("❌ Some commits failed validation\n")
		builder.WriteString(fmt.Sprintf("   Failed: %d/%d\n", results.TotalCommits-results.PassedCommits, results.TotalCommits))
	}

	// Write the summary
	_, err := g.options.Writer.Write([]byte(builder.String()))
	if err != nil {
		return err
	}

	// If the validation failed and this is the final output, write an error code
	if !results.AllPassed() && g.options.Writer == os.Stdout {
		// This is a convention to indicate failure to the shell
		return fmt.Errorf("validation failed: %d of %d commits failed", results.TotalCommits-results.PassedCommits, results.TotalCommits)
	}

	return nil
}

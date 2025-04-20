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
	"github.com/itiquette/gommitlint/internal/infrastructure/output"
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
type Generator struct {
	options *Options
}

// NewGenerator creates a new report generator.
func NewGenerator(options *Options) *Generator {
	return &Generator{
		options: options,
	}
}

// Generate generates a report from validation results.
func (g *Generator) Generate(results *domain.ValidationResults) error {
	switch g.options.Format {
	case FormatJSON:
		return g.generateJSONReport(results)
	case FormatText:
		return g.generateTextReport(results)
	case FormatGitHubActions:
		return g.generateGitHubActionsReport(results)
	case FormatGitLabCI:
		return g.generateGitLabCIReport(results)
	default:
		return g.generateTextReport(results)
	}
}

// generateJSONReport generates a JSON report.
func (g *Generator) generateJSONReport(results *domain.ValidationResults) error {
	formatter := output.NewJSONFormatter()
	report := formatter.Format(results)

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

// generateTextReport generates a text report.
func (g *Generator) generateTextReport(results *domain.ValidationResults) error {
	formatter := output.NewTextFormatter(g.options.Verbose, g.options.ShowHelp, g.options.LightMode)

	// If a specific rule is requested for help, show only that
	if g.options.RuleToShowHelp != "" {
		helpText := formatter.FormatRuleHelp(g.options.RuleToShowHelp, results)
		_, err := g.options.Writer.Write([]byte(helpText))

		return err
	}

	// Format full report
	report := formatter.Format(results)

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

// generateGitHubActionsReport generates a GitHub Actions compatible report.
func (g *Generator) generateGitHubActionsReport(results *domain.ValidationResults) error {
	// Create a GitHub Actions formatter
	formatter := output.NewGitHubActionsFormatter(g.options.Verbose, g.options.ShowHelp)

	// Format the results
	report := formatter.Format(results)

	// Write the report
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

// generateGitLabCIReport generates a GitLab CI compatible report.
func (g *Generator) generateGitLabCIReport(results *domain.ValidationResults) error {
	// Create a GitLab CI formatter
	formatter := output.NewGitLabCIFormatter(g.options.Verbose, g.options.ShowHelp)

	// Format the results
	report := formatter.Format(results)

	// Write the report
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

// GenerateSummary generates a brief summary report.
func (g *Generator) GenerateSummary(results *domain.ValidationResults) error {
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

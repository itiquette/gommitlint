// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package format

import (
	"context"
	"fmt"
	"io"

	"github.com/itiquette/gommitlint/internal/domain"
)

// Generator creates formatted reports using pure functional composition.
// This replaces the complex ReportService with a simple, focused approach.
type Generator struct {
	formatter Formatter
	writer    io.Writer
}

// NewGenerator creates a generator with explicit dependencies.
// No hidden dependencies or complex configuration.
func NewGenerator(formatter Formatter, writer io.Writer) Generator {
	return Generator{
		formatter: formatter,
		writer:    writer,
	}
}

// GenerateFromReport creates and writes a report from a domain Report.
// This is the main entry point that coordinates the generation process.
func (g Generator) GenerateFromReport(ctx context.Context, report domain.Report) error {
	// Format using adapter (pure function)
	content := g.formatter.Format(ctx, report)

	// Write using pure function
	return writeContent(g.writer, content)
}

// Generate creates and writes a report using functional composition.
// Legacy method for backward compatibility - prefer GenerateFromReport.
func (g Generator) Generate(ctx context.Context, commitResults []domain.ValidationResult, repoErrors []domain.ValidationError, options domain.ReportOptions) error {
	// Build report using pure domain function
	report := domain.BuildReport(commitResults, repoErrors, options)

	// Format using adapter (pure function)
	content := g.formatter.Format(ctx, report)

	// Write using pure function
	return writeContent(g.writer, content)
}

// GenerateSummary creates a brief summary report for quick feedback.
func (g Generator) GenerateSummary(_ context.Context, report domain.Report) error {
	// Create summary content using pure function
	summary := buildSummaryContent(report)

	// Write using pure function
	return writeContent(g.writer, summary)
}

// writeContent is a pure function to write content to a writer.
func writeContent(writer io.Writer, content string) error {
	if len(content) == 0 {
		return nil // Empty reports are valid
	}

	bytesWritten, err := writer.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to write %d bytes: %w", len(content), err)
	}

	if bytesWritten != len(content) {
		return fmt.Errorf("short write: wrote %d of %d bytes", bytesWritten, len(content))
	}

	return nil
}

// buildSummaryContent creates summary content from a report (pure function).
func buildSummaryContent(report domain.Report) string {
	if report.Summary.AllPassed {
		return "✅ All commits passed validation\n"
	}

	return fmt.Sprintf("❌ Some commits failed validation\n   Failed: %d/%d\n",
		report.Summary.FailedCommits, report.Summary.TotalCommits)
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"io"

	"github.com/itiquette/gommitlint/internal/adapters/output"
	"github.com/itiquette/gommitlint/internal/domain"
)

// OutputOptions represents how validation results should be formatted and displayed.
// This is a focused value type with single responsibility for output concerns.
type OutputOptions struct {
	Format       string    // "text", "json", "github", "gitlab"
	Verbose      bool      // Show detailed validation results
	ExtraVerbose bool      // Show extra detailed validation results
	RuleHelp     string    // Show detailed help for a specific rule
	LightMode    bool      // Use light background color scheme
	Writer       io.Writer // Where to write output
}

// NewOutputOptions creates OutputOptions with sensible defaults.
func NewOutputOptions(writer io.Writer) OutputOptions {
	return OutputOptions{
		Format:       "text",
		Verbose:      false,
		ExtraVerbose: false,
		RuleHelp:     "",
		LightMode:    false,
		Writer:       writer,
	}
}

// WithFormat returns a new OutputOptions with the specified format.
func (o OutputOptions) WithFormat(format string) OutputOptions {
	o.Format = format

	return o
}

// WithVerbose returns a new OutputOptions with verbose enabled/disabled.
func (o OutputOptions) WithVerbose(verbose bool) OutputOptions {
	o.Verbose = verbose

	return o
}

// WithExtraVerbose returns a new OutputOptions with extra verbose enabled/disabled.
func (o OutputOptions) WithExtraVerbose(extraVerbose bool) OutputOptions {
	o.ExtraVerbose = extraVerbose

	return o
}

// WithRuleHelp returns a new OutputOptions with rule help.
func (o OutputOptions) WithRuleHelp(ruleHelp string) OutputOptions {
	o.RuleHelp = ruleHelp

	return o
}

// WithLightMode returns a new OutputOptions with light mode enabled/disabled.
func (o OutputOptions) WithLightMode(lightMode bool) OutputOptions {
	o.LightMode = lightMode

	return o
}

// ShowHelp returns true if help should be shown.
func (o OutputOptions) ShowHelp() bool {
	return o.ExtraVerbose || o.RuleHelp != ""
}

// FormatReport formats a domain report using the specified options (pure function).
func (o OutputOptions) FormatReport(report domain.Report) string {
	switch o.Format {
	case "json":
		return output.JSON(report)
	case "github":
		return output.GitHub(report)
	case "gitlab":
		return output.GitLab(report)
	case "text":
		fallthrough
	default:
		textOptions := output.TextOptions{
			Verbose:   o.Verbose,
			ShowHelp:  o.ShowHelp(),
			LightMode: o.LightMode,
		}

		return output.Text(report, textOptions)
	}
}

// WriteReport formats and writes a report to the configured writer.
func (o OutputOptions) WriteReport(report domain.Report) error {
	content := o.FormatReport(report)
	_, err := o.Writer.Write([]byte(content))

	return err
}

// ToReportOptions converts OutputOptions to domain.ReportOptions.
func (o OutputOptions) ToReportOptions() domain.ReportOptions {
	return domain.ReportOptions{
		Format:       o.Format,
		Verbose:      o.Verbose,
		ExtraVerbose: o.ExtraVerbose,
		ShowHelp:     o.ShowHelp(),
		LightMode:    o.LightMode,
		Writer:       o.Writer,
	}
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package format

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
)

// Formatter converts domain reports to different output formats.
// This is the hexagonal port interface - adapters implement this.
type Formatter interface {
	// Format converts a domain report to a formatted string (pure function)
	Format(ctx context.Context, report domain.Report) string

	// ContentType returns the MIME type of the formatted output
	ContentType() string
}

// Format represents the output format type.
type Format string

// Supported output formats.
const (
	FormatText   Format = "text"
	FormatJSON   Format = "json"
	FormatGitHub Format = "github"
	FormatGitLab Format = "gitlab"
)

// NewFormatter creates a formatter for the specified format.
// This is the clean factory function that replaces complex patterns.
func NewFormatter(format Format, verbose, showHelp, lightMode bool) Formatter {
	switch format {
	case FormatJSON:
		return NewJSONFormatter()
	case FormatGitHub:
		return NewGitHubFormatter()
	case FormatGitLab:
		return NewGitLabFormatter()
	case FormatText:
		fallthrough
	default:
		return NewTextFormatter(verbose, showHelp, lightMode)
	}
}

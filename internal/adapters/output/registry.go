// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import "github.com/itiquette/gommitlint/internal/domain"

// Formatter is a function type for formatting validation reports.
type Formatter func(report domain.Report, options interface{}) string

// TextOnlyFormatter is for formatters that don't need options.
type TextOnlyFormatter func(report domain.Report) string

// formatters maps format names to their corresponding formatter functions.
var formatters = map[string]interface{}{
	"text":   Text,   // func(domain.Report, TextOptions) string
	"json":   JSON,   // func(domain.Report) string
	"github": GitHub, // func(domain.Report) string
	"gitlab": GitLab, // func(domain.Report) string
}

// Format formats a report using the specified format (main entry point).
// For text format, options should be TextOptions, for others it can be nil.
func Format(format string, report domain.Report, options interface{}) string {
	switch format {
	case "text":
		if textOpts, ok := options.(TextOptions); ok {
			return Text(report, textOpts)
		}

		return Text(report, TextOptions{})
	case "json":
		return JSON(report)
	case "github":
		return GitHub(report)
	case "gitlab":
		return GitLab(report)
	default:
		// Default to text format
		if textOpts, ok := options.(TextOptions); ok {
			return Text(report, textOpts)
		}

		return Text(report, TextOptions{})
	}
}

// SupportedFormats returns a list of all supported output formats.
func SupportedFormats() []string {
	formats := make([]string, 0, len(formatters))
	for format := range formatters {
		formats = append(formats, format)
	}

	return formats
}

// IsValidFormat checks if the given format is supported.
func IsValidFormat(format string) bool {
	_, exists := formatters[format]

	return exists
}

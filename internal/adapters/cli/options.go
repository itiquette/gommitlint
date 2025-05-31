// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
)

// Options represents the command-line options for the application.
type Options struct {
	// Verbosity represents the log level (trace, debug, brief, quiet)
	Verbosity string

	// Quiet indicates whether quiet mode is enabled (errors only)
	Quiet bool

	// VerbosityWithCaller indicates whether to include caller info in logs
	VerbosityWithCaller bool

	// OutputFormat represents the output format (text, json, github, gitlab)
	OutputFormat string
}

// WithCLIOptions adds Options to the context.
func WithCLIOptions(ctx context.Context, options Options) context.Context {
	return context.WithValue(ctx, CLIOptionsKey, options)
}

// OptionsFromContext retrieves Options from the context.
// If no options are present, it returns default options.
func OptionsFromContext(ctx context.Context) Options {
	if ctx == nil {
		return DefaultOptions()
	}

	value := ctx.Value(CLIOptionsKey)
	if value == nil {
		return DefaultOptions()
	}

	options, ok := value.(Options)
	if !ok {
		return DefaultOptions()
	}

	return options
}

// DefaultOptions returns the default CLI options.
func DefaultOptions() Options {
	return Options{
		Verbosity:           "brief",
		Quiet:               false,
		VerbosityWithCaller: false,
		OutputFormat:        "text",
	}
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package domain

import "context"

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

// Context keys.
const (
	// CLIOptionsKey is the context key for CLI options.
	CLIOptionsKey ContextKey = "cli_options"
)

// CLIOptions represents the command-line options for the application.
type CLIOptions struct {
	// Verbosity represents the log level (trace, debug, brief, quiet)
	Verbosity string

	// Quiet indicates whether quiet mode is enabled (errors only)
	Quiet bool

	// VerbosityWithCaller indicates whether to include caller info in logs
	VerbosityWithCaller bool

	// OutputFormat represents the output format (text, json, github, gitlab)
	OutputFormat string
}

// WithCLIOptions adds CLIOptions to the context.
func WithCLIOptions(ctx context.Context, options CLIOptions) context.Context {
	return context.WithValue(ctx, CLIOptionsKey, options)
}

// CLIOptionsFromContext retrieves CLIOptions from the context.
// If no options are present, it returns default options.
func CLIOptionsFromContext(ctx context.Context) CLIOptions {
	if ctx == nil {
		return DefaultCLIOptions()
	}

	value := ctx.Value(CLIOptionsKey)
	if value == nil {
		return DefaultCLIOptions()
	}

	options, ok := value.(CLIOptions)
	if !ok {
		return DefaultCLIOptions()
	}

	return options
}

// DefaultCLIOptions returns the default CLI options.
func DefaultCLIOptions() CLIOptions {
	return CLIOptions{
		Verbosity:           "brief",
		Quiet:               false,
		VerbosityWithCaller: false,
		OutputFormat:        "text",
	}
}

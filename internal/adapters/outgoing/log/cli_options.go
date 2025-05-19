// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"context"

	"github.com/itiquette/gommitlint/internal/common/contextkeys"
)

// CLIOptions defines the interface for accessing CLI options.
// This avoids direct dependency on domain.CLIOptions.
type CLIOptions interface {
	// GetVerbosity returns the log level.
	GetVerbosity() string

	// GetQuiet returns whether quiet mode is enabled.
	GetQuiet() bool

	// GetVerbosityWithCaller returns whether to include caller info in logs.
	GetVerbosityWithCaller() bool

	// GetOutputFormat returns the output format.
	GetOutputFormat() string
}

// DefaultCLIOptions provides default CLI options.
type DefaultCLIOptions struct{}

// GetVerbosity implements CLIOptions.GetVerbosity.
func (d DefaultCLIOptions) GetVerbosity() string {
	return "brief"
}

// GetQuiet implements CLIOptions.GetQuiet.
func (d DefaultCLIOptions) GetQuiet() bool {
	return false
}

// GetVerbosityWithCaller implements CLIOptions.GetVerbosityWithCaller.
func (d DefaultCLIOptions) GetVerbosityWithCaller() bool {
	return false
}

// GetOutputFormat implements CLIOptions.GetOutputFormat.
func (d DefaultCLIOptions) GetOutputFormat() string {
	return "text"
}

// CLIOptionsFromContext extracts CLI options from the context.
// It tries to find a CLIOptions implementation stored under CLIOptionsKey.
func CLIOptionsFromContext(ctx context.Context) CLIOptions {
	if ctx == nil {
		return DefaultCLIOptions{}
	}

	if optionsVal := ctx.Value(contextkeys.CLIOptionsKey); optionsVal != nil {
		if options, ok := optionsVal.(CLIOptions); ok {
			return options
		}
	}

	return DefaultCLIOptions{}
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"context"

	"github.com/spf13/cobra"
)

// InitLoggerContext initializes the logger for the application.
// This is a convenience function for use in the application's main entry point.
func InitLoggerContext(ctx context.Context, cmd *cobra.Command) context.Context {
	// Initialize the logger
	cliOptions := CLIOptionsFromContext(ctx)

	return InitLogger(ctx, cmd, cliOptions.GetVerbosityWithCaller(), cliOptions.GetOutputFormat())
}

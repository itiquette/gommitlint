// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cmd

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func newRootCommand(_ context.Context, versionString string) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "gommitlint",
		Version: versionString,
		Short:   "Commit validator.",
		Long:    `A tool to validate git commit messages against configurable rules.`,
	}

	rootCmd.AddCommand(newValidateCmd())

	return rootCmd
}

func Execute(version, commitSHA, buildDate string) {
	ctx := context.Background()

	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"
	err := newRootCommand(ctx, versionString).Execute()

	HandleError(ctx, err)
}

// HandleError processes errors in a consistent way across the application.
// It logs the error with appropriate context and exits with the correct status code.
func HandleError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	logger := zerolog.Ctx(ctx)

	// Get exit code - default to 1 for general errors
	exitCode := 1

	// Check for known error types that might have specific exit codes
	if exitErr, ok := err.(interface{ ExitCode() int }); ok {
		exitCode = exitErr.ExitCode()
	}

	// Log with appropriate level based on severity
	logEvent := logger.Error().Err(err)

	// Add context information if available
	if ctxErr, ok := err.(interface{ Context() map[string]string }); ok {
		for k, v := range ctxErr.Context() {
			logEvent = logEvent.Str(k, v)
		}
	}

	// Log the error
	logEvent.Msg("Command execution failed")

	// Exit with the determined status code
	os.Exit(exitCode)
}

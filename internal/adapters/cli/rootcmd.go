// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"os"

	log "github.com/itiquette/gommitlint/internal/adapters/logging"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/spf13/cobra"
)

// Execute executes the root command with the provided context, version information, and config.
func Execute(ctx context.Context, version, commitSHA, buildDate string, config config.Config) {
	rootCmd := createRootCommand(version, commitSHA, buildDate)

	// Add commands
	rootCmd.AddCommand(
		newValidateCmd(ctx, config),
		newInstallHookCmd(),
		newRemoveHookCmd(),
	)

	err := rootCmd.Execute()
	loggerInterface := GetLogger(ctx)

	if logger, ok := loggerInterface.(log.Logger); ok {
		HandleError(logger, err)
	} else if err != nil {
		// Fallback if logger is not available
		os.Exit(1)
	}
}

// createRootCommand creates the root cobra command with common setup.
func createRootCommand(version, commitSHA, buildDate string) *cobra.Command {
	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"

	rootCmd := &cobra.Command{
		Use:               "gommitlint",
		Version:           versionString,
		Short:             "Commit validator.",
		Long:              `A tool to validate git commit messages against configurable rules.`,
		PersistentPreRunE: setupLogging,
	}

	// Add common flags
	rootCmd.PersistentFlags().String("verbosity", "brief", "Log level (quiet, brief, trace)")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().Bool("caller", false, "Include caller information in logs")
	rootCmd.PersistentFlags().String("output", "text", "Output format (text, json, github, gitlab)")

	return rootCmd
}

// setupLogging initializes logging from command flags.
func setupLogging(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Get output format for logger initialization
	output, _ := cmd.Flags().GetString("output")
	ctx = log.InitLogger(ctx, cmd, output)

	// Ensure logger is available in context
	if ctx.Value(LoggerKey) == nil {
		zlog := log.GetLogger(ctx)
		logger := log.NewLogger(*zlog)
		ctx = WithLogger(ctx, logger)
	}

	cmd.SetContext(ctx)

	return nil
}

// HandleError processes errors in a consistent way across the application.
// It logs the error with appropriate context and exits with the correct status code.
func HandleError(logger log.Logger, err error) {
	if err == nil {
		return
	}

	// Get exit code - default to 1 for general errors
	exitCode := 1

	// Check for known error types that might have specific exit codes
	if exitErr, ok := err.(interface{ ExitCode() int }); ok {
		exitCode = exitErr.ExitCode()
	}

	// Log with appropriate level based on severity
	// Prepare error context if available
	if ctxErr, ok := err.(interface{ Context() map[string]string }); ok {
		ctxData := ctxErr.Context()
		// Build key-value pairs for logging
		kvPairs := make([]interface{}, 0, 2+len(ctxData)*2)
		kvPairs = append(kvPairs, "error", err)

		for k, v := range ctxData {
			kvPairs = append(kvPairs, k, v)
		}

		logger.Error("Command execution failed", kvPairs...)
	} else {
		logger.Error("Command execution failed", "error", err)
	}

	// Exit with the determined status code
	os.Exit(exitCode)
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"os"

	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
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

	// Get logger from context and handle error
	zerologLogger := logadapter.GetLogger(ctx)
	logger := logadapter.NewDomainLogger(zerologLogger)

	if err != nil {
		logger.Error("Command execution failed", "error", err)
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
	ctx = logadapter.InitLogger(ctx, cmd, output)

	// Logger is already set up by log.InitLogger, no additional context setup needed

	cmd.SetContext(ctx)

	return nil
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/application/options"
	"github.com/itiquette/gommitlint/internal/common/contextkeys"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/composition"
	configTypes "github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
	"github.com/spf13/cobra"
)

// DependencyContainer defines the interface for the dependency injection container.
type DependencyContainer interface {
	GetCreateValidationService() func(context.Context, string) (incoming.ValidationService, error)
	CreateValidationOrchestrator(ctx context.Context, repoPath string, formatter outgoing.ResultFormatter) (incoming.ValidationOrchestrator, error)
	GetLogger() outgoing.Logger
}

// Execute executes the root command with the provided context, version information, and config.
func Execute(ctx context.Context, version, commitSHA, buildDate string, config configTypes.Config) {
	// Create version string
	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"

	// Container will be created after logger is initialized
	var container DependencyContainer

	// Create root command
	rootCmd := &cobra.Command{
		Use:     "gommitlint",
		Version: versionString,
		Short:   "Commit validator.",
		Long:    `A tool to validate git commit messages against configurable rules.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Extract CLI options directly from command flags
			verbosity, _ := cmd.Flags().GetString("verbosity")
			quiet, _ := cmd.Flags().GetBool("quiet")
			caller, _ := cmd.Flags().GetBool("caller")
			output, _ := cmd.Flags().GetString("output")

			cliOptions := options.CLIOptions{
				Verbosity:           verbosity,
				Quiet:               quiet,
				VerbosityWithCaller: caller,
				OutputFormat:        output,
			}

			ctx = log.InitLogger(ctx, cmd, cliOptions)

			// Get the logger adapter from context
			logger, ok := ctx.Value(contextkeys.LoggerKey).(outgoing.Logger)
			if !ok {
				// Create adapter from zerolog logger
				zlog := log.Logger(ctx)
				logger = log.NewAdapter(*zlog)
				// Store the logger adapter in context
				ctx = context.WithValue(ctx, contextkeys.LoggerKey, logger)
			}

			// Create container with proper logger and config
			containerImpl := composition.NewContainer(logger, config)
			container = &containerImpl

			// Propagate the context to all commands
			cmd.SetContext(ctx)

			return nil
		},
	}

	// Add common flags
	rootCmd.PersistentFlags().String("verbosity", "brief", "Log level (quiet, brief, trace)")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().Bool("caller", false, "Include caller information in logs")
	rootCmd.PersistentFlags().String("output", "text", "Output format (text, json, github, gitlab)")

	// Add commands
	rootCmd.AddCommand(newValidateCmd(func(_ context.Context) DependencyContainer { return container })) //nolint:contextcheck // container is captured from closure
	rootCmd.AddCommand(newInstallHookCmd())
	rootCmd.AddCommand(newRemoveHookCmd())

	// Execute
	err := rootCmd.Execute()
	if container != nil {
		HandleError(container.GetLogger(), err)
	} else {
		// Fallback to context logger if container is not available
		HandleError(contextx.GetLogger(ctx), err)
	}
}

// HandleError processes errors in a consistent way across the application.
// It logs the error with appropriate context and exits with the correct status code.
func HandleError(logger outgoing.Logger, err error) {
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

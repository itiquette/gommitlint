// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"os"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
	"github.com/spf13/cobra"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// Context keys.
const (
	dependenciesKey contextKey = "dependencies"
)

// AppDependencies holds dependencies that can be injected into the application.
// It uses domain interfaces instead of concrete implementations to follow
// the Dependency Inversion Principle.
type AppDependencies struct {
	// ConfigManager provides configuration
	ConfigManager *config.Manager
}

func newRootCommand(ctx context.Context, versionString string, deps *AppDependencies) *cobra.Command {
	// Store dependencies in the context
	ctx = context.WithValue(ctx, dependenciesKey, deps)

	var rootCmd = &cobra.Command{
		Use:     "gommitlint",
		Version: versionString,
		Short:   "Commit validator.",
		Long:    `A tool to validate git commit messages against configurable rules.`,
		// Set the context for the command
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Initialize logger with command flags
			cliOptions := domain.CLIOptionsFromContext(ctx)
			ctx = log.InitLogger(ctx, cmd, cliOptions.VerbosityWithCaller, cliOptions.OutputFormat)

			// Log initialization (trace level)
			logger := log.Logger(ctx)
			logger.Trace().Msg("Logger initialized")

			// Propagate the context to all commands
			cmd.SetContext(ctx)

			return nil
		},
	}

	// Add common flags
	rootCmd.PersistentFlags().String("verbosity", "brief", "Log level (quiet, brief, debug, trace)")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().Bool("caller", false, "Include caller information in logs")
	rootCmd.PersistentFlags().String("output", "text", "Output format (text, json, github, gitlab)")

	// Create validate command
	validateCmd := newValidateCmd(ctx)

	// Add the validate command
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(newInstallHookCmd())
	rootCmd.AddCommand(newRemoveHookCmd())

	return rootCmd
}

func Execute(version, commitSHA, buildDate string) {
	// Create root context
	ctx := context.Background()

	// Add default CLI options to context
	ctx = domain.WithCLIOptions(ctx, domain.DefaultCLIOptions())

	// Initialize basic logger
	logger := log.InitBasicLogger()
	ctx = logger.WithContext(ctx)

	// Create config manager with the context
	configManager, err := config.NewManager(ctx)
	if err != nil {
		log.Logger(ctx).Error().Err(err).Msg("Failed to create configuration manager")
		os.Exit(1)
	}

	// Use the ExecuteWithDependencies function with the context
	ExecuteWithDependencies(ctx, version, commitSHA, buildDate, configManager)
}

// ExecuteWithDependencies executes the root command with explicit dependencies.
// This allows for better testability by injecting mock dependencies.
// It follows the Dependency Inversion Principle by accepting domain interfaces
// rather than concrete implementations.
func ExecuteWithDependencies(
	ctx context.Context,
	version,
	commitSHA,
	buildDate string,
	configManager *config.Manager,
) {
	// Create a logger
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering ExecuteWithDependencies")

	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"

	// Create dependencies container
	deps := &AppDependencies{
		ConfigManager: configManager,
	}

	// Create and execute root command with dependencies
	err := newRootCommand(ctx, versionString, deps).Execute()

	HandleError(ctx, err)
}

// HandleError processes errors in a consistent way across the application.
// It logs the error with appropriate context and exits with the correct status code.
func HandleError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	logger := log.Logger(ctx)

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

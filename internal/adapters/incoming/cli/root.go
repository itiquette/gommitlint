// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/spf13/cobra"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// Context keys.
const (
	cliDependenciesKey contextKey = "cli_dependencies"
)

// AppDependencies holds dependencies that can be injected into the application.
// It uses domain interfaces instead of concrete implementations to follow
// the Dependency Inversion Principle.
type AppDependencies struct {
	// ConfigManager provides access to configuration in a more structured way
	ConfigManager *config.Manager
}

func newRootCommand(ctx context.Context, versionString string, deps *AppDependencies) *cobra.Command {
	// Store dependencies in the context
	ctx = context.WithValue(ctx, cliDependenciesKey, deps)

	// Create root command
	var rootCmd = &cobra.Command{
		Use:     "gommitlint",
		Version: versionString,
		Short:   "Commit validator.",
		Long:    `A tool to validate git commit messages against configurable rules.`,
		// Set the context for the command
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Initialize logger with command flags
			ctx = log.InitLoggerContext(ctx, cmd)

			// Log initialization (trace level)
			logger := contextx.GetLogger(ctx)
			logger.Debug("Logger initialized")

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

// ExecuteWithContext executes the root command with the provided context and version information.
// The context should be created with context.Background() ONLY in main.go.
func ExecuteWithContext(ctx context.Context, version, commitSHA, buildDate string) {
	// Initialize basic logger
	logger := log.InitBasicLogger()
	ctx = logger.WithContext(ctx)

	// Register the domain logger provider if not already registered

	// Create a config manager with the propagated context
	configManager, err := config.NewManager(ctx)
	if err != nil {
		errorLogger := contextx.GetLogger(ctx)
		errorLogger.Error("Failed to create configuration manager", "error", err)
		os.Exit(1)
	}

	// Get the config from the manager
	cfg := configManager.GetConfig()

	// Print loaded config information to verify
	configLogger := contextx.GetLogger(ctx)
	configLogger.Info("Loaded configuration",
		"enabled_rules", cfg.Rules.EnabledRules,
		"disabled_rules", cfg.Rules.DisabledRules)

	configLogger.Debug("Added config directly to context using contextx.WithConfig")

	// Use the ExecuteWithDependencies function with the context
	ExecuteWithDependencies(
		ctx,
		version,
		commitSHA,
		buildDate,
		&AppDependencies{
			ConfigManager: configManager,
		},
	)
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
	deps *AppDependencies,
) {
	// Create a logger
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering ExecuteWithDependencies")

	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"

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

	logger := contextx.GetLogger(ctx)

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

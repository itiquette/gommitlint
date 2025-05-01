// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"os"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/rs/zerolog"
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
			// Propagate the context to all commands
			cmd.SetContext(ctx)

			return nil
		},
	}

	// Create validate command
	validateCmd := newValidateCmd()

	// Add the validate command
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(newInstallHookCmd())
	rootCmd.AddCommand(newRemoveHookCmd())

	return rootCmd
}

func Execute(version, commitSHA, buildDate string) {
	ctx := context.Background()

	// Create config manager
	configManager, err := config.NewManager()
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to create configuration manager")
		os.Exit(1)
	}

	// Use the ExecuteWithDependencies function
	ExecuteWithDependencies(version, commitSHA, buildDate, configManager)
}

// ExecuteWithDependencies executes the root command with explicit dependencies.
// This allows for better testability by injecting mock dependencies.
// It follows the Dependency Inversion Principle by accepting domain interfaces
// rather than concrete implementations.
func ExecuteWithDependencies(
	version,
	commitSHA,
	buildDate string,
	configManager *config.Manager,
) {
	ctx := context.Background()

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

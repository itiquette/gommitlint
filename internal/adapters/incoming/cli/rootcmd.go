// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"context"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/application/options"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
	"github.com/spf13/cobra"
)

// containerKey is the context key for the dependency injection container.
type containerKey struct{}

// DependencyContainer defines the interface for the dependency injection container.
type DependencyContainer interface {
	GetCreateValidationService() func(context.Context, string) (incoming.ValidationService, error)
	CreateValidationOrchestrator(ctx context.Context, repoPath string, formatter outgoing.ResultFormatter) (incoming.ValidationOrchestrator, error)
}

// getContainer retrieves the dependency injection container from context.
func getContainer(ctx context.Context) DependencyContainer {
	if container, ok := ctx.Value(containerKey{}).(DependencyContainer); ok {
		return container
	}

	return nil
}

func newRootCommand(ctx context.Context, versionString string) *cobra.Command {
	// Create root command
	var rootCmd = &cobra.Command{
		Use:     "gommitlint",
		Version: versionString,
		Short:   "Commit validator.",
		Long:    `A tool to validate git commit messages against configurable rules.`,
		// Set the context for the command
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cliOptions := options.CLIOptionsFromContext(ctx)
			ctx = log.InitLogger(ctx, cmd, cliOptions.GetVerbosityWithCaller(), cliOptions.GetOutputFormat())

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

	// Create validate command
	validateCmd := newValidateCmd(ctx)

	// Add the validate command
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(newInstallHookCmd())
	rootCmd.AddCommand(newRemoveHookCmd())

	return rootCmd
}

// Execute executes the root command with the provided context, version information, and dependency container.
func Execute(ctx context.Context, version, commitSHA, buildDate string, container DependencyContainer) {
	// Store container in context for later use
	ctx = context.WithValue(ctx, containerKey{}, container)

	// Create version string
	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"

	// Create and execute root command
	err := newRootCommand(ctx, versionString).Execute()

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

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package integtest contains end-to-end integration tests for gommitlint workflows.
// These tests verify that the application's components work together correctly.
package integtest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
)

// SetupTestRepository creates a temporary Git repository for testing.
// It initializes a git repository with a commit containing the given message.
// Returns the repository path and a cleanup function.
func SetupTestRepository(t *testing.T, commitMessage string) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gommitlint-integration-test-*")
	require.NoError(t, err)

	// Initialize git repository
	err = runGitCommand(tempDir, "init")
	require.NoError(t, err)

	// Configure git identity for commits
	err = runGitCommand(tempDir, "config", "user.name", "Test User")
	require.NoError(t, err)

	err = runGitCommand(tempDir, "config", "user.email", "test@example.com")
	require.NoError(t, err)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0600)
	require.NoError(t, err)

	// Add and commit the file
	err = runGitCommand(tempDir, "add", "test.txt")
	require.NoError(t, err)

	err = runGitCommand(tempDir, "commit", "-m", commitMessage)
	require.NoError(t, err)

	// Return the repository path and cleanup function
	return tempDir, func() {
		os.RemoveAll(tempDir)
	}
}

// runGitCommand runs a git command in the specified directory.
func runGitCommand(dir string, args ...string) error {
	// Use the exec.Command but implement using git command args
	cmd := execCommand("git", args...)
	cmd.Dir = dir

	return cmd.Run()
}

// execCommand is used to allow mocking in tests.
var execCommand = exec.Command

// IsGitAvailable checks if git is available on the system.
func IsGitAvailable() bool {
	cmd := exec.Command("git", "--version")

	return cmd.Run() == nil
}

// createTestConfigManager creates a config manager with a specific configuration file
// and no defaults or standard paths loaded.
// It automatically enhances the provided context with logger for tracing.
// Returns both the config manager and the enhanced context.
func createTestConfigManager(ctx context.Context, t *testing.T, configPath string) (*config.Manager, context.Context) {
	t.Helper()

	// Always enhance context with logger for testing
	ctx = enhanceContextForTesting(ctx, t)

	// Create manager with the enhanced context
	manager, err := config.NewManager(ctx)
	require.NoError(t, err)

	// Reset config to explicitly load only our test config
	err = manager.LoadFromFile(ctx, configPath)
	require.NoError(t, err)

	// For testing purposes, ensure we're using the default config settings
	// This helps with tests that are dependent on all rules being registered
	manager.SetConfig(config.DefaultConfig())

	// Now load the test config on top of the default config
	err = manager.LoadFromFile(ctx, configPath)
	require.NoError(t, err)

	return manager, ctx
}

// enhanceContextForTesting adds test-specific logging settings to an existing context.
// If the provided context is nil, a new background context will be created.
func enhanceContextForTesting(ctx context.Context, t *testing.T) context.Context {
	t.Helper()

	// Use existing context or create a new one if nil
	if ctx == nil {
		ctx = context.Background()
	}

	// Create a zerolog logger with trace level for tests
	writer := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	logger := zerolog.New(writer).Level(zerolog.TraceLevel).With().Timestamp().Logger()

	// Add CLI options to context
	// Get existing options or use defaults
	options := domain.CLIOptionsFromContext(ctx)
	options.Verbosity = "trace"
	options.VerbosityWithCaller = true
	ctx = domain.WithCLIOptions(ctx, options)

	// Add logger to context
	ctx = logger.WithContext(ctx)

	// Also add it through our contextx key system
	ctx = contextx.WithValue(ctx, contextx.LoggerKey, &logger)

	return ctx
}

// Helper function to extract validation errors from a commit result.
func getValidationErrors(result domain.CommitResult) []string {
	var errors []string

	for _, ruleResult := range result.RuleResults {
		if ruleResult.Status == domain.StatusFailed {
			for _, err := range ruleResult.Errors {
				errors = append(errors, err.Error())
			}
		}
	}

	return errors
}

// Helper function to extract validation errors from a commit result, excluding a specific rule.
func getValidationErrorsExcludingRule(result domain.CommitResult, excludeRule string) []string {
	var errors []string

	for _, ruleResult := range result.RuleResults {
		if ruleResult.Status == domain.StatusFailed && ruleResult.RuleName != excludeRule {
			for _, err := range ruleResult.Errors {
				errors = append(errors, err.Error())
			}
		}
	}

	return errors
}

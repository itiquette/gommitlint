// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package integtest contains end-to-end integration tests for gommitlint workflows.
// These tests verify that the application's components work together correctly.
package integtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
)

// setupTestRepository creates a temporary Git repository for testing.
// It initializes a git repository with a commit containing the given message.
// Returns the repository path and a cleanup function.
func setupTestRepository(t *testing.T, commitMessage string) (string, func()) {
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

// isGitAvailable checks if git is available on the system.
func isGitAvailable() bool {
	cmd := exec.Command("git", "--version")

	return cmd.Run() == nil
}

// createTestConfigManager creates a config manager with a specific configuration file
// and no defaults or standard paths loaded.
func createTestConfigManager(t *testing.T, configPath string) *config.Manager {
	t.Helper()

	// Create manager with defaults
	manager, err := config.New()
	require.NoError(t, err)

	// Reset config to explicitly load only our test config
	err = manager.LoadFromFile(configPath)
	require.NoError(t, err)

	return manager
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

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package integtest contains end-to-end integration tests for gommitlint workflows.
// These tests verify that the application's components work together correctly.
// THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE.
package integtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

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

// GetValidationErrors extracts validation errors from a commit result.
// It returns a slice of error messages from all failed rule results.
func GetValidationErrors(result domain.CommitResult) []string {
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

// GetValidationErrorsExcludingRule extracts validation errors from a commit result, excluding a specific rule.
// This is typically used to exclude CommitsAhead rule errors which can be problematic in test environments.
func GetValidationErrorsExcludingRule(result domain.CommitResult, excludeRule string) []string {
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

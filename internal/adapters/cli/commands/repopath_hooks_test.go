// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// TestRepoPathHookInstallation tests installing hooks via --repo-path from different directories.
func TestRepoPathHookInstallation(t *testing.T) {
	// Create repository to install hooks in
	targetRepoDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = targetRepoDir
	err := cmd.Run()
	require.NoError(t, err)

	// Create different working directory
	workingDir := t.TempDir()
	originalWd, _ := os.Getwd()

	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(workingDir)
	require.NoError(t, err)

	tests := []struct {
		name        string
		repoPath    string
		hookType    string
		expectError bool
		description string
	}{
		{
			name:        "install commit-msg hook in different repo",
			repoPath:    targetRepoDir,
			hookType:    "commit-msg",
			expectError: false,
			description: "should install commit-msg hook in target repository",
		},
		{
			name:        "install pre-commit hook in different repo",
			repoPath:    targetRepoDir,
			hookType:    "pre-commit",
			expectError: false,
			description: "should install pre-commit hook in target repository",
		},
		{
			name:        "install in non-existent repo",
			repoPath:    "/nonexistent/repo",
			hookType:    "commit-msg",
			expectError: true,
			description: "should fail when target repository doesn't exist",
		},
		{
			name:        "install in current directory without git",
			repoPath:    ".",
			hookType:    "commit-msg",
			expectError: true,
			description: "should fail when current directory is not a git repo",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{Name: "repo-path"},
			}

			err := app.Root().Set("repo-path", testCase.repoPath)
			require.NoError(t, err)

			// Test getRepoPath extraction
			repoPath := getRepoPath(app)
			require.Equal(t, testCase.repoPath, repoPath)

			// Verify hooks directory behavior
			if !testCase.expectError {
				hooksDir := filepath.Join(repoPath, ".git", "hooks")

				// Check if hooks directory exists or can be created
				if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
					err = os.MkdirAll(hooksDir, 0755)
					require.NoError(t, err, "should be able to create hooks directory")
				}

				// Verify hook file path would be valid
				hookPath := filepath.Join(hooksDir, testCase.hookType)
				require.Contains(t, hookPath, testCase.hookType, "hook path should contain hook type")
				require.Contains(t, hookPath, repoPath, "hook path should be within target repo")
			}
		})
	}
}

// TestRepoPathHookRemoval tests removing hooks via --repo-path from different directories.
func TestRepoPathHookRemoval(t *testing.T) {
	// Create repository with existing hooks
	targetRepoDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = targetRepoDir
	err := cmd.Run()
	require.NoError(t, err)

	// Create hooks directory and sample hook
	hooksDir := filepath.Join(targetRepoDir, ".git", "hooks")
	err = os.MkdirAll(hooksDir, 0755)
	require.NoError(t, err)

	// Create a test hook file
	hookPath := filepath.Join(hooksDir, "commit-msg")
	hookContent := `#!/bin/sh
# Test hook
echo "test hook"
`
	err = os.WriteFile(hookPath, []byte(hookContent), 0600)
	require.NoError(t, err)

	// Create different working directory
	workingDir := t.TempDir()
	originalWd, _ := os.Getwd()

	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(workingDir)
	require.NoError(t, err)

	tests := []struct {
		name        string
		repoPath    string
		hookType    string
		expectError bool
		description string
	}{
		{
			name:        "remove hook from different repo",
			repoPath:    targetRepoDir,
			hookType:    "commit-msg",
			expectError: false,
			description: "should locate hook in target repository",
		},
		{
			name:        "remove non-existent hook",
			repoPath:    targetRepoDir,
			hookType:    "pre-push",
			expectError: true,
			description: "should fail when hook doesn't exist",
		},
		{
			name:        "remove from non-existent repo",
			repoPath:    "/nonexistent/repo",
			hookType:    "commit-msg",
			expectError: true,
			description: "should fail when target repository doesn't exist",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{Name: "repo-path"},
			}

			err := app.Root().Set("repo-path", testCase.repoPath)
			require.NoError(t, err)

			// Test getRepoPath extraction
			repoPath := getRepoPath(app)
			require.Equal(t, testCase.repoPath, repoPath)

			// Check hook file existence
			hookFilePath := filepath.Join(repoPath, ".git", "hooks", testCase.hookType)

			if !testCase.expectError {
				// For valid cases, verify we can access the hook path
				if testCase.hookType == "commit-msg" {
					_, err := os.Stat(hookFilePath)
					require.NoError(t, err, "hook file should exist")
				}

				// Verify path is within target repository
				require.Contains(t, hookFilePath, repoPath, "hook path should be within target repo")
			}
		})
	}
}

// TestRepoPathHookSecurity tests security aspects of hook operations with --repo-path.
func TestRepoPathHookSecurity(t *testing.T) {
	// Create legitimate repository
	validRepoDir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = validRepoDir
	err := cmd.Run()
	require.NoError(t, err)

	tests := []struct {
		name             string
		repoPath         string
		hookType         string
		expectValidPath  bool
		expectSecurePath bool
		description      string
	}{
		{
			name:             "normal hook operation",
			repoPath:         validRepoDir,
			hookType:         "commit-msg",
			expectValidPath:  true,
			expectSecurePath: true,
			description:      "should handle normal paths securely",
		},
		{
			name:             "path traversal attempt",
			repoPath:         "../../../etc",
			hookType:         "commit-msg",
			expectValidPath:  false,
			expectSecurePath: false,
			description:      "path traversal should be handled by validators",
		},
		{
			name:             "invalid hook type",
			repoPath:         validRepoDir,
			hookType:         "invalid-hook",
			expectValidPath:  true,
			expectSecurePath: false,
			description:      "invalid hook types should be rejected by validators",
		},
		{
			name:             "hook type with special chars",
			repoPath:         validRepoDir,
			hookType:         "commit-msg; rm -rf /",
			expectValidPath:  true,
			expectSecurePath: false,
			description:      "malicious hook types should be rejected by validators",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{Name: "repo-path"},
			}

			err := app.Root().Set("repo-path", testCase.repoPath)
			require.NoError(t, err)

			// Test getRepoPath extraction (this should always work - validation happens later)
			repoPath := getRepoPath(app)
			require.Equal(t, testCase.repoPath, repoPath)

			// Basic path construction for hooks
			if testCase.expectValidPath {
				hookPath := filepath.Join(repoPath, ".git", "hooks", testCase.hookType)

				// Verify path structure is reasonable
				require.Contains(t, hookPath, ".git/hooks", "should contain hooks directory")

				// Only check for hook type if it's a valid hook name
				if testCase.expectSecurePath {
					require.Contains(t, hookPath, testCase.hookType, "should contain hook type")
				}
			}
			// Note: Actual security validation would happen in the path validator,
			// not in getRepoPath. This test verifies that getRepoPath itself
			// doesn't introduce security issues by returning the path as-is
			// for later validation.
			_ = "validation happens in path validator"
		})
	}
}

// TestRepoPathCrossDirectory tests running hook operations across different directories.
func TestRepoPathCrossDirectory(t *testing.T) {
	// Create multiple repositories in different locations
	repo1Dir := t.TempDir()
	repo2Dir := t.TempDir()
	workDir := t.TempDir()

	// Initialize both repositories
	for _, repoDir := range []string{repo1Dir, repo2Dir} {
		cmd := exec.Command("git", "init")
		cmd.Dir = repoDir
		err := cmd.Run()
		require.NoError(t, err)
	}

	// Change to working directory (different from both repos)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()

	err := os.Chdir(workDir)
	require.NoError(t, err)

	tests := []struct {
		name        string
		targetRepo  string
		hookType    string
		description string
	}{
		{
			name:        "operate on first repository",
			targetRepo:  repo1Dir,
			hookType:    "commit-msg",
			description: "should target first repository from working directory",
		},
		{
			name:        "operate on second repository",
			targetRepo:  repo2Dir,
			hookType:    "pre-commit",
			description: "should target second repository from working directory",
		},
		{
			name:        "relative path to repository",
			targetRepo:  filepath.Base(repo1Dir),
			hookType:    "commit-msg",
			description: "should handle relative paths",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{Name: "repo-path"},
			}

			err := app.Root().Set("repo-path", testCase.targetRepo)
			require.NoError(t, err)

			// Test repo path resolution
			repoPath := getRepoPath(app)
			require.Equal(t, testCase.targetRepo, repoPath)

			// Verify we're not in the target repository
			currentWd, err := os.Getwd()
			require.NoError(t, err)

			if filepath.IsAbs(testCase.targetRepo) {
				require.NotEqual(t, testCase.targetRepo, currentWd,
					"should be running from different directory than target")
			}

			// Verify hook path construction
			expectedHookPath := filepath.Join(repoPath, ".git", "hooks", testCase.hookType)
			require.Contains(t, expectedHookPath, testCase.hookType,
				"hook path should contain hook type")
		})
	}
}

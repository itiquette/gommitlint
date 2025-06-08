// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/cli"
	"github.com/itiquette/gommitlint/internal/adapters/git"
)

// TestRepoPathIntegration tests complete --repo-path functionality across different directories.
func TestRepoPathIntegration(t *testing.T) {
	// Create target repository in one location
	targetRepoDir := t.TempDir()

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = targetRepoDir
	err := cmd.Run()
	require.NoError(t, err, "failed to initialize git repository")

	// Create test commit with conventional format
	testFile := filepath.Join(targetRepoDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = targetRepoDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "-c", "user.name=Test User", "-c", "user.email=test@example.com",
		"commit", "-m", "feat: add test file")
	cmd.Dir = targetRepoDir
	err = cmd.Run()
	require.NoError(t, err)

	// Create working directory in different location
	workingDir := t.TempDir()
	originalWd, _ := os.Getwd()

	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(workingDir)
	require.NoError(t, err)

	tests := []struct {
		name         string
		repoPath     string
		gitReference string
		expectValid  bool
		description  string
	}{
		{
			name:         "validate HEAD commit in remote repository",
			repoPath:     targetRepoDir,
			gitReference: "HEAD",
			expectValid:  true,
			description:  "should validate HEAD commit in target repository from different directory",
		},
		{
			name:         "validate HEAD again in remote repository",
			repoPath:     targetRepoDir,
			gitReference: "HEAD",
			expectValid:  true,
			description:  "should validate HEAD reference multiple times in target repository",
		},
		{
			name:         "fail validation in current directory without git",
			repoPath:     ".",
			gitReference: "HEAD",
			expectValid:  false,
			description:  "should fail when current directory is not a git repository",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()

			// Test git repository creation with target path
			gitRepo, err := git.NewRepository(testCase.repoPath)

			if !testCase.expectValid {
				// For invalid cases, expect error in repository creation
				require.Error(t, err, testCase.description)

				return
			}

			require.NoError(t, err, testCase.description)
			require.NotNil(t, gitRepo, "should create git repository adapter")

			// Test getting commit from target repository
			commit, err := gitRepo.GetCommit(ctx, testCase.gitReference)

			if testCase.expectValid {
				require.NoError(t, err, testCase.description)
				require.NotNil(t, commit, "should retrieve commit from target repository")
				require.NotEmpty(t, commit.Hash, "commit should have hash")
				require.NotEmpty(t, commit.Message, "commit should have message")
			} else {
				require.Error(t, err, testCase.description)
			}
		})
	}
}

// TestRepoPathHookIntegration tests hook installation/removal with --repo-path.
func TestRepoPathHookIntegration(t *testing.T) {
	// Create target repository
	targetRepoDir := t.TempDir()

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = targetRepoDir
	err := cmd.Run()
	require.NoError(t, err)

	// Create working directory in different location
	workingDir := t.TempDir()
	originalWd, _ := os.Getwd()

	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(workingDir)
	require.NoError(t, err)

	// Test path validator with target repository
	pathValidator := cli.StrictPathValidator{}

	tests := []struct {
		name        string
		repoPath    string
		hookType    string
		expectError bool
		description string
	}{
		{
			name:        "validate commit-msg hook path in remote repo",
			repoPath:    targetRepoDir,
			hookType:    "commit-msg",
			expectError: false,
			description: "should validate hook path in target repository",
		},
		{
			name:        "validate pre-commit hook path in remote repo",
			repoPath:    targetRepoDir,
			hookType:    "pre-commit",
			expectError: false,
			description: "should validate different hook types in target repository",
		},
		{
			name:        "reject hook installation in non-git directory",
			repoPath:    workingDir,
			hookType:    "commit-msg",
			expectError: true,
			description: "should reject hook installation in non-git directory",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			hookPath, err := pathValidator.ValidateHookPath(testCase.repoPath, testCase.hookType)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
				require.Empty(t, hookPath)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, hookPath)

				// Verify hook path is in target repository
				require.Contains(t, hookPath, testCase.repoPath,
					"hook path should be within target repository")
				require.Contains(t, hookPath, testCase.hookType,
					"hook path should contain hook type")
				require.Contains(t, hookPath, ".git/hooks",
					"hook path should be in hooks directory")

				// Verify hooks directory creation
				err = pathValidator.EnsureHooksDirectory(testCase.repoPath)
				require.NoError(t, err, "should be able to ensure hooks directory exists")
			}
		})
	}
}

// TestRepoPathSecurityIntegration tests security aspects of --repo-path in integration scenarios.
func TestRepoPathSecurityIntegration(t *testing.T) {
	// Create legitimate repository
	validRepoDir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = validRepoDir
	err := cmd.Run()
	require.NoError(t, err)

	// Test path validator security
	pathValidator := cli.StrictPathValidator{}

	tests := []struct {
		name          string
		repoPath      string
		hookType      string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:        "normal repository path",
			repoPath:    validRepoDir,
			hookType:    "commit-msg",
			expectError: false,
			description: "should accept normal repository paths",
		},
		{
			name:          "path traversal attack",
			repoPath:      "../../../etc",
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject path traversal attempts",
		},
		{
			name:          "null byte injection",
			repoPath:      "/tmp/repo\x00/malicious",
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject paths with null bytes",
		},
		{
			name:          "command injection in hook type",
			repoPath:      validRepoDir,
			hookType:      "commit-msg; rm -rf /",
			expectError:   true,
			errorContains: "invalid hook type",
			description:   "should reject malicious hook types",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			hookPath, err := pathValidator.ValidateHookPath(testCase.repoPath, testCase.hookType)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
				require.Contains(t, err.Error(), testCase.errorContains, testCase.description)
				require.Empty(t, hookPath)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, hookPath)
			}
		})
	}
}

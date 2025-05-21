// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package fsutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// MaxPathLength defines a safe maximum path length limit to prevent resource exhaustion.
// Most filesystems have a limit of 4096, but we're being more conservative.
const MaxPathLength = 1024

// ValidateGitRepoPath performs extensive validation of repository paths with additional
// checks specific to git repositories. This function is a comprehensive, secure path
// validator that handles all edge cases and prevents security vulnerabilities.
// Returns the canonical path to the repository root.
func ValidateGitRepoPath(repoPath string) (string, error) {
	// Empty path validation
	if repoPath == "" {
		return "", errors.New("repository path cannot be empty")
	}

	// Normalize the path to handle .. sequences properly
	cleanPath := filepath.Clean(repoPath)

	// Convert to absolute path for consistent validation
	if !filepath.IsAbs(cleanPath) {
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", fmt.Errorf("could not resolve absolute path: %w", err)
		}

		cleanPath = absPath
	}

	// Check maximum path length to prevent resource exhaustion
	if len(cleanPath) > MaxPathLength {
		return "", fmt.Errorf("repository path exceeds maximum length (%d characters)", MaxPathLength)
	}

	// Resolve symlinks to get canonical path
	canonPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return "", fmt.Errorf("path evaluation failed: %w", err)
	}

	// Validate the path exists and has proper permissions using a file descriptor
	// to prevent TOCTOU race conditions
	dirFile, err := os.Open(canonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("repository path does not exist: %s", canonPath)
		}

		if os.IsPermission(err) {
			return "", fmt.Errorf("insufficient permissions to access repository path: %s", canonPath)
		}

		return "", fmt.Errorf("invalid repository path: %w", err)
	}
	defer dirFile.Close()

	// Get file info directly from the file descriptor
	fileInfo, err := dirFile.Stat()
	if err != nil {
		return "", fmt.Errorf("could not stat repository path: %w", err)
	}

	// Ensure it's a directory
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("repository path is not a directory: %s", canonPath)
	}

	// Check for read/write permissions
	if fileInfo.Mode().Perm()&0600 != 0600 {
		return "", fmt.Errorf("insufficient permissions to read/write in repository path: %s", canonPath)
	}

	// Check for .git directory to verify it's likely a git repository
	gitPath := filepath.Join(canonPath, ".git")

	gitInfo, err := os.Stat(gitPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not a git repository (no .git directory found): %s", canonPath)
		}

		return "", fmt.Errorf("error accessing .git directory: %w", err)
	}

	// Verify .git is a directory (could be a file in submodules, but that's a special case)
	if !gitInfo.IsDir() {
		// This is a special case for git submodules where .git is a file
		// We could add additional handling for this case if needed
		return "", fmt.Errorf("invalid git repository structure: %s", canonPath)
	}

	return canonPath, nil
}

// FindGitRoot attempts to find the git repository root starting from the given path.
// It searches up the directory tree looking for a .git directory.
// Returns the absolute, canonical path to the repository root.
func FindGitRoot(startPath string) (string, error) {
	// First, make sure the start path is valid
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("invalid start path: %w", err)
	}

	// Search for .git directory by walking up the tree
	currentPath := absPath

	for {
		// Check if .git exists in the current directory
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			// Found git root - return canonical path
			return filepath.EvalSymlinks(currentPath)
		}

		// Go up one directory
		parentPath := filepath.Dir(currentPath)

		// If we've reached the top of the filesystem without finding .git
		if parentPath == currentPath {
			return "", fmt.Errorf("no git repository found in %s or its parent directories", absPath)
		}

		currentPath = parentPath
	}
}

// IsValidGitHookType verifies if a hook type is valid to prevent command injection.
// Git hooks need to have specific, well-known names.
func IsValidGitHookType(hookType string) bool {
	validHooks := map[string]bool{
		"commit-msg":            true,
		"pre-commit":            true,
		"pre-push":              true,
		"pre-receive":           true,
		"update":                true,
		"post-update":           true,
		"post-receive":          true,
		"post-merge":            true,
		"pre-rebase":            true,
		"pre-applypatch":        true,
		"post-applypatch":       true,
		"post-checkout":         true,
		"post-commit":           true,
		"pre-auto-gc":           true,
		"post-rewrite":          true,
		"sendemail-validate":    true,
		"fsmonitor-watchman":    true,
		"p4-pre-submit":         true,
		"p4-prepare-changelist": true,
		"p4-changelist":         true,
		"p4-post-changelist":    true,
		"prepare-commit-msg":    true,
	}

	// Check if hookType is in the valid hooks map
	_, exists := validHooks[hookType]

	return exists
}

// GetGitHookPath safely computes the path to a git hook.
// Given a repository path and hook type, returns the absolute path to the hook.
// This function handles path validation and security checks.
func GetGitHookPath(repoPath, hookType string) (string, error) {
	// Validate repository path
	canonPath, err := ValidateGitRepoPath(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Validate hook type
	if !IsValidGitHookType(hookType) {
		return "", fmt.Errorf("invalid hook type: %s", hookType)
	}

	// Use SafeJoin to safely get hooks directory path
	hooksDir, err := SafeJoin(canonPath, ".git", "hooks")
	if err != nil {
		return "", fmt.Errorf("invalid hooks directory: %w", err)
	}

	// Use SafeJoin to safely get hook path
	hookPath, err := SafeJoin(hooksDir, hookType)
	if err != nil {
		return "", fmt.Errorf("invalid hook path: %w", err)
	}

	// Additional verification that hook path is within git directory
	isWithin, err := IsWithinDirectory(hookPath, canonPath)
	if err != nil {
		return "", fmt.Errorf("path validation error: %w", err)
	}

	if !isWithin {
		return "", errors.New("invalid hook path: path traversal detected")
	}

	return hookPath, nil
}

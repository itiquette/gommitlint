// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package fsutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MaxPathLength defines a safe maximum path length limit to prevent resource exhaustion.
// Most filesystems have a limit of 4096, but we're being more conservative.
const MaxPathLength = 1024

// MaxSymlinkDepth defines the maximum depth of symlink resolution to prevent
// infinite loops from circular symlinks or excessive resource consumption.
const MaxSymlinkDepth = 10

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

	// Reject suspicious path patterns
	if strings.Contains(cleanPath, "\\") ||
		strings.Contains(cleanPath, "%2e%2e") || // URL-encoded ..
		strings.Contains(cleanPath, "%2E%2E") {
		return "", fmt.Errorf("repository path contains suspicious character sequences: %s", repoPath)
	}

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

	// Resolve symlinks to get canonical path with custom depth limiting
	canonPath, err := safeResolveSymlinks(cleanPath, MaxSymlinkDepth)
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

	// Validate that the .git directory exists
	gitDirPath := filepath.Join(canonPath, ".git")

	gitPath, err := verifyGitDirectory(gitDirPath)
	if err != nil {
		return "", err
	}

	// Validate that the git directory is within the repository root
	isWithin, err := IsWithinDirectory(gitPath, canonPath)
	if err != nil {
		return "", fmt.Errorf("path validation error: %w", err)
	}

	if !isWithin {
		return "", errors.New("invalid git directory: path traversal detected")
	}

	return canonPath, nil
}

// safeResolveSymlinks resolves symlinks in a path with a maximum depth limit
// to prevent infinite loops and excessive resource consumption.
func safeResolveSymlinks(path string, maxDepth int) (string, error) {
	if maxDepth <= 0 {
		return "", errors.New("maximum symlink resolution depth exceeded")
	}

	// Get file info to check if it's a symlink
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	// If not a symlink, we're done with this component
	if info.Mode()&os.ModeSymlink == 0 {
		// If it's a directory, it's a final path component
		if info.IsDir() {
			return filepath.Abs(path)
		}

		// If it's a file, resolve its directory
		dir := filepath.Dir(path)

		resolvedDir, err := safeResolveSymlinks(dir, maxDepth)
		if err != nil {
			return "", err
		}

		return filepath.Join(resolvedDir, filepath.Base(path)), nil
	}

	// It's a symlink, read the link target
	linkTarget, err := os.Readlink(path)
	if err != nil {
		return "", err
	}

	// If relative link, join with parent directory
	if !filepath.IsAbs(linkTarget) {
		dir := filepath.Dir(path)
		linkTarget = filepath.Join(dir, linkTarget)
	}

	// Clean up the path
	linkTarget = filepath.Clean(linkTarget)

	// Recurse to handle multi-level symlinks, decrementing depth counter
	return safeResolveSymlinks(linkTarget, maxDepth-1)
}

// verifyGitDirectory checks that the .git directory is valid.
// It handles both regular git repositories and git submodules.
func verifyGitDirectory(gitDirPath string) (string, error) {
	// First check if .git exists
	fileInfo, err := os.Lstat(gitDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("not a git repository (no .git directory found)")
		}

		return "", fmt.Errorf("error accessing .git directory: %w", err)
	}

	// Check if it's a directory (normal git repository)
	if fileInfo.IsDir() {
		// Verify essential git files exist to confirm it's a real git repository
		configPath := filepath.Join(gitDirPath, "config")
		if _, err := os.Stat(configPath); err != nil {
			return "", errors.New("invalid git repository (missing config file)")
		}

		// Looks like a valid git directory
		return gitDirPath, nil
	}

	// Not a directory - could be a git submodule with a gitfile link
	if fileInfo.Mode().IsRegular() {
		// Try to read it as a gitfile (used in submodules)
		content, err := os.ReadFile(gitDirPath)
		if err != nil {
			return "", fmt.Errorf("error reading git file: %w", err)
		}

		// Parse the gitfile content
		gitfileContent := strings.TrimSpace(string(content))
		if strings.HasPrefix(gitfileContent, "gitdir: ") {
			// Extract the actual git directory path
			actualGitDir := strings.TrimPrefix(gitfileContent, "gitdir: ")

			// Resolve the path if relative
			if !filepath.IsAbs(actualGitDir) {
				repoRoot := filepath.Dir(gitDirPath)
				actualGitDir = filepath.Join(repoRoot, actualGitDir)
			}

			// Confirm this is a real git directory
			if _, err := os.Stat(filepath.Join(actualGitDir, "config")); err != nil {
				return "", errors.New("invalid git submodule reference")
			}

			// Validate that this path isn't trying to escape
			cleanedPath := filepath.Clean(actualGitDir)
			if strings.Contains(cleanedPath, "..") {
				return "", errors.New("invalid git submodule path (contains path traversal)")
			}

			return cleanedPath, nil
		}

		return "", errors.New("invalid git file format")
	}

	// Neither a directory nor a regular file
	return "", errors.New("invalid git repository structure")
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

		// Use Lstat to avoid following symlinks directly
		info, err := os.Lstat(gitPath)
		if err == nil {
			// .git exists - check if it's a directory or gitfile
			if info.IsDir() || info.Mode().IsRegular() {
				// Verify it's a legitimate git directory/file
				_, err := verifyGitDirectory(gitPath)
				if err == nil {
					// Found valid git root - resolve path safely
					return safeResolveSymlinks(currentPath, MaxSymlinkDepth)
				}
			}
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

	// Get git directory path - this handles submodules correctly
	gitDir := filepath.Join(canonPath, ".git")

	gitDirPath, err := verifyGitDirectory(gitDir)
	if err != nil {
		return "", fmt.Errorf("invalid git directory: %w", err)
	}

	// Use SafeJoin to safely get hooks directory path
	hooksDir, err := SafeJoin(gitDirPath, "hooks")
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

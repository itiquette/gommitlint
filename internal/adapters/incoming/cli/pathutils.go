// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// validateRepositoryPath performs extensive validation of repository paths to prevent security issues.
// This pure function implements security best practices for path validation.
func validateRepositoryPath(repoPath string) (string, error) {
	// Empty path validation
	if repoPath == "" {
		return "", errors.New("repository path cannot be empty")
	}

	// Normalize the path to prevent path traversal
	repoPath = filepath.Clean(repoPath)

	// Convert to absolute path for consistent validation
	if !filepath.IsAbs(repoPath) {
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return "", fmt.Errorf("could not resolve absolute path: %w", err)
		}

		repoPath = absPath
	}

	// Check maximum path length to prevent resource exhaustion
	// Most filesystems have a limit of 4096, but we'll be more conservative
	const maxPathLength = 1024
	if len(repoPath) > maxPathLength {
		return "", fmt.Errorf("repository path exceeds maximum length (%d characters)", maxPathLength)
	}

	// Validate the path exists and has proper permissions
	fileInfo, err := os.Stat(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("repository path does not exist: %s", repoPath)
		}

		if os.IsPermission(err) {
			return "", fmt.Errorf("insufficient permissions to access repository path: %s", repoPath)
		}

		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Ensure it's a directory
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("repository path is not a directory: %s", repoPath)
	}

	// Optional: Check if there are read/write permissions
	// This is a basic check; actual permissions will be fully tested when accessing files
	if fileInfo.Mode().Perm()&0600 != 0600 {
		return "", fmt.Errorf("insufficient permissions to read/write in repository path: %s", repoPath)
	}

	return repoPath, nil
}

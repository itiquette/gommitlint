// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolvePath resolves a configuration path, handling relative paths
// and empty values by returning a default. It follows these rules:
// 1. If the path is empty, return the defaultPath
// 2. If the path is relative, join it with the working directory
// 3. If the path is absolute, return it as is
//
// If path retrieval fails, the defaultPath is returned.
func ResolvePath(config Config, key string, defaultPath string) string {
	// Retrieve path from config
	path := config.GetString(key)
	if path == "" {
		return defaultPath
	}

	// Clean the path to normalize it
	path = filepath.Clean(path)

	// Handle relative vs absolute paths
	if !filepath.IsAbs(path) {
		// Get current working directory
		wd, err := os.Getwd()
		if err != nil {
			return defaultPath
		}

		return filepath.Join(wd, path)
	}

	return path
}

// EnsureDirectory ensures a directory exists with the specified permissions.
// It returns an error if the path is empty or if there's a failure creating the directory.
// It does nothing if the directory already exists.
func EnsureDirectory(path string, perm os.FileMode) error {
	if path == "" {
		return os.ErrInvalid
	}

	// Clean and normalize the path
	path = filepath.Clean(path)

	// Check if directory already exists
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return nil // Directory exists
		}

		return fmt.Errorf("path exists but is not a directory: %s", path)
	} else if !os.IsNotExist(err) {
		return err // Some other error occurred
	}

	// Create directory and all parent directories
	return os.MkdirAll(path, perm)
}

// ResolveFilePath is similar to ResolvePath but ensures the parent directory
// exists with the given permissions. It's useful when working with file paths
// where the directory may need to be created.
func ResolveFilePath(config Config, key, defaultPath string, perm os.FileMode) (string, error) {
	path := ResolvePath(config, key, defaultPath)
	dir := filepath.Dir(path)

	// Ensure the directory exists
	if err := EnsureDirectory(dir, perm); err != nil {
		return "", fmt.Errorf("failed to ensure directory exists: %w", err)
	}

	return path, nil
}

// Note: SafeJoin has been moved to the fsutils package
// Use "github.com/itiquette/gommitlint/internal/common/fsutils" for path joining operations

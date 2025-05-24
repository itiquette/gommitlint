// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides test utilities for configuration handling.
// This package is intended for testing only and should not be used in production code.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itiquette/gommitlint/internal/common/config"
)

// EnsureDirectory ensures a directory exists with the specified permissions.
// It returns an error if the path is empty or if there's a failure creating the directory.
// It does nothing if the directory already exists.
//
// This function is intended for testing only.
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
//
// This function is intended for testing only.
func ResolveFilePath(cfg config.Config, key, defaultPath string, perm os.FileMode) (string, error) {
	path := config.ResolvePath(cfg, key, defaultPath)
	dir := filepath.Dir(path)

	// Ensure the directory exists
	if err := EnsureDirectory(dir, perm); err != nil {
		return "", fmt.Errorf("failed to ensure directory exists: %w", err)
	}

	return path, nil
}

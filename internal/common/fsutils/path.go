// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package fsutils provides file system utility functions for path handling,
// file operations, and directory management with security as a priority.
package fsutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SanitizePath sanitizes and validates a directory path with robust security checks.
// It resolves symbolic links, ensures the path is a directory, and performs
// several security validations to prevent common vulnerabilities.
// This function follows value semantics - it doesn't modify the input path.
func SanitizePath(path string) (string, error) {
	// Normalize path to handle .. sequences properly
	cleanPath := filepath.Clean(path)

	// Get absolute path (applies further normalization)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if path has appropriate canonical form and resolve symlinks
	canonPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", fmt.Errorf("path evaluation failed: %w", err)
	}

	// Check if path exists and is a directory (after symlink resolution)
	fileInfo, err := os.Stat(canonPath)
	if err != nil {
		return "", fmt.Errorf("path error: %w", err)
	}

	if !fileInfo.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", canonPath)
	}

	return canonPath, nil
}

// IsWithinDirectory checks if a path is contained within a base directory.
// This function helps prevent directory traversal attacks by ensuring
// one path is contained within another. It resolves symlinks to their
// real paths before checking containment.
func IsWithinDirectory(path, baseDir string) (bool, error) {
	// Ensure both paths are absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("invalid path: %w", err)
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return false, fmt.Errorf("invalid base directory: %w", err)
	}

	// Check if paths exist before trying to resolve symlinks
	_, err = os.Stat(absPath)
	if err != nil {
		// Allow checking paths that don't exist yet, but fail on other errors
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("path error: %w", err)
		}
	} else {
		// Path exists, resolve symlinks
		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return false, fmt.Errorf("path evaluation failed: %w", err)
		}

		absPath = resolvedPath
	}

	_, err = os.Stat(absBaseDir)
	if err != nil {
		return false, fmt.Errorf("base directory error: %w", err)
	}

	// Always resolve symlinks in the base directory
	canonBaseDir, err := filepath.EvalSymlinks(absBaseDir)
	if err != nil {
		return false, fmt.Errorf("base directory evaluation failed: %w", err)
	}

	// Calculate a relative path to verify containment
	relPath, err := filepath.Rel(canonBaseDir, absPath)
	if err != nil {
		return false, err
	}

	// A path is within base dir if the relative path doesn't start with ".."
	return !strings.HasPrefix(relPath, ".."), nil
}

// SafeJoin safely joins path elements, preventing path traversal attacks.
// It rejects any element containing ".." sequences and ensures the final
// path remains within the base directory.
func SafeJoin(baseDir string, elements ...string) (string, error) {
	// Verify baseDir exists and is a directory
	fi, err := os.Stat(baseDir)
	if err != nil {
		return "", fmt.Errorf("invalid base directory: %w", err)
	}

	if !fi.IsDir() {
		return "", fmt.Errorf("base path is not a directory: %s", baseDir)
	}

	// Clean base path
	cleanBase := filepath.Clean(baseDir)

	// Check each element and build path
	result := cleanBase

	for _, element := range elements {
		// Reject path traversal attempts
		if strings.Contains(element, "..") {
			return "", fmt.Errorf("path element contains illegal sequence: %s", element)
		}

		result = filepath.Join(result, element)
	}

	// Final safety check - ensure result is within baseDir
	relPath, err := filepath.Rel(cleanBase, result)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("resulting path escapes base directory: %s", result)
	}

	return result, nil
}

// SafeReadDirectoryContents reads directory contents using file descriptors
// to prevent Time-of-Check/Time-of-Use (TOCTOU) race conditions.
func SafeReadDirectoryContents(dirPath string) ([]os.FileInfo, error) {
	// Open the directory to get a file descriptor
	dirFile, err := os.Open(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}
	defer dirFile.Close()

	// Get file info on the directory descriptor to verify it's a directory
	fileInfo, err := dirFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dirPath)
	}

	// Read directory contents using the file descriptor
	entries, err := dirFile.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	return entries, nil
}

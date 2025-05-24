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
// It rejects any element containing ".." sequences, guards against symlink attacks,
// and ensures the final path remains within the base directory using multiple
// layers of validation.
func SafeJoin(baseDir string, elements ...string) (string, error) {
	// Verify baseDir exists and is a directory using a file descriptor
	// to prevent TOCTOU race conditions
	dirFile, err := os.Open(baseDir)
	if err != nil {
		return "", fmt.Errorf("invalid base directory: %w", err)
	}
	defer dirFile.Close()

	// Get file info directly from the file descriptor
	fi, err := dirFile.Stat()
	if err != nil {
		return "", fmt.Errorf("cannot stat base directory: %w", err)
	}

	if !fi.IsDir() {
		return "", fmt.Errorf("base path is not a directory: %s", baseDir)
	}

	// Clean base path and get its absolute form
	absBaseDir, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", fmt.Errorf("cannot get absolute base directory: %w", err)
	}

	// Resolve base directory symlinks for security
	canonBase, err := filepath.EvalSymlinks(absBaseDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base directory symlinks: %w", err)
	}

	// Check each element and build path incrementally
	// Start with the canonical base path
	result := canonBase

	// Additional security check - verify each component individually
	for elementIndex, element := range elements {
		// Reject empty elements
		if element == "" {
			continue
		}

		// First clean the element to normalize it
		cleanElement := filepath.Clean(element)

		// Comprehensive path traversal blocking
		// 1. Check for parent directory references (..)
		if strings.Contains(element, "..") || strings.Contains(cleanElement, "..") {
			return "", fmt.Errorf("path element contains illegal sequence: %s", element)
		}

		// 2. Check for other suspicious patterns
		if strings.Contains(cleanElement, "\\") ||
			strings.Contains(cleanElement, "%2e%2e") || // URL-encoded ..
			strings.Contains(cleanElement, "%2E%2E") {
			return "", fmt.Errorf("path element contains suspicious character sequence: %s", element)
		}

		// 3. Reject absolute paths in elements
		if filepath.IsAbs(cleanElement) {
			return "", fmt.Errorf("path element cannot be absolute: %s", element)
		}

		// Safely join the element
		nextResult := filepath.Join(result, cleanElement)

		// Perform thorough path containment verification
		// 4. Check path containment using relative path
		relPath, err := filepath.Rel(canonBase, nextResult)
		if err != nil {
			return "", fmt.Errorf("path calculation error: %w", err)
		}

		if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, "/../") {
			return "", fmt.Errorf("path element causes directory traversal: %s", element)
		}

		// 5. Check path containment using string prefix (belt and suspenders approach)
		// Ensure resulting path has canonical base as prefix
		if !strings.HasPrefix(nextResult, canonBase+string(filepath.Separator)) && nextResult != canonBase {
			return "", fmt.Errorf("path escapes base directory: %s", element)
		}

		// 6. For all but the last element, verify each component if it exists
		// This prevents symlink trickery where intermediate components are symlinks
		if elementIndex < len(elements)-1 {
			if _, err := os.Lstat(nextResult); err == nil {
				// Path exists, check if it's a symlink
				realPath, err := filepath.EvalSymlinks(nextResult)
				if err != nil {
					return "", fmt.Errorf("failed to evaluate potential symlink: %w", err)
				}

				// Verify the resolved path is still within the base directory
				if !strings.HasPrefix(realPath, canonBase) {
					return "", fmt.Errorf("symlink in path escapes base directory: %s", element)
				}

				// Use the resolved path going forward if it exists
				nextResult = realPath
			}
		}

		// Update our result for the next iteration
		result = nextResult
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

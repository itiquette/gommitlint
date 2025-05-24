// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package fsutils

import (
	"fmt"
	"os"
	"path/filepath"
)

// SafeWriteFile writes a file atomically with proper permissions and error handling.
// This prevents TOCTOU vulnerabilities by writing to a temporary file first and then
// atomically moving it to the destination. It sets file permissions before writing any data.
func SafeWriteFile(path string, data []byte, perm os.FileMode) error {
	// Get the directory portion of the destination path
	dir := filepath.Dir(path)

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create a temporary file in the same directory for atomic operations
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpPath := tmpFile.Name()

	// Ensure temporary file is cleaned up on error
	successful := false
	defer func() {
		// Close the file if it's still open
		tmpFile.Close()

		// If the operation wasn't successful, remove the temp file
		if !successful {
			os.Remove(tmpPath)
		}
	}()

	// Set permissions before writing content (important for security)
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Write data to the temporary file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file data: %w", err)
	}

	// Close the file explicitly before rename
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically move the temporary file to the destination
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to move temporary file to destination: %w", err)
	}

	// Mark operation as successful so the deferred cleanup doesn't remove the file
	successful = true

	return nil
}

// FindFilesWithExtensions returns all files in a directory with the specified extensions.
// It does not recurse into subdirectories.
func FindFilesWithExtensions(dir string, extensions []string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Convert extensions slice to a map for O(1) lookups
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[ext] = true
	}

	// Pre-allocate slice with a reasonable capacity
	// A good estimate is to allocate for half the entries
	// since not all entries will match our extensions
	files := make([]string, 0, len(entries)/2)

	// Filter and map entries to full paths
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if extMap[ext] {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

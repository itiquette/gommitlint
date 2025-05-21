// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package fsutils

import (
	"fmt"
	"os"
	"path/filepath"
)

// SafeReadFile reads a file safely, with proper error handling.
// It follows value semantics by returning the file content without modifying any state.
func SafeReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file %s could not be read: %w", path, err)
	}

	return content, nil
}

// FindFilesWithExtensions returns all files in a directory with the specified extensions.
// It does not recurse into subdirectories.
func FindFilesWithExtensions(dir string, extensions []string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Create extension map for faster lookup
	extMap := make(map[string]bool, len(extensions))
	for _, ext := range extensions {
		extMap[ext] = true
	}

	// Pre-allocate slice with a reasonable capacity
	// A good estimate is to allocate for half the entries
	// since not all entries will match our extensions
	files := make([]string, 0, len(entries)/2)

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

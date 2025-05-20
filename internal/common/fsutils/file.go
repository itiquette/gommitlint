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
// It does not recurse into subdirectories and maintains value semantics by not modifying input.
func FindFilesWithExtensions(dir string, extensions []string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range extensions {
			if filepath.Ext(name) == ext {
				files = append(files, filepath.Join(dir, name))

				break
			}
		}
	}

	return files, nil
}

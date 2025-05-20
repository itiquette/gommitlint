// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package fsutils

import (
	"fmt"
	"os"
	"path/filepath"
)

// SanitizePath converts a path to an absolute path and confirms it is a directory.
// It returns the absolute path if valid, or an error if the path does not exist or is not a directory.
// This function follows value semantics - it doesn't modify the input path.
func SanitizePath(path string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Verify the path exists and is a directory
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("path error: %w", err)
	}

	if !fileInfo.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", absPath)
	}

	return absPath, nil
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
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

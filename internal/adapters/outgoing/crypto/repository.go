// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package crypto provides adapters for cryptographic operations.
package crypto

import (
	"os"
	"path/filepath"
)

// KeyRepository defines a repository for accessing cryptographic keys.
type KeyRepository interface {
	// GetKeyDirectory returns the directory containing trusted keys.
	GetKeyDirectory() string

	// FindKeyFiles finds key files matching the given extensions.
	FindKeyFiles(extensions []string) ([]string, error)
}

// FileSystemKeyRepository implements KeyRepository using the local filesystem.
type FileSystemKeyRepository struct {
	keyDir string
}

// NewFileSystemKeyRepository creates a new repository for accessing keys from a directory.
func NewFileSystemKeyRepository(keyDir string) *FileSystemKeyRepository {
	return &FileSystemKeyRepository{
		keyDir: keyDir,
	}
}

// GetKeyDirectory returns the directory containing trusted keys.
func (r *FileSystemKeyRepository) GetKeyDirectory() string {
	return r.keyDir
}

// FindKeyFiles finds key files matching the given extensions.
func (r *FileSystemKeyRepository) FindKeyFiles(extensions []string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(r.keyDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range extensions {
			if filepath.Ext(name) == ext {
				files = append(files, filepath.Join(r.keyDir, name))

				break
			}
		}
	}

	return files, nil
}

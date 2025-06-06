// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package crypto provides adapters for cryptographic operations.
package signing

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileSystemKeyRepository implements domain.CryptoKeyRepository using the local filesystem.
type FileSystemKeyRepository struct {
	keyDir string
}

// FileSystemKeyRepository is kept as an internal implementation detail.
// It no longer implements any domain interface following the functional simplification.

// RepositoryOption configures a FileSystemKeyRepository.
type RepositoryOption func(*FileSystemKeyRepository)

// NewFileSystemKeyRepository creates a new repository for accessing keys from a directory.
func NewFileSystemKeyRepository(keyDir string) *FileSystemKeyRepository {
	return NewFileSystemKeyRepositoryWithOptions(keyDir)
}

// NewFileSystemKeyRepositoryWithOptions creates a new repository with options.
// This allows for creating test repositories with different security settings.
func NewFileSystemKeyRepositoryWithOptions(keyDir string, options ...RepositoryOption) *FileSystemKeyRepository {
	// Create repository
	repo := &FileSystemKeyRepository{
		keyDir: keyDir,
	}

	// Apply options
	for _, option := range options {
		option(repo)
	}

	return repo
}

// GetKeyDirectory returns the directory containing trusted keys.
func (r FileSystemKeyRepository) GetKeyDirectory() string {
	return r.keyDir
}

// FindKeyFiles finds key files matching the given extensions.
// Implements improved path security by using safe directory access and path validation.
func (r FileSystemKeyRepository) FindKeyFiles(extensions []string) ([]string, error) {
	// Use the safer directory reading function to prevent TOCTOU race conditions
	entries, err := r.readDirSafely()
	if err != nil {
		return nil, fmt.Errorf("failed to read key directory: %w", err)
	}

	// Create an initial empty slice - each append operation will create a new slice
	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range extensions {
			if filepath.Ext(name) == ext {
				// Use SafeJoin for path safety
				safePath, err := SafeJoin(r.keyDir, name)
				if err != nil {
					// Log the error but continue processing other files
					continue
				}

				// Create a new slice with the additional path
				files = append(files, safePath)

				break
			}
		}
	}

	return files, nil
}

// ReadKeyFile reads a key file with appropriate security checks based on the file type.
func (r FileSystemKeyRepository) ReadKeyFile(path string) ([]byte, error) {
	// Determine security level based on file type and extension
	level := GetSecurityLevelForFile(path)

	// Check permissions
	secure, err := IsSecurePermissions(path, level)
	if err != nil {
		return nil, fmt.Errorf("security check failed for %s: %w", path, err)
	}

	// Report insecure permissions
	if !secure {
		return nil, fmt.Errorf("insecure permissions on key file %s", path)
	}

	// For private key files, also verify ownership
	if level == Private {
		owned, err := VerifyOwnership(path)
		if err != nil {
			return nil, fmt.Errorf("ownership check failed: %w", err)
		}

		if !owned {
			return nil, fmt.Errorf("private key file %s is not owned by current user", path)
		}
	}

	// Now safe to read
	return os.ReadFile(path)
}

// readDirSafely reads the key directory using file descriptor operations
// to prevent Time-of-Check/Time-of-Use (TOCTOU) race conditions.
func (r FileSystemKeyRepository) readDirSafely() ([]os.DirEntry, error) {
	// Open the directory to get a file descriptor
	dirFile, err := os.Open(r.keyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open key directory: %w", err)
	}
	defer dirFile.Close()

	// Get file info on the directory descriptor to verify it's a directory
	fileInfo, err := dirFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat key directory: %w", err)
	}

	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("key path is not a directory: %s", r.keyDir)
	}

	// Read directory contents using the file descriptor
	entries, err := dirFile.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	return entries, nil
}

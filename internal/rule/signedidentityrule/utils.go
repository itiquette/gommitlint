// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
)

// sanitizePath validates and sanitizes a directory path.
func sanitizePath(path string) (string, error) {
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

// safeReadFile reads a file with file locking to prevent race conditions.
func safeReadFile(path string) ([]byte, error) {
	// Create a flock
	fileLock := flock.New(path)

	locked, err := fileLock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to lock file: %w", err)
	}

	if !locked {
		return nil, errors.New("file is currently locked by another process")
	}

	defer func() {
		err := fileLock.Unlock()
		if err != nil {
			fmt.Printf("failed to unlock file: %v", err)
		}
	}()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file %s could not be read: %w", path, err)
	}

	return content, nil
}

// getCommitBytes returns the commit data as bytes for signature verification.
func getCommitBytes(commit *object.Commit) ([]byte, error) {
	encoded := &plumbing.MemoryObject{}
	if err := commit.EncodeWithoutSignature(encoded); err != nil {
		return nil, fmt.Errorf("failed to encode commit: %w", err)
	}

	reader, err := encoded.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to read commit: %w", err)
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// findKeyFiles returns all files in dir with any of the given extensions.
func findKeyFiles(dir string, extensions []string, fileType string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != dir {
			return filepath.SkipDir // Don't recurse into subdirectories
		}

		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				// Validate key file content before adding
				if isValidKeyFile(path, fileType) {
					files = append(files, path)
				}

				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// isValidKeyFile validates key file content before processing.
func isValidKeyFile(path string, fileType string) bool {
	data, err := safeReadFile(path)
	if err != nil {
		return false
	}

	content := string(data)

	// Basic validation based on file type
	switch fileType {
	case GPG:
		return strings.Contains(content, "BEGIN PGP PUBLIC KEY BLOCK") ||
			strings.Contains(content, "BEGIN PGP PRIVATE KEY BLOCK")
	case SSH:
		return strings.HasPrefix(content, "ssh-") ||
			strings.HasPrefix(content, "ecdsa-") ||
			strings.HasPrefix(content, "sk-") ||
			strings.Contains(content, " ssh-")
	default:
		return false
	}
}

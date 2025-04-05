// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gofrs/flock"
)

// sanitizePath validates and sanitizes a directory path for security.
//
// Parameters:
//   - path: The file path to validate and sanitize
//
// The function performs several security checks:
//  1. Converts the path to an absolute path to eliminate relative path issues
//  2. Verifies that the path exists in the filesystem
//  3. Confirms that the path is a directory, not a file
//
// These checks help prevent path traversal attacks and ensure the path
// is valid before attempting cryptographic operations that depend on it.
//
// Returns:
//   - string: The sanitized absolute path
//   - error: Any error encountered during validation
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
//
// Parameters:
//   - path: Path to the file to read
//
// The function uses file locking to ensure thread-safe reading, which is
// especially important when multiple processes might access the same key files.
// This prevents potential race conditions that could occur during signature
// verification operations.
//
// Returns:
//   - []byte: The file content as a byte slice
//   - error: Any error encountered during file reading or locking
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
//
// Parameters:
//   - commit: The git commit object to encode
//
// The function extracts the commit data without its signature, which is
// necessary for signature verification. When verifying a signature, we need
// the exact bytes that were originally signed, which excludes the signature
// itself.
//
// Returns:
//   - []byte: The encoded commit data ready for signature verification
//   - error: Any error encountered during the encoding process
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

// findKeyFiles returns all files in a directory with specified extensions.
//
// Parameters:
//   - dir: Directory to search for key files
//   - extensions: List of file extensions to look for (e.g., ".pub", ".ssh")
//   - fileType: Type of key file to look for ("GPG" or "SSH")
//
// The function walks through the specified directory (not recursively) and
// identifies files that match the given extensions. It also validates the
// file content to ensure it actually contains a key of the specified type.
//
// Returns:
//   - []string: List of paths to valid key files
//   - error: Any error encountered during directory traversal
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
//
// Parameters:
//   - path: Path to the key file to validate
//   - fileType: Type of key to validate ("GPG" or "SSH")
//
// The function performs basic content validation to ensure a file with the right
// extension actually contains key data of the expected type. This helps filter out
// files that might have the right extension but contain invalid or incorrect data.
//
// Returns:
//   - bool: true if the file contains valid key data of the specified type, false otherwise
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

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signing

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Level defines required permission security levels.
type Level int

const (
	// Basic allows default permissions.
	Basic Level = iota
	// Safe prevents world-writable permissions.
	Safe
	// Restricted limits access to owner and possibly group.
	Restricted
	// Private requires owner-only permissions.
	Private
)

// File permission security functions

// IsSecurePermissions checks if a file has appropriately secure permissions.
func IsSecurePermissions(path string, level Level) (bool, error) {
	// Open file to get a file descriptor - prevents TOCTOU issues
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("cannot access file: %w", err)
	}
	defer file.Close()

	// Get file info directly from the file descriptor
	info, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("cannot stat file: %w", err)
	}

	mode := info.Mode().Perm()

	switch level {
	case Private:
		// 0600 or stricter - owner read/write only
		return mode&0077 == 0, nil
	case Restricted:
		// No world access, group read at most
		return (mode&0007 == 0) && (mode&0020 == 0), nil
	case Safe:
		// Not world-writable
		return mode&0002 == 0, nil
	case Basic:
		return true, nil
	}

	// Default case - most permissive
	return true, nil
}

// VerifyOwnership checks if current user owns the file.
func VerifyOwnership(path string) (bool, error) {
	// Open file to get a file descriptor - prevents TOCTOU issues
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("cannot access file: %w", err)
	}
	defer file.Close()

	// Get file info directly from the file descriptor
	info, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("cannot stat file: %w", err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false, errors.New("could not get detailed file info")
	}

	// Get current user's UID
	currentUID := os.Getuid()
	if currentUID < 0 {
		return false, errors.New("invalid user ID")
	}

	// Get the file's UID
	fileUID := stat.Uid

	// Instead of converting the int to uint32, which the linter flags as potentially
	// dangerous, we'll convert the uint32 to int64 which can safely hold both values
	// This avoids the conversion warning while still performing the same check
	return int64(fileUID) == int64(currentUID), nil
}

// SetSecurePermissions sets appropriate permissions for a file based on security level.
func SetSecurePermissions(path string, level Level) error {
	var mode os.FileMode

	switch level {
	case Private:
		// Owner read/write only
		mode = 0600
	case Restricted:
		// Owner read/write, group read
		mode = 0640
	case Safe:
		// Owner read/write, group read, world read
		mode = 0644
	case Basic:
		// Standard permissions
		mode = 0644
	default:
		// Conservative default
		mode = 0644
	}

	// For executable files, add execute permission where read permission exists
	if strings.HasSuffix(path, ".sh") || strings.HasSuffix(path, ".bash") {
		mode |= 0111
	}

	return os.Chmod(path, mode)
}

// SecureVerifyFileExists checks if a file exists and has appropriate permissions.
func SecureVerifyFileExists(path string, level Level) (bool, error) {
	// Try to open the file directly - prevents TOCTOU issues
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("error checking file: %w", err)
	}
	defer file.Close()

	// File exists, now check permissions
	isSecure, err := IsSecurePermissions(path, level)
	if err != nil {
		return false, err
	}

	return isSecure, nil
}

// GetSecurityLevelForFile determines the appropriate security level for a file based on its extension and name.
func GetSecurityLevelForFile(path string) Level {
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	extLower := strings.ToLower(ext)
	baseLower := strings.ToLower(base)

	// Private/sensitive key files
	if extLower == ".key" || extLower == ".gpg" || extLower == ".pem" || extLower == ".priv" ||
		strings.Contains(baseLower, "private") || strings.Contains(baseLower, "secret") {
		return Private
	}

	// Public keys still need some protection
	if extLower == ".pub" || strings.Contains(baseLower, "public") {
		return Restricted
	}

	// Config files
	if extLower == ".conf" || extLower == ".config" || extLower == ".yaml" ||
		extLower == ".yml" || extLower == ".json" {
		return Safe
	}

	// Shell scripts should be Safe level at minimum
	if extLower == ".sh" || extLower == ".bash" {
		return Safe
	}

	return Basic
}

// IsPathWritable securely checks if a path is writable by the current user.
func IsPathWritable(path string) (bool, error) {
	// Check if directory exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If directory doesn't exist, check if parent is writable
			parent := filepath.Dir(path)

			return IsPathWritable(parent)
		}

		return false, fmt.Errorf("error checking path: %w", err)
	}

	// If it's a file, check if we can open it for writing
	if !fileInfo.IsDir() {
		file, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err == nil {
			file.Close()

			return true, nil
		}

		// We don't care about the specific error, just that it's not writable
		return false, nil
	}

	// It's a directory, try to create a temporary file with secure random name
	tmpFile, err := os.CreateTemp(path, ".tmp_write_test_*")
	if err != nil {
		// Directory is not writable
		return false, fmt.Errorf("directory not writable: %w", err)
	}

	// Clean up - get the name before closing
	tmpName := tmpFile.Name()
	tmpFile.Close()
	os.Remove(tmpName)

	return true, nil
}

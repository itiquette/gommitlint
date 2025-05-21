// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package security provides file system security operations
package security

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

// FileSecurityService handles file permission security checks.
type FileSecurityService struct{}

// IsSecurePermissions checks if a file has appropriately secure permissions.
func (s *FileSecurityService) IsSecurePermissions(path string, level Level) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("cannot access file: %w", err)
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

// Note: TestSecurityChecker is defined in testutils.go

// VerifyOwnership checks if current user owns the file.
func (s *FileSecurityService) VerifyOwnership(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
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

// GetSecurityLevelForFile determines the appropriate security level for a file
// based on its extension and name.
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

	return Basic
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
)

// SecurityValidator provides secure validation for all user inputs.
type SecurityValidator struct{}

// NewSecurityValidator creates a new security validator.
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{}
}

// ValidateRepoPath securely validates a repository path.
func (s *SecurityValidator) ValidateRepoPath(repoPath string) (string, error) {
	if repoPath == "" {
		repoPath = "."
	}

	// Use the existing robust git repo validation
	validatedPath, err := signing.ValidateGitRepoPath(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	return validatedPath, nil
}

// ValidateOutputFilePath securely validates a file path for output.
func (s *SecurityValidator) ValidateOutputFilePath(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("file path cannot be empty")
	}

	// Validate against common injection patterns
	if err := s.validatePathSecurity(filePath); err != nil {
		return "", fmt.Errorf("invalid output file path: %w", err)
	}

	// Clean and get absolute path
	cleanPath := filepath.Clean(filePath)

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve absolute path: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create output directory: %w", err)
	}

	return absPath, nil
}

// ValidateMessageFilePath securely validates a message file path for reading.
func (s *SecurityValidator) ValidateMessageFilePath(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("message file path cannot be empty")
	}

	// Validate against common injection patterns
	if err := s.validatePathSecurity(filePath); err != nil {
		return "", fmt.Errorf("invalid message file path: %w", err)
	}

	// Clean and get absolute path
	cleanPath := filepath.Clean(filePath)

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve absolute path: %w", err)
	}

	// Check if file exists and is readable
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("message file does not exist: %s", filePath)
		}

		return "", fmt.Errorf("cannot access message file: %w", err)
	}

	return absPath, nil
}

// ValidateGitReference securely validates git references (commit hashes, branch names, etc).
func (s *SecurityValidator) ValidateGitReference(ref string) error {
	if ref == "" {
		return errors.New("git reference cannot be empty")
	}

	// Check length (git refs are typically under 255 chars)
	if len(ref) > 255 {
		return errors.New("git reference too long")
	}

	// Check for null bytes and control characters
	if strings.ContainsRune(ref, 0) {
		return errors.New("git reference contains null byte")
	}

	for _, r := range ref {
		if r < 32 && r != 9 { // Allow tab but reject other control chars
			return errors.New("git reference contains control characters")
		}
	}

	// Validate git reference format (basic validation)
	// Git refs cannot contain certain characters: space, ~, ^, :, ?, *, [, \, ..
	invalidChars := regexp.MustCompile(`[\s~^:?*[\\\x00-\x1f\x7f]|\.\.`)
	if invalidChars.MatchString(ref) {
		return errors.New("git reference contains invalid characters")
	}

	// Additional safety checks for common injection patterns
	if strings.Contains(ref, "$(") || strings.Contains(ref, "`") || strings.Contains(ref, ";") {
		return errors.New("git reference contains potentially dangerous patterns")
	}

	return nil
}

// ValidateCommitRange securely validates git commit ranges.
func (s *SecurityValidator) ValidateCommitRange(commitRange string) error {
	if commitRange == "" {
		return errors.New("commit range cannot be empty")
	}

	// Split on .. or ... (git range syntax)
	var parts []string
	if strings.Contains(commitRange, "...") {
		parts = strings.Split(commitRange, "...")
	} else if strings.Contains(commitRange, "..") {
		parts = strings.Split(commitRange, "..")
	} else {
		return errors.New("invalid commit range format (use .. or ...)")
	}

	if len(parts) != 2 {
		return errors.New("commit range must have exactly two parts")
	}

	// Validate each part as a git reference
	for _, part := range parts {
		if err := s.ValidateGitReference(strings.TrimSpace(part)); err != nil {
			return fmt.Errorf("invalid commit reference '%s': %w", part, err)
		}
	}

	return nil
}

// validatePathSecurity performs common security validations on file paths.
func (s *SecurityValidator) validatePathSecurity(path string) error {
	// Check for null bytes
	if strings.ContainsRune(path, 0) {
		return errors.New("path contains null byte")
	}

	// Check for control characters
	for _, r := range path {
		if r < 32 && r != 9 { // Allow tab but reject other control chars
			return errors.New("path contains control characters")
		}
	}

	// Check path length
	if len(path) > 1000 {
		return errors.New("path too long")
	}

	// Check for obvious path traversal attempts
	if strings.Contains(path, "..") {
		return errors.New("path traversal detected")
	}

	// Check for URL-encoded traversal attempts
	if strings.Contains(path, "%2e%2e") || strings.Contains(path, "%2E%2E") {
		return errors.New("encoded path traversal detected")
	}

	return nil
}

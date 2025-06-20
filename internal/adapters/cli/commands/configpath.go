// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigPathValidationError represents a config path validation error.
type ConfigPathValidationError struct {
	Path   string
	Reason string
}

func (e ConfigPathValidationError) Error() string {
	return fmt.Sprintf("invalid config path '%s': %s", e.Path, e.Reason)
}

// validateConfigPath validates a config file path for security.
// Returns the cleaned absolute path if valid, or an error if invalid.
func validateConfigPath(configPath string) (string, error) {
	if configPath == "" {
		return "", ConfigPathValidationError{
			Path:   configPath,
			Reason: "empty path",
		}
	}

	// Check for null bytes (path injection)
	if strings.ContainsRune(configPath, 0) {
		return "", ConfigPathValidationError{
			Path:   configPath,
			Reason: "invalid character (null byte)",
		}
	}

	// Check for control characters
	for _, r := range configPath {
		if r < 32 && r != 9 { // Allow tab (9) but reject other control chars
			return "", ConfigPathValidationError{
				Path:   configPath,
				Reason: "invalid character (control character)",
			}
		}
	}

	// Check path length
	if len(configPath) > 1000 {
		return "", ConfigPathValidationError{
			Path:   configPath,
			Reason: "path too long",
		}
	}

	// Reject absolute paths to system directories
	if filepath.IsAbs(configPath) {
		return "", ConfigPathValidationError{
			Path:   configPath,
			Reason: "absolute path not allowed",
		}
	}

	// Clean the path and check for traversal
	cleanPath := filepath.Clean(configPath)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", ConfigPathValidationError{
			Path:   configPath,
			Reason: "path traversal detected",
		}
	}

	// Check if it's a symlink
	if info, err := os.Lstat(cleanPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", ConfigPathValidationError{
				Path:   configPath,
				Reason: "symlink not allowed",
			}
		}
	}

	// Get absolute path for further validation
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", ConfigPathValidationError{
			Path:   configPath,
			Reason: fmt.Sprintf("cannot resolve absolute path: %v", err),
		}
	}

	return absPath, nil
}

// validateConfigFilePermissions validates that a config file has secure permissions.
func validateConfigFilePermissions(configPath string) error {
	// Open file to get a file descriptor - prevents TOCTOU issues
	file, err := os.Open(configPath)
	if err != nil {
		return err // File doesn't exist or other access error
	}
	defer file.Close()

	// Get file info directly from the file descriptor
	info, err := file.Stat()
	if err != nil {
		return ConfigPathValidationError{
			Path:   configPath,
			Reason: fmt.Sprintf("cannot stat file: %v", err),
		}
	}

	mode := info.Mode()
	perm := mode.Perm()

	// Check if file is executable
	if perm&0111 != 0 {
		return ConfigPathValidationError{
			Path:   configPath,
			Reason: "config file should not be executable",
		}
	}

	// Check if file is world-writable
	if perm&0002 != 0 {
		return ConfigPathValidationError{
			Path:   configPath,
			Reason: "insecure permissions (world-writable)",
		}
	}

	// Check if file is group-writable
	if perm&0020 != 0 {
		return ConfigPathValidationError{
			Path:   configPath,
			Reason: "insecure permissions (group-writable)",
		}
	}

	return nil
}

// secureConfigPathValidation performs comprehensive security validation of a config path.
// This is a pure function that validates without side effects.
func secureConfigPathValidation(configPath string) error {
	// Step 1: Validate the path format and security
	validatedPath, err := validateConfigPath(configPath)
	if err != nil {
		return err
	}

	// Step 2: Check if file exists
	if _, err := os.Stat(validatedPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file '%s' not found", configPath)
		}

		return fmt.Errorf("cannot access config file '%s': %w", configPath, err)
	}

	// Step 3: Validate file permissions
	if err := validateConfigFilePermissions(validatedPath); err != nil {
		return err
	}

	return nil
}

// isConfigPathSecure checks if a config path is secure without side effects.
// Returns true if the path is safe to use, false otherwise.
func isConfigPathSecure(configPath string) bool {
	return secureConfigPathValidation(configPath) == nil
}

// sanitizeConfigPath returns a sanitized version of the config path.
// This is a pure function that doesn't modify the original path.
func sanitizeConfigPath(configPath string) string {
	if configPath == "" {
		return ""
	}

	// Remove null bytes
	sanitized := strings.ReplaceAll(configPath, "\x00", "")

	// Remove other control characters except tab
	var result strings.Builder

	for _, r := range sanitized {
		if r >= 32 || r == 9 { // Keep printable chars and tab
			result.WriteRune(r)
		}
	}

	// Clean the path
	cleaned := filepath.Clean(result.String())

	// Limit length
	if len(cleaned) > 1000 {
		cleaned = cleaned[:1000]
	}

	return cleaned
}

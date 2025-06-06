// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package testdata

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
)

// SimplePathValidator uses simplified validation for test environments.
// This is a test-only implementation that provides basic path validation
// without the full security checks used in production.
type SimplePathValidator struct{}

// ValidateHookPath performs simple path construction for basic Git directories.
func (SimplePathValidator) ValidateHookPath(repoPath, hookType string) (string, error) {
	// Validate hook type to prevent command injection
	if !signing.IsValidGitHookType(hookType) {
		return "", fmt.Errorf("invalid hook type: %s", hookType)
	}

	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	hooksDir := filepath.Join(gitDir, "hooks")

	return filepath.Join(hooksDir, hookType), nil
}

// EnsureHooksDirectory ensures the hooks directory exists with simple validation.
func (SimplePathValidator) EnsureHooksDirectory(repoPath string) error {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	hooksDir := filepath.Join(gitDir, "hooks")

	// Check if hooks directory exists
	fi, err := os.Stat(hooksDir)
	if err == nil {
		// Path exists - verify it's a directory
		if !fi.IsDir() {
			return fmt.Errorf("hooks path exists but is not a directory: %s", hooksDir)
		}
		// Directory exists, no need to create it
		return nil
	} else if !os.IsNotExist(err) {
		// Some error other than "not exists"
		return fmt.Errorf("could not check hooks directory: %w", err)
	}

	// Create hooks directory
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("could not create hooks directory: %w", err)
	}

	return nil
}


// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
)

// PathValidator validates and resolves Git hook paths.
type PathValidator interface {
	// ValidateHookPath validates and returns the hook file path.
	ValidateHookPath(repoPath, hookType string) (string, error)
	// EnsureHooksDirectory ensures the hooks directory exists.
	EnsureHooksDirectory(repoPath string) error
}

// StrictPathValidator uses full validation for production environments.
type StrictPathValidator struct{}

// ValidateHookPath performs strict validation of Git repository and hook paths.
func (StrictPathValidator) ValidateHookPath(repoPath, hookType string) (string, error) {
	// Validate hook type to prevent command injection
	if !signing.IsValidGitHookType(hookType) {
		return "", fmt.Errorf("invalid hook type: %s", hookType)
	}

	// Validate the repository path using fsutils
	validatedRepo, err := signing.ValidateGitRepoPath(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Use the fsutils to get the hook path securely
	hookPath, err := signing.GetGitHookPath(validatedRepo, hookType)
	if err != nil {
		return "", err
	}

	return hookPath, nil
}

// EnsureHooksDirectory ensures the hooks directory exists with strict validation.
func (StrictPathValidator) EnsureHooksDirectory(repoPath string) error {
	// Validate the repository path using fsutils
	validatedRepo, err := signing.ValidateGitRepoPath(repoPath)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	// Create a safe hooks directory path
	gitDirPath, err := signing.SafeJoin(validatedRepo, ".git")
	if err != nil {
		return fmt.Errorf("invalid git directory: %w", err)
	}

	hooksDir, err := signing.SafeJoin(gitDirPath, "hooks")
	if err != nil {
		return fmt.Errorf("invalid hooks directory: %w", err)
	}

	// Verify directory containment
	isWithin, err := signing.IsWithinDirectory(hooksDir, gitDirPath)
	if err != nil {
		return fmt.Errorf("path validation error: %w", err)
	}

	if !isWithin {
		return errors.New("invalid hooks directory: path traversal detected")
	}

	// Check if the hooks directory already exists
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

	// Create hooks directory with secure permissions
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("could not create hooks directory: %w", err)
	}

	return nil
}

// DefaultPathValidator returns the standard path validator for production use.
func DefaultPathValidator() PathValidator {
	return StrictPathValidator{}
}

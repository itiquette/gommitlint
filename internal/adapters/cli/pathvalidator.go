// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	if !crypto.IsValidGitHookType(hookType) {
		return "", fmt.Errorf("invalid hook type: %s", hookType)
	}

	// Validate the repository path using fsutils
	validatedRepo, err := crypto.ValidateGitRepoPath(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Use the fsutils to get the hook path securely
	hookPath, err := crypto.GetGitHookPath(validatedRepo, hookType)
	if err != nil {
		return "", err
	}

	return hookPath, nil
}

// EnsureHooksDirectory ensures the hooks directory exists with strict validation.
func (StrictPathValidator) EnsureHooksDirectory(repoPath string) error {
	// Validate the repository path using fsutils
	validatedRepo, err := crypto.ValidateGitRepoPath(repoPath)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	// Create a safe hooks directory path
	gitDirPath, err := crypto.SafeJoin(validatedRepo, ".git")
	if err != nil {
		return fmt.Errorf("invalid git directory: %w", err)
	}

	hooksDir, err := crypto.SafeJoin(gitDirPath, "hooks")
	if err != nil {
		return fmt.Errorf("invalid hooks directory: %w", err)
	}

	// Verify directory containment
	isWithin, err := crypto.IsWithinDirectory(hooksDir, gitDirPath)
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

// SimplePathValidator uses simplified validation for environments with basic Git setups.
type SimplePathValidator struct{}

// ValidateHookPath performs simple path construction for basic Git directories.
func (SimplePathValidator) ValidateHookPath(repoPath, hookType string) (string, error) {
	// Validate hook type to prevent command injection
	if !crypto.IsValidGitHookType(hookType) {
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

// DefaultPathValidator returns the standard path validator for production use.
func DefaultPathValidator() PathValidator {
	return StrictPathValidator{}
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// newInstallHookCmd creates a new command for installing Git hooks.
func newInstallHookCmd() *cobra.Command {
	var installHookCmd = &cobra.Command{
		Use:   "install-hook",
		Short: "Install Git commit-msg hook for validation",
		Long: `Installs a Git commit-msg hook to automatically validate commit messages.

The --force flag allows overwriting existing hooks, which is useful for updating hooks
or replacing custom hooks with standardized ones.

Examples:
  # Install commit-msg hook in the current repository
  gommitlint install-hook

  # Install commit-msg hook with specific options
  gommitlint install-hook --force
  
  # Install hook in a different repository
  gommitlint install-hook --repo-path=/path/to/my/repo`,
		Run: func(cmd *cobra.Command, _ []string) {
			// Get flags
			force, _ := cmd.Flags().GetBool("force")
			repoPath, _ := cmd.Flags().GetString("repo-path")

			// If repo path not provided, use current directory
			if repoPath == "" {
				repoPath = "."
			}

			// Install the hook
			if err := installHook(force, repoPath); err != nil {
				cmd.PrintErrf("Error: %v\n", err)
				os.Exit(1)
			}

			cmd.Println("✅ Git commit-msg hook installed successfully!")
		},
	}

	// Add flags
	installHookCmd.Flags().BoolP("force", "f", false, "overwrite existing hook if it exists")
	installHookCmd.Flags().String("repo-path", "", "path to the Git repository (defaults to current directory)")

	return installHookCmd
}

// HookInstallationParameters contains all parameters needed for hook installation.
// This structure follows the functional patterns of immutability and value semantics.
type HookInstallationParameters struct {
	Force    bool
	RepoPath string
	HookType string
}

// NewHookInstallationParameters creates HookInstallationParameters with defaults.
func NewHookInstallationParameters(force bool, repoPath string) HookInstallationParameters {
	// Default to commit-msg hook type (currently the only supported type)
	return HookInstallationParameters{
		Force:    force,
		RepoPath: repoPath,
		HookType: "commit-msg",
	}
}

// WithHookType returns a new HookInstallationParameters with the hook type updated.
func (p HookInstallationParameters) WithHookType(hookType string) HookInstallationParameters {
	result := p
	result.HookType = hookType

	return result
}

// WithForce returns a new HookInstallationParameters with the force flag updated.
func (p HookInstallationParameters) WithForce(force bool) HookInstallationParameters {
	result := p
	result.Force = force

	return result
}

// WithRepoPath returns a new HookInstallationParameters with the repo path updated.
func (p HookInstallationParameters) WithRepoPath(repoPath string) HookInstallationParameters {
	result := p
	result.RepoPath = repoPath

	return result
}

// FindHookPath determines the hook file path based on the parameters.
// Implements security best practices for safe path handling.
func (p HookInstallationParameters) FindHookPath() (string, error) {
	// Clean the repository path to prevent path traversal
	repoPath := filepath.Clean(p.RepoPath)

	// Find Git root directory
	gitDir, err := findGitDirectory(repoPath)
	if err != nil {
		return "", fmt.Errorf("could not find Git directory: %w", err)
	}

	// Validate hook type to prevent command injection
	hookType := p.HookType
	if !isValidHookType(hookType) {
		return "", fmt.Errorf("invalid hook type: %s", hookType)
	}

	// Hooks directory path
	hooksDir := filepath.Clean(filepath.Join(gitDir, "hooks"))

	// Path to the hook file
	hookPath := filepath.Clean(filepath.Join(hooksDir, hookType))

	// Verify the hook path is within the git directory to prevent path traversal
	if !strings.HasPrefix(hookPath, gitDir) {
		return "", errors.New("invalid hook path: path traversal detected")
	}

	return hookPath, nil
}

// isValidHookType checks if a hook type is valid to prevent command injection.
func isValidHookType(hookType string) bool {
	validHooks := map[string]bool{
		"commit-msg":         true,
		"pre-commit":         true,
		"pre-push":           true,
		"pre-receive":        true,
		"update":             true,
		"post-update":        true,
		"post-receive":       true,
		"post-checkout":      true,
		"post-merge":         true,
		"pre-applypatch":     true,
		"post-applypatch":    true,
		"pre-rebase":         true,
		"prepare-commit-msg": true,
	}

	return validHooks[hookType]
}

// EnsureHooksDirectory ensures the hooks directory exists.
// Implements security best practices for safe path handling.
func (p HookInstallationParameters) EnsureHooksDirectory() error {
	// Clean the repository path to prevent path traversal
	repoPath := filepath.Clean(p.RepoPath)

	// Find Git root directory
	gitDir, err := findGitDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("could not find Git directory: %w", err)
	}

	// Create a safe hooks directory path
	hooksDir := filepath.Clean(filepath.Join(gitDir, "hooks"))

	// Verify the hooks directory is within the git directory
	if !strings.HasPrefix(hooksDir, gitDir) {
		return errors.New("invalid hooks directory: path traversal detected")
	}

	// Ensure the directory doesn't already exist or is a directory
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

// GetHookContent returns the content for the hook based on its type.
func (p HookInstallationParameters) GetHookContent() string {
	// Currently, only commit-msg hooks are supported
	return generateCommitMsgHook()
}

// CanInstallHook checks if a hook can be installed based on parameters.
func (p HookInstallationParameters) CanInstallHook() error {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return err
	}

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil && !p.Force {
		return fmt.Errorf("hook already exists at %s (use --force to overwrite)", hookPath)
	}

	return nil
}

// installHook installs a Git commit-msg hook in the specified repository.
func installHook(force bool, repoPath string) error {
	// Validate and normalize the repository path
	validatedPath, err := validateRepositoryPath(repoPath)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	// Create parameters with defaults
	params := NewHookInstallationParameters(force, validatedPath)

	// Ensure hooks directory exists
	if err := params.EnsureHooksDirectory(); err != nil {
		return err
	}

	// Check if we can install the hook
	if err := params.CanInstallHook(); err != nil {
		return err
	}

	// Get the hook path
	hookPath, err := params.FindHookPath()
	if err != nil {
		return err
	}

	// Get the hook content
	hookContent := params.GetHookContent()

	// Write the hook file with safe permissions (read/write for owner only)
	if err := os.WriteFile(hookPath, []byte(hookContent), 0600); err != nil {
		return fmt.Errorf("could not write hook file: %w", err)
	}

	// Make the hook executable
	if err := os.Chmod(hookPath, 0700); err != nil {
		return fmt.Errorf("could not make hook file executable: %w", err)
	}

	return nil
}

// findGitDirectory finds the .git directory for a repository.
// Implements security best practices for path handling using go-git library.
func findGitDirectory(repoPath string) (string, error) {
	// Use our enhanced validation function for repository path validation
	validatedPath, err := validateRepositoryPath(repoPath)
	if err != nil {
		return "", err
	}

	// Set a reasonable limit to prevent excessive traversal
	const maxLevels = 20

	level := 0
	current := validatedPath

	// Find the git directory by traversing up the directory tree
	for level < maxLevels {
		// Check if the current directory is a git repository
		repo, err := git.PlainOpen(current)
		if err == nil {
			// We found the repository
			// Get the worktree to find the root directory (we don't actually use it,
			// but this verifies it's a valid repository)
			_, err := repo.Worktree()
			if err != nil {
				return "", fmt.Errorf("found git repository but could not access worktree: %w", err)
			}

			// Return the .git directory path
			gitDir := filepath.Join(current, ".git")

			// Verify the directory exists and is a valid git directory
			gitDirInfo, err := os.Lstat(gitDir) // Use Lstat to avoid following symlinks
			if err != nil {
				return "", fmt.Errorf("invalid git directory: %w", err)
			}

			// Check if it's a symlink and reject if it is - prevents symlink attacks
			if gitDirInfo.Mode()&os.ModeSymlink != 0 {
				return "", errors.New("git directory is a symlink which is not supported for security reasons")
			}

			if !gitDirInfo.IsDir() {
				// Git submodules can have a file instead of a directory
				// Read it to ensure it's a valid git reference
				content, err := os.ReadFile(gitDir)
				if err != nil {
					return "", fmt.Errorf("could not read git directory file: %w", err)
				}

				if !strings.HasPrefix(string(content), "gitdir:") {
					return "", errors.New("invalid git directory structure")
				}

				// Extract the actual gitdir path from the file for submodules
				parts := strings.SplitN(string(content), ":", 2)
				if len(parts) != 2 {
					return "", errors.New("invalid git directory reference format")
				}

				submoduleGitDir := strings.TrimSpace(parts[1])

				// Convert to absolute path if needed
				if !filepath.IsAbs(submoduleGitDir) {
					submoduleGitDir = filepath.Clean(filepath.Join(filepath.Dir(gitDir), submoduleGitDir))
				}

				// Verify submodule path exists and is a directory
				subInfo, err := os.Lstat(submoduleGitDir) // Use Lstat to avoid following symlinks
				if err != nil {
					return "", fmt.Errorf("invalid submodule git directory: %w", err)
				}

				if !subInfo.IsDir() {
					return "", errors.New("submodule gitdir path is not a directory")
				}

				// Return the actual gitdir path for submodules
				return submoduleGitDir, nil
			}

			return gitDir, nil
		}

		// Go up one level
		parent := filepath.Dir(current)
		if parent == current {
			// Reached the root directory, .git not found
			return "", fmt.Errorf("git repository not found in %s or any parent directory", repoPath)
		}

		current = parent
		level++
	}

	return "", fmt.Errorf("exceeded maximum directory traversal levels (%d) without finding git repository", maxLevels)
}

// generateCommitMsgHook generates content for the commit-msg hook.
func generateCommitMsgHook() string {
	return CreateDefaultHookScript()
}

// CreateDefaultHookScript creates a shell script for the commit-msg hook.
func CreateDefaultHookScript() string {
	return `#!/bin/sh
#
# gommitlint commit-msg hook for validating commit messages.
# Generated by gommitlint install-hook command.
#

# Get the commit message file path
COMMIT_MSG_FILE=$1

# Run validation using gommitlint
gommitlint validate --message-file="$COMMIT_MSG_FILE"
RESULT=$?

# Return the validation result
exit $RESULT
`
}

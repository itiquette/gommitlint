// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
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
// This structure supports safe operations.
type HookInstallationParameters struct {
	Force         bool
	RepoPath      string
	HookType      string
	PathValidator PathValidator
}

// NewHookInstallationParameters creates HookInstallationParameters with defaults.
func NewHookInstallationParameters(force bool, repoPath string) HookInstallationParameters {
	// Default to commit-msg hook type (currently the only supported type)
	return HookInstallationParameters{
		Force:         force,
		RepoPath:      repoPath,
		HookType:      "commit-msg",
		PathValidator: DefaultPathValidator(),
	}
}

// FindHookPath determines the hook file path based on the parameters.
// Implements security best practices for safe path handling.
func (p HookInstallationParameters) FindHookPath() (string, error) {
	if p.PathValidator == nil {
		p.PathValidator = DefaultPathValidator()
	}

	return p.PathValidator.ValidateHookPath(p.RepoPath, p.HookType)
}

// EnsureHooksDirectory ensures the hooks directory exists.
// Implements security best practices for safe path handling.
func (p HookInstallationParameters) EnsureHooksDirectory() error {
	if p.PathValidator == nil {
		p.PathValidator = DefaultPathValidator()
	}

	return p.PathValidator.EnsureHooksDirectory(p.RepoPath)
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

	// Check if hook already exists using file descriptor to prevent TOCTOU
	file, err := os.Open(hookPath)
	if err == nil {
		// File exists
		file.Close()

		if !p.Force {
			return fmt.Errorf("hook already exists at %s (use --force to overwrite)", hookPath)
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("cannot check hook existence: %w", err)
	}

	return nil
}

// WithForce returns a new HookInstallationParameters with Force enabled.
func (p HookInstallationParameters) WithForce() HookInstallationParameters {
	p.Force = true

	return p
}

// WithoutForce returns a new HookInstallationParameters with Force disabled.
func (p HookInstallationParameters) WithoutForce() HookInstallationParameters {
	p.Force = false

	return p
}

// Hook returns a new HookInstallationParameters with HookType set.
func (p HookInstallationParameters) Hook(hookType string) HookInstallationParameters {
	p.HookType = hookType

	return p
}

// Repo returns a new HookInstallationParameters with RepoPath set.
func (p HookInstallationParameters) Repo(repoPath string) HookInstallationParameters {
	p.RepoPath = repoPath

	return p
}

// installHook installs a Git commit-msg hook in the specified repository.
// It uses atomic file operations and proper permission management to ensure security.
func installHook(force bool, repoPath string) error {
	// Validate and normalize the repository path using fsutils
	validatedPath, err := signing.ValidateGitRepoPath(repoPath)
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

	// Write the hook file using our secure file writing function
	// This handles creating a temporary file, setting permissions, and atomically
	// renaming it to the final destination to prevent TOCTOU vulnerabilities
	if err := signing.SafeWriteFile(hookPath, []byte(hookContent), 0700); err != nil {
		return fmt.Errorf("could not write hook file: %w", err)
	}

	return nil
}

// generateCommitMsgHook generates content for the commit-msg hook.
func generateCommitMsgHook() string {
	return createDefaultHookScript()
}

// createDefaultHookScript creates a shell script for the commit-msg hook.
// It is used internally by generateCommitMsgHook.
func createDefaultHookScript() string {
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

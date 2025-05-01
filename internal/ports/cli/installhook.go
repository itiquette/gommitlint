// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
func (p HookInstallationParameters) FindHookPath() (string, error) {
	// Find Git root directory
	gitDir, err := findGitDirectory(p.RepoPath)
	if err != nil {
		return "", fmt.Errorf("could not find Git directory: %w", err)
	}

	// Hooks directory path
	hooksDir := filepath.Join(gitDir, "hooks")

	// Path to the hook file
	hookPath := filepath.Join(hooksDir, p.HookType)

	return hookPath, nil
}

// EnsureHooksDirectory ensures the hooks directory exists.
func (p HookInstallationParameters) EnsureHooksDirectory() error {
	// Find Git root directory
	gitDir, err := findGitDirectory(p.RepoPath)
	if err != nil {
		return fmt.Errorf("could not find Git directory: %w", err)
	}

	// Create hooks directory if it doesn't exist
	hooksDir := filepath.Join(gitDir, "hooks")
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
// This implementation follows functional programming patterns with immutable parameters.
func installHook(force bool, repoPath string) error {
	// Create parameters with defaults
	params := NewHookInstallationParameters(force, repoPath)

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

	// Write the hook file
	if err := os.WriteFile(hookPath, []byte(hookContent), 0600); err != nil {
		return fmt.Errorf("could not write hook file: %w", err)
	}

	return nil
}

// findGitDirectory finds the .git directory for a repository.
func findGitDirectory(repoPath string) (string, error) {
	// Try to find .git directory
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--git-dir")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not find Git repository: %w", err)
	}

	gitDir := strings.TrimSpace(string(output))

	// If it's a relative path, join with repo path
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoPath, gitDir)
	}

	return gitDir, nil
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

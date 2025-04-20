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

// installHook installs a Git commit-msg hook in the specified repository.
func installHook(force bool, repoPath string) error {
	// Set hook type to commit-msg (only supported type)
	hookType := "commit-msg"

	// Find Git root directory
	gitDir, err := findGitDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("could not find Git directory: %w", err)
	}

	// Create hooks directory if it doesn't exist
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("could not create hooks directory: %w", err)
	}

	// Path to the hook file
	hookPath := filepath.Join(hooksDir, hookType)

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil && !force {
		return fmt.Errorf("hook already exists at %s (use --force to overwrite)", hookPath)
	}

	// Get the hook content
	hookContent := generateCommitMsgHook()

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
	return `#!/bin/sh
# gommitlint commit-msg hook
# Validates commit messages against defined rules

# Get the commit message file path
commit_msg_file="$1"

# Run gommitlint validation
gommitlint validate --message-file="$commit_msg_file"

# Get the exit code
result=$?

# If validation failed, abort the commit
if [ $result -ne 0 ]; then
    echo "❌ Commit message validation failed. Commit aborted."
    exit 1
fi

# Validation passed
echo "✅ Commit message validation passed."
exit 0
`
}

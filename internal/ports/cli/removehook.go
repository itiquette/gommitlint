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

	"github.com/spf13/cobra"
)

// newRemoveHookCmd creates a new command for removing Git hooks.
func newRemoveHookCmd() *cobra.Command {
	var removeHookCmd = &cobra.Command{
		Use:   "remove-hook",
		Short: "Remove Git hooks for commit validation",
		Long: `Removes Git hooks previously installed by gommitlint.

The command will check if the hook was actually installed by gommitlint before removing it.
If the hook was not installed by gommitlint, you'll be prompted for confirmation
unless the --yes flag is provided.

Examples:
  # Remove commit-msg hook from the current repository
  gommitlint remove-hook
  
  # Remove hook from a specific repository
  gommitlint remove-hook --repo-path=/path/to/repo
  
  # Remove hook without confirmation prompt
  gommitlint remove-hook --yes`,
		Run: func(cmd *cobra.Command, _ []string) {
			// Get flags
			repoPath, _ := cmd.Flags().GetString("repo-path")
			skipConfirm, _ := cmd.Flags().GetBool("yes")

			// If repo path not provided, use current directory
			if repoPath == "" {
				repoPath = "."
			}

			// Remove the hook
			if err := removeHook(cmd, repoPath, skipConfirm); err != nil {
				cmd.PrintErrf("Error: %v\n", err)
				os.Exit(1)
			}
			cmd.Println("✅ Git hook removed successfully!")
		},
	}

	// Add flags
	removeHookCmd.Flags().String("repo-path", "", "path to the Git repository (defaults to current directory)")
	removeHookCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")

	return removeHookCmd
}

// removeHook removes a Git hook from the specified repository.
func removeHook(cmd *cobra.Command, repoPath string, skipConfirm bool) error {
	// Set hook type to commit-msg (only supported type)
	hookType := "commit-msg"

	// Find Git root directory
	gitDir, err := findGitDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("could not find Git directory: %w", err)
	}

	// Path to the hook file
	hookPath := filepath.Join(gitDir, "hooks", hookType)

	// Check if hook exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return fmt.Errorf("hook does not exist at %s", hookPath)
	}

	// Verify the hook is actually one of ours
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("could not read hook file: %w", err)
	}

	// Check if the hook was installed by gommitlint
	if !strings.Contains(string(content), "gommitlint") {
		if !skipConfirm {
			cmd.Println("⚠️ Warning: The hook doesn't appear to be installed by gommitlint.")
			cmd.Print("Do you still want to remove it? [y/N]: ")

			var response string

			// Check if response is affirmative
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				return errors.New("operation cancelled by user")
			}
		}
	}

	// Remove the hook file
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("could not remove hook file: %w", err)
	}

	return nil
}

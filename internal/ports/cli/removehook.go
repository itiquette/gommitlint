// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"io"
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

// HookRemovalParameters contains all parameters needed for hook removal.
// This structure follows functional programming principles of immutability.
type HookRemovalParameters struct {
	RepoPath    string
	SkipConfirm bool
	HookType    string
	Output      io.Writer
	Input       io.Reader
}

// NewHookRemovalParameters creates HookRemovalParameters with defaults.
func NewHookRemovalParameters(cmd *cobra.Command, repoPath string, skipConfirm bool) HookRemovalParameters {
	return HookRemovalParameters{
		RepoPath:    repoPath,
		SkipConfirm: skipConfirm,
		HookType:    "commit-msg", // Currently the only supported type
		Output:      cmd.OutOrStdout(),
		Input:       cmd.InOrStdin(),
	}
}

// WithHookType returns a new HookRemovalParameters with the hook type updated.
func (p HookRemovalParameters) WithHookType(hookType string) HookRemovalParameters {
	result := p
	result.HookType = hookType

	return result
}

// WithSkipConfirm returns a new HookRemovalParameters with the skip confirm flag updated.
func (p HookRemovalParameters) WithSkipConfirm(skipConfirm bool) HookRemovalParameters {
	result := p
	result.SkipConfirm = skipConfirm

	return result
}

// FindHookPath determines the hook file path based on the parameters.
func (p HookRemovalParameters) FindHookPath() (string, error) {
	// Find Git root directory
	gitDir, err := findGitDirectory(p.RepoPath)
	if err != nil {
		return "", fmt.Errorf("could not find Git directory: %w", err)
	}

	// Path to the hook file
	hookPath := filepath.Join(gitDir, "hooks", p.HookType)

	return hookPath, nil
}

// VerifyHookExists checks if the hook file exists.
func (p HookRemovalParameters) VerifyHookExists() error {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return err
	}

	// Check if hook exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return fmt.Errorf("hook does not exist at %s", hookPath)
	}

	return nil
}

// IsGommitlintHook checks if the hook was installed by gommitlint.
func (p HookRemovalParameters) IsGommitlintHook() (bool, error) {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return false, err
	}

	// Verify the hook is actually one of ours
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return false, fmt.Errorf("could not read hook file: %w", err)
	}

	return strings.Contains(string(content), "gommitlint"), nil
}

// ConfirmRemoval asks the user to confirm removing a non-gommitlint hook.
func (p HookRemovalParameters) ConfirmRemoval() (bool, error) {
	if p.SkipConfirm {
		return true, nil
	}

	fmt.Fprintln(p.Output, "⚠️ Warning: The hook doesn't appear to be installed by gommitlint.")
	fmt.Fprint(p.Output, "Do you still want to remove it? [y/N]: ")

	// Read user input
	var response string

	_, err := fmt.Fscanln(p.Input, &response)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	// Check if response is affirmative
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes", nil
}

// RemoveHookFile removes the hook file from the filesystem.
func (p HookRemovalParameters) RemoveHookFile() error {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return err
	}

	// Remove the hook file
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("could not remove hook file: %w", err)
	}

	return nil
}

// removeHook removes a Git hook from the specified repository.
// This implementation follows functional programming patterns with immutable parameters.
func removeHook(cmd *cobra.Command, repoPath string, skipConfirm bool) error {
	// Create parameters with defaults
	params := NewHookRemovalParameters(cmd, repoPath, skipConfirm)

	// Verify the hook exists
	if err := params.VerifyHookExists(); err != nil {
		return err
	}

	// Check if the hook was installed by gommitlint
	isGommitlintHook, err := params.IsGommitlintHook()
	if err != nil {
		return err
	}

	// If not a gommitlint hook, ask for confirmation
	if !isGommitlintHook {
		confirmed, err := params.ConfirmRemoval()
		if err != nil {
			return err
		}

		if !confirmed {
			return errors.New("operation cancelled by user")
		}
	}

	// Remove the hook file
	return params.RemoveHookFile()
}

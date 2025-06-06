// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
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
type HookRemovalParameters struct {
	RepoPath      string
	SkipConfirm   bool
	HookType      string
	Output        io.Writer
	Input         io.Reader
	PathValidator PathValidator
}

// NewHookRemovalParameters creates HookRemovalParameters with defaults.
func NewHookRemovalParameters(cmd *cobra.Command, repoPath string, skipConfirm bool) HookRemovalParameters {
	return HookRemovalParameters{
		RepoPath:      repoPath,
		SkipConfirm:   skipConfirm,
		HookType:      "commit-msg", // Currently the only supported type
		Output:        cmd.OutOrStdout(),
		Input:         cmd.InOrStdin(),
		PathValidator: DefaultPathValidator(),
	}
}

// FindHookPath determines the hook file path based on the parameters.
// Applies security best practices for safe path handling using signing.
func (p HookRemovalParameters) FindHookPath() (string, error) {
	if p.PathValidator == nil {
		p.PathValidator = DefaultPathValidator()
	}

	return p.PathValidator.ValidateHookPath(p.RepoPath, p.HookType)
}

// isValidHookType is now defined in installhook.go to avoid duplication
// It validates hook types to prevent command injection attacks

// VerifyHookExists checks if the hook file exists.
func (p HookRemovalParameters) VerifyHookExists() error {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return err
	}

	// Check if hook exists using file descriptor to prevent TOCTOU
	file, err := os.Open(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("hook does not exist at %s", hookPath)
		}

		return fmt.Errorf("cannot check hook existence: %w", err)
	}

	file.Close()

	return nil
}

// IsGommitlintHook checks if the hook was installed by gommitlint.
func (p HookRemovalParameters) IsGommitlintHook() (bool, error) {
	hookPath, err := p.FindHookPath()
	if err != nil {
		return false, err
	}

	// Open file to check size first
	file, err := os.Open(hookPath)
	if err != nil {
		return false, fmt.Errorf("could not open hook file: %w", err)
	}
	defer file.Close()

	// Check file size to prevent reading huge files
	info, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("could not stat hook file: %w", err)
	}

	// Limit to 10KB - hooks should be small scripts
	const maxHookSize = 10 * 1024
	if info.Size() > maxHookSize {
		return false, fmt.Errorf("hook file too large: %d bytes (max %d)", info.Size(), maxHookSize)
	}

	// Now safe to read the file
	content := make([]byte, info.Size())

	_, err = io.ReadFull(file, content)
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

// GetHookPath returns the hook file path to be removed.
func (p HookRemovalParameters) GetHookPath() (string, error) {
	return p.FindHookPath()
}

// RemoveHookFile removes the hook file from the filesystem.
func RemoveHookFile(hookPath string) error {
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("could not remove hook file: %w", err)
	}

	return nil
}

// WithSkipConfirm returns a new HookRemovalParameters with SkipConfirm enabled.
func (p HookRemovalParameters) WithSkipConfirm() HookRemovalParameters {
	p.SkipConfirm = true

	return p
}

// WithoutSkipConfirm returns a new HookRemovalParameters with SkipConfirm disabled.
func (p HookRemovalParameters) WithoutSkipConfirm() HookRemovalParameters {
	p.SkipConfirm = false

	return p
}

// Hook returns a new HookRemovalParameters with HookType set.
func (p HookRemovalParameters) Hook(hookType string) HookRemovalParameters {
	p.HookType = hookType

	return p
}

// removeHook removes a Git hook from the specified repository.
func removeHook(cmd *cobra.Command, repoPath string, skipConfirm bool) error {
	// Validate and normalize the repository path using fsutils
	validatedPath, err := signing.ValidateGitRepoPath(repoPath)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	// Create parameters with defaults
	params := NewHookRemovalParameters(cmd, validatedPath, skipConfirm)

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

	// Get hook path (pure function)
	hookPath, err := params.GetHookPath()
	if err != nil {
		return err
	}

	// Remove the hook file (side effect isolated)
	return RemoveHookFile(hookPath)
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
	"github.com/itiquette/gommitlint/internal/adapters/signing"
	"github.com/urfave/cli/v3"
)

// NewRemoveHookCommand creates the remove-hook subcommand.
func NewRemoveHookCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove-hook",
		Usage: "Remove Git hooks for commit validation",
		Description: `Removes Git hooks previously installed by gommitlint.

Examples:
  # Remove commit-msg hook from the current repository
  gommitlint remove-hook
  
  # Remove hook without confirmation prompt
  gommitlint remove-hook --yes`,

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "skip confirmation prompt",
			},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			return ExecuteRemoveHook(ctx, cmd)
		},
	}
}

// ExecuteRemoveHook orchestrates the hook removal process.
func ExecuteRemoveHook(ctx context.Context, cmd *cli.Command) error {
	// Get flags
	skipConfirm := cmd.Bool("yes")
	repoPath := getRepoPath(cmd)

	// Create logger from context
	zerologLogger := logadapter.GetLogger(ctx)
	logger := logadapter.NewDomainLogger(zerologLogger)

	// Remove the hook
	if err := removeHook(cmd, repoPath, skipConfirm); err != nil {
		logger.Error("Hook removal failed", "error", err)

		return err
	}

	fmt.Fprintln(cmd.Writer, "✅ Git hook removed successfully!")

	return nil
}

// removeHook removes a Git hook from the specified repository.
func removeHook(cmd *cli.Command, repoPath string, skipConfirm bool) error {
	// Validate and normalize the repository path using signing utilities
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
	hookPath, err := FindHookPath(params.RepoPath, params.HookType, params.PathValidator)
	if err != nil {
		return err
	}

	// Remove the hook file (side effect isolated)
	return RemoveHookFile(hookPath)
}

// HookRemovalParameters contains all parameters needed for hook removal.
type HookRemovalParameters struct {
	RepoPath      string
	SkipConfirm   bool
	HookType      string
	Output        io.Writer
	Input         io.Reader
	PathValidator cliAdapter.PathValidator
}

// NewHookRemovalParameters creates HookRemovalParameters with defaults.
func NewHookRemovalParameters(cmd *cli.Command, repoPath string, skipConfirm bool) HookRemovalParameters {
	return HookRemovalParameters{
		RepoPath:      repoPath,
		SkipConfirm:   skipConfirm,
		HookType:      "commit-msg", // Currently the only supported type
		Output:        cmd.Writer,
		Input:         cmd.Reader,
		PathValidator: cliAdapter.DefaultPathValidator(),
	}
}

// VerifyHookExists checks if the hook file exists.
func (p HookRemovalParameters) VerifyHookExists() error {
	hookPath, err := FindHookPath(p.RepoPath, p.HookType, p.PathValidator)
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
	hookPath, err := FindHookPath(p.RepoPath, p.HookType, p.PathValidator)
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

// RemoveHookFile removes the hook file from the filesystem.
func RemoveHookFile(hookPath string) error {
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("could not remove hook file: %w", err)
	}

	return nil
}

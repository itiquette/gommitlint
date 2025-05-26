// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newValidateCmd creates a new validate command.
func newValidateCmd(containerProvider func(context.Context) DependencyContainer) *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validates commit messages",
		Long: `Validates commit message/s against a set of rules.

Examples:
  # Validate commits in the current branch against main
  gommitlint validate --base-branch=main
  
  # Validate a specific commit
  gommitlint validate --git-reference=HEAD
  
  # Validate a commit message from a file
  gommitlint validate --message-file=/path/to/commit-msg.txt
  
  # Validate a range of commits
  gommitlint validate --revision-range=main..HEAD`,
		Run: func(cmd *cobra.Command, _ []string) {
			// Get context from command (set by root command)
			ctx := cmd.Context()

			// Process validation request
			exitCode, err := runNewValidation(ctx, cmd, containerProvider(ctx))
			if err != nil {
				cmd.PrintErrf("Error: %v\n", err)
				os.Exit(1)
			}

			// Special exit codes:
			// 0 - success
			// 1 - system error
			// 2 - validation failure
			if exitCode > 0 {
				os.Exit(exitCode)
			}
		},
	}

	// Add flags to the command
	validateCmd.Flags().String("message-file", "", "commit message file path to validate")
	validateCmd.Flags().String("git-reference", "", "git reference to validate (defaults to HEAD)")
	validateCmd.Flags().Int("commit-count", 1, "number of commits from HEAD to validate")
	validateCmd.Flags().String("revision-range", "", "range of commits to validate (<commit1>..<commit2>)")
	validateCmd.Flags().String("base-branch", "", "base branch to compare with (sets revision-range to <base-branch>..HEAD)")
	validateCmd.Flags().BoolP("verbose", "v", false, "show detailed validation results")
	validateCmd.Flags().Bool("extra-verbose", false, "show extra detailed validation results")
	validateCmd.Flags().Bool("light-mode", false, "use light background color scheme")
	validateCmd.Flags().String("rulehelp", "", "show detailed help for a specific rule")
	validateCmd.Flags().String("format", "text", "output format: text, json, github, or gitlab")
	validateCmd.Flags().Bool("skip-merge-commits", true, "skip merge commits in validation")
	validateCmd.Flags().String("repo-path", "", "path to the repository to validate")

	return validateCmd
}

// runNewValidation handles the core validation logic and returns an exit code.
func runNewValidation(ctx context.Context, cmd *cobra.Command, container DependencyContainer) (int, error) {
	// Create parameters object to encapsulate all inputs
	params := NewValidateParams(cmd)

	// Validate container
	if container == nil {
		return 1, errors.New("dependency container is nil")
	}

	// Get repository path
	repoPath := params.GetRepoPath()
	if repoPath == "" {
		// Use current directory if not specified
		var err error

		repoPath, err = os.Getwd()
		if err != nil {
			return 1, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Create formatter based on parameters
	formatter := params.CreateFormatter()

	// Create validation orchestrator using dependency container
	orchestrator, err := container.CreateValidationOrchestrator(ctx, repoPath, formatter)
	if err != nil {
		return 1, fmt.Errorf("failed to create validation orchestrator: %w", err)
	}

	// Get report options from parameters
	reportOptions := params.ToReportOptions()

	// Determine what to validate
	targetType, target1, target2, err := params.GetValidationTarget()
	if err != nil {
		return 1, err
	}

	// Perform validation based on target type
	switch targetType {
	case "message":
		// Read message from file
		message, err := os.ReadFile(target1)
		if err != nil {
			return 1, fmt.Errorf("failed to read message file: %w", err)
		}

		return orchestrator.ValidateMessageAndReport(ctx, string(message), reportOptions)

	case "range":
		// Validate commit range
		return orchestrator.ValidateRangeAndReport(ctx, target1, target2, params.SkipMergeCommits, reportOptions)

	case "commit":
		// Validate single commit
		return orchestrator.ValidateAndReport(ctx, target1, params.SkipMergeCommits, reportOptions)

	case "count":
		// For commit count, we need to convert to a range
		// This is a limitation of the current interface - we'd need to extend the orchestrator
		// to support commit count directly. For now, just validate HEAD.
		container.GetLogger().Warn("Commit count validation not fully implemented, validating HEAD instead")

		return orchestrator.ValidateAndReport(ctx, "HEAD", params.SkipMergeCommits, reportOptions)

	default:
		return 1, fmt.Errorf("unknown validation target type: %s", targetType)
	}
}

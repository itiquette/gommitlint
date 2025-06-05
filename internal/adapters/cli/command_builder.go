// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"os"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/spf13/cobra"
)

// BuildValidationCommand creates a command from CLI flags using pure functions.
// This follows functional programming principles with no side effects.
func BuildValidationCommand(cmd *cobra.Command, cfg config.Config) (ValidationCommand, error) {
	// Extract validation target flags
	messageFile, _ := cmd.Flags().GetString("message-file")
	gitReference, _ := cmd.Flags().GetString("git-reference")
	commitCount, _ := cmd.Flags().GetInt("commit-count")
	revisionRange, _ := cmd.Flags().GetString("revision-range")
	baseBranch, _ := cmd.Flags().GetString("base-branch")

	// Extract output flags
	format, _ := cmd.Flags().GetString("format")
	verbose, _ := cmd.Flags().GetBool("verbose")
	extraVerbose, _ := cmd.Flags().GetBool("extra-verbose")
	ruleHelp, _ := cmd.Flags().GetString("rulehelp")
	lightMode, _ := cmd.Flags().GetBool("light-mode")

	// Extract other flags
	repoPath, _ := cmd.Flags().GetString("repo-path")
	skipMergeCommits, _ := cmd.Flags().GetBool("skip-merge-commits")

	// Build validation target using focused builder (immutable)
	target, err := NewValidationTargetBuilder().
		WithMessageFile(messageFile).
		WithGitReference(gitReference).
		WithCommitCount(commitCount).
		WithRevisionRange(revisionRange).
		WithBaseBranch(baseBranch).
		Build()
	if err != nil {
		return ValidationCommand{}, err
	}

	// Build output options using focused value type (immutable)
	output := NewOutputOptions(cmd.OutOrStdout()).
		WithFormat(format).
		WithVerbose(verbose).
		WithExtraVerbose(extraVerbose).
		WithRuleHelp(ruleHelp).
		WithLightMode(lightMode)

	// Set default repo path if empty
	if repoPath == "" {
		var err error

		repoPath, err = os.Getwd()
		if err != nil {
			repoPath = "." // Fallback to current directory
		}
	}

	// Return immutable command value type
	return ValidationCommand{
		Target:           target,
		Output:           output,
		Config:           cfg,
		RepoPath:         repoPath,
		SkipMergeCommits: skipMergeCommits,
	}, nil
}

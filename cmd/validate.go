// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cmd

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/configuration"
	gitService "github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/validation"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validates commit messages against configured rules",
		Long:  `Validates commit messages in the repository against the rules defined in your configuration file.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get configuration
			gommitLintConf, err := configuration.New()
			if err != nil {
				return fmt.Errorf("failed to create validator: %w", err)
			}

			// Create Git service
			git, err := gitService.NewService()
			if err != nil {
				return fmt.Errorf("failed to initialize git service: %w", err)
			}

			// Process flags
			opts, err := processFlags(cmd, git)
			if err != nil {
				return err
			}

			// Validate
			report, err := validation.NewValidator(opts, gommitLintConf.GommitConf)
			if err != nil {
				return fmt.Errorf("failed to create validator: %w", err)
			}

			r, err := report.Validate()
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			return internal.PrintReport(r.All(), &internal.CommitInfo{})
		},
	}

	// Add flags
	validateCmd.Flags().String("message-file", "", "commit message file path to validate")
	validateCmd.Flags().String("git-reference", "", "git reference to validate (defaults to auto-detected main branch)")
	validateCmd.Flags().String("revision-range", "", "range of commits to validate (<commit1>..<commit2>)")
	validateCmd.Flags().String("base-branch", "", "base branch to compare with (sets revision-range to <base-branch>..HEAD and overrides git-reference)")

	return validateCmd
}

// processFlags handles all flag logic with clear precedence rules.
func processFlags(cmd *cobra.Command, git gitService.Service) (*model.Options, error) {
	opts := model.NewOptions()

	mainBranch, err := git.DetectMainBranch()
	if err != nil || mainBranch == "" {
		return nil, fmt.Errorf("failed to detect main branch: %w", err)
	}

	if mainBranch != "" {
		opts.CommitRef = "refs/heads/" + mainBranch
		cmd.Printf("Auto-detected main branch: %s\n", mainBranch)
	}

	// 1. First check for commit message file
	msgFromFile, err := cmd.Flags().GetString("message-file")
	if err != nil {
		return nil, fmt.Errorf("failed to get message-file flag: %w", err)
	}

	if msgFromFile != "" {
		opts.MsgFromFile = &msgFromFile

		return opts, nil
	}

	// 2. Check for base branch
	baseBranch, err := cmd.Flags().GetString("base-branch")
	if err != nil {
		return nil, fmt.Errorf("failed to get base-branch flag: %w", err)
	}

	if baseBranch != "" {
		opts.RevisionRange = baseBranch + "..HEAD"
		opts.CommitRef = "refs/heads/" + baseBranch
		cmd.Printf("Using base-branch: %s (overrides git-reference if provided)\n", baseBranch)

		return opts, nil
	}

	// 3. Check for revision range if base branch not provided
	revisionRange, err := cmd.Flags().GetString("revision-range")
	if err != nil {
		return nil, fmt.Errorf("failed to get revision-range flag: %w", err)
	}

	if revisionRange != "" {
		opts.RevisionRange = revisionRange

		return opts, nil
	}

	commitRef, err := cmd.Flags().GetString("git-reference")
	if err != nil {
		return nil, fmt.Errorf("failed to get git-reference flag: %w", err)
	}

	if commitRef != "" {
		refExist := git.RefExists(commitRef)
		if !refExist {
			return nil, fmt.Errorf("failed to find git-reference: %s", commitRef)
		}

		opts.CommitRef = commitRef

		return opts, nil
	}

	return opts, nil
}

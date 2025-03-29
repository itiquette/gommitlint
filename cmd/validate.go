// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package cmd

import (
	"errors"
	"fmt"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/validation"
	"github.com/spf13/cobra"
)

const main = "main"

func newValidateCmd() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("the validate command does not take arguments")
			}
			// Done validating the arguments, do not print usage for errors
			// after this point
			cmd.SilenceUsage = true

			gommitLintConf, err := configuration.New()
			if err != nil {
				return fmt.Errorf("failed to create validator: %w", err)
			}

			opts := model.NewOptions()

			if commitMsgFile := cmd.Flags().Lookup("commit-msg-file").Value.String(); commitMsgFile != "" {
				opts.CommitMsgFile = &commitMsgFile
			}

			if commitRef := cmd.Flags().Lookup("commit-ref").Value.String(); commitRef != "" {
				opts.CommitRef = commitRef
			} else {
				mainBranch, err := detectMainBranch()
				if err != nil {
					return fmt.Errorf("failed to detect main branch: %w", err)
				}
				if mainBranch != "" {
					opts.CommitRef = "refs/heads/" + mainBranch
				}
			}

			if baseBranch := cmd.Flags().Lookup("base-branch").Value.String(); baseBranch != "" {
				opts.RevisionRange = baseBranch + "..HEAD"
				opts.CommitRef = "refs/heads/" + baseBranch
			} else if revisionRange := cmd.Flags().Lookup("revision-range").Value.String(); revisionRange != "" {
				opts.RevisionRange = revisionRange
			}

			report, err := validation.NewValidator(opts, gommitLintConf.GommitConf)
			if err != nil {
				return err
			}
			r, _ := report.Validate()

			return internal.PrintReport(r.All())
		},
	}

	validateCmd.Flags().String("commit-msg-file", "", "the path to the temporary commit message file")
	validateCmd.Flags().String("commit-ref", "", "the ref to compare git policies against")
	validateCmd.Flags().String("revision-range", "", "<commit1>..<commit2>")
	validateCmd.Flags().String("base-branch", "", "base branch to compare with")

	return validateCmd
}

// 4. Fall back to "main" if no main branch can be determined.
func detectMainBranch() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		// not a git repo, ignore
		return "", nil //nolint:nilerr
	}

	// Step 1: Check if "main" or "master" branches exist locally
	// This is the fastest way to determine the main branch
	commonMainBranches := []string{"main", "master"}
	for _, branchName := range commonMainBranches {
		branchRef := plumbing.NewBranchReferenceName(branchName)

		_, err := repo.Reference(branchRef, true)
		if err == nil {
			// Branch exists locally
			return branchName, nil
		}
	}

	// Step 2: If no local main branches, try to find the default branch from remote
	// Most Git hosting services set a HEAD symref pointing to the default branch
	remote, err := repo.Remote(git.DefaultRemoteName)
	if err != nil {
		// No origin remote, fall back to "main"
		return main, nil //nolint
	}

	// Get remote references
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return main, err
	}

	// Check for HEAD symref in remote refs to identify default branch
	// The origin/HEAD ref typically points to the default branch (e.g., refs/remotes/origin/HEAD -> refs/remotes/origin/main)
	for _, ref := range refs {
		if ref.Name().String() == "refs/remotes/origin/HEAD" {
			// The HEAD symref usually points to the default branch
			if target, err := repo.Reference(ref.Name(), true); err == nil {
				// Extract branch name from "refs/remotes/origin/branchname"
				targetName := target.Target().String()

				return strings.TrimPrefix(targetName, "refs/remotes/origin/"), nil
			}
		}
	}

	// Step 3: Check if remote has main or master branches
	// Even if they don't exist locally, they might exist on the remote
	for _, branchName := range commonMainBranches {
		remoteRef := plumbing.NewRemoteReferenceName(git.DefaultRemoteName, branchName)

		_, err := repo.Reference(remoteRef, true)
		if err == nil {
			// Remote branch exists
			return branchName, nil
		}
	}

	// Step 4: Default fallback to "main"
	// If all else fails, "main" is the modern default
	return "main", nil
}

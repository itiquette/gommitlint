// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/janderssonse/gommitlint/internal/configuration"
	"github.com/spf13/cobra"
)

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

			opts := []configuration.Option{}

			if commitMsgFile := cmd.Flags().Lookup("commit-msg-file").Value.String(); commitMsgFile != "" {
				opts = append(opts, configuration.WithCommitMsgFile(&commitMsgFile))
			}

			if commitRef := cmd.Flags().Lookup("commit-ref").Value.String(); commitRef != "" {
				opts = append(opts, configuration.WithCommitRef(commitRef))
			} else {
				mainBranch, err := detectMainBranch()
				if err != nil {
					return fmt.Errorf("failed to detect main branch: %w", err)
				}
				if mainBranch != "" {
					opts = append(opts, configuration.WithCommitRef("refs/heads/"+mainBranch))
				}
			}

			if baseBranch := cmd.Flags().Lookup("base-branch").Value.String(); baseBranch != "" {
				opts = append(opts, configuration.WithRevisionRange(baseBranch+"..HEAD"))
			} else if revisionRange := cmd.Flags().Lookup("revision-range").Value.String(); revisionRange != "" {
				opts = append(opts, configuration.WithRevisionRange(revisionRange))
			}

			s := configuration.NewDefaultOptions(opts...)
			report, err := configuration.Compliance(s, gommitLintConf.GommitConf)
			if err != nil {
				return err
			}

			return configuration.Validate(report.Checks())
		},
	}

	validateCmd.Flags().String("commit-msg-file", "", "the path to the temporary commit message file")
	validateCmd.Flags().String("commit-ref", "", "the ref to compare git policies against")
	validateCmd.Flags().String("revision-range", "", "<commit1>..<commit2>")
	validateCmd.Flags().String("base-branch", "", "base branch to compare with")

	return validateCmd
}

func detectMainBranch() (string, error) {
	mainBranch := "main"

	repo, err := git.PlainOpen(".")
	if err != nil {
		// not a git repo, ignore
		return "", nil //nolint:nilerr
	}

	repoConfig, err := repo.Config()
	if err != nil {
		return "", fmt.Errorf("failed to get repository configuration: %w", err)
	}

	rawConfig := repoConfig.Raw

	const branchSectionName = "branch"

	branchSection := rawConfig.Section(branchSectionName)
	for _, b := range branchSection.Subsections {
		remote := b.Option("remote")
		if remote == git.DefaultRemoteName {
			mainBranch = b.Name

			break
		}
	}

	return mainBranch, nil
}

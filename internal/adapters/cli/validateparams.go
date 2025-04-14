// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/spf13/cobra"
)

// ValidationParameters represents the complete set of CLI validation parameters.
type ValidationParameters struct {
	// Validation target (what to validate)
	Target ValidationTarget

	// Output options (how to format results)
	Output OutputOptions

	// Repository options
	RepoPath         string
	SkipMergeCommits bool
}

// NewValidateParams creates a ValidationParameters from command flags.
func NewValidateParams(cmd *cobra.Command) (ValidationParameters, error) {
	// Extract flag values
	messageFile, _ := cmd.Flags().GetString("message-file")
	gitReference, _ := cmd.Flags().GetString("git-reference")
	commitCount, _ := cmd.Flags().GetInt("commit-count")
	revisionRange, _ := cmd.Flags().GetString("revision-range")
	baseBranch, _ := cmd.Flags().GetString("base-branch")
	repoPath, _ := cmd.Flags().GetString("repo-path")
	skipMergeCommits, _ := cmd.Flags().GetBool("skip-merge-commits")

	format, _ := cmd.Flags().GetString("format")
	verbose, _ := cmd.Flags().GetBool("verbose")
	extraVerbose, _ := cmd.Flags().GetBool("extra-verbose")
	ruleHelp, _ := cmd.Flags().GetString("rulehelp")
	lightMode, _ := cmd.Flags().GetBool("light-mode")

	// Build validation target using pure function
	target, err := NewValidationTarget(messageFile, gitReference, revisionRange, baseBranch, commitCount)
	if err != nil {
		return ValidationParameters{}, err
	}

	// Build output options using focused value type
	output := NewOutputOptions(cmd.OutOrStdout()).
		WithFormat(format).
		WithVerbose(verbose).
		WithExtraVerbose(extraVerbose).
		WithRuleHelp(ruleHelp).
		WithLightMode(lightMode)

	return ValidationParameters{
		Target:           target,
		Output:           output,
		RepoPath:         repoPath,
		SkipMergeCommits: skipMergeCommits,
	}, nil
}

// GetRepoPath returns the repository path, with empty string indicating current directory.
func (p ValidationParameters) GetRepoPath() string {
	return p.RepoPath
}

// ToReportOptions converts ValidationParameters to domain.ReportOptions.
func (p ValidationParameters) ToReportOptions() domain.ReportOptions {
	return p.Output.ToReportOptions()
}

// WriteReport writes a formatted report using the configured output options.
func (p ValidationParameters) WriteReport(report domain.Report) error {
	return p.Output.WriteReport(report)
}

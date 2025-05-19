// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"fmt"
	"os"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/git"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/spf13/cobra"
)

// newValidateCmd creates a new validate command.
func newValidateCmd(ctx context.Context) *cobra.Command {
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
			// Process validation request
			exitCode, err := runNewValidation(ctx, cmd)
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
func runNewValidation(ctx context.Context, cmd *cobra.Command) (int, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering runNewValidation")
	// Create parameters object to encapsulate all inputs
	params := NewValidationParameters(cmd)

	// Create validation service using the context
	service, err := params.CreateValidationService(ctx)
	if err != nil {
		return 1, fmt.Errorf("failed to create validation service: %w", err)
	}

	// We'll directly validate HEAD instead of using options
	// No need to convert parameters

	// Try to get a specific commit directly
	commitResult, err := service.ValidateCommit(ctx, "HEAD")
	if err != nil {
		return 1, fmt.Errorf("failed to validate HEAD commit: %w", err)
	}

	// Create a validation results object with this commit
	results := domain.NewValidationResults()
	results = results.WithResult(commitResult)

	// Create formatter and report generator
	formatter := params.CreateFormatter()
	reportOptions := params.ToReportOptions()
	reportGenerator := report.NewGenerator(reportOptions, formatter)

	// Generate report using the domain interface
	err = reportGenerator.GenerateReport(ctx, results)
	if err != nil {
		return 1, fmt.Errorf("failed to generate report: %w", err)
	}

	// Return success if all rules passed, validation failure code otherwise
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// constructValidationService creates a validation service using injected dependencies.
func constructValidationService(ctx context.Context, deps *AppDependencies, repoPath string) (validate.ValidationService, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("repo_path", repoPath).Msg("Entering constructValidationService")

	// Get config from manager
	unifiedConfig := deps.ConfigManager.GetConfig()

	// Add the config to the context using the standard pattern via adapter
	ctx = contextx.WithConfig(ctx, infraConfig.NewAdapter(unifiedConfig))

	// Create a repository adapter
	repoAdapter, err := git.NewRepositoryAdapter(ctx, repoPath)
	if err != nil {
		return validate.ValidationService{}, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	// Create a validation service that uses the registry directly
	return validate.CreateValidationService(
		ctx,         // Context with configuration
		repoAdapter, // CommitRepository
		repoAdapter, // RepositoryInfoProvider
		repoAdapter, // CommitAnalyzer
	), nil
}

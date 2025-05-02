// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
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
	// Create parameters object to encapsulate all inputs
	params := NewValidationParameters(cmd)

	// Create validation service
	service, err := params.CreateValidationService()
	if err != nil {
		return 1, fmt.Errorf("failed to create validation service: %w", err)
	}

	// Convert parameters to validation options
	opts, err := params.ToValidationOptions()
	if err != nil {
		return 1, err
	}

	results, err := service.ValidateWithOptions(ctx, opts)
	if err != nil {
		return 1, fmt.Errorf("validation failed: %w", err)
	}

	// Create formatter and report generator
	formatter := params.CreateFormatter()
	reportOptions := params.ToReportOptions()
	reportGenerator := report.NewGenerator(reportOptions, formatter)

	// Generate report using the domain interface
	err = reportGenerator.GenerateReport(results)
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
func constructValidationService(deps *AppDependencies, repoPath string) (validate.ValidationService, error) {
	// Get config from manager
	unifiedConfig := deps.ConfigManager.GetValidationConfig()

	// Adapt the config to match the ValidationConfig interface
	validationConfig := NewConfigAdapter(unifiedConfig)

	// Create a repository adapter
	repoAdapter, err := git.NewRepositoryAdapter(repoPath)
	if err != nil {
		return validate.ValidationService{}, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	// Return the validation service with dependencies
	return validate.CreateValidationService(
		validationConfig,
		repoAdapter, // GitCommitService
		repoAdapter, // RepositoryInfoProvider
		repoAdapter, // CommitAnalyzer
	), nil
}

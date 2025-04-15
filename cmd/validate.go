// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/configuration"
	gitService "github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/results"
	"github.com/itiquette/gommitlint/internal/validation"
	"github.com/spf13/cobra"
)

// Note: Error handling is now done in runValidation function

func newValidateCmd() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validates commit message/s against configured rules",
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
			exitCode, err := runValidation(cmd)
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
	validateCmd.Flags().String("git-reference", "", "git reference to validate (defaults to auto-detected main branch)")
	validateCmd.Flags().String("revision-range", "", "range of commits to validate (<commit1>..<commit2>)")
	validateCmd.Flags().String("base-branch", "", "base branch to compare with (sets revision-range to <base-branch>..HEAD and overrides git-reference)")
	validateCmd.Flags().BoolP("verbose", "v", false, "show detailed validation results")
	validateCmd.Flags().Bool("extra-verbose", false, "show extra detailed validation results")
	validateCmd.Flags().Bool("light-mode", false, "use light background color scheme")
	validateCmd.Flags().String("rulehelp", "", "show detailed help for a specific rule (e.g., --rulehelp=signature)")
	validateCmd.Flags().String("format", "text", "output format: text or json")

	return validateCmd
}

// runValidation handles the core validation logic and returns an exit code.
func runValidation(cmd *cobra.Command) (int, error) {
	// Setup phase
	validator, _, err := setupValidation(cmd)
	if err != nil {
		return 1, err
	}

	// Validation phase
	aggregator, err := validateCommits(cmd, validator)
	if err != nil {
		return 1, err
	}

	// Reporting phase
	if err := generateReport(cmd, aggregator); err != nil {
		return 1, err
	}

	// Return success if all rules passed, validation failure code otherwise
	if aggregator.AllRulesPassed() {
		return 0, nil
	}

	return 2, nil
}

// setupValidation prepares the validation environment.
func setupValidation(cmd *cobra.Command) (*validation.Validator, gitService.Service, error) {
	// Get configuration
	gommitLintConf, err := configuration.New()
	if err != nil {
		return nil, nil, internal.NewConfigError(fmt.Errorf("failed to load configuration: %w", err))
	}

	// Create Git service
	git, err := gitService.NewService()
	if err != nil {
		return nil, nil, internal.NewGitError(fmt.Errorf("failed to initialize git service: %w", err))
	}

	// Process flags
	opts, err := processFlags(cmd, git)
	if err != nil {
		// Error is already wrapped with specific type in processFlags functions
		return nil, nil, err
	}

	// Create validator
	validator, err := validation.NewValidator(opts, gommitLintConf.GommitConf)
	if err != nil {
		return nil, nil, internal.NewConfigError(fmt.Errorf("failed to create validator: %w", err))
	}

	return validator, git, nil
}

// validateCommits processes all commits and collects validation results.
func validateCommits(cmd *cobra.Command, validator *validation.Validator) (*results.Aggregator, error) {
	// Create context for the validation with cancellation
	ctx := cmd.Context()

	// Get commits to validate using context
	commits, err := validator.GetCommitsToValidateWithContext(ctx)
	if err != nil {
		return nil, internal.NewValidationError(fmt.Errorf("failed to get commits: %w", err))
	}

	if len(commits) == 0 {
		return nil, internal.NewInputError(errors.New("no commits found to validate"))
	}

	// Create result aggregator
	aggregator := results.NewAggregator()

	// For each commit, validate and add results to aggregator
	for _, commitInfo := range commits {
		// Check for context cancellation
		if ctx.Err() != nil {
			return nil, internal.NewApplicationError(ctx.Err(), 1, "Validation")
		}

		// Validate the commit with context
		rules, err := validator.ValidateCommitWithContext(ctx, commitInfo)
		if err != nil {
			// Just log error and continue with next commit
			hash := "unknown"
			if commitInfo.RawCommit != nil {
				hash = commitInfo.RawCommit.Hash.String()
			}

			cmd.PrintErrf("Error validating commit %s: %v\n", hash, err)

			continue
		}

		// Add results to aggregator
		aggregator.AddCommitResult(commitInfo, rules.All())
	}

	return aggregator, nil
}

// generateReport creates and displays the validation report.
func generateReport(cmd *cobra.Command, aggregator *results.Aggregator) error {
	// Create reporter options
	reporterOptions, err := createReporterOptions(cmd)
	if err != nil {
		return internal.NewConfigError(fmt.Errorf("failed to create reporter options: %w", err))
	}

	// Create reporter and generate report
	reporter := results.NewReporter(aggregator, reporterOptions)
	if err := reporter.GenerateReport(); err != nil {
		return internal.NewApplicationError(
			fmt.Errorf("failed to generate report: %w", err),
			1,
			"Reporting",
		)
	}

	return nil
}

// createReporterOptions builds the options for the validation reporter.
func createReporterOptions(cmd *cobra.Command) (results.ReporterOptions, error) {
	// Determine output format
	formatFlag, err := cmd.Flags().GetString("format")
	if err != nil {
		return results.ReporterOptions{}, fmt.Errorf("failed to get format flag: %w", err)
	}

	outputFormat := results.FormatText
	if formatFlag == "json" {
		outputFormat = results.FormatJSON
	}

	// Get display flags
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return results.ReporterOptions{}, fmt.Errorf("failed to get verbose flag: %w", err)
	}

	extraVerbose, err := cmd.Flags().GetBool("extra-verbose")
	if err != nil {
		return results.ReporterOptions{}, fmt.Errorf("failed to get extra-verbose flag: %w", err)
	}

	lightMode, err := cmd.Flags().GetBool("light-mode")
	if err != nil {
		return results.ReporterOptions{}, fmt.Errorf("failed to get light-mode flag: %w", err)
	}

	ruleHelp, err := cmd.Flags().GetString("rulehelp")
	if err != nil {
		return results.ReporterOptions{}, fmt.Errorf("failed to get rulehelp flag: %w", err)
	}

	return results.ReporterOptions{
		Format:         outputFormat,
		Verbose:        verbose,
		ShowHelp:       extraVerbose || ruleHelp != "",
		RuleToShowHelp: ruleHelp,
		LightMode:      lightMode,
		Writer:         cmd.OutOrStdout(),
	}, nil
}

// processFlags handles all flag logic with clear precedence rules.
func processFlags(cmd *cobra.Command, git gitService.Service) (*model.Options, error) {
	opts := model.NewOptions()

	// Process display and help flags
	if err := processDisplayFlags(cmd, opts); err != nil {
		return nil, internal.NewInputError(fmt.Errorf("failed to process display flags: %w", err))
	}

	// Detect main branch
	if err := detectAndSetMainBranch(cmd, git, opts); err != nil {
		return nil, err
	}

	// Apply validation source with precedence order:
	// 1. Message from file (highest priority)
	// 2. Base branch comparison
	// 3. Revision range
	// 4. Single git reference
	// 5. Default to main branch (already set above)
	if processed, err := processMessageFromFile(cmd, opts); err != nil {
		return nil, err
	} else if processed {
		return opts, nil
	}

	if processed, err := processBaseBranch(cmd, opts); err != nil {
		return nil, err
	} else if processed {
		return opts, nil
	}

	if processed, err := processRevisionRange(cmd, opts); err != nil {
		return nil, err
	} else if processed {
		return opts, nil
	}

	if processed, err := processGitReference(cmd, git, opts); err != nil {
		return nil, err
	} else if processed {
		return opts, nil
	}

	return opts, nil
}

// processDisplayFlags handles flags related to output display options.
func processDisplayFlags(cmd *cobra.Command, opts *model.Options) error {
	// Process verbose flag
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	opts.Verbose = verbose

	// Process extra-verbose flag
	extraVerbose, err := cmd.Flags().GetBool("extra-verbose")
	if err != nil {
		return fmt.Errorf("failed to get extra-verbose flag: %w", err)
	}

	opts.ShowHelp = extraVerbose

	// Process rule help flag
	helpRule, err := cmd.Flags().GetString("rulehelp")
	if err != nil {
		return fmt.Errorf("failed to get rulehelp flag: %w", err)
	}

	if helpRule != "" {
		opts.ShowHelp = true
		opts.RuleToShowHelp = helpRule
	}

	// Process light mode flag
	lightMode, err := cmd.Flags().GetBool("light-mode")
	if err != nil {
		return fmt.Errorf("failed to get light-mode flag: %w", err)
	}

	opts.LightMode = lightMode

	return nil
}

// detectAndSetMainBranch detects the main branch and sets it as default commit reference.
func detectAndSetMainBranch(cmd *cobra.Command, git gitService.Service, opts *model.Options) error {
	ctx := cmd.Context()

	mainBranch, err := git.DetectMainBranchWithContext(ctx)
	if err != nil || mainBranch == "" {
		return internal.NewGitError(
			fmt.Errorf("failed to detect main branch: %w", err),
			map[string]string{"action": "detect_main_branch"},
		)
	}

	opts.CommitRef = "refs/heads/" + mainBranch
	cmd.Printf("Auto-detected main branch: %s\n", mainBranch)

	return nil
}

// processMessageFromFile checks for commit message file and configures options if present.
// Returns true if processed and should exit the flag processing pipeline.
func processMessageFromFile(cmd *cobra.Command, opts *model.Options) (bool, error) {
	msgFromFile, err := cmd.Flags().GetString("message-file")
	if err != nil {
		return false, internal.NewInputError(
			fmt.Errorf("failed to get message-file flag: %w", err),
			map[string]string{"flag": "message-file"},
		)
	}

	if msgFromFile != "" {
		opts.MsgFromFile = &msgFromFile

		return true, nil
	}

	return false, nil
}

// processBaseBranch checks for base branch flag and configures options if present.
// Returns true if processed and should exit the flag processing pipeline.
func processBaseBranch(cmd *cobra.Command, opts *model.Options) (bool, error) {
	baseBranch, err := cmd.Flags().GetString("base-branch")
	if err != nil {
		return false, internal.NewInputError(
			fmt.Errorf("failed to get base-branch flag: %w", err),
			map[string]string{"flag": "base-branch"},
		)
	}

	if baseBranch != "" {
		opts.RevisionRange = baseBranch + "..HEAD"
		opts.CommitRef = "refs/heads/" + baseBranch
		cmd.Printf("Using base-branch: %s (overrides git-reference if provided)\n", baseBranch)

		return true, nil
	}

	return false, nil
}

// processRevisionRange checks for revision range flag and configures options if present.
// Returns true if processed and should exit the flag processing pipeline.
func processRevisionRange(cmd *cobra.Command, opts *model.Options) (bool, error) {
	revisionRange, err := cmd.Flags().GetString("revision-range")
	if err != nil {
		return false, internal.NewInputError(
			fmt.Errorf("failed to get revision-range flag: %w", err),
			map[string]string{"flag": "revision-range"},
		)
	}

	if revisionRange != "" {
		opts.RevisionRange = revisionRange

		return true, nil
	}

	return false, nil
}

// processGitReference checks for git reference flag and configures options if present.
// Returns true if processed and should exit the flag processing pipeline.
func processGitReference(cmd *cobra.Command, git gitService.Service, opts *model.Options) (bool, error) {
	ctx := cmd.Context()

	commitRef, err := cmd.Flags().GetString("git-reference")
	if err != nil {
		return false, internal.NewInputError(
			fmt.Errorf("failed to get git-reference flag: %w", err),
			map[string]string{"flag": "git-reference"},
		)
	}

	if commitRef != "" {
		refExists := git.RefExistsWithContext(ctx, commitRef)
		if !refExists {
			return false, internal.NewGitError(
				fmt.Errorf("git reference not found: %s", commitRef),
				map[string]string{"reference": commitRef},
			)
		}

		opts.CommitRef = commitRef

		return true, nil
	}

	return false, nil
}

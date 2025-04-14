// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cmd

import (
	"fmt"
	"os"

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
	// Get configuration
	gommitLintConf, err := configuration.New()
	if err != nil {
		return 1, fmt.Errorf("failed to create validator: %w", err)
	}

	// Create Git service
	git, err := gitService.NewService()
	if err != nil {
		return 3, fmt.Errorf("failed to initialize git service: %w", err)
	}

	// Process flags
	opts, err := processFlags(cmd, git)
	if err != nil {
		return 1, fmt.Errorf("failed to process flags: %w", err)
	}

	// Create validator
	validator, err := validation.NewValidator(opts, gommitLintConf.GommitConf)
	if err != nil {
		return 1, fmt.Errorf("failed to create validator: %w", err)
	}

	// Get commits to validate
	commits, err := validator.GetCommitsToValidate()
	if err != nil {
		return 1, fmt.Errorf("failed to get commits: %w", err)
	}

	// Create result aggregator
	aggregator := results.NewAggregator()

	// For each commit, validate and add results to aggregator
	for _, commitInfo := range commits {
		// Validate the commit
		rules, err := validator.ValidateCommit(commitInfo)
		if err != nil {
			// Just log error and continue with next commit
			cmd.PrintErrf("Error validating commit %s: %v\n", commitInfo.RawCommit.Hash.String(), err)

			continue
		}

		// Add results to aggregator
		aggregator.AddCommitResult(commitInfo, rules.All())
	}

	// Determine output format
	formatFlag, _ := cmd.Flags().GetString("format")
	outputFormat := results.FormatText

	if formatFlag == "json" {
		outputFormat = results.FormatJSON
	}

	// Create reporter options
	verbose, _ := cmd.Flags().GetBool("verbose")
	extraVerbose, _ := cmd.Flags().GetBool("extra-verbose")
	lightMode, _ := cmd.Flags().GetBool("light-mode")
	ruleHelp, _ := cmd.Flags().GetString("rulehelp")

	reporterOptions := results.ReporterOptions{
		Format:         outputFormat,
		Verbose:        verbose,
		ShowHelp:       extraVerbose || ruleHelp != "",
		RuleToShowHelp: ruleHelp,
		LightMode:      lightMode,
		Writer:         cmd.OutOrStdout(),
	}

	// Create reporter and generate report
	reporter := results.NewReporter(aggregator, reporterOptions)
	if err := reporter.GenerateReport(); err != nil {
		return 1, fmt.Errorf("failed to generate report: %w", err)
	}

	// Return success if all rules passed, validation failure code otherwise
	if aggregator.AllRulesPassed() {
		return 0, nil
	}

	return 2, nil
}

// processFlags handles all flag logic with clear precedence rules.
func processFlags(cmd *cobra.Command, git gitService.Service) (*model.Options, error) {
	opts := model.NewOptions()

	verbose, err := cmd.Flags().GetBool("verbose")
	if err == nil {
		opts.Verbose = verbose
	}

	extraVerbose, err := cmd.Flags().GetBool("extra-verbose")
	if err == nil {
		opts.ShowHelp = extraVerbose
	}

	// Check for help option
	helpRule, err := cmd.Flags().GetString("rulehelp")
	if err == nil && helpRule != "" {
		opts.ShowHelp = true
		opts.RuleToShowHelp = helpRule
	}

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

	lightMode, err := cmd.Flags().GetBool("light-mode")
	if err == nil {
		opts.LightMode = lightMode
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

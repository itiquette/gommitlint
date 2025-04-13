// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/configuration"
	gitService "github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/validation"
	"github.com/spf13/cobra"
)

// exitCode 2 = validation failure.
func handleCommandError(err error, message string, exitCode int) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
		os.Exit(exitCode)
	}
}

func newValidateCmd() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validates commit message/s against configured rules",
		Long:  `Validates commit message/s against the a set of rules.`,
		Run: func(cmd *cobra.Command, _ []string) {
			// Get configuration
			gommitLintConf, err := configuration.New()
			if err != nil {
				handleCommandError(err, "Failed to create validator", 1)
			}

			// Create Git service
			git, err := gitService.NewService()
			if err != nil {
				handleCommandError(err, "Failed to initialize git service", 3)
			}

			// Process flags
			opts, err := processFlags(cmd, git)
			if err != nil {
				handleCommandError(err, "Failed to process flags", 1)
			}

			// Validate
			validator, err := validation.NewValidator(opts, gommitLintConf.GommitConf)
			if err != nil {
				handleCommandError(err, "Failed to create validator", 1)
			}

			// Get commits to validate
			commits, err := validator.GetCommitsToValidate()
			if err != nil {
				handleCommandError(err, "Failed to get commits", 1)
			}

			// Create printer options with proper verbose/help settings
			printOpts := &internal.PrintOptions{
				Verbose:        opts.Verbose,
				ShowHelp:       opts.ShowHelp,
				RuleToShowHelp: opts.RuleToShowHelp,
				LightMode:      opts.LightMode,
			}

			passedCommits := 0

			// For each commit, validate and print results
			for _, commitInfo := range commits {
				// Validate the commit
				rules, err := validator.ValidateCommit(commitInfo)
				if err != nil {
					continue
				}

				// Print report for this commit
				err = internal.PrintReport(rules.All(), &commitInfo, printOpts)
				if err != nil {
					continue
				}

				// Track if this commit passed (all rules passed)
				commitPassed := true
				for _, rule := range rules.All() {
					if len(rule.Errors()) > 0 {
						commitPassed = false

						break
					}
				}

				if commitPassed {
					passedCommits++
				}
			}

			// Print overall summary if multiple commits were validated
			if len(commits) > 1 {
				printOverallSummary(
					len(commits),
					passedCommits,
					color.NoColor,  // Use the current global color setting
					opts.LightMode, // Use light mode setting from options
				)
			}

			// Check for validation failures
			if len(commits) != passedCommits {
				// Exit with status 2 for validation failures
				fmt.Fprintln(os.Stderr, color.New(color.FgRed, color.Bold).Sprint("Validation failed: some commits did not pass all rules"))
				os.Exit(2)
			}
		},
	}

	validateCmd.Flags().String("message-file", "", "commit message file path to validate")
	validateCmd.Flags().String("git-reference", "", "git reference to validate (defaults to auto-detected main branch)")
	validateCmd.Flags().String("revision-range", "", "range of commits to validate (<commit1>..<commit2>)")
	validateCmd.Flags().String("base-branch", "", "base branch to compare with (sets revision-range to <base-branch>..HEAD and overrides git-reference)")
	validateCmd.Flags().BoolP("verbose", "v", false, "show detailed validation results")
	validateCmd.Flags().Bool("extra-verbose", false, "show extra detailed validation results")
	validateCmd.Flags().Bool("light-mode", false, "use light background color scheme")
	validateCmd.Flags().String("rulehelp", "", "show detailed help for a specific rule (e.g., --rulehelp=signature)")

	return validateCmd
}

// Print overall summary focused on commit success/failure.
func printOverallSummary(totalCommits int, passedCommits int, noColor bool, lightMode bool) {
	// Create a divider line
	divider := strings.Repeat("=", 80)

	// Define colors based on settings
	var summaryColor func(format string, a ...interface{}) string

	if noColor {
		// No color mode - just use plain text
		summaryColor = fmt.Sprintf
	} else if lightMode {
		// Light mode - use blue bold (works well on light backgrounds)
		summaryColor = color.New(color.FgBlue, color.Bold).SprintfFunc()
	} else {
		// Dark mode - use cyan bold (works well on dark backgrounds)
		summaryColor = color.New(color.FgHiBlue, color.Bold).SprintfFunc()
	}

	// Calculate failed commits
	failedCommits := totalCommits - passedCommits

	// Print the summary
	fmt.Println(summaryColor(divider))
	fmt.Println(summaryColor("OVERALL SUMMARY"))
	fmt.Println(summaryColor(divider))
	fmt.Printf("%s Validated %d commits\n", summaryColor("Result:"), totalCommits)
	fmt.Printf("  %s %d commits passed\n", summaryColor("Passed:"), passedCommits)
	fmt.Printf("  %s %d commits failed\n", summaryColor("Failed:"), failedCommits)
	fmt.Println()
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

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/adapters/git"
	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/spf13/cobra"
)

// newValidateCmd creates a new validate command.
//
//nolint:contextcheck // Context is retrieved from cmd.Context() in the Run function
func newValidateCmd(_ context.Context, config config.Config) *cobra.Command {
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
			// Get context from command (set by root command)
			ctx := cmd.Context()

			// Process validation request
			exitCode, err := runNewValidation(ctx, cmd, config)
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

// runNewValidation handles validation using direct composition.
func runNewValidation(ctx context.Context, cmd *cobra.Command, config config.Config) (int, error) {
	// Extract flags directly - pure function
	flags := extractValidationFlags(cmd)

	// Determine validation target - pure function
	target, err := determineValidationTarget(flags)
	if err != nil {
		return 1, fmt.Errorf("failed to determine validation target: %w", err)
	}

	// Execute validation using direct composition
	err = executeValidation(ctx, target, flags, config, cmd.OutOrStdout())
	if err != nil {
		// Check if it's a validation failure or system error
		if strings.Contains(err.Error(), "validation failed") {
			return 2, nil // Validation failure exit code
		}

		return 1, err // System error
	}

	return 0, nil // Success
}

// ValidationFlags holds all CLI flags.
type ValidationFlags struct {
	MessageFile      string
	GitReference     string
	CommitCount      int
	RevisionRange    string
	BaseBranch       string
	Format           string
	Verbose          bool
	ExtraVerbose     bool
	RuleHelp         string
	LightMode        bool
	RepoPath         string
	SkipMergeCommits bool
}

// extractValidationFlags extracts all flags from cobra command.
func extractValidationFlags(cmd *cobra.Command) ValidationFlags {
	messageFile, _ := cmd.Flags().GetString("message-file")
	gitReference, _ := cmd.Flags().GetString("git-reference")
	commitCount, _ := cmd.Flags().GetInt("commit-count")
	revisionRange, _ := cmd.Flags().GetString("revision-range")
	baseBranch, _ := cmd.Flags().GetString("base-branch")
	format, _ := cmd.Flags().GetString("format")
	verbose, _ := cmd.Flags().GetBool("verbose")
	extraVerbose, _ := cmd.Flags().GetBool("extra-verbose")
	ruleHelp, _ := cmd.Flags().GetString("rulehelp")
	lightMode, _ := cmd.Flags().GetBool("light-mode")
	repoPath, _ := cmd.Flags().GetString("repo-path")
	skipMergeCommits, _ := cmd.Flags().GetBool("skip-merge-commits")

	// Set default repo path if empty
	if repoPath == "" {
		var err error

		repoPath, err = os.Getwd()
		if err != nil {
			repoPath = "." // Fallback to current directory
		}
	}

	return ValidationFlags{
		MessageFile:      messageFile,
		GitReference:     gitReference,
		CommitCount:      commitCount,
		RevisionRange:    revisionRange,
		BaseBranch:       baseBranch,
		Format:           format,
		Verbose:          verbose,
		ExtraVerbose:     extraVerbose,
		RuleHelp:         ruleHelp,
		LightMode:        lightMode,
		RepoPath:         repoPath,
		SkipMergeCommits: skipMergeCommits,
	}
}

// determineValidationTarget determines what to validate based on flags.
func determineValidationTarget(flags ValidationFlags) (ValidationTarget, error) {
	// Use the existing pure function from validationtarget.go
	return NewValidationTarget(
		flags.MessageFile,
		flags.GitReference,
		flags.RevisionRange,
		flags.BaseBranch,
		flags.CommitCount,
	)
}

// executeValidation performs validation using direct composition.
func executeValidation(ctx context.Context, target ValidationTarget, flags ValidationFlags, cfg config.Config, output io.Writer) error {
	// Create repository adapter
	repo, err := git.NewRepository(flags.RepoPath)
	if err != nil {
		return fmt.Errorf("repository creation failed: %w", err)
	}

	// Create rules
	commitRules := rules.CreateCommitRules(cfg)
	repoRules := rules.CreateRepositoryRules(cfg)

	// Create domain logger adapter
	zerologLogger := logadapter.GetLogger(ctx)
	logger := logadapter.NewDomainLogger(zerologLogger)

	// Execute validation using pure domain functions
	report, err := validateTarget(ctx, target, commitRules, repoRules, repo, cfg, logger, flags.SkipMergeCommits)
	if err != nil {
		return err
	}

	// Create output options for formatting
	outputOpts := createOutputOptions(output, flags)

	// Generate output
	err = outputOpts.WriteReport(report)
	if err != nil {
		return err
	}

	// Return validation failure if rules failed
	if !report.Summary.AllPassed {
		return fmt.Errorf("validation failed: %d of %d commits passed", report.Summary.PassedCommits, report.Summary.TotalCommits)
	}

	return nil
}

// createOutputOptions creates output options from flags.
func createOutputOptions(output io.Writer, flags ValidationFlags) OutputOptions {
	return NewOutputOptions(output).
		WithFormat(flags.Format).
		WithVerbose(flags.Verbose).
		WithExtraVerbose(flags.ExtraVerbose).
		WithRuleHelp(flags.RuleHelp).
		WithLightMode(flags.LightMode)
}

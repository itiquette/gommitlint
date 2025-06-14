// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"context"
	"fmt"
	"os"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
	"github.com/itiquette/gommitlint/internal/adapters/git"
	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
	"github.com/itiquette/gommitlint/internal/adapters/output"
	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/urfave/cli/v3"
)

// NewValidateCommand creates the validate subcommand.
func NewValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate commit messages",
		Description: `Validates commit message/s against a set of rules.

Examples:
  # Validate commits in the current branch against main
  gommitlint validate --base-branch=main
  
  # Validate a specific commit
  gommitlint validate --ref=HEAD~1
  
  # Validate a commit message from a file
  gommitlint validate --message-file=/path/to/commit-msg.txt
  
  # Validate a range of commits
  gommitlint validate --range=main..feature
  
  # Validate last 5 commits
  gommitlint validate --count=5`,

		Flags: []cli.Flag{
			// Validation Target flags (choose one)
			&cli.StringFlag{
				Name:     "message-file",
				Aliases:  []string{"f"},
				Usage:    "validate commit message from `FILE`",
				Category: "Validation Target (choose one)",
			},
			&cli.StringFlag{
				Name:     "ref",
				Aliases:  []string{"r"},
				Usage:    "git `REF` to validate (default: HEAD)",
				Category: "Validation Target (choose one)",
			},
			&cli.IntFlag{
				Name:     "count",
				Aliases:  []string{"n"},
				Value:    1,
				Usage:    "number of commits from HEAD to validate",
				Category: "Validation Target (choose one)",
			},
			&cli.StringFlag{
				Name:     "range",
				Usage:    "validate commit `RANGE` (e.g., main..feature)",
				Category: "Validation Target (choose one)",
			},
			&cli.StringFlag{
				Name:     "base-branch",
				Usage:    "validate commits in `BRANCH`..HEAD",
				Category: "Validation Target (choose one)",
			},

			// Output flags
			&cli.IntFlag{
				Name:     "verbose",
				Aliases:  []string{"v"},
				Usage:    "show detailed validation results (1 for verbose, 2 for extra verbose)",
				Category: "Output Options",
			},
			&cli.StringFlag{
				Name:     "rule-help",
				Usage:    "show detailed help for `RULE`",
				Category: "Output Options",
			},
			&cli.StringFlag{
				Name:     "report-file",
				Usage:    "write results to `FILE`",
				Category: "Output Options",
			},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			return ExecuteValidation(ctx, cmd)
		},
	}
}

// ExecuteValidation orchestrates the validation process.
func ExecuteValidation(ctx context.Context, cmd *cli.Command) error {
	// Load configuration
	cfgResult, err := LoadConfigFromCommand(cmd.Root())
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := cfgResult.Config

	// Create logger from context
	zerologLogger := logadapter.GetLogger(ctx)
	logger := logadapter.NewDomainLogger(zerologLogger)

	// Create validation target from CLI flags
	target, err := createValidationTarget(cmd)
	if err != nil {
		return fmt.Errorf("failed to create validation target: %w", err)
	}

	// Create output options from CLI flags
	outputOptions, err := createOutputOptions(cmd)
	if err != nil {
		return fmt.Errorf("failed to create output options: %w", err)
	}

	// Handle rule help if requested
	if outputOptions.ShowRuleHelp() {
		return handleRuleHelp(outputOptions, cfg)
	}

	// Create Git repository
	repoPath := getRepoPath(cmd)

	repo, err := git.NewRepository(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Create rules from configuration
	commitRules := rules.CreateCommitRules(cfg)
	repoRules := rules.CreateRepositoryRules(cfg)

	// Execute validation
	report, err := cliAdapter.ValidateTarget(ctx, target, commitRules, repoRules, repo, cfg, logger)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Write output
	err = outputOptions.WriteReport(report)
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Return non-zero exit code if validation failed
	if !report.Summary.AllPassed {
		os.Exit(1)
	}

	return nil
}

// createValidationTarget creates a ValidationTarget from CLI flags.
func createValidationTarget(cmd *cli.Command) (cliAdapter.ValidationTarget, error) {
	messageFile := cmd.String("message-file")
	gitRef := cmd.String("ref")
	commitRange := cmd.String("range")
	baseBranch := cmd.String("base-branch")
	commitCount := cmd.Int("count")

	return cliAdapter.NewValidationTarget(messageFile, gitRef, commitRange, baseBranch, commitCount)
}

// createOutputOptions creates OutputOptions from CLI flags.
func createOutputOptions(cmd *cli.Command) (cliAdapter.OutputOptions, error) {
	// Determine output writer
	var writer *os.File

	reportFile := cmd.String("report-file")
	if reportFile != "" {
		file, err := os.Create(reportFile)
		if err != nil {
			return cliAdapter.OutputOptions{}, fmt.Errorf("failed to create report file: %w", err)
		}

		writer = file
	} else {
		writer = os.Stdout
	}

	// Get format from root command flags
	format := cmd.Root().String("format")

	// Validate format is supported
	if !output.IsValidFormat(format) {
		return cliAdapter.OutputOptions{}, fmt.Errorf("unsupported format '%s', supported formats: %v",
			format, output.SupportedFormats())
	}

	color := cmd.Root().String("color")
	quiet := cmd.Root().Bool("quiet")

	// Create base options
	options := cliAdapter.NewOutputOptions(writer).
		WithFormat(format).
		WithColor(color)

	// Handle verbose flags (command-specific)
	verboseLevel := cmd.Int("verbose")
	if verboseLevel > 0 && !quiet {
		options = options.WithVerbose(true)
	}

	// Handle rule help
	ruleHelp := cmd.String("rule-help")
	if ruleHelp != "" {
		options = options.WithRuleHelp(ruleHelp)
		if err := options.ValidateRuleHelp(); err != nil {
			return cliAdapter.OutputOptions{}, err
		}
	}

	return options, nil
}

// getRepoPath gets the repository path from CLI flags or defaults to current directory.
func getRepoPath(cmd *cli.Command) string {
	repoPath := cmd.Root().String("repo-path")
	if repoPath == "" {
		repoPath = "."
	}

	return repoPath
}

// handleRuleHelp shows help for a specific rule and exits.
func handleRuleHelp(options cliAdapter.OutputOptions, _ configTypes.Config) error {
	// For rule help, we create a minimal report showing rule information
	// This is a simplified implementation - in a full implementation you might
	// want to create a dedicated help system
	fmt.Printf("Help for rule: %s\n", options.GetRuleHelp())
	fmt.Println("(Rule help display not yet fully implemented)")

	return nil
}

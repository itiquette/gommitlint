// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/spf13/cobra"
)

// newValidateCmd creates a new validate command.
func newValidateCmd() *cobra.Command {
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
			exitCode, err := runNewValidation(cmd)
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
func runNewValidation(cmd *cobra.Command) (int, error) {
	// Get flags
	messageFile, _ := cmd.Flags().GetString("message-file")
	gitReference, _ := cmd.Flags().GetString("git-reference")
	commitCount, _ := cmd.Flags().GetInt("commit-count")
	revisionRange, _ := cmd.Flags().GetString("revision-range")
	baseBranch, _ := cmd.Flags().GetString("base-branch")
	verbose, _ := cmd.Flags().GetBool("verbose")
	extraVerbose, _ := cmd.Flags().GetBool("extra-verbose")
	lightMode, _ := cmd.Flags().GetBool("light-mode")
	ruleHelp, _ := cmd.Flags().GetString("rulehelp")
	format, _ := cmd.Flags().GetString("format")
	skipMergeCommits, _ := cmd.Flags().GetBool("skip-merge-commits")
	repoPath, _ := cmd.Flags().GetString("repo-path")

	// Create context
	ctx := cmd.Context()

	// Create validation service
	service, err := validate.CreateDefaultValidationService(repoPath)
	if err != nil {
		return 1, fmt.Errorf("failed to create validation service: %w", err)
	}

	// Process validation flags with precedence
	opts := validate.ValidationOptions{
		SkipMergeCommits: skipMergeCommits,
	}

	// Default to validating HEAD if no other option is provided
	if gitReference == "" && revisionRange == "" && baseBranch == "" && messageFile == "" && commitCount <= 0 {
		// Default behavior uses HEAD
		opts.CommitHash = "HEAD"
	}

	// Apply validation source with precedence order
	if messageFile != "" {
		// 1. Message from file (highest priority)
		opts.MessageFile = messageFile
	} else if baseBranch != "" {
		// 2. Base branch comparison
		opts.FromHash = baseBranch
		opts.ToHash = "HEAD"
	} else if revisionRange != "" {
		// 3. Revision range
		// Parse revision range (format: from..to)
		parts := parseRevisionRange(revisionRange)
		if len(parts) == 2 {
			opts.FromHash = parts[0]
			opts.ToHash = parts[1]
		} else {
			return 1, fmt.Errorf("invalid revision range format: %s (expected format: from..to)", revisionRange)
		}
	} else if gitReference != "" {
		// 4. Single git reference
		opts.CommitHash = gitReference
	} else if commitCount > 0 {
		// 5. Commit count
		opts.CommitCount = commitCount
	}

	// Validate according to options
	results, err := service.ValidateWithOptions(ctx, opts)
	if err != nil {
		return 1, fmt.Errorf("validation failed: %w", err)
	}

	// Create report generator
	reportOptions := report.Options{
		Format:         getReportFormat(format),
		Verbose:        verbose,
		ShowHelp:       extraVerbose || ruleHelp != "",
		RuleToShowHelp: ruleHelp,
		LightMode:      lightMode,
		Writer:         cmd.OutOrStdout(),
	}

	reportGenerator := report.NewGenerator(reportOptions)

	// Generate report
	err = reportGenerator.Generate(results)
	if err != nil {
		return 1, fmt.Errorf("failed to generate report: %w", err)
	}

	// Return success if all rules passed, validation failure code otherwise
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// parseRevisionRange parses a revision range string (e.g., "main..HEAD") into a slice of parts.
func parseRevisionRange(revisionRange string) []string {
	// Split by ".." to get from and to
	parts := []string{}

	// Handle both ".." and "..." formats
	if idx := indexOf(revisionRange, "..."); idx >= 0 {
		parts = append(parts, revisionRange[:idx], revisionRange[idx+3:])
	} else if idx := indexOf(revisionRange, ".."); idx >= 0 {
		parts = append(parts, revisionRange[:idx], revisionRange[idx+2:])
	}

	return parts
}

// indexOf returns the index of the first occurrence of substring in s, or -1 if not found.
func indexOf(s, substring string) int {
	for i := 0; i <= len(s)-len(substring); i++ {
		if s[i:i+len(substring)] == substring {
			return i
		}
	}

	return -1
}

// getReportFormat converts a string format to the report.Format type.
func getReportFormat(format string) report.Format {
	switch format {
	case "json":
		return report.FormatJSON
	case "github":
		return report.FormatGitHubActions
	case "gitlab":
		return report.FormatGitLabCI
	default:
		return report.FormatText
	}
}

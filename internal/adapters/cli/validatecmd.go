// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/itiquette/gommitlint/internal"
	log "github.com/itiquette/gommitlint/internal/adapters/logging"
	format "github.com/itiquette/gommitlint/internal/adapters/output"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/spf13/cobra"
)

// Logger provides structured logging capabilities.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// ValidationContext holds validation dependencies.
type ValidationContext struct {
	Validator domain.ValidatorWithDeps
	Formatter format.Formatter
	Logger    Logger
	Options   domain.ReportOptions
}

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

// runNewValidation handles the core validation logic and returns an exit code.
func runNewValidation(ctx context.Context, cmd *cobra.Command, config config.Config) (int, error) {
	// Create parameters object to encapsulate all inputs
	params := NewValidateParams(cmd)

	// Get repository path
	repoPath := params.GetRepoPath()
	if repoPath == "" {
		var err error

		repoPath, err = os.Getwd()
		if err != nil {
			return 1, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Direct service creation - no factory needed
	// Get the concrete logger from context
	loggerInterface := ctx.Value(LoggerKey)
	if loggerInterface == nil {
		return 1, errors.New("logger is nil")
	}

	// Cast to concrete logger type needed by wire
	logger, ok := loggerInterface.(log.Logger)
	if !ok {
		return 1, errors.New("logger is not the expected type")
	}

	validator, err := internal.CreateValidator(ctx, &config, repoPath, logger)
	if err != nil {
		return 1, fmt.Errorf("failed to create validator: %w", err)
	}

	// Create formatter based on parameters
	formatter := params.CreateFormatter()

	// Get report options from parameters
	reportOptions := params.ToReportOptions()

	// Determine what to validate
	targetType, target1, target2, err := params.GetValidationTarget()
	if err != nil {
		return 1, err
	}

	// Create validation context with common dependencies
	validateCtx := ValidationContext{
		Validator: validator,
		Formatter: formatter,
		Logger:    logger,
		Options:   reportOptions,
	}

	// Perform validation based on target type
	switch targetType {
	case "message":
		message, err := os.ReadFile(target1)
		if err != nil {
			return 1, fmt.Errorf("failed to read message file: %w", err)
		}

		return validateMessage(ctx, validateCtx, string(message))

	case "range":
		return validateRange(ctx, validateCtx, target1, target2, params.SkipMergeCommits)

	case "commit":
		return validateCommit(ctx, validateCtx, target1, params.SkipMergeCommits)

	case "count":
		validateCtx.Logger.Warn("Commit count validation not fully implemented, validating HEAD instead")

		return validateCommit(ctx, validateCtx, "HEAD", params.SkipMergeCommits)

	default:
		return 1, fmt.Errorf("unknown validation target type: %s", targetType)
	}
}

// Simple validation functions following direct service call pattern

// validateMessage validates a commit message and generates a report.
func validateMessage(ctx context.Context, vctx ValidationContext, message string) (int, error) {
	// Validate the message
	result, err := vctx.Validator.ValidateMessage(message)
	if err != nil {
		return 1, fmt.Errorf("failed to validate message: %w", err)
	}

	// Convert to results format for reporting
	results := domain.NewValidationResults()
	commitResult := convertValidationResult(result, vctx.Validator)
	results = results.AddResult(commitResult)

	// Generate report
	if err := generateReport(ctx, vctx, results); err != nil {
		return 1, err
	}

	// Return appropriate exit code
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// validateRange validates a range of commits and generates a report.
func validateRange(ctx context.Context, vctx ValidationContext, fromHash, toHash string, skipMerge bool) (int, error) {
	// Get commits in range
	commits, err := vctx.Validator.Deps.Repository.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return 1, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Filter merge commits if requested
	commits = domain.FilterMergeCommits(commits, skipMerge)

	// Validate commits
	validationResults := vctx.Validator.ValidateCommits(commits)

	// Convert to results format
	results := domain.NewValidationResults()

	for _, result := range validationResults {
		commitResult := convertValidationResult(result, vctx.Validator)
		results = results.AddResult(commitResult)
	}

	// Add repository validation
	repoFailures := vctx.Validator.ValidateRepository()
	if len(repoFailures) > 0 {
		repoResults := convertRepoFailures(repoFailures)
		results = results.AddRepositoryResults(repoResults)
	}

	// Generate report
	if err := generateReport(ctx, vctx, results); err != nil {
		return 1, err
	}

	// Return appropriate exit code
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// validateCommit validates a single commit and generates a report.
func validateCommit(ctx context.Context, vctx ValidationContext, ref string, skipMerge bool) (int, error) {
	// Get the commit
	commit, err := vctx.Validator.Deps.Repository.GetCommit(ctx, ref)
	if err != nil {
		return 1, fmt.Errorf("failed to get commit: %w", err)
	}

	// Skip if it's a merge commit and skipMerge is true
	if skipMerge && commit.IsMergeCommit {
		// Create empty results for skipped merge commit
		results := domain.NewValidationResults()
		emptyResult := domain.NewCommitResult(commit)
		results = results.AddResult(emptyResult)

		if err := generateReport(ctx, vctx, results); err != nil {
			return 1, err
		}

		return 0, nil
	}

	// Validate the commit
	validationResult := vctx.Validator.ValidateCommit(commit)

	// Convert to results format
	results := domain.NewValidationResults()
	commitResult := convertValidationResult(validationResult, vctx.Validator)
	results = results.AddResult(commitResult)

	// Add repository validation
	repoFailures := vctx.Validator.ValidateRepository()
	if len(repoFailures) > 0 {
		repoResults := convertRepoFailures(repoFailures)
		results = results.AddRepositoryResults(repoResults)
	}

	// Generate report
	if err := generateReport(ctx, vctx, results); err != nil {
		return 1, err
	}

	// Return appropriate exit code
	if results.AllPassed() {
		return 0, nil
	}

	return 2, nil
}

// generateReport generates a report using the provided formatter and options.
func generateReport(ctx context.Context, vctx ValidationContext, results domain.ValidationResults) error {
	// Convert domain.ReportOptions to format.Options
	formatOptions := convertReportOptions(vctx.Options)

	// Create report service
	reportService := format.NewReportService(formatOptions, vctx.Formatter, vctx.Logger)

	// Generate the report
	return reportService.GenerateReport(ctx, results)
}

// convertReportOptions converts domain report options to format options.
func convertReportOptions(options domain.ReportOptions) format.Options {
	// Convert format string to format.Format
	var formatType format.Format

	switch options.Format {
	case "json":
		formatType = format.FormatJSON
	case "github":
		formatType = format.FormatGitHubActions
	case "gitlab":
		formatType = format.FormatGitLabCI
	default:
		formatType = format.FormatText
	}

	return format.Options{
		Format:         formatType,
		Verbose:        options.Verbose,
		ShowHelp:       options.ShowHelp || options.ExtraVerbose,
		RuleToShowHelp: options.RuleToShowHelp,
		LightMode:      options.LightMode,
		Writer:         options.Writer,
	}
}

// convertValidationResult converts a domain ValidationResult to a domain CommitResult.
// This needs access to the validator to show both passing and failing rules.
func convertValidationResult(result domain.ValidationResult, validator domain.ValidatorWithDeps) domain.CommitResult {
	commitResult := domain.NewCommitResult(result.Commit)

	// Get all commit rules that were executed
	allRules := domain.FilterCommitRules(validator.Validator.Rules())

	// Create a map of failures for quick lookup
	failureMap := make(map[string]domain.RuleFailure)
	for _, failure := range result.Failures {
		failureMap[failure.Rule] = failure
	}

	// Add result for each rule that was executed
	for _, rule := range allRules {
		ruleName := rule.Name()
		if failure, exists := failureMap[ruleName]; exists {
			// Rule failed
			err := domain.New(failure.Rule, domain.ErrUnknown, failure.Message)
			commitResult = commitResult.AddRuleResult(ruleName, []domain.ValidationError{err})
		} else {
			// Rule passed
			commitResult = commitResult.AddRuleResult(ruleName, []domain.ValidationError{})
		}
	}

	return commitResult
}

// convertRepoFailures converts repository rule failures to domain RuleResults.
func convertRepoFailures(failures []domain.RuleFailure) []domain.RuleResult {
	if len(failures) == 0 {
		return nil
	}

	results := make([]domain.RuleResult, len(failures))

	for i, failure := range failures {
		err := domain.New(failure.Rule, domain.ErrUnknown, failure.Message)
		results[i] = domain.RuleResult{
			RuleName: failure.Rule,
			Status:   domain.StatusFailed,
			Errors:   []domain.ValidationError{err},
		}
	}

	return results
}

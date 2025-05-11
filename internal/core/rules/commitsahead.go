// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// CommitsAheadRule validates that the current branch is not too many commits ahead
// of the reference branch.
type CommitsAheadRule struct {
	baseRule         BaseRule
	maxCommitsAhead  int
	reference        string
	repositoryGetter func() domain.CommitAnalyzer
	commitsAhead     int
}

// CommitsAheadOption is a function that configures a CommitsAheadRule.
type CommitsAheadOption func(CommitsAheadRule) CommitsAheadRule

// WithMaxCommitsAhead sets the maximum number of commits allowed ahead of the reference branch.
func WithMaxCommitsAhead(maxCount int) CommitsAheadOption {
	return func(r CommitsAheadRule) CommitsAheadRule {
		result := r
		result.maxCommitsAhead = maxCount

		return result
	}
}

// WithReference sets the reference branch to compare against.
func WithReference(ref string) CommitsAheadOption {
	return func(r CommitsAheadRule) CommitsAheadRule {
		result := r
		result.reference = ref

		return result
	}
}

// WithRepositoryGetter sets the function used to get the repository.
func WithRepositoryGetter(getter func() domain.CommitAnalyzer) CommitsAheadOption {
	return func(r CommitsAheadRule) CommitsAheadRule {
		result := r
		result.repositoryGetter = getter

		return result
	}
}

// NewCommitsAheadRule creates a new rule for checking commits ahead of a reference branch.
func NewCommitsAheadRule(options ...CommitsAheadOption) CommitsAheadRule {
	// Create a rule with default values
	rule := CommitsAheadRule{
		baseRule:        NewBaseRule("CommitsAhead"),
		maxCommitsAhead: 10,     // Default: 10 commits ahead is the maximum
		reference:       "main", // Default reference branch
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate checks that the current branch is not too many commits ahead of the reference branch.
// This implementation uses context to get configuration values.
func (r CommitsAheadRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating commits ahead using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the existing validation logic
	errors, _ := ValidateCommitsAheadWithState(ctx, rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r CommitsAheadRule) withContextConfig(ctx context.Context) CommitsAheadRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Extract configuration values
	maxCommitsAhead := cfg.Repository.MaxCommitsAhead
	ref := cfg.Repository.ReferenceBranch

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Int("max_commits_ahead", maxCommitsAhead).
		Str("reference", ref).
		Msg("CommitsAhead rule configuration from context")

	// Create a copy of the rule
	result := r

	// Update with context configuration
	result.maxCommitsAhead = maxCommitsAhead
	result.reference = ref

	// Keep the repository getter from the original rule
	// (We don't replace it with contextual configuration)

	return result
}

// ValidateCommitsAheadWithState validates the commits ahead and returns both the errors and an updated rule.
func ValidateCommitsAheadWithState(ctx context.Context, rule CommitsAheadRule, commit domain.CommitInfo) ([]appErrors.ValidationError, CommitsAheadRule) {
	// Log validation at trace level
	log.Logger(ctx).Trace().
		Str("rule", rule.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating commits ahead")
	// Start with a clean rule
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Get the repository
	if rule.repositoryGetter == nil {
		// Without a repository getter, we can't validate
		return result.baseRule.Errors(), result
	}

	repository := rule.repositoryGetter()
	if repository == nil {
		// Without a repository, we can't validate
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrInvalidRepo,
			"Repository object is nil",
		)
		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// Get the number of commits ahead
	commitsAhead, err := repository.GetCommitsAhead(ctx, rule.reference)
	if err != nil {
		// Handle errors from the repository
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrGitOperationFailed,
			"Failed to get commits ahead of reference",
		).WithContext("reference", rule.reference).WithContext("error", err.Error())

		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// Update the rule with the number of commits ahead
	result.commitsAhead = commitsAhead

	// Validate against the maximum
	if rule.maxCommitsAhead > 0 && commitsAhead > rule.maxCommitsAhead {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrTooManyCommits,
			fmt.Sprintf("Branch is %d commits ahead of reference branch '%s'",
				commitsAhead, rule.reference),
		).
			WithContext("reference", rule.reference).
			WithContext("commits_ahead", strconv.Itoa(commitsAhead)).
			WithContext("max_allowed", strconv.Itoa(rule.maxCommitsAhead))

		result.baseRule = result.baseRule.WithError(validationErr)
	}

	// Update the rule with the number of commits ahead even if no error
	result.commitsAhead = commitsAhead

	return result.baseRule.Errors(), result
}

// Name returns the rule name.
func (r CommitsAheadRule) Name() string {
	return r.baseRule.Name()
}

// Errors returns all validation errors found by this rule.
func (r CommitsAheadRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r CommitsAheadRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r CommitsAheadRule) SetErrors(errors []appErrors.ValidationError) CommitsAheadRule {
	result := r
	result.baseRule = r.baseRule.WithClearedErrors()

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)
	}

	return result
}

// Result returns a concise validation result.
func (r CommitsAheadRule) Result(errors []appErrors.ValidationError) string {
	// Check for repository error first
	for _, err := range errors {
		if err.Code == string(appErrors.ErrInvalidRepo) {
			return "Repository object is nil"
		}
	}

	if len(errors) > 0 {
		return fmt.Sprintf("Branch is %d commits ahead of reference branch '%s'",
			r.commitsAhead, r.reference)
	}

	return fmt.Sprintf("Branch is %d commits ahead of reference branch '%s'",
		r.commitsAhead, r.reference)
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r CommitsAheadRule) VerboseResult(errors []appErrors.ValidationError) string {
	// Check for repository error first
	for _, err := range errors {
		if err.Code == string(appErrors.ErrInvalidRepo) {
			return "Repository object is nil - unable to analyze commits ahead"
		}
	}

	if len(errors) > 0 {
		return fmt.Sprintf("❌ Branch is %d commits ahead of reference branch '%s' (maximum allowed: %d)",
			r.commitsAhead, r.reference, r.maxCommitsAhead)
	}

	return fmt.Sprintf("✓ Branch is %d commits ahead of reference branch '%s' (within limit of %d)",
		r.commitsAhead, r.reference, r.maxCommitsAhead)
}

// Help returns guidance for fixing rule violations.
func (r CommitsAheadRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return `This rule compares your current branch with a reference branch (usually 'main').
It ensures your branch doesn't diverge too much from the reference, which can make integrating 
your changes more difficult. If your branch is too far ahead, consider merging or rebasing.`
	}

	// Check for specific error types
	for _, err := range errors {
		if err.Code == string(appErrors.ErrInvalidRepo) {
			return `Repository Access Error: Unable to access the Git repository for commit analysis.
Make sure you are in a valid Git repository and have appropriate permissions.`
		}
	}

	return fmt.Sprintf(`Commits Ahead Error: Your branch has too many commits compared to reference branch '%s'.
Your branch is %d commits ahead of reference branch '%s', which exceeds the maximum of %d.

Consider one of the following:
1. Merge '%s' into your branch to reduce the divergence
2. Rebase your branch onto '%s'
3. Adjust the maxCommitsAhead configuration if this limit is too restrictive`,
		r.reference, r.commitsAhead, r.reference, r.maxCommitsAhead, r.reference, r.reference)
}

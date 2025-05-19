// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// BranchAheadRule validates that the current branch is not too many commits ahead
// of the reference branch.
type BranchAheadRule struct {
	name             string
	maxCommitsAhead  int
	reference        string
	repositoryGetter func() domain.CommitAnalyzer
	commitsAhead     int
}

// BranchAheadOption is a function that configures a BranchAheadRule.
type BranchAheadOption func(BranchAheadRule) BranchAheadRule

// WithMaxCommitsAhead sets the maximum number of commits allowed ahead of the reference branch.
func WithMaxCommitsAhead(maxCount int) BranchAheadOption {
	return func(r BranchAheadRule) BranchAheadRule {
		result := r
		result.maxCommitsAhead = maxCount

		return result
	}
}

// WithReference sets the reference branch to compare against.
func WithReference(ref string) BranchAheadOption {
	return func(r BranchAheadRule) BranchAheadRule {
		result := r
		result.reference = ref

		return result
	}
}

// WithRepositoryGetter sets the function used to get the repository.
func WithRepositoryGetter(getter func() domain.CommitAnalyzer) BranchAheadOption {
	return func(r BranchAheadRule) BranchAheadRule {
		result := r
		result.repositoryGetter = getter

		return result
	}
}

// NewBranchAheadRule creates a new rule for checking commits ahead of a reference branch.
func NewBranchAheadRule(options ...BranchAheadOption) BranchAheadRule {
	// Create a rule with default values
	rule := BranchAheadRule{
		name:            "BranchAhead",
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
func (r BranchAheadRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating commits ahead using context configuration", "rule", r.Name(), "commit_hash", commit.Hash)

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Get the repository
	if rule.repositoryGetter == nil {
		// This should not happen in normal operation
		// Create a clear error when repository getter is unavailable
		return []appErrors.ValidationError{
			appErrors.New(
				"BranchAhead",
				appErrors.ErrInvalidRepo,
				"Repository analyzer unavailable - cannot verify commit distance",
			),
		}
	}

	repository := rule.repositoryGetter()
	if repository == nil {
		// This is another error that should not happen in normal operation
		// Without a repository, we can't validate
		return []appErrors.ValidationError{
			appErrors.New(
				"BranchAhead",
				appErrors.ErrInvalidRepo,
				"Repository analyzer unavailable - cannot verify commit distance",
			),
		}
	}

	// Get the number of commits ahead
	commitsAhead, err := repository.GetCommitsAhead(ctx, rule.reference)
	if err != nil {
		// Handle errors from the repository
		return []appErrors.ValidationError{
			appErrors.New(
				"BranchAhead",
				appErrors.ErrGitOperationFailed,
				"Failed to get commits ahead of reference",
			).WithContext("reference", rule.reference).WithContext("error", err.Error()),
		}
	}

	// Store the commits ahead for result formatting
	r.commitsAhead = commitsAhead

	// Validate against the maximum
	if rule.maxCommitsAhead > 0 && commitsAhead > rule.maxCommitsAhead {
		return []appErrors.ValidationError{
			appErrors.New(
				"BranchAhead",
				appErrors.ErrTooManyCommits,
				fmt.Sprintf("Branch is %d commits ahead of reference branch '%s'",
					commitsAhead, rule.reference),
			).
				WithContext("reference", rule.reference).
				WithContext("commits_ahead", strconv.Itoa(commitsAhead)).
				WithContext("max_allowed", strconv.Itoa(rule.maxCommitsAhead)),
		}
	}

	return nil
}

// withContextConfig creates a new rule with configuration from context.
func (r BranchAheadRule) withContextConfig(ctx context.Context) BranchAheadRule {
	// Get configuration directly from context
	cfg := contextx.GetConfig(ctx)

	// Extract configuration values
	maxCommitsAhead := cfg.GetInt("repository.max_commits_ahead")
	ref := cfg.GetString("repository.reference_branch")

	// Log configuration at debug level
	logger := contextx.GetLogger(ctx)
	logger.Debug("CommitsAhead rule configuration from context", "max_commits_ahead", maxCommitsAhead, "reference", ref)

	// Create a copy of the rule
	result := r

	// Update with context configuration
	result.maxCommitsAhead = maxCommitsAhead
	result.reference = ref

	// Keep the repository getter from the original rule
	// (We don't replace it with contextual configuration)

	return result
}

// Name returns the rule name.
func (r BranchAheadRule) Name() string {
	return r.name
}

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

// WithContext implements the ConfigurableRule interface for BranchAheadRule.
// It returns a new rule with configuration from the provided context.
func (r BranchAheadRule) WithContext(ctx context.Context) domain.Rule {
	// Get configuration directly from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return r
	}

	// Extract configuration values
	maxCommitsAhead := cfg.GetInt("repo.max_commits_ahead")
	ref := cfg.GetString("repo.branch")

	// Create a copy of the rule
	result := r

	// Update with context configuration
	result.maxCommitsAhead = maxCommitsAhead
	result.reference = ref

	// Keep the repository getter from the original rule
	// (We don't replace it with contextual configuration)

	return result
}

// Validate checks that the current branch is not too many commits ahead of the reference branch.
// This implementation uses the rule's pre-configured state set by WithContext.
func (r BranchAheadRule) Validate(ctx context.Context, _ domain.CommitInfo) []appErrors.ValidationError {
	// Get the repository
	if r.repositoryGetter == nil {
		// This should not happen in normal operation
		// Create a clear error when repository getter is unavailable
		return []appErrors.ValidationError{
			appErrors.NewBranchError(
				appErrors.ErrInvalidRepo,
				"BranchAhead",
				"Repository analyzer unavailable - cannot verify commit distance",
				"Ensure the repository is properly initialized",
			),
		}
	}

	repository := r.repositoryGetter()
	if repository == nil {
		// This is another error that should not happen in normal operation
		// Without a repository, we can't validate
		return []appErrors.ValidationError{
			appErrors.NewBranchError(
				appErrors.ErrInvalidRepo,
				"BranchAhead",
				"Repository analyzer unavailable - cannot verify commit distance",
				"Ensure the repository is properly initialized",
			),
		}
	}

	// Get the number of commits ahead
	commitsAhead, err := repository.GetCommitsAhead(ctx, r.reference)
	if err != nil {
		// Handle errors from the repository
		return []appErrors.ValidationError{
			appErrors.NewBranchError(
				appErrors.ErrGitOperationFailed,
				"BranchAhead",
				"Failed to get commits ahead of reference",
				fmt.Sprintf("Check if the reference branch '%s' exists", r.reference),
			).WithContextMap(map[string]string{
				"reference": r.reference,
				"error":     err.Error(),
			}),
		}
	}

	// The commitsAhead field is not used since the rule is stateless
	// Keeping it as a local variable instead

	// Validate against the maximum
	if r.maxCommitsAhead > 0 && commitsAhead > r.maxCommitsAhead {
		return []appErrors.ValidationError{
			appErrors.NewBranchError(
				appErrors.ErrTooManyCommits,
				"BranchAhead",
				fmt.Sprintf("Your branch is %d commits ahead of '%s' (max allowed: %d)",
					commitsAhead, r.reference, r.maxCommitsAhead),
				fmt.Sprintf("Rebase on %s or squash some commits to reduce the distance", r.reference),
			).WithContextMap(map[string]string{
				"reference":     r.reference,
				"commits_ahead": strconv.Itoa(commitsAhead),
				"max_allowed":   strconv.Itoa(r.maxCommitsAhead),
			}),
		}
	}

	return nil
}

// Name returns the rule name.
func (r BranchAheadRule) Name() string {
	return r.name
}

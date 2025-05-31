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
)

// BranchAheadRule validates that the current branch is not too many commits ahead
// of the reference branch.
type BranchAheadRule struct {
	name             string
	maxCommitsAhead  int
	reference        string
	repositoryGetter func() domain.Repository
}

// NewBranchAheadRule creates a new rule for checking commits ahead of a reference branch from config.
func NewBranchAheadRule(cfg config.Config, deps domain.RuleDependencies) BranchAheadRule {
	// Set defaults if not configured
	maxCommitsAhead := cfg.Repo.MaxCommitsAhead
	if maxCommitsAhead <= 0 {
		maxCommitsAhead = 10
	}

	reference := cfg.Repo.ReferenceBranch
	if reference == "" {
		reference = "main"
	}

	return BranchAheadRule{
		name:            "BranchAhead",
		maxCommitsAhead: maxCommitsAhead,
		reference:       reference,
		repositoryGetter: func() domain.Repository {
			return deps.Repository
		},
	}
}

// Validate checks that the current branch is not too many commits ahead of the reference branch.
// This implementation uses the rule's pre-configured state set by WithContext.
func (r BranchAheadRule) Validate(ctx context.Context, _ domain.CommitInfo) []domain.ValidationError {
	// Get the repository
	if r.repositoryGetter == nil {
		// This should not happen in normal operation
		// Create a clear error when repository getter is unavailable
		return []domain.ValidationError{
			domain.New(
				"BranchAhead",
				domain.ErrInvalidRepo,
				"Repository analyzer unavailable - cannot verify commit distance",
			).WithHelp("Ensure the repository is properly initialized"),
		}
	}

	repository := r.repositoryGetter()
	if repository == nil {
		// This is another error that should not happen in normal operation
		// Without a repository, we can't validate
		return []domain.ValidationError{
			domain.New(
				"BranchAhead",
				domain.ErrInvalidRepo,
				"Cannot access repository - validation impossible",
			).WithHelp("Check repository permissions and initialization state"),
		}
	}

	// Get the number of commits ahead
	commitsAhead, err := repository.GetCommitsAhead(ctx, r.reference)
	if err != nil {
		// Handle errors from the repository
		return []domain.ValidationError{
			domain.New(
				"BranchAhead",
				domain.ErrGitOperationFailed,
				"Failed to get commits ahead of reference",
			).WithHelp(fmt.Sprintf("Check if the reference branch '%s' exists", r.reference)).WithContextMap(map[string]string{
				"reference": r.reference,
				"error":     err.Error(),
			}),
		}
	}

	// The commitsAhead field is not used since the rule is stateless
	// Keeping it as a local variable instead

	// Validate against the maximum
	if r.maxCommitsAhead > 0 && commitsAhead > r.maxCommitsAhead {
		return []domain.ValidationError{
			domain.New(
				"BranchAhead",
				domain.ErrTooManyCommits,
				fmt.Sprintf("Your branch is %d commits ahead of '%s' (max allowed: %d)",
					commitsAhead, r.reference, r.maxCommitsAhead),
			).WithHelp(fmt.Sprintf("Rebase on %s or squash some commits to reduce the distance", r.reference)).WithContextMap(map[string]string{
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

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// BranchAheadRule validates that the current branch is not too many commits ahead
// of the reference branch.
type BranchAheadRule struct {
	maxCommitsAhead int
	reference       string
}

// NewBranchAheadRule creates a new rule for checking commits ahead of a reference branch from config.
func NewBranchAheadRule(cfg config.Config) BranchAheadRule {
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
		maxCommitsAhead: maxCommitsAhead,
		reference:       reference,
	}
}

// Validate checks that the current branch is not too many commits ahead of the reference branch.
// This rule requires repository access, so it checks if repository is available.
func (r BranchAheadRule) Validate(_ domain.Commit, repo domain.Repository, _ config.Config) []domain.ValidationError {
	// Skip if no repository is provided
	if repo == nil {
		return nil
	}

	// Get the number of commits ahead
	commitsAhead, err := repo.GetCommitsAheadCount(context.Background(), r.reference)
	if err != nil {
		// Skip on error rather than fail - repository might not have the reference branch
		return nil
	}

	// Validate against the maximum
	if r.maxCommitsAhead > 0 && commitsAhead > r.maxCommitsAhead {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrTooManyCommits,
				fmt.Sprintf("Branch is %d commits ahead of %s (max: %d)", commitsAhead, r.reference, r.maxCommitsAhead)).
				WithHelp(fmt.Sprintf("Consider rebasing or creating a pull request when ahead by more than %d commits", r.maxCommitsAhead)),
		}
	}

	return nil
}

// Name returns the rule name.
func (r BranchAheadRule) Name() string {
	return "BranchAhead"
}

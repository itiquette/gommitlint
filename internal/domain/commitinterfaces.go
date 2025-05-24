// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import "context"

// CommitRepository provides access to Git commit information.
// This interface combines repository access and analysis capabilities.
type CommitRepository interface {
	// GetCommit retrieves a commit by its hash.
	GetCommit(ctx context.Context, hash string) (CommitInfo, error)

	// GetCommits retrieves a list of commits.
	GetCommits(ctx context.Context, limit int) ([]CommitInfo, error)

	// GetCommitRange retrieves a range of commits.
	GetCommitRange(ctx context.Context, fromHash, toHash string) ([]CommitInfo, error)

	// GetHeadCommits returns the specified number of commits from HEAD.
	GetHeadCommits(ctx context.Context, count int) ([]CommitInfo, error)
}

// RepositoryInfoProvider provides general information about the repository.
type RepositoryInfoProvider interface {
	// GetCurrentBranch returns the name of the current branch.
	GetCurrentBranch(ctx context.Context) (string, error)

	// GetRepositoryName returns the name of the repository.
	GetRepositoryName(ctx context.Context) string

	// IsValid checks if the repository is a valid Git repository.
	IsValid(ctx context.Context) (bool, error)
}

// CommitAnalyzer provides commit analysis operations.
type CommitAnalyzer interface {
	// GetCommitsAhead returns the number of commits ahead of a reference.
	GetCommitsAhead(ctx context.Context, ref string) (int, error)
}

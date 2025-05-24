// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
)

// GitRepositoryAdapter combines CommitRepository, RepositoryInfoProvider, and CommitAnalyzer
// into a single repository implementation. It acts as an adapter that unifies multiple
// Git-related interfaces into a cohesive whole.
type GitRepositoryAdapter struct {
	commitService  domain.CommitRepository
	infoProvider   domain.RepositoryInfoProvider
	commitAnalyzer domain.CommitAnalyzer
}

// Ensure GitRepositoryAdapter implements all repository interfaces.
var (
	_ domain.CommitRepository       = GitRepositoryAdapter{}
	_ domain.RepositoryInfoProvider = GitRepositoryAdapter{}
	_ domain.CommitAnalyzer         = GitRepositoryAdapter{}
)

// GetCommit retrieves a commit by its hash.
func (r GitRepositoryAdapter) GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error) {
	return r.commitService.GetCommit(ctx, hash)
}

// GetCommitRange returns all commits in the given range.
func (r GitRepositoryAdapter) GetCommitRange(ctx context.Context, fromHash, toHash string) ([]domain.CommitInfo, error) {
	return r.commitService.GetCommitRange(ctx, fromHash, toHash)
}

// GetHeadCommits returns the specified number of commits from HEAD.
func (r GitRepositoryAdapter) GetHeadCommits(ctx context.Context, count int) ([]domain.CommitInfo, error) {
	return r.commitService.GetHeadCommits(ctx, count)
}

// GetCurrentBranch returns the name of the current branch.
func (r GitRepositoryAdapter) GetCurrentBranch(ctx context.Context) (string, error) {
	return r.infoProvider.GetCurrentBranch(ctx)
}

// GetRepositoryName returns the name of the repository.
func (r GitRepositoryAdapter) GetRepositoryName(ctx context.Context) string {
	return r.infoProvider.GetRepositoryName(ctx)
}

// IsValid checks if the repository is a valid Git repository.
func (r GitRepositoryAdapter) IsValid(ctx context.Context) (bool, error) {
	return r.infoProvider.IsValid(ctx)
}

// GetCommitsAhead returns the number of commits ahead of the given reference.
func (r GitRepositoryAdapter) GetCommitsAhead(ctx context.Context, reference string) (int, error) {
	return r.commitAnalyzer.GetCommitsAhead(ctx, reference)
}

// GetCommits retrieves a list of commits.
func (r GitRepositoryAdapter) GetCommits(ctx context.Context, limit int) ([]domain.CommitInfo, error) {
	return r.commitService.GetCommits(ctx, limit)
}

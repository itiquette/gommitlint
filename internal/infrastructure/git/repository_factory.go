// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// RepositoryFactory creates domain-specific repository interfaces.
type RepositoryFactory struct {
	adapter        *RepositoryAdapter
	commitReader   *CommitReaderAdapter
	historyReader  *HistoryReaderAdapter
	infoProvider   *InfoProviderAdapter
	commitAnalyzer *CommitAnalyzerAdapter
}

// NewRepositoryFactory creates a new repository factory for the given path.
func NewRepositoryFactory(path string) (*RepositoryFactory, error) {
	// Create the base repository adapter
	adapter, err := NewRepositoryAdapter(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	// Create specialized adapters
	commitReader := &CommitReaderAdapter{adapter: adapter}
	historyReader := &HistoryReaderAdapter{adapter: adapter}
	infoProvider := &InfoProviderAdapter{adapter: adapter}
	commitAnalyzer := &CommitAnalyzerAdapter{adapter: adapter}

	return &RepositoryFactory{
		adapter:        adapter,
		commitReader:   commitReader,
		historyReader:  historyReader,
		infoProvider:   infoProvider,
		commitAnalyzer: commitAnalyzer,
	}, nil
}

// CreateCommitReader creates a CommitReader interface.
func (f *RepositoryFactory) CreateCommitReader() domain.CommitReader {
	return f.commitReader
}

// CreateHistoryReader creates a CommitHistoryReader interface.
func (f *RepositoryFactory) CreateHistoryReader() domain.CommitHistoryReader {
	return f.historyReader
}

// CreateInfoProvider creates a RepositoryInfoProvider interface.
func (f *RepositoryFactory) CreateInfoProvider() domain.RepositoryInfoProvider {
	return f.infoProvider
}

// CreateCommitAnalyzer creates a CommitAnalyzer interface.
func (f *RepositoryFactory) CreateCommitAnalyzer() domain.CommitAnalyzer {
	return f.commitAnalyzer
}

// CreateFullService creates a complete GitRepositoryService.
func (f *RepositoryFactory) CreateFullService() domain.GitRepositoryService {
	return f.adapter
}

// CommitReaderAdapter adapts the repository to the CommitReader interface.
type CommitReaderAdapter struct {
	adapter *RepositoryAdapter
}

// GetCommit retrieves a commit by its hash.
func (a *CommitReaderAdapter) GetCommit(ctx context.Context, hash string) (*domain.CommitInfo, error) {
	return a.adapter.GetCommit(ctx, hash)
}

// HistoryReaderAdapter adapts the repository to the CommitHistoryReader interface.
type HistoryReaderAdapter struct {
	adapter *RepositoryAdapter
}

// GetHeadCommits retrieves the specified number of commits from HEAD.
func (a *HistoryReaderAdapter) GetHeadCommits(ctx context.Context, count int) ([]*domain.CommitInfo, error) {
	return a.adapter.GetHeadCommits(ctx, count)
}

// GetCommitRange retrieves all commits in the given range.
func (a *HistoryReaderAdapter) GetCommitRange(ctx context.Context, fromHash, toHash string) ([]*domain.CommitInfo, error) {
	return a.adapter.GetCommitRange(ctx, fromHash, toHash)
}

// InfoProviderAdapter adapts the repository to the RepositoryInfoProvider interface.
type InfoProviderAdapter struct {
	adapter *RepositoryAdapter
}

// GetCurrentBranch returns the name of the current branch.
func (a *InfoProviderAdapter) GetCurrentBranch(ctx context.Context) (string, error) {
	return a.adapter.GetCurrentBranch(ctx)
}

// GetRepositoryName returns the name of the repository.
func (a *InfoProviderAdapter) GetRepositoryName(ctx context.Context) string {
	return a.adapter.GetRepositoryName(ctx)
}

// IsValid checks if the repository is valid.
func (a *InfoProviderAdapter) IsValid(ctx context.Context) bool {
	return a.adapter.IsValid(ctx)
}

// CommitAnalyzerAdapter adapts the repository to the CommitAnalyzer interface.
type CommitAnalyzerAdapter struct {
	adapter *RepositoryAdapter
}

// GetCommitsAhead returns the number of commits ahead of the given reference.
func (a *CommitAnalyzerAdapter) GetCommitsAhead(ctx context.Context, reference string) (int, error) {
	return a.adapter.GetCommitsAhead(ctx, reference)
}

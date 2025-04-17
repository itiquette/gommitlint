// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
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
func (a *CommitReaderAdapter) GetCommit(hash string) (*domain.CommitInfo, error) {
	return a.adapter.GetCommit(hash)
}

// HistoryReaderAdapter adapts the repository to the CommitHistoryReader interface.
type HistoryReaderAdapter struct {
	adapter *RepositoryAdapter
}

// GetHeadCommits retrieves the specified number of commits from HEAD.
func (a *HistoryReaderAdapter) GetHeadCommits(count int) ([]*domain.CommitInfo, error) {
	return a.adapter.GetHeadCommits(count)
}

// GetCommitRange retrieves all commits in the given range.
func (a *HistoryReaderAdapter) GetCommitRange(fromHash, toHash string) ([]*domain.CommitInfo, error) {
	return a.adapter.GetCommitRange(fromHash, toHash)
}

// InfoProviderAdapter adapts the repository to the RepositoryInfoProvider interface.
type InfoProviderAdapter struct {
	adapter *RepositoryAdapter
}

// GetCurrentBranch returns the name of the current branch.
func (a *InfoProviderAdapter) GetCurrentBranch() (string, error) {
	return a.adapter.GetCurrentBranch()
}

// GetRepositoryName returns the name of the repository.
func (a *InfoProviderAdapter) GetRepositoryName() string {
	return a.adapter.GetRepositoryName()
}

// IsValid checks if the repository is valid.
func (a *InfoProviderAdapter) IsValid() bool {
	return a.adapter.IsValid()
}

// CommitAnalyzerAdapter adapts the repository to the CommitAnalyzer interface.
type CommitAnalyzerAdapter struct {
	adapter *RepositoryAdapter
}

// GetCommitsAhead returns the number of commits ahead of the given reference.
func (a *CommitAnalyzerAdapter) GetCommitsAhead(reference string) (int, error) {
	return a.adapter.GetCommitsAhead(reference)
}

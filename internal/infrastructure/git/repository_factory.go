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
	adapter *RepositoryAdapter
}

// NewRepositoryFactory creates a new repository factory for the given path.
func NewRepositoryFactory(path string) (*RepositoryFactory, error) {
	// Create the base repository adapter
	adapter, err := NewRepositoryAdapter(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	return &RepositoryFactory{
		adapter: adapter,
	}, nil
}

// CreateGitCommitService creates a unified GitCommitService interface.
func (f *RepositoryFactory) CreateGitCommitService() domain.GitCommitService {
	return f.adapter
}

// CreateInfoProvider creates a RepositoryInfoProvider interface.
func (f *RepositoryFactory) CreateInfoProvider() domain.RepositoryInfoProvider {
	return f.adapter
}

// CreateCommitAnalyzer creates a CommitAnalyzer interface.
func (f *RepositoryFactory) CreateCommitAnalyzer() domain.CommitAnalyzer {
	return f.adapter
}

// CreateFullService creates a complete GitRepositoryService.
func (f *RepositoryFactory) CreateFullService() domain.GitRepositoryService {
	return f.adapter
}

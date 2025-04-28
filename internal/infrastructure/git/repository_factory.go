// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// RepositoryFactory creates and manages Git repository services.
// It implements the domain.RepositoryFactory interface.
// This version uses value semantics throughout.
type RepositoryFactory struct {
	repoPath string
	adapter  RepositoryAdapter
}

// Ensure RepositoryFactory implements domain.RepositoryFactory.
var _ domain.RepositoryFactory = RepositoryFactory{}

// NewRepositoryFactory creates a new repository factory for the given path.
func NewRepositoryFactory(path string) (RepositoryFactory, error) {
	adapter, err := NewRepositoryAdapter(path)
	if err != nil {
		return RepositoryFactory{}, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	return RepositoryFactory{
		repoPath: path,
		adapter:  adapter,
	}, nil
}

// CreateGitCommitService returns an implementation of domain.GitCommitService.
func (f RepositoryFactory) CreateGitCommitService() domain.GitCommitService {
	return f.adapter
}

// CreateInfoProvider returns an implementation of domain.RepositoryInfoProvider.
func (f RepositoryFactory) CreateInfoProvider() domain.RepositoryInfoProvider {
	return f.adapter
}

// CreateCommitAnalyzer returns an implementation of domain.CommitAnalyzer.
func (f RepositoryFactory) CreateCommitAnalyzer() domain.CommitAnalyzer {
	return f.adapter
}

// CreateFullService returns an implementation of domain.GitRepositoryService,
// which combines all Git-related interfaces.
func (f RepositoryFactory) CreateFullService() domain.GitRepositoryService {
	return f.adapter
}

// NewRepositoryServices creates all repository-related services in a single call.
// This simplifies dependency injection by providing all necessary components at once.
// This is the preferred method for getting repository services.
// This version uses value semantics throughout.
func NewRepositoryServices(path string) (domain.GitCommitService, domain.RepositoryInfoProvider, domain.CommitAnalyzer, error) {
	adapter, err := NewRepositoryAdapter(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	return adapter, adapter, adapter, nil
}

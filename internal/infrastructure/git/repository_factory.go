// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// RepositoryAdapter is used directly in most cases with the NewRepositoryAdapter function.
// RepositoryFactory is being kept for backwards compatibility but will be deprecated in the future.

// RepositoryFactory creates domain-specific repository interfaces.
// It implements the domain.RepositoryFactory interface to provide
// access to Git-specific repository services.
type RepositoryFactory struct {
	adapter *RepositoryAdapter
}

// Ensure RepositoryFactory implements domain.RepositoryFactory.
var _ domain.RepositoryFactory = (*RepositoryFactory)(nil)

// NewRepositoryFactory creates a new repository factory for the given path.
// DEPRECATED: Use NewRepositoryAdapter or NewRepositoryServices directly instead.
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
// DEPRECATED: Use the adapter directly as it already implements this interface.
func (f *RepositoryFactory) CreateGitCommitService() domain.GitCommitService {
	return f.adapter
}

// CreateInfoProvider creates a RepositoryInfoProvider interface.
// DEPRECATED: Use the adapter directly as it already implements this interface.
func (f *RepositoryFactory) CreateInfoProvider() domain.RepositoryInfoProvider {
	return f.adapter
}

// CreateCommitAnalyzer creates a CommitAnalyzer interface.
// DEPRECATED: Use the adapter directly as it already implements this interface.
func (f *RepositoryFactory) CreateCommitAnalyzer() domain.CommitAnalyzer {
	return f.adapter
}

// CreateFullService creates a complete GitRepositoryService.
// DEPRECATED: Use the adapter directly as it already implements this interface.
func (f *RepositoryFactory) CreateFullService() domain.GitRepositoryService {
	return f.adapter
}

// NewRepositoryServices creates all repository-related services in a single call.
// This simplifies dependency injection by providing all necessary components at once.
// This is the preferred method for getting repository services.
func NewRepositoryServices(path string) (domain.GitCommitService, domain.RepositoryInfoProvider, domain.CommitAnalyzer, error) {
	adapter, err := NewRepositoryAdapter(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	return adapter, adapter, adapter, nil
}

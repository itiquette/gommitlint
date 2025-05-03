// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides Git repository adapters for the domain model.
package git

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
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
func NewRepositoryFactory(ctx context.Context, path string) (RepositoryFactory, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("path", path).Msg("Entering NewRepositoryFactory")

	adapter, err := NewRepositoryAdapter(ctx, path)
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
func NewRepositoryServices(ctx context.Context, path string) (domain.GitCommitService, domain.RepositoryInfoProvider, domain.CommitAnalyzer, error) {
	logger := log.Logger(ctx)
	logger.Trace().Str("path", path).Msg("Entering NewRepositoryServices")

	adapter, err := NewRepositoryAdapter(ctx, path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	return adapter, adapter, adapter, nil
}

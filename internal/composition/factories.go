// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/git"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/output"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// OutgoingAdapterFactory creates outgoing adapters.
type OutgoingAdapterFactory struct {
	config types.Config
	logger outgoing.Logger
}

// NewOutgoingAdapterFactory creates a new factory for outgoing adapters.
func NewOutgoingAdapterFactory(config types.Config, logger outgoing.Logger) *OutgoingAdapterFactory {
	return &OutgoingAdapterFactory{
		config: config,
		logger: logger,
	}
}

// CreateGitRepository creates a Git repository adapter.
func (f OutgoingAdapterFactory) CreateGitRepository(ctx context.Context, repoPath string) (domain.CommitRepository, error) {
	// Create Git repository factory
	gitRepoFactory, err := git.NewRepositoryFactory(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository factory: %w", err)
	}

	// Create service adapters
	gitCommitService := gitRepoFactory.CreateCommitRepository()
	repoInfoProvider := gitRepoFactory.CreateRepositoryInfoProvider()
	commitAnalyzer := gitRepoFactory.CreateCommitAnalyzer()

	// Create composite repository
	compositeRepo := &GitRepositoryAdapter{
		commitService:  gitCommitService,
		infoProvider:   repoInfoProvider,
		commitAnalyzer: commitAnalyzer,
	}

	return compositeRepo, nil
}

// CreateConfigAdapter creates a configuration adapter.
func (f OutgoingAdapterFactory) CreateConfigAdapter() *config.Adapter {
	return config.NewAdapter(f.config)
}

// CreateOutputFormatters creates all output formatters.
func (f OutgoingAdapterFactory) CreateOutputFormatters() map[string]outgoing.ResultFormatter {
	return map[string]outgoing.ResultFormatter{
		"json":   output.NewJSONFormatter(),
		"text":   output.NewTextFormatter(),
		"github": output.NewGitHubFormatter(),
		"gitlab": output.NewGitLabFormatter(),
	}
}

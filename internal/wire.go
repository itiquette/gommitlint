// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package internal provides simple factory functions for the functional validation system.
package internal

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/adapters/git"
	"github.com/itiquette/gommitlint/internal/application"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// CreateRepository creates a git repository instance.
// This is a simple factory function with no hidden dependencies.
func CreateRepository(ctx context.Context, repoPath string) (domain.Repository, error) {
	gitRepo, err := git.NewRepository(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create git repository: %w", err)
	}

	return gitRepo, nil
}

// CreateValidator creates a validator with all dependencies.
// This is the main factory function that wires everything together.
func CreateValidator(ctx context.Context, cfg *config.Config, repoPath string, logger domain.Logger) (domain.ValidatorWithDeps, error) {
	repo, err := CreateRepository(ctx, repoPath)
	if err != nil {
		return domain.ValidatorWithDeps{}, fmt.Errorf("failed to create repository: %w", err)
	}

	return application.CreateValidator(repo, cfg, logger), nil
}

// CreateRules creates all enabled rules based on configuration.
// This is a pure function that returns a new slice of rules.
func CreateRules(cfg *config.Config) []domain.Rule {
	return application.CreateEnabledRules(cfg)
}

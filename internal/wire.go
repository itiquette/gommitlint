// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package internal provides simple factory functions for dependency injection.
// This follows functional programming principles with explicit dependencies.
package internal

import (
	"context"
	"fmt"

	cryptofactory "github.com/itiquette/gommitlint/internal/adapters/signing"

	"github.com/itiquette/gommitlint/internal/adapters/git"
	"github.com/itiquette/gommitlint/internal/adapters/logging"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/rules"
)

// NewValidationService creates a validation service with direct construction.
// Simplified dependency injection - no complex factory patterns.
func NewValidationService(ctx context.Context, config config.Config, repoPath string, logger log.Logger) (*domain.Service, error) {
	// Create git repository - single responsibility, direct construction
	gitRepo, err := git.NewRepository(ctx, repoPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create git repository: %w", err)
	}

	// Create enabled rules with minimal dependencies
	enabledRules := createEnabledRules(config, gitRepo)

	// Create service with direct repository and rules - clean and simple
	service := domain.NewService(gitRepo, enabledRules)

	return &service, nil
}

// createEnabledRules creates enabled rules with simplified dependencies.
// Direct construction - crypto services created only when needed.
func createEnabledRules(config config.Config, gitRepo domain.Repository) []domain.Rule {
	// Create minimal rule dependencies
	var ruleDeps domain.RuleDependencies

	// Always provide git repository for branchahead rule
	ruleDeps.Repository = gitRepo

	// Only create crypto services if actually needed
	if domain.ShouldRunRule("identity", config.Rules.Enabled, config.Rules.Disabled) {
		// Simple, direct crypto construction
		cryptoRepo := cryptofactory.NewFileSystemKeyRepository(config.Signing.KeyDirectory)
		gpgVerifier := cryptofactory.NewDefaultGPGVerificationService()
		sshVerifier := cryptofactory.NewDefaultSSHVerificationService()
		verificationSvc := cryptofactory.NewVerificationService(gpgVerifier, sshVerifier)
		cryptoVerifier := cryptofactory.NewVerificationAdapterWithOptions(
			cryptofactory.WithVerificationService(verificationSvc),
			cryptofactory.WithKeyRepository(cryptoRepo),
		)

		ruleDeps.CryptoVerifier = cryptoVerifier
		ruleDeps.CryptoRepository = cryptoRepo
	}

	// Create rules with direct construction
	return rules.CreateEnabledRules(&config, ruleDeps)
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"errors"
	"fmt"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/application/orchestration"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	configTypes "github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/factories"
	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// ValidationServiceAdapter adapts the core validation service to the incoming port interface.
type ValidationServiceAdapter struct {
	service validation.Service
}

// ValidateCommit implements incoming.ValidationService.
func (a ValidationServiceAdapter) ValidateCommit(ctx context.Context, ref string, _ bool) (domain.CommitResult, error) {
	return a.service.ValidateCommit(ctx, ref)
}

// ValidateCommits implements incoming.ValidationService.
func (a ValidationServiceAdapter) ValidateCommits(ctx context.Context, commitHashes []string, _ bool) (domain.ValidationResults, error) {
	// For now, validate commits one by one - this could be optimized in the future
	results := domain.NewValidationResults()

	for _, hash := range commitHashes {
		result, err := a.service.ValidateCommit(ctx, hash)
		if err != nil {
			return results, err
		}

		results = results.WithResult(result)
	}

	return results, nil
}

// ValidateCommitRange implements incoming.ValidationService.
func (a ValidationServiceAdapter) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMergeCommits bool) (domain.ValidationResults, error) {
	return a.service.ValidateCommitRange(ctx, fromHash, toHash, skipMergeCommits)
}

// ValidateMessage implements incoming.ValidationService.
func (a ValidationServiceAdapter) ValidateMessage(ctx context.Context, message string) (domain.ValidationResults, error) {
	return a.service.ValidateMessage(ctx, message)
}

// Container represents the dependency injection container for the application.
// It follows functional programming principles with immutable state.
// This is the composition root that wires up all application dependencies.
type Container struct {
	logger       outgoing.Logger
	actualConfig configTypes.Config
	ruleRegistry *domain.RuleRegistry
	factory      *OutgoingAdapterFactory
}

// NewContainer creates a new dependency injection container with pre-initialized dependencies.
// This follows functional programming by initializing all state upfront.
func NewContainer(logger outgoing.Logger, config configTypes.Config) Container {
	// Initialize rule registry upfront
	ruleRegistry := domain.NewRuleRegistry()
	factory := NewOutgoingAdapterFactory(config, logger)

	return Container{
		logger:       logger,
		actualConfig: config,
		ruleRegistry: ruleRegistry,
		factory:      factory,
	}
}

// CreateValidationService creates a validation service for the given repository path.
// This is a pure function that doesn't mutate the receiver.
func (c Container) CreateValidationService(ctx context.Context, repoPath string) (incoming.ValidationService, error) {
	// Add the config to the context using the standard pattern via adapter
	ctx = contextx.WithConfig(ctx, config.NewAdapter(c.actualConfig))

	// Create git repository adapter
	gitRepo, err := c.factory.CreateGitRepository(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create git repository: %w", err)
	}

	// Cast to required interfaces
	analyzer, isAnalyzer := gitRepo.(domain.CommitAnalyzer)
	if !isAnalyzer {
		return nil, errors.New("git repository does not implement CommitAnalyzer interface")
	}

	infoProvider, isInfoProvider := gitRepo.(domain.RepositoryInfoProvider)
	if !isInfoProvider {
		return nil, errors.New("git repository does not implement RepositoryInfoProvider interface")
	}

	// Initialize rule registry with rules if not already done
	if !c.ruleRegistry.HasRules() {
		// Create rule factory with priority service from registry
		ruleFactory := factories.NewSimpleRuleFactory().
			WithConfig(&c.actualConfig).
			WithPriorityService(c.ruleRegistry.GetPriorityService())

		// Create crypto dependencies if signing is configured
		if c.actualConfig.Signing.KeyDirectory != "" {
			keyRepo := crypto.NewFileSystemKeyRepository(c.actualConfig.Signing.KeyDirectory)
			verifier := crypto.NewVerificationAdapterWithOptions(
				crypto.WithKeyRepository(keyRepo),
			)

			ruleFactory = ruleFactory.
				WithCryptoVerifier(verifier).
				WithCryptoRepository(keyRepo)
		}

		// Register basic rules
		basicRules := ruleFactory.CreateBasicRules()
		for name, ruleFunc := range basicRules {
			c.ruleRegistry.RegisterWithContext(ctx, name, ruleFunc)
		}

		// Register analyzer rules if analyzer is available
		if analyzer != nil {
			analyzerRules := ruleFactory.CreateAnalyzerRules(analyzer)
			for name, ruleFunc := range analyzerRules {
				c.ruleRegistry.RegisterWithContext(ctx, name, ruleFunc)
			}
		}

		// Initialize rules with context that has configuration
		c.ruleRegistry.InitializeRules(ctx)
	}

	// Create validation config from actual config
	validationConfig := validation.Config{
		EnabledRules:  c.actualConfig.Rules.Enabled,
		DisabledRules: c.actualConfig.Rules.Disabled,
	}

	// Create validation engine with injected dependencies
	engine := validation.CreateEngine(validationConfig, analyzer, c.ruleRegistry)

	// Create validation service dependencies
	deps := validation.ServiceDependencies{
		Engine:        engine,
		CommitService: gitRepo,
		InfoProvider:  infoProvider,
		Analyzer:      analyzer,
	}

	// Create and return service with proper dependency injection
	service := validation.NewService(deps, c.actualConfig)

	return ValidationServiceAdapter{service: service}, nil
}

// Getters for dependencies.

// GetCreateValidationService returns a function that creates validation services.
// This allows the CLI to create services for different repository paths.
func (c Container) GetCreateValidationService() func(context.Context, string) (incoming.ValidationService, error) {
	return c.CreateValidationService
}

// CreateValidationOrchestrator creates a validation orchestrator for the given repository path.
// This orchestrates validation and report generation.
func (c Container) CreateValidationOrchestrator(ctx context.Context, repoPath string, formatter outgoing.ResultFormatter) (incoming.ValidationOrchestrator, error) {
	// Create validation service
	validationService, err := c.CreateValidationService(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation service: %w", err)
	}

	// Create and return orchestrator
	return orchestration.NewValidationOrchestrator(validationService, formatter, c.logger), nil
}

// GetCreateValidationOrchestrator returns a function that creates validation orchestrators.
func (c Container) GetCreateValidationOrchestrator() func(context.Context, string, outgoing.ResultFormatter) (incoming.ValidationOrchestrator, error) {
	return c.CreateValidationOrchestrator
}

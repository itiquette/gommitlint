// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/application/orchestration"
	configTypes "github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/factories"
	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// Container represents the dependency injection container for the application.
// It follows functional programming principles with controlled state management.
// This is the composition root that wires up all application dependencies.
type Container struct {
	logger       outgoing.Logger
	actualConfig configTypes.Config
	factory      *OutgoingAdapterFactory

	// Registry state is managed carefully in the composition root
	registryMu          sync.RWMutex
	ruleRegistry        domain.RuleRegistry
	registryInitialized bool
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

// initializeRegistry initializes the rule registry if not already done.
// This is called once per container instance and is thread-safe.
func (c *Container) initializeRegistry(ctx context.Context, analyzer domain.CommitAnalyzer) domain.RuleRegistry {
	c.registryMu.Lock()
	defer c.registryMu.Unlock()

	// Double-check under lock
	if c.registryInitialized {
		return c.ruleRegistry
	}

	registry := c.ruleRegistry

	// Create rule factory with priority service from registry
	ruleFactory := factories.NewSimpleRuleFactory().
		WithConfig(&c.actualConfig).
		WithPriorityService(registry.GetPriorityService())

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
		registry = registry.Register(name, ruleFunc)
	}

	// Register analyzer rules if analyzer is available
	if analyzer != nil {
		analyzerRules := ruleFactory.CreateAnalyzerRules(analyzer)
		for name, ruleFunc := range analyzerRules {
			registry = registry.Register(name, ruleFunc)
		}
	}

	// Initialize rules with context that has configuration
	registry = registry.WithInitializedRules(ctx)

	// Update container state
	c.ruleRegistry = registry
	c.registryInitialized = true

	return registry
}

// CreateValidationService creates a validation service for the given repository path.
// This is mostly a pure function, with controlled registry initialization.
func (c *Container) CreateValidationService(ctx context.Context, repoPath string) (incoming.ValidationService, error) {
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

	// Get initialized registry (thread-safe)
	registry := c.initializeRegistry(ctx, analyzer)

	// Create validation config from actual config
	validationConfig := validation.Config{
		EnabledRules:  c.actualConfig.Rules.Enabled,
		DisabledRules: c.actualConfig.Rules.Disabled,
	}

	// Create validation engine with injected dependencies
	engine := validation.CreateEngine(validationConfig, analyzer, registry)

	// Create validation service dependencies
	deps := validation.ServiceDependencies{
		Engine:        engine,
		CommitService: gitRepo,
		InfoProvider:  infoProvider,
		Analyzer:      analyzer,
	}

	// Create and return service with proper dependency injection
	service := validation.NewService(deps, c.actualConfig)

	return service, nil
}

// Getters for dependencies.

// GetCreateValidationService returns a function that creates validation services.
// This allows the CLI to create services for different repository paths.
func (c *Container) GetCreateValidationService() func(context.Context, string) (incoming.ValidationService, error) {
	return c.CreateValidationService
}

// CreateValidationOrchestrator creates a validation orchestrator for the given repository path.
// This orchestrates validation and report generation.
func (c *Container) CreateValidationOrchestrator(ctx context.Context, repoPath string, formatter outgoing.ResultFormatter) (incoming.ValidationOrchestrator, error) {
	// Create validation service
	validationService, err := c.CreateValidationService(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation service: %w", err)
	}

	// Create and return orchestrator
	return orchestration.NewValidationOrchestrator(validationService, formatter, c.logger), nil
}

// GetCreateValidationOrchestrator returns a function that creates validation orchestrators.
func (c *Container) GetCreateValidationOrchestrator() func(context.Context, string, outgoing.ResultFormatter) (incoming.ValidationOrchestrator, error) {
	return c.CreateValidationOrchestrator
}

// GetLogger returns the logger instance for dependency injection.
// This follows functional programming by exposing the logger as a pure getter.
func (c *Container) GetLogger() outgoing.Logger {
	return c.logger
}

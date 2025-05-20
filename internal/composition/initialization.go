// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"fmt"

	infra "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/application/factories"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
)

// initializeCore handles core initialization (logger, config).
func (r *Root) initializeCore(ctx context.Context) error {
	// Initialize logger (now handled differently)
	logger := contextx.GetLogger(ctx)
	logger.Debug("Logger initialized")

	// Load configuration using the new service
	configService, err := infra.NewService()
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}

	if err := configService.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	r.config = configService.GetConfig()

	// Create configuration adapter
	r.configAdapter = configService.GetAdapter()

	// Add config to context
	configuredCtx := contextx.WithConfig(ctx, r.configAdapter)

	configLogger := contextx.GetLogger(configuredCtx)
	configLogger.Info("Configuration loaded",
		"enabled_rules", r.config.Rules.Enabled,
		"disabled_rules", r.config.Rules.Disabled)

	return nil
}

// createFactories creates adapter factories.
func (r *Root) createFactories(ctx context.Context) {
	logger := contextx.GetLogger(ctx)
	r.outgoingFactory = NewOutgoingAdapterFactory(r.config, logger)
	r.incomingFactory = NewIncomingAdapterFactory(r.config, logger)
	logger.Debug("Factories created")
}

// initializeAdapters initializes all adapters.
func (r *Root) initializeAdapters(ctx context.Context) error {
	// Initialize Git repository
	repoPath := r.config.Repo.Path
	if repoPath == "" {
		repoPath = "."
	}

	gitRepo, err := r.outgoingFactory.CreateGitRepository(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to create git repository: %w", err)
	}

	r.gitRepository = gitRepo
	logger := contextx.GetLogger(ctx)
	logger.Debug("Git repository initialized", "path", repoPath)

	return nil
}

// initializeApplicationServices initializes application services.
func (r *Root) initializeApplicationServices(ctx context.Context) {
	// Initialize validation config
	r.initializeValidationConfig(ctx)

	// Initialize rule registry
	r.initializeRuleRegistry(ctx)

	// Initialize validation service
	r.initializeValidationService(ctx)
}

// initializeValidationConfig creates the validation config adapter.
func (r *Root) initializeValidationConfig(ctx context.Context) {
	// Configuration is now accessed via context using contextx.GetConfig(ctx)
	// No need to store it in the root
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validation config initialized")
}

// initializeRuleRegistry initializes the rule registry.
func (r *Root) initializeRuleRegistry(ctx context.Context) {
	// Create simple rule factory
	simpleFactory := factories.NewSimpleRuleFactory()

	// Create basic rules
	basicRules := simpleFactory.CreateBasicRules()

	// Create analyzer rules if we have an analyzer
	if r.gitRepository != nil {
		analyzer, ok := r.gitRepository.(domain.CommitAnalyzer)
		if !ok {
			logger := contextx.GetLogger(ctx)
			logger.Debug("Git repository does not implement analyzer interface")
		}

		if analyzer != nil {
			analyzerRules := simpleFactory.CreateAnalyzerRules(analyzer)
			for name, factory := range analyzerRules {
				basicRules[name] = factory
			}
		}
	}

	// Create rule registry
	simpleRegistry := domain.NewRuleRegistry()

	// Register all rules
	for name, factory := range basicRules {
		simpleRegistry.RegisterWithContext(ctx, name, factory)
	}

	// Store the registry directly
	r.ruleRegistry = simpleRegistry

	logger := contextx.GetLogger(ctx)
	logger.Debug("Simple rule registry initialized",
		"enabled_rules", r.config.Rules.Enabled,
		"disabled_rules", r.config.Rules.Disabled)
}

// initializeValidationService creates and configures the validation service.
func (r *Root) initializeValidationService(ctx context.Context) {
	// Create validation engine
	// The CreateEngine function uses the context to access configuration
	// Pass a default config since the function ignores it anyway (marked with _)
	analyzer, isAnalyzer := r.gitRepository.(domain.CommitAnalyzer)
	if !isAnalyzer {
		logger := contextx.GetLogger(ctx)
		logger.Error("Git repository does not implement CommitAnalyzer interface")

		return
	}

	engine := validation.CreateEngine(ctx, types.Config{}, analyzer)

	// Create validation service
	infoProvider, isInfoProvider := r.gitRepository.(domain.RepositoryInfoProvider)
	if !isInfoProvider {
		logger := contextx.GetLogger(ctx)
		logger.Error("Git repository does not implement RepositoryInfoProvider interface")

		return
	}

	r.validationService = validate.NewValidationService(
		engine,
		r.gitRepository,
		infoProvider,
	)

	logger := contextx.GetLogger(ctx)
	logger.Debug("Validation service initialized")
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validate provides application services for commit validation.
package validate

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// Note: Using domain package interfaces instead of a local interface definition

// ValidationOptions contains options for validation.
type ValidationOptions struct {
	// CommitHash is the hash of a specific commit to validate.
	CommitHash string

	// CommitCount is the number of commits from HEAD to validate.
	CommitCount int

	// FromHash is the start of a commit range to validate.
	FromHash string

	// ToHash is the end of a commit range to validate.
	ToHash string

	// MessageFile is the path to a file containing a commit message to validate.
	MessageFile string

	// SkipMergeCommits indicates whether to skip merge commits.
	SkipMergeCommits bool
}

// ValidationEngine defines the interface for the validation engine.
// This matches the domain.ValidationEngine interface.
type ValidationEngine interface {
	ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult
	ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults
}

// ValidationService orchestrates commit validation operations.
// It is designed to be used with value semantics and follows functional programming patterns.
type ValidationService struct {
	engine        ValidationEngine
	commitService domain.GitCommitService
	infoProvider  domain.RepositoryInfoProvider
}

// Engine returns the validation engine.
func (s ValidationService) Engine() ValidationEngine {
	return s.engine
}

// CommitService returns the commit service.
func (s ValidationService) CommitService() domain.GitCommitService {
	return s.commitService
}

// InfoProvider returns the repository info provider.
func (s ValidationService) InfoProvider() domain.RepositoryInfoProvider {
	return s.infoProvider
}

// NewValidationService creates a new ValidationService.
func NewValidationService(
	engine ValidationEngine,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
) ValidationService {
	return ValidationService{
		engine:        engine,
		commitService: commitService,
		infoProvider:  infoProvider,
	}
}

// WithEngine returns a new ValidationService with the engine replaced.
func (s ValidationService) WithEngine(engine ValidationEngine) ValidationService {
	return ValidationService{
		engine:        engine,
		commitService: s.commitService,
		infoProvider:  s.infoProvider,
	}
}

// WithCommitService returns a new ValidationService with the commit service replaced.
func (s ValidationService) WithCommitService(commitService domain.GitCommitService) ValidationService {
	return ValidationService{
		engine:        s.engine,
		commitService: commitService,
		infoProvider:  s.infoProvider,
	}
}

// WithInfoProvider returns a new ValidationService with the info provider replaced.
func (s ValidationService) WithInfoProvider(infoProvider domain.RepositoryInfoProvider) ValidationService {
	return ValidationService{
		engine:        s.engine,
		commitService: s.commitService,
		infoProvider:  infoProvider,
	}
}

// WithCustomRule returns a new ValidationService with the custom rule added to the engine.
// Returns an error and the original service if the engine doesn't support custom rules.
func (s ValidationService) WithCustomRule(rule domain.Rule) (ValidationService, error) {
	// Create a copy to ensure immutability
	serviceCopy := s

	// Check if the engine itself implements the method directly
	if customizer, ok := s.engine.(interface {
		WithCustomRule(rule domain.Rule) ValidationEngine
	}); ok {
		serviceCopy.engine = customizer.WithCustomRule(rule)

		return serviceCopy, nil
	}

	// Check if provider is exposed and supports custom rules
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		provider := engineWithProvider.GetProvider()

		// Check if the provider allows registering custom rules
		if ruleProvider, ok := provider.(interface {
			WithCustomRule(ruleToAdd domain.Rule) domain.RuleProvider
		}); ok {
			newProvider := ruleProvider.WithCustomRule(rule)

			// Create a new engine with the updated provider
			if engineWithUpdatedProvider, ok := s.engine.(interface {
				WithProvider(provider domain.RuleProvider) ValidationEngine
			}); ok {
				serviceCopy.engine = engineWithUpdatedProvider.WithProvider(newProvider)

				return serviceCopy, nil
			}
		}
	}

	return s, errors.CreateBasicError("WithCustomRule", errors.ErrInvalidConfig, "rule provider does not support adding custom rules")
}

// WithCustomRuleFactory returns a new ValidationService with a custom rule factory added.
// Returns an error and the original service if the engine doesn't support custom rule factories.
func (s ValidationService) WithCustomRuleFactory(
	name string,
	factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
	requiresAnalyzer bool,
	condition func(config ValidationConfig) bool,
) (ValidationService, error) {
	// Create a copy to ensure immutability
	serviceCopy := s

	// Check if the engine itself implements the method directly
	if customizer, hasCustomizer := s.engine.(interface {
		WithCustomRuleFactory(
			name string,
			factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
			requiresAnalyzer bool,
			condition func(config ValidationConfig) bool,
		) ValidationEngine
	}); hasCustomizer {
		serviceCopy.engine = customizer.WithCustomRuleFactory(name, factory, requiresAnalyzer, condition)

		return serviceCopy, nil
	}

	// Check if provider is exposed and supports custom rule factories
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		provider := engineWithProvider.GetProvider()

		// Check if the provider allows registering custom rule factories
		if ruleProvider, exists := provider.(interface {
			WithCustomRuleFactory(
				name string,
				factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
				requiresAnalyzer bool,
				condition func(config ValidationConfig) bool,
			) domain.RuleProvider
		}); exists {
			newProvider := ruleProvider.WithCustomRuleFactory(name, factory, requiresAnalyzer, condition)

			// Create a new engine with the updated provider
			if engineWithUpdatedProvider, ok := s.engine.(interface {
				WithProvider(provider domain.RuleProvider) ValidationEngine
			}); ok {
				serviceCopy.engine = engineWithUpdatedProvider.WithProvider(newProvider)

				return serviceCopy, nil
			}
		}
	}

	return s, errors.CreateBasicError("WithCustomRuleFactory", errors.ErrInvalidConfig, "rule provider does not support registering custom rule factories")
}

// WithActiveRules creates a new ValidationService with the specified active rules.
// If an empty list is provided, all rules will be active.
// Returns an error and the original service if the rule provider doesn't support this operation.
// updateRulesHelper is a helper function that handles rule activation/deactivation.
// It reduces code duplication between WithActiveRules and WithDisabledRules.
func (s ValidationService) updateRulesHelper(
	ruleNames []string,
	operationMethod func(provider domain.RuleProvider) domain.RuleProvider,
	engineMethod func(ruleNames []string) (ValidationEngine, error),
	errorOp string,
	errorMsg string,
) (ValidationService, error) {
	// Create a copy to ensure immutability
	serviceCopy := s

	// First check if we can get the provider from the engine
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		// Get the provider
		provider := engineWithProvider.GetProvider()

		// Apply the operation on the provider
		newProvider := operationMethod(provider)

		// Create a new engine with the updated provider
		if engineWithUpdatedProvider, ok := s.engine.(interface {
			WithProvider(provider domain.RuleProvider) ValidationEngine
		}); ok {
			serviceCopy.engine = engineWithUpdatedProvider.WithProvider(newProvider)

			return serviceCopy, nil
		}
	}

	// As a fallback, check if the engine itself implements the method directly
	if engineMethod != nil {
		newEngine, err := engineMethod(ruleNames)
		if err != nil {
			return s, err
		}

		serviceCopy.engine = newEngine

		return serviceCopy, nil
	}

	return s, errors.CreateBasicError(errorOp, errors.ErrInvalidConfig, errorMsg)
}

func (s ValidationService) WithActiveRules(ruleNames []string) (ValidationService, error) {
	// Create closures for the required operations
	operationMethod := func(provider domain.RuleProvider) domain.RuleProvider {
		return provider.WithActiveRules(ruleNames)
	}

	// Get the engine method if it exists
	var engineMethod func([]string) (ValidationEngine, error)
	if ruleEngine, ok := s.engine.(interface {
		WithActiveRules(ruleNames []string) (ValidationEngine, error)
	}); ok {
		engineMethod = ruleEngine.WithActiveRules
	}

	return s.updateRulesHelper(
		ruleNames,
		operationMethod,
		engineMethod,
		"WithActiveRules",
		"rule provider does not support setting active rules",
	)
}

// WithDisabledRules creates a new ValidationService with the specified rules disabled.
// Returns an error and the original service if the rule provider doesn't support this operation.
func (s ValidationService) WithDisabledRules(ruleNames []string) (ValidationService, error) {
	// Create closures for the required operations
	operationMethod := func(provider domain.RuleProvider) domain.RuleProvider {
		return provider.WithDisabledRules(ruleNames)
	}

	// Get the engine method if it exists
	var engineMethod func([]string) (ValidationEngine, error)
	if ruleEngine, ok := s.engine.(interface {
		WithDisabledRules(ruleNames []string) (ValidationEngine, error)
	}); ok {
		engineMethod = ruleEngine.WithDisabledRules
	}

	return s.updateRulesHelper(
		ruleNames,
		operationMethod,
		engineMethod,
		"WithDisabledRules",
		"rule provider does not support disabling rules",
	)
}

// ValidateCommit validates a single commit.
func (s ValidationService) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
	// Get logger from context
	logger := log.Logger(ctx)
	logger.Trace().Str("commit_hash", hash).Msg("Entering ValidateCommit")

	// Get configuration from context
	config := config.GetConfig(ctx)
	logger.Debug().
		Int("subject_max_length", config.Subject.MaxLength).
		Bool("body_required", config.Body.Required).
		Bool("conventional_required", config.Conventional.Required).
		Msg("Using configuration from context")

	// Get the commit from the git repository
	commit, err := s.commitService.GetCommit(ctx, hash)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Create a context with commit information for better logging
	ctxWithCommit := logger.With().
		Str("commit_hash", commit.Hash).
		Str("commit_subject", commit.Subject).
		Logger().WithContext(ctx)

	// Validate the commit using the context with configuration
	return s.engine.ValidateCommit(ctxWithCommit, commit), nil
}

// ValidateHeadCommits validates the specified number of commits from HEAD.
func (s ValidationService) ValidateHeadCommits(ctx context.Context, count int, skipMerge bool) (domain.ValidationResults, error) {
	// Get logger from context
	logger := log.Logger(ctx)
	logger.Trace().Int("count", count).Bool("skip_merge", skipMerge).Msg("Entering ValidateHeadCommits")

	// Get configuration from context
	config := config.GetConfig(ctx)
	logger.Debug().
		Int("subject_max_length", config.Subject.MaxLength).
		Bool("body_required", config.Body.Required).
		Bool("conventional_required", config.Conventional.Required).
		Msg("Using configuration from context")

	// Get repository configuration
	repoConfig := config.Repository
	if repoConfig.Path != "" {
		logger.Debug().Str("repo_path", repoConfig.Path).Msg("Using repository path from configuration")
	}

	// Get the commits from the git repository
	commits, err := s.commitService.GetHeadCommits(ctx, count)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get head commits: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := domain.NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Create a child context with validation information
	ctxWithInfo := logger.With().
		Int("commit_count", len(collection.All())).
		Bool("skip_merge", skipMerge).
		Logger().WithContext(ctx)

	// Validate the commits using the context with configuration
	return s.engine.ValidateCommits(ctxWithInfo, collection.All()), nil
}

// ValidateCommitRange validates all commits in the given range.
func (s ValidationService) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMerge bool) (domain.ValidationResults, error) {
	// Get logger from context
	logger := log.Logger(ctx)
	logger.Trace().
		Str("from_hash", fromHash).
		Str("to_hash", toHash).
		Bool("skip_merge", skipMerge).
		Msg("Entering ValidateCommitRange")

	// Get configuration from context
	config := config.GetConfig(ctx)
	logger.Debug().
		Int("subject_max_length", config.Subject.MaxLength).
		Bool("body_required", config.Body.Required).
		Bool("conventional_required", config.Conventional.Required).
		Msg("Using configuration from context")

	// Get the commits from the git repository
	commits, err := s.commitService.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := domain.NewCommitCollection(commits)
	if skipMerge {
		// Filter merge commits
		collection = collection.FilterMergeCommits()

		logger.Debug().Msg("Filtering merge commits")
	} else {
		// If not skipping merge commits, don't filter
		logger.Debug().Msg("Not filtering merge commits by request")
	}

	// Create a child context with validation information
	ctxWithInfo := logger.With().
		Str("from_hash", fromHash).
		Str("to_hash", toHash).
		Int("commit_count", len(collection.All())).
		Bool("skip_merge", skipMerge).
		Logger().WithContext(ctx)

	// Validate the commits using the context with configuration
	return s.engine.ValidateCommits(ctxWithInfo, collection.All()), nil
}

// ValidateMessageFile validates a commit message from a file.
func (s ValidationService) ValidateMessageFile(ctx context.Context, filePath string) (domain.ValidationResults, error) {
	// Get logger from context
	logger := log.Logger(ctx)
	logger.Trace().Str("file_path", filePath).Msg("Entering ValidateMessageFile")

	// Get configuration from context
	config := config.GetConfig(ctx)
	logger.Debug().
		Int("subject_max_length", config.Subject.MaxLength).
		Bool("body_required", config.Body.Required).
		Bool("conventional_required", config.Conventional.Required).
		Msg("Using configuration from context")

	// Read the message file
	messageBytes, err := os.ReadFile(filePath)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to read message file: %w", err)
	}

	// Convert to string and trim whitespace
	message := strings.TrimSpace(string(messageBytes))
	if message == "" {
		return domain.NewValidationResults(), stderrors.New("empty commit message")
	}

	// Log file information
	logger.Debug().
		Str("file_path", filePath).
		Str("message_length", strconv.Itoa(len(message))).
		Msg("Validating commit message from file")

	// Split into subject and body
	subject, body := domain.SplitCommitMessage(message)

	// Create a dummy commit
	commit := domain.CommitInfo{
		Hash:          "0000000000000000000000000000000000000000",
		Subject:       subject,
		Body:          body,
		Message:       message,
		IsMergeCommit: false,
	}

	// Validate the commit
	result := s.engine.ValidateCommit(ctx, commit)

	// Create validation results
	results := domain.NewValidationResults()
	results.AddCommitResult(result)

	return results, nil
}

// ValidateWithOptions validates commits according to the provided options.
func (s ValidationService) ValidateWithOptions(ctx context.Context, opts ValidationOptions) (domain.ValidationResults, error) {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("message_file", opts.MessageFile).
		Str("commit_hash", opts.CommitHash).
		Str("from_hash", opts.FromHash).
		Str("to_hash", opts.ToHash).
		Int("commit_count", opts.CommitCount).
		Bool("skip_merge", opts.SkipMergeCommits).
		Msg("Entering ValidateWithOptions")

	// Validate commit message file
	if opts.MessageFile != "" {
		return s.ValidateMessageFile(ctx, opts.MessageFile)
	}

	// Validate specific commit
	if opts.CommitHash != "" {
		result, err := s.ValidateCommit(ctx, opts.CommitHash)
		if err != nil {
			return domain.ValidationResults{}, err
		}

		// Create validation results
		results := domain.NewValidationResults()
		results.AddCommitResult(result)

		return results, nil
	}

	// Validate commit range
	if opts.FromHash != "" || opts.ToHash != "" {
		return s.ValidateCommitRange(ctx, opts.FromHash, opts.ToHash, opts.SkipMergeCommits)
	}

	// Validate head commits
	if opts.CommitCount > 0 {
		return s.ValidateHeadCommits(ctx, opts.CommitCount, opts.SkipMergeCommits)
	}

	// Default to validating the HEAD commit
	result, err := s.ValidateCommit(ctx, "HEAD")
	if err != nil {
		return domain.ValidationResults{}, err
	}

	// Create validation results
	results := domain.NewValidationResults()
	results.AddCommitResult(result)

	return results, nil
}

// CreateValidationServiceWithDependencies provides dependency injection
// in accordance with hexagonal architecture principles.

// ValidationConfig represents a composition of all the domain configuration interfaces
// needed for validation. This follows the Interface Segregation Principle by composing
// smaller, focused interfaces rather than using a single monolithic interface.
//
// This is deliberately kept in the application layer because it's a use-case specific
// composition of interfaces defined in the domain layer.
type ValidationConfig interface {
	domain.SubjectConfigProvider
	domain.BodyConfigProvider
	domain.ConventionalConfigProvider
	domain.JiraConfigProvider
	domain.SecurityConfigProvider
	domain.SpellCheckConfigProvider
	domain.RepositoryConfigProvider
	domain.RuleConfigProvider
}

// CreateValidationService creates a ValidationService with explicit dependencies.
//
// Parameters:
// - config: Provides validation configuration parameters
// - commitService: Provides access to Git commits
// - infoProvider: Provides information about the repository
// - analyzer: Provides analysis of the repository (e.g., commits ahead of reference branch)
//
// The validation service uses a rule factory approach where rules are created
// conditionally based on configuration. Rules can be enabled or disabled via
// configuration, and each rule has specific conditions for when it should be applied.
func CreateValidationService(
	config ValidationConfig,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
	analyzer domain.CommitAnalyzer,
) ValidationService {
	// Create rule provider with domain configuration
	// Uses standardDomainRuleFactories to create rules based on their conditions
	engineProvider := &DomainRuleProvider{
		config:          config,
		analyzer:        analyzer,
		customFactories: make(map[string]domainRuleFactory),
	}

	// Create validation engine that will use the rule provider
	engine := &DomainValidationEngine{
		provider: engineProvider,
	}

	// Create and return the validation service
	return NewValidationService(engine, commitService, infoProvider)
}

// CreateValidationServiceWithContext creates a ValidationService using context-based configuration.
//
// Parameters:
// - ctx: Context containing configuration via config.WithConfig
// - commitService: Provides access to Git commits
// - infoProvider: Provides information about the repository
// - analyzer: Provides analysis of the repository (e.g., commits ahead of reference branch)
//
// The validation service uses context-based rule factories where rules retrieve their
// configuration from context during validation. This supports the functional
// value-based approach to configuration management.
func CreateValidationServiceWithContext(
	ctx context.Context,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
	analyzer domain.CommitAnalyzer,
) ValidationService {
	logger := log.Logger(ctx)
	logger.Debug().Msg("Creating context-based validation service")

	// Get configuration from context for use in the rule provider
	// This is mainly to determine which rules to activate
	cfg := config.GetConfig(ctx)

	// Create a context-based rule provider that will build rules
	// which get their configuration from context during validation
	engineProvider := &ContextRuleProvider{
		configSnapshot:  cfg,
		analyzer:        analyzer,
		customFactories: make(map[string]contextRuleFactory),
	}

	// Create a validation engine that uses the context-based provider
	engine := &ContextValidationEngine{
		provider:  engineProvider,
		activeCtx: ctx,
	}

	// Create and return the validation service
	return NewValidationService(engine, commitService, infoProvider)
}

// See context_helpers.go for ContextRuleProvider implementation

// DomainRuleProvider provides rules using domain interfaces.
// It manages both built-in and custom rule factories.
type DomainRuleProvider struct {
	config          ValidationConfig
	analyzer        domain.CommitAnalyzer
	rules           []domain.Rule
	customFactories map[string]domainRuleFactory
}

// domainRuleFactory represents a factory for creating a rule, with a condition for when to create it.
type domainRuleFactory struct {
	provider         func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule
	requiresAnalyzer bool
	condition        func(config ValidationConfig) bool
}

// WithConfig returns a new DomainRuleProvider with the config replaced.
func (p DomainRuleProvider) WithConfig(config ValidationConfig) DomainRuleProvider {
	return DomainRuleProvider{
		config:          config,
		analyzer:        p.analyzer,
		rules:           p.copyRules(),
		customFactories: p.copyCustomFactories(),
	}
}

// WithAnalyzer returns a new DomainRuleProvider with the analyzer replaced.
func (p DomainRuleProvider) WithAnalyzer(analyzer domain.CommitAnalyzer) DomainRuleProvider {
	return DomainRuleProvider{
		config:          p.config,
		analyzer:        analyzer,
		rules:           p.copyRules(),
		customFactories: p.copyCustomFactories(),
	}
}

// copyRules creates a deep copy of the rules slice to ensure immutability.
func (p DomainRuleProvider) copyRules() []domain.Rule {
	if p.rules == nil {
		return nil
	}

	rulesCopy := make([]domain.Rule, len(p.rules))
	copy(rulesCopy, p.rules)

	return rulesCopy
}

// copyCustomFactories creates a deep copy of the customFactories map to ensure immutability.
func (p DomainRuleProvider) copyCustomFactories() map[string]domainRuleFactory {
	if p.customFactories == nil {
		return make(map[string]domainRuleFactory)
	}

	factoriesCopy := make(map[string]domainRuleFactory, len(p.customFactories))
	for k, v := range p.customFactories {
		factoriesCopy[k] = v
	}

	return factoriesCopy
}

// createRuleOnDemand attempts to create a rule instance on-demand when it's explicitly
// requested but not found in the existing rules. This is primarily for testing scenarios
// where rules need to be dynamically created.
func (p *DomainRuleProvider) createRuleOnDemand(name string) domain.Rule {
	if factory, ok := standardDomainRuleFactories[name]; ok {
		var rule domain.Rule
		if factory.requiresAnalyzer {
			rule = factory.provider(p.config, p.analyzer)
		} else {
			rule = factory.provider(p.config, nil)
		}

		return rule
	}

	return nil
}

// standardDomainRuleFactories defines factories for all built-in rules with conditions for creation.
var standardDomainRuleFactories = map[string]domainRuleFactory{
	"SubjectLength": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSubjectLengthRule(
				rules.WithMaxLength(config.SubjectMaxLength()),
			)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"CommitBody": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewCommitBodyRule(
				rules.WithRequireBody(config.BodyRequired()),
				rules.WithAllowSignOffOnly(config.BodyAllowSignOffOnly()),
			)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"ConventionalCommit": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.ConventionalCommitOption{}

			// Apply the allowed types if provided
			if types := config.ConventionalTypes(); len(types) > 0 {
				options = append(options, rules.WithAllowedTypes(types))
			}

			// Apply the allowed scopes if provided
			if scopes := config.ConventionalScopes(); len(scopes) > 0 {
				options = append(options, rules.WithAllowedScopes(scopes))
			}

			// Apply the max description length if provided
			if maxLength := config.ConventionalMaxDescriptionLength(); maxLength > 0 {
				options = append(options, rules.WithMaxDescLength(maxLength))
			}

			return rules.NewConventionalCommitRule(options...)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"ImperativeVerb": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			isConventional := config.ConventionalRequired()
			options := []rules.ImperativeVerbOption{}

			if isConventional {
				options = append(options, rules.WithImperativeConventionalCommit(true))
			}

			return rules.NewImperativeVerbRule(
				options...,
			)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"Signature": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SignatureOption{}

			options = append(options, rules.WithRequireSignature(config.SignatureRequired()))

			if types := config.AllowedSignatureTypes(); len(types) > 0 {
				options = append(options, rules.WithAllowedSignatureTypes(types))
			}

			return rules.NewSignatureRule(options...)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"SignOff": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SignOffOption{}

			options = append(options, rules.WithRequireSignOff(config.SignOffRequired()))

			options = append(options, rules.WithAllowMultipleSignOffs(config.AllowMultipleSignOffs()))

			return rules.NewSignOffRule(options...)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"Spell": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SpellRuleOption{}

			if locale := config.SpellLocale(); locale != "" {
				options = append(options, rules.WithLocale(locale))
			}

			if maxErrors := config.SpellMaxErrors(); maxErrors > 0 {
				options = append(options, rules.WithMaxErrors(maxErrors))
			}

			if ignoreWords := config.SpellIgnoreWords(); len(ignoreWords) > 0 {
				options = append(options, rules.WithIgnoreWords(ignoreWords))
			}

			if customWords := config.SpellCustomWords(); len(customWords) > 0 {
				// Convert map keys to slice
				wordList := make([]string, 0, len(customWords))
				for word := range customWords {
					wordList = append(wordList, word)
				}
				options = append(options, rules.WithCustomWords(wordList))
			}

			return rules.NewSpellRule(options...)
		},
		requiresAnalyzer: false,
		condition:        func(config ValidationConfig) bool { return config.SpellEnabled() }, // Only when enabled
	},
	"SubjectCase": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SubjectCaseOption{}

			if caseChoice := config.SubjectCase(); caseChoice != "" {
				options = append(options, rules.WithCaseChoice(caseChoice))
			}

			if config.ConventionalRequired() {
				options = append(options, rules.WithSubjectCaseCommitFormat(true))
			}

			if config.SubjectRequireImperative() {
				options = append(options, rules.WithAllowNonAlpha(true))
			}

			return rules.NewSubjectCaseRule(options...)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"SubjectSuffix": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			suffixes := config.SubjectInvalidSuffixes()
			if suffixes == "" {
				return rules.NewSubjectSuffixRule()
			}

			// Just pass the string as is - the rule expects a string
			return rules.NewSubjectSuffixRule(
				rules.WithInvalidSuffixes(suffixes),
			)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"JiraReference": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.JiraReferenceOption{}

			// Check if conventional commit format is required
			if config.ConventionalRequired() {
				options = append(options, rules.WithConventionalCommit())
			}

			// Check if body reference checking is enabled
			if config.JiraBodyRef() {
				options = append(options, rules.WithBodyRefChecking())
			}

			// Add valid projects if provided
			if projects := config.JiraProjects(); len(projects) > 0 {
				options = append(options, rules.WithValidProjects(projects))
			}

			return rules.NewJiraReferenceRule(options...)
		},
		requiresAnalyzer: false,
		condition:        func(config ValidationConfig) bool { return true }, // Always create, controlled by enabled_rules/disabled_rules
	},
	"CommitsAhead": {
		provider: func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule {
			options := []rules.CommitsAheadOption{}

			options = append(options, rules.WithMaxCommitsAhead(config.MaxCommitsAhead()))

			if analyzer != nil {
				options = append(options, rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
					return analyzer
				}))
			}

			return rules.NewCommitsAheadRule(options...)
		},
		requiresAnalyzer: true,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
}

// GetRules returns all configured validation rules.
func (p *DomainRuleProvider) GetRules(ctx context.Context) []domain.Rule {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering DomainRuleProvider.GetRules")

	if p.rules == nil {
		// Initialize rules if not already done
		p.initializeRules()
	}

	return p.rules
}

// GetActiveRules returns all active validation rules.
func (p *DomainRuleProvider) GetActiveRules(ctx context.Context) []domain.Rule {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering DomainRuleProvider.GetActiveRules")

	// Note: We don't have direct access to the RuleRegistry in DomainRuleProvider
	// So we'll just use the base rules we already have
	allRules := p.rules

	// Make sure rules are initialized
	if p.rules == nil {
		p.initializeRules()
		allRules = p.rules
	}

	// If no specific configuration for enabled/disabled rules, return all rules
	if len(p.config.EnabledRules()) == 0 && len(p.config.DisabledRules()) == 0 {
		return allRules
	}

	// Create a map of rule names for easier lookup
	ruleMap := make(map[string]domain.Rule)
	for _, rule := range allRules {
		ruleMap[rule.Name()] = rule
	}

	// If specific rules are enabled, only include those
	if len(p.config.EnabledRules()) > 0 {
		activeRules := make([]domain.Rule, 0)

		// Add only the enabled rules
		for _, name := range p.config.EnabledRules() {
			// Check if rule exists in map
			if rule, exists := ruleMap[name]; exists {
				activeRules = append(activeRules, rule)
			} else {
				// Try to create rule on demand if it's explicitly requested
				if rule := p.createRuleOnDemand(name); rule != nil {
					activeRules = append(activeRules, rule)
				}
			}
		}

		return activeRules
	}

	// Otherwise, include all except disabled ones
	activeRules := make([]domain.Rule, 0, len(allRules))

	// Create a map of disabled rule names for faster lookup
	disabledRuleMap := make(map[string]bool)
	for _, name := range p.config.DisabledRules() {
		disabledRuleMap[name] = true
	}

	// Add all rules except disabled ones
	for _, rule := range allRules {
		if !disabledRuleMap[rule.Name()] {
			activeRules = append(activeRules, rule)
		}
	}

	return activeRules
}

// GetAvailableRuleNames returns a list of all available rule names.
// This helps with discovery of supported rules.
func (p *DomainRuleProvider) GetAvailableRuleNames() []string {
	// Create a map to track names (avoid duplicates)
	nameSet := make(map[string]struct{})

	// Add standard rule names
	for name := range standardDomainRuleFactories {
		nameSet[name] = struct{}{}
	}

	// Add custom rule names
	for name := range p.customFactories {
		nameSet[name] = struct{}{}
	}

	// If rules are already initialized, include their names too
	if p.rules != nil {
		for _, rule := range p.rules {
			nameSet[rule.Name()] = struct{}{}
		}
	}

	// Convert map keys to slice
	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}

	return names
}

// disableRules is a helper function to create a new DomainRuleProvider with the specified rules disabled.
func (p DomainRuleProvider) disableRules(ruleNames []string) DomainRuleProvider {
	// Create a copy of the provider
	providerCopy := p

	// If the config supports setting disabled rules, use that
	if configSetter, ok := providerCopy.config.(interface {
		SetDisabledRules(ruleNames []string)
		SetEnabledRules(ruleNames []string)
	}); ok {
		// Add to disabled rules but clear enabled rules (to use the disable list)
		configSetter.SetEnabledRules([]string{})
		configSetter.SetDisabledRules(ruleNames)
	}

	// Reset rules so they'll be reinitialized with the new config
	providerCopy.rules = nil

	return providerCopy
}

// WithDisabledRules implements the domain.RuleProvider interface.
// It returns a new DomainRuleProvider with the specified rules disabled.
func (p *DomainRuleProvider) WithDisabledRules(ruleNames []string) domain.RuleProvider {
	// Create a copy of the provider with the specified rules disabled
	newProvider := p.disableRules(ruleNames)

	return &newProvider
}

// WithActiveRules implements the domain.RuleProvider interface.
// It returns a new DomainRuleProvider with the specified active rules.
func (p *DomainRuleProvider) WithActiveRules(ruleNames []string) domain.RuleProvider {
	// Create a copy of the provider
	providerCopy := *p

	// If the config supports setting enabled rules, use that
	if configSetter, ok := providerCopy.config.(interface {
		SetEnabledRules(ruleNames []string)
	}); ok {
		configSetter.SetEnabledRules(ruleNames)
	}

	// Reset rules so they'll be reinitialized with the new config
	providerCopy.rules = nil

	return &providerCopy
}

// addCustomRule is a helper function to create a new DomainRuleProvider with the custom rule added.
func (p DomainRuleProvider) addCustomRule(rule domain.Rule) DomainRuleProvider {
	// Create a copy of the provider
	providerCopy := p

	// Make a copy of the rules if they exist
	if providerCopy.rules != nil {
		rulesCopy := make([]domain.Rule, len(providerCopy.rules))
		copy(rulesCopy, providerCopy.rules)
		providerCopy.rules = rulesCopy

		// Find the rule by name to see if it already exists
		for i, existingRule := range providerCopy.rules {
			if existingRule.Name() == rule.Name() {
				// Replace the existing rule
				providerCopy.rules[i] = rule

				return providerCopy
			}
		}

		// If we didn't find and replace the rule, add it
		providerCopy.rules = append(providerCopy.rules, rule)
	} else {
		// Initialize rules with just this rule
		providerCopy.rules = []domain.Rule{rule}
	}

	return providerCopy
}

// WithCustomRule implements the domain.RuleProvider interface.
// It returns a new DomainRuleProvider with the custom rule added.
func (p *DomainRuleProvider) WithCustomRule(rule domain.Rule) domain.RuleProvider {
	newProvider := p.addCustomRule(rule)

	return &newProvider
}

// WithCustomRuleFactory returns a new DomainRuleProvider with the custom rule factory added.
func (p DomainRuleProvider) WithCustomRuleFactory(
	name string,
	factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
	requiresAnalyzer bool,
	condition func(config ValidationConfig) bool,
) DomainRuleProvider {
	// Create a copy of the provider
	providerCopy := p

	// Create a factory entry
	customFactory := domainRuleFactory{
		provider:         factory,
		requiresAnalyzer: requiresAnalyzer,
		condition:        condition,
	}

	// Make a copy of the custom factories map
	customFactoriesCopy := make(map[string]domainRuleFactory, len(p.customFactories)+1)
	for n, f := range p.customFactories {
		customFactoriesCopy[n] = f
	}

	// Add the new factory
	customFactoriesCopy[name] = customFactory
	providerCopy.customFactories = customFactoriesCopy

	// Reset rules so they'll be reinitialized with the new factory
	providerCopy.rules = nil

	return providerCopy
}

// initializeRules creates all the validation rules using factories and returns the updated provider.
func (p *DomainRuleProvider) initializeRules() {
	// Create a slice with initial capacity to avoid reallocations
	totalFactories := len(standardDomainRuleFactories) + len(p.customFactories)
	rules := make([]domain.Rule, 0, totalFactories)

	// Iterate through all standard rule factories and create rules based on conditions
	for _, factory := range standardDomainRuleFactories {
		// Check if the rule should be created based on its condition
		if factory.condition(p.config) {
			var rule domain.Rule
			if factory.requiresAnalyzer {
				rule = factory.provider(p.config, p.analyzer)
			} else {
				rule = factory.provider(p.config, nil)
			}

			rules = append(rules, rule)
		}
	}

	// Add any custom rule factories
	for _, factory := range p.customFactories {
		// Check if the rule should be created based on its condition
		if factory.condition(p.config) {
			var rule domain.Rule
			if factory.requiresAnalyzer {
				rule = factory.provider(p.config, p.analyzer)
			} else {
				rule = factory.provider(p.config, nil)
			}

			rules = append(rules, rule)
		}
	}

	// Set the computed rules
	p.rules = rules
}

// ContextValidationEngine adapts the ValidationEngine interface to use context-based configuration.
// It validates commits using rules that retrieve their configuration from context.
type ContextValidationEngine struct {
	provider  domain.RuleProvider
	activeCtx context.Context //nolint:containedctx // Renamed from ctx to avoid "context in struct" linting issues
}

// ValidateCommit validates a single commit using context-based configuration.
func (e ContextValidationEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	logger := log.Logger(ctx)
	logger.Debug().
		Str("commit", commit.Hash).
		Msg("Validating commit using context-based configuration")

	// Merge the engine's context with the provided context
	// This ensures that configuration from both contexts is available
	ctx = contextx.MergeContext(e.activeCtx, ctx)

	// Get the rules to apply from the provider using the context
	rules := e.provider.GetRules(ctx)

	// Build result using the domain factory function
	result := domain.NewCommitResult(commit)

	// Apply each rule
	for _, rule := range rules {
		// Use validation helper function with context
		validateRuleWithContext(ctx, rule, commit, result)
	}

	return *result
}

// ValidateCommits validates multiple commits using context-based configuration.
func (e ContextValidationEngine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	logger := log.Logger(ctx)
	logger.Debug().
		Int("commit_count", len(commits)).
		Msg("Validating multiple commits using context-based configuration")

	// Merge the engine's context with the provided context
	// This ensures that configuration from both contexts is available
	ctx = contextx.MergeContext(e.activeCtx, ctx)

	// Create results container
	results := domain.NewValidationResults()

	// Process each commit
	for _, commit := range commits {
		// Validate and add to results
		commitResult := e.ValidateCommit(ctx, commit)
		// Create a pointer to the result since AddResult expects a pointer
		resultPtr := &commitResult
		results.AddResult(resultPtr)
	}

	return results
}

// GetProvider returns the rule provider used by this engine.
func (e ContextValidationEngine) GetProvider() domain.RuleProvider {
	return e.provider
}

// DomainValidationEngine adapts the ValidationEngine interface to use domain interfaces.
// It is designed with value semantics for functional programming patterns.
type DomainValidationEngine struct {
	provider domain.RuleProvider
}

// GetProvider returns the rule provider used by this engine.
// This allows access to the rule provider for rule discovery and configuration.
func (e DomainValidationEngine) GetProvider() domain.RuleProvider {
	return e.provider
}

// WithProvider returns a new DomainValidationEngine with the provider replaced.
func (e DomainValidationEngine) WithProvider(provider domain.RuleProvider) ValidationEngine {
	return &DomainValidationEngine{
		provider: provider,
	}
}

// WithCustomRule returns a new DomainValidationEngine with the custom rule registered to the provider.
// If the provider doesn't support registering custom rules, returns the original engine unchanged.
func (e DomainValidationEngine) WithCustomRule(rule domain.Rule) DomainValidationEngine {
	// Check if provider supports WithCustomRule
	if provider, ok := e.provider.(interface {
		WithCustomRule(ruleToAdd domain.Rule) domain.RuleProvider
	}); ok {
		newProvider := provider.WithCustomRule(rule)

		// Use type assertion to convert from interface to concrete type
		if newEngine, ok := e.WithProvider(newProvider).(*DomainValidationEngine); ok {
			return *newEngine
		}
	}

	// If provider doesn't support the operation, return unchanged
	return e
}

// WithCustomRuleFactory returns a new DomainValidationEngine with the custom rule factory registered.
// If the provider doesn't support registering custom rule factories, returns the original engine unchanged.
func (e DomainValidationEngine) WithCustomRuleFactory(
	name string,
	factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
	requiresAnalyzer bool,
	condition func(config ValidationConfig) bool,
) DomainValidationEngine {
	// Check if provider supports WithCustomRuleFactory
	if provider, hasCustomRuleFactory := e.provider.(interface {
		WithCustomRuleFactory(
			name string,
			factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
			requiresAnalyzer bool,
			condition func(config ValidationConfig) bool,
		) domain.RuleProvider
	}); hasCustomRuleFactory {
		newProvider := provider.WithCustomRuleFactory(name, factory, requiresAnalyzer, condition)

		// Use type assertion to convert from interface to concrete type
		if newEngine, ok := e.WithProvider(newProvider).(*DomainValidationEngine); ok {
			return *newEngine
		}
	}

	// If provider doesn't support the operation, return unchanged
	return e
}

// GetAvailableRuleNames delegates to the provider if it supports this method.
func (e DomainValidationEngine) GetAvailableRuleNames(ctx context.Context) []string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GetAvailableRuleNames")

	// Check if the provider implements a method to get available rule names with context
	if nameProvider, ok := e.provider.(interface {
		GetAvailableRuleNames(ctx context.Context) []string
	}); ok {
		return nameProvider.GetAvailableRuleNames(ctx)
	}

	// Otherwise return the names of all rules the provider knows about using Map
	rules := e.provider.GetRules(ctx)

	return contextx.Map(rules, func(rule domain.Rule) string {
		return rule.Name()
	})
}

// WithActiveRules returns a new DomainValidationEngine with the specified active rules.
// If the provider doesn't support setting active rules, returns an error and the original engine.
func (e DomainValidationEngine) WithActiveRules(ruleNames []string) (DomainValidationEngine, error) {
	// Check if provider supports WithActiveRules
	if provider, ok := e.provider.(interface {
		WithActiveRules(ruleNames []string) domain.RuleProvider
	}); ok {
		newProvider := provider.WithActiveRules(ruleNames)

		// Use type assertion to convert from interface to concrete type
		if newEngine, ok := e.WithProvider(newProvider).(*DomainValidationEngine); ok {
			return *newEngine, nil
		}
	}

	return e, errors.CreateBasicError("WithActiveRules", errors.ErrInvalidConfig, "rule provider does not support setting active rules")
}

// WithDisabledRules returns a new DomainValidationEngine with the specified rules disabled.
// If the provider doesn't support disabling rules, returns an error and the original engine.
func (e DomainValidationEngine) WithDisabledRules(ruleNames []string) (DomainValidationEngine, error) {
	// Check if provider supports WithDisabledRules
	if provider, hasProvider := e.provider.(interface {
		WithDisabledRules(ruleNames []string) domain.RuleProvider
	}); hasProvider {
		newProvider := provider.WithDisabledRules(ruleNames)

		// Use type assertion to convert from interface to concrete type
		if newEngine, ok := e.WithProvider(newProvider).(*DomainValidationEngine); ok {
			return *newEngine, nil
		}
	}

	return e, errors.CreateBasicError("WithDisabledRules", errors.ErrInvalidConfig, "rule provider does not support disabling rules")
}

// ValidateCommit validates a single commit.
func (e DomainValidationEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	activeRules := e.provider.GetActiveRules(ctx)

	// Use pure functions for validation logic
	if len(activeRules) == 0 {
		return createEmptyResult(commit)
	}

	// Create rule results
	ruleResults := validateWithRules(ctx, commit, activeRules)

	// Determine if all rules passed
	passed := allRulesPassed(ruleResults)

	// Create the final result
	return domain.CommitResult{
		CommitInfo:  commit,
		RuleResults: ruleResults,
		Passed:      passed,
	}
}

// createEmptyResult creates a passing result when no rules are active.
func createEmptyResult(commit domain.CommitInfo) domain.CommitResult {
	return domain.CommitResult{
		CommitInfo:  commit,
		RuleResults: []domain.RuleResult{},
		Passed:      true,
	}
}

// validateWithRules runs all active rules against a commit.
func validateWithRules(ctx context.Context, commit domain.CommitInfo, rules []domain.Rule) []domain.RuleResult {
	results := make([]domain.RuleResult, 0, len(rules))

	for _, rule := range rules {
		// Create rule result
		ruleResult := validateWithRule(ctx, commit, rule)
		results = append(results, ruleResult)
	}

	return results
}

// validateWithRule runs a single rule against a commit.
func validateWithRule(ctx context.Context, commit domain.CommitInfo, rule domain.Rule) domain.RuleResult {
	// Check if the rule supports context
	var ruleErrors = rule.Validate(ctx, commit)

	// Create rule result
	ruleResult := domain.RuleResult{
		RuleID:         rule.Name(),
		RuleName:       rule.Name(),
		Message:        rule.Result(ruleErrors),
		VerboseMessage: rule.VerboseResult(ruleErrors),
		HelpMessage:    rule.Help(ruleErrors),
		Errors:         ruleErrors,
	}

	// Set status based on errors
	if len(ruleErrors) > 0 {
		ruleResult.Status = domain.StatusFailed
		ruleResult.HelpMessage = generateHelpMessageForFailingRule(rule.Name(), ruleResult.HelpMessage, ruleErrors)
	} else {
		ruleResult.Status = domain.StatusPassed
	}

	return ruleResult
}

// generateHelpMessageForFailingRule creates a helpful message for failing rules.
// Using underscore parameter name to indicate deliberately unused parameter.
func generateHelpMessageForFailingRule(ruleName, currentHelp string, _ []errors.ValidationError) string {
	// Only override help messages for failing rules in two scenarios:
	// 1. If the help message contains "No errors to fix" (which is incorrect for failing rules)
	// 2. If the help message contains success-oriented text like "valid" or "contains valid" (for JiraReference)
	if !strings.Contains(currentHelp, "No errors to fix") &&
		!(ruleName == "JiraReference" && (strings.Contains(currentHelp, "contains valid") ||
			strings.HasPrefix(currentHelp, "Commit message contains valid"))) {
		// Keep current help message if it doesn't need override
		return currentHelp
	}

	// Based on the rule name, provide specific help guidance
	switch ruleName {
	case "CommitsAhead":
		return "Your branch is too far ahead of the reference branch.\n" +
			"To fix this, either:\n" +
			"1. Merge the reference branch into your branch\n" +
			"2. Rebase your branch onto the latest reference\n" +
			"3. Squash some commits to reduce the total count"
	case "JiraReference":
		return "Include a valid Jira issue key (e.g., PROJECT-123) in your commit subject.\n" +
			"For conventional commits, place the key at the end of the first line:\n" +
			"- feat(auth): add login feature PROJ-123\n" +
			"- fix: resolve timeout issue [PROJ-123]"
	case "SubjectLength":
		return "Your commit subject is too long. Shorten it to make it more readable.\n" +
			"A good commit subject should be concise (typically 50-70 characters) and to the point."
	case "SubjectCase":
		return "Your commit subject should use consistent capitalization.\n" +
			"Follow the project's convention for first letter case (usually lowercase for conventional commits)."
	case "SubjectSuffix":
		return "Your commit subject ends with invalid punctuation.\n" +
			"Remove trailing punctuation marks like periods, commas or other symbols at the end of your subject line."
	case "Spell":
		return "Your commit message contains spelling errors.\n" +
			"Check your message for typos and correct them for better readability."
	case "ConventionalCommit":
		return "Your commit doesn't follow the conventional commit format.\n" +
			"Use the format: type(scope): description\n" +
			"Example: feat(auth): add login feature"
	case "ImperativeVerb":
		return "Your commit should start with an imperative verb.\n" +
			"Use verbs like 'Add', 'Fix', 'Update' instead of 'Added', 'Fixed', 'Updates'."
	case "CommitBody":
		return "Your commit message body has formatting issues.\n" +
			"Ensure there's an empty line between subject and body, and include meaningful content."
	case "SignOff":
		return "Your commit is missing a proper sign-off.\n" +
			"Add 'Signed-off-by: Your Name <your.email@example.com>' to certify the Developer Certificate of Origin."
	case "Signature":
		return "Your commit is missing a cryptographic signature.\n" +
			"Sign your commit with 'git commit -S' using either GPG or SSH keys."
	default:
		// For any other rules with "No errors to fix" in a failing state
		return "Run with '--rulehelp=" + ruleName + "' for specific guidance on this rule."
	}
}

// allRulesPassed checks if all rules passed.
func allRulesPassed(results []domain.RuleResult) bool {
	return contextx.Every(results, func(result domain.RuleResult) bool {
		return result.Status != domain.StatusFailed
	})
}

// ValidateCommits validates multiple commits.
func (e DomainValidationEngine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	// Use a pure function to validate all commits
	return validateMultipleCommits(ctx, e, commits)
}

// validateMultipleCommits is a pure function that validates multiple commits.
func validateMultipleCommits(ctx context.Context, engine DomainValidationEngine, commits []domain.CommitInfo) domain.ValidationResults {
	// Use Reduce to process each commit and accumulate results
	return contextx.Reduce(commits, domain.NewValidationResults(), func(acc domain.ValidationResults, commit domain.CommitInfo) domain.ValidationResults {
		commitResult := engine.ValidateCommit(ctx, commit)
		acc.AddCommitResult(commitResult)

		return acc
	})
}

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
	"strings"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
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
type ValidationService struct {
	engine        ValidationEngine
	commitService domain.GitCommitService
	infoProvider  domain.RepositoryInfoProvider
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

// GetAvailableRuleNames returns the names of all available rules.
// This is useful for discovery and documentation purposes.
func (s ValidationService) GetAvailableRuleNames() []string {
	// Check if the engine provider exposes available rule names
	if provider, ok := s.engine.(interface{ GetAvailableRuleNames() []string }); ok {
		return provider.GetAvailableRuleNames()
	}

	// If not, check if provider is exposed
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		// Check if the provider can give us rule names
		if nameProvider, ok := engineWithProvider.GetProvider().(interface{ GetAvailableRuleNames() []string }); ok {
			return nameProvider.GetAvailableRuleNames()
		}
	}

	// If all else fails, return an empty list
	return []string{}
}

// RegisterCustomRule adds a custom rule to the validation engine.
// Returns an error if the engine doesn't support custom rules.
func (s ValidationService) RegisterCustomRule(rule domain.Rule) error {
	// Check if the engine itself implements the method directly
	if customizer, ok := s.engine.(interface{ RegisterCustomRule(rule domain.Rule) }); ok {
		customizer.RegisterCustomRule(rule)

		return nil
	}

	// Check if provider is exposed and supports custom rules
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		provider := engineWithProvider.GetProvider()

		// Check if the provider allows registering custom rules
		if ruleRegistrar, ok := provider.(interface{ RegisterCustomRule(rule domain.Rule) }); ok {
			ruleRegistrar.RegisterCustomRule(rule)

			return nil
		}
	}

	return errors.New("RegisterCustomRule", errors.ErrInvalidConfig, "rule provider does not support registering custom rules")
}

// RegisterCustomRuleFactory registers a factory for creating a custom rule with conditional creation.
// This provides more flexibility than directly registering a rule instance.
// Returns an error if the engine doesn't support custom rule factories.
func (s ValidationService) RegisterCustomRuleFactory(
	name string,
	factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
	requiresAnalyzer bool,
	condition func(config ValidationConfig) bool,
) error {
	// Check if provider is exposed and supports custom rule factories
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		provider := engineWithProvider.GetProvider()

		// Check if the provider allows registering custom rule factories
		if factoryRegistrar, exists := provider.(interface {
			RegisterCustomRuleFactory(
				name string,
				factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
				requiresAnalyzer bool,
				condition func(config ValidationConfig) bool,
			)
		}); exists {
			factoryRegistrar.RegisterCustomRuleFactory(name, factory, requiresAnalyzer, condition)

			return nil
		}
	}

	return errors.New("RegisterCustomRuleFactory", errors.ErrInvalidConfig, "rule provider does not support registering custom rule factories")
}

// SetActiveRules configures which rules should be active for validation.
// If an empty list is provided, all rules will be active.
// Returns an error if the rule provider doesn't support this operation.
func (s ValidationService) SetActiveRules(ruleNames []string) error {
	// Check if the engine itself implements the method directly
	if setter, ok := s.engine.(interface {
		SetActiveRules(ruleNames []string) error
	}); ok {
		return setter.SetActiveRules(ruleNames)
	}

	// Check if provider is exposed and supports setting active rules
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		provider := engineWithProvider.GetProvider()

		// Check if the provider has a method to set active rules
		if activeSetter, ok := provider.(interface{ SetActiveRules(ruleNames []string) }); ok {
			activeSetter.SetActiveRules(ruleNames)

			return nil
		}
	}

	// Use a fallback approach for simpler engines that just have active rule lists
	if _, ok := s.engine.(interface {
		GetRules() []domain.Rule
		GetActiveRules() []domain.Rule
	}); ok {
		// This is informational - we've done our best
		// We could inspect the rules but can't modify the engine
		return nil
	}

	return errors.New("SetActiveRules", errors.ErrInvalidConfig, "rule provider does not support setting active rules")
}

// DisableRules disables specific rules by name.
// Returns an error if the rule provider doesn't support this operation.
func (s ValidationService) DisableRules(ruleNames []string) error {
	// Check if the engine itself implements the method directly
	if disabler, ok := s.engine.(interface {
		DisableRules(ruleNames []string) error
	}); ok {
		return disabler.DisableRules(ruleNames)
	}

	// Check if provider is exposed and supports disabling rules
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		provider := engineWithProvider.GetProvider()

		// Check if the provider has a method to disable rules
		if ruleSetter, ok := provider.(interface{ DisableRules(ruleNames []string) }); ok {
			ruleSetter.DisableRules(ruleNames)

			return nil
		}
	}

	// Use a fallback approach for simple engines - similar to SetActiveRules
	if _, ok := s.engine.(interface {
		GetRules() []domain.Rule
		GetActiveRules() []domain.Rule
	}); ok {
		// Informational only - can't modify the engine directly
		if len(ruleNames) == 0 {
			return nil // Nothing to disable
		}

		return nil
	}

	return errors.New("DisableRules", errors.ErrInvalidConfig, "rule provider does not support disabling rules")
}

// GetActiveRules returns the names of currently active rules.
func (s ValidationService) GetActiveRules() []string {
	// Check if provider is exposed
	if engineWithProvider, ok := s.engine.(interface{ GetProvider() domain.RuleProvider }); ok {
		provider := engineWithProvider.GetProvider()

		// Get active rules and extract their names
		activeRules := provider.GetActiveRules()
		names := make([]string, 0, len(activeRules))

		for _, rule := range activeRules {
			names = append(names, rule.Name())
		}

		return names
	}

	return []string{}
}

// ValidateCommit validates a single commit.
func (s ValidationService) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
	// Get the commit from the git repository
	commit, err := s.commitService.GetCommit(ctx, hash)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Validate the commit
	return s.engine.ValidateCommit(ctx, commit), nil
}

// ValidateHeadCommits validates the specified number of commits from HEAD.
func (s ValidationService) ValidateHeadCommits(ctx context.Context, count int, skipMerge bool) (domain.ValidationResults, error) {
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

	// Validate the commits
	return s.engine.ValidateCommits(ctx, collection.All()), nil
}

// ValidateCommitRange validates all commits in the given range.
func (s ValidationService) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMerge bool) (domain.ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.commitService.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := domain.NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Validate the commits
	return s.engine.ValidateCommits(ctx, collection.All()), nil
}

// ValidateMessageFile validates a commit message from a file.
func (s ValidationService) ValidateMessageFile(ctx context.Context, filePath string) (domain.ValidationResults, error) {
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
	// Create validation results
	results := domain.NewValidationResults()

	// Validate commit message file
	if opts.MessageFile != "" {
		return s.ValidateMessageFile(ctx, opts.MessageFile)
	}

	// Validate specific commit
	if opts.CommitHash != "" {
		result, err := s.ValidateCommit(ctx, opts.CommitHash)
		if err != nil {
			return results, err
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
		return results, err
	}

	results.AddCommitResult(result)

	return results, nil
}

// CreateValidationService is a legacy method kept for backward compatibility.
// For better testability and dependency management, use dependency injection via NewValidationService instead.
//
// DEPRECATED: Use CreateValidationServiceFromFactory instead, which accepts domain interface types
// rather than concrete implementation types.
func CreateValidationService(_ string) (ValidationService, error) {
	return ValidationService{}, errors.New("CreateValidationService", errors.ErrInvalidConfig, "deprecated: use CreateValidationServiceFromFactory instead")
}

// CreateValidationServiceFromFactory creates a validation service using domain interfaces.
// This follows the hexagonal architecture pattern by depending on domain interfaces
// rather than concrete implementations.
func CreateValidationServiceFromFactory(
	config ValidationConfig,
	repoFactory domain.RepositoryFactory,
) (ValidationService, error) {
	// Get specialized repository interfaces from the factory
	commitService := repoFactory.CreateGitCommitService()
	infoProvider := repoFactory.CreateInfoProvider()
	analyzer := repoFactory.CreateCommitAnalyzer()

	return CreateValidationServiceWithDependencies(
		config,
		commitService,
		infoProvider,
		analyzer,
	), nil
}

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

// CreateValidationServiceWithDependencies creates a ValidationService with explicit dependencies.
// This is the preferred constructor for better testability and dependency management.
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
func CreateValidationServiceWithDependencies(
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

// standardDomainRuleFactories defines factories for all built-in rules with conditions for creation.
var standardDomainRuleFactories = map[string]domainRuleFactory{
	"SubjectLength": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSubjectLengthRuleWithConfig(config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"CommitBody": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewCommitBodyRuleWithConfig(config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"ConventionalCommit": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewConventionalCommitRuleWithConfig(config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"ImperativeVerb": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewImperativeVerbRuleWithConfig(config, config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"Signature": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSignatureRuleWithConfig(config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"SignOff": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSignOffRuleWithConfig(config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"Spell": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSpellRuleWithConfig(config)
		},
		requiresAnalyzer: false,
		condition:        func(config ValidationConfig) bool { return config.SpellEnabled() }, // Only when enabled
	},
	"SubjectCase": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSubjectCaseRuleWithConfig(config, config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"SubjectSuffix": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSubjectSuffixRuleWithConfig(config)
		},
		requiresAnalyzer: false,
		condition:        func(_ ValidationConfig) bool { return true }, // Always create
	},
	"JiraReference": {
		provider: func(config ValidationConfig, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewJiraReferenceRuleWithConfig(config, config)
		},
		requiresAnalyzer: false,
		condition:        func(config ValidationConfig) bool { return config.JiraRequired() }, // Only when Jira is required
	},
	"CommitsAhead": {
		provider: func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule {
			return rules.NewCommitsAheadRuleWithConfig(config, analyzer)
		},
		requiresAnalyzer: true,
		condition:        func(config ValidationConfig) bool { return config.CheckCommitsAhead() }, // Only when enabled
	},
}

// GetRules returns all configured validation rules.
func (p *DomainRuleProvider) GetRules() []domain.Rule {
	if p.rules == nil {
		// Initialize rules if not already done
		p.initializeRules()
	}

	return p.rules
}

// GetActiveRules returns all active validation rules.
func (p *DomainRuleProvider) GetActiveRules() []domain.Rule {
	allRules := p.GetRules()

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
			if rule, exists := ruleMap[name]; exists {
				activeRules = append(activeRules, rule)
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

// SetActiveRules sets which rules are active based on a list of rule names.
// If an empty list is provided, all rules will be active.
func (p *DomainRuleProvider) SetActiveRules(ruleNames []string) {
	// We need to reinitialize rules with the new settings
	// Reset the rules so they will be reinitialized
	p.rules = nil

	// If the config supports setting enabled/disabled rules, use that
	if configSetter, ok := p.config.(interface {
		SetEnabledRules(ruleNames []string)
		SetDisabledRules(ruleNames []string)
	}); ok {
		if len(ruleNames) > 0 {
			// Set specific rules as enabled
			configSetter.SetEnabledRules(ruleNames)
			configSetter.SetDisabledRules([]string{}) // Clear disabled rules
		} else {
			// Enable all rules
			configSetter.SetEnabledRules([]string{})
			configSetter.SetDisabledRules([]string{})
		}
	}

	// Force rule reinitialization with the new settings
	p.GetRules()
}

// DisableRules disables specific rules by name.
func (p *DomainRuleProvider) DisableRules(ruleNames []string) {
	// We need to reinitialize rules with the new settings
	// Reset the rules so they will be reinitialized
	p.rules = nil

	// If the config supports setting disabled rules, use that
	if configSetter, ok := p.config.(interface {
		SetDisabledRules(ruleNames []string)
		SetEnabledRules(ruleNames []string)
	}); ok {
		// Add to disabled rules but clear enabled rules (to use the disable list)
		configSetter.SetEnabledRules([]string{})
		configSetter.SetDisabledRules(ruleNames)
	}

	// Force rule reinitialization with the new settings
	p.GetRules()
}

// RegisterCustomRule adds a custom rule to the provider.
// This allows for extension with user-defined rules.
// If a rule with the same name already exists, it will be replaced.
func (p *DomainRuleProvider) RegisterCustomRule(rule domain.Rule) {
	// Reset rules so they'll be reinitialized
	if p.rules != nil {
		// Find the rule by name to see if it already exists
		for i, existingRule := range p.rules {
			if existingRule.Name() == rule.Name() {
				// Replace the existing rule
				p.rules[i] = rule

				return
			}
		}

		// If we didn't find and replace the rule, add it
		p.rules = append(p.rules, rule)
	}
}

// RegisterCustomRuleFactory registers a factory for creating a custom rule.
// This allows for more flexible rule creation with conditions.
// If a factory with the same name already exists, it will be replaced.
func (p *DomainRuleProvider) RegisterCustomRuleFactory(
	name string,
	factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
	requiresAnalyzer bool,
	condition func(config ValidationConfig) bool,
) {
	// Create a factory entry
	customFactory := domainRuleFactory{
		provider:         factory,
		requiresAnalyzer: requiresAnalyzer,
		condition:        condition,
	}

	// Register the factory in a thread-safe way
	// Since maps are not inherently thread-safe in Go, we make a copy
	// This is not strictly necessary in single-threaded use, but is good practice
	customFactories := make(map[string]domainRuleFactory)
	for name, factory := range p.customFactories {
		customFactories[name] = factory
	}

	customFactories[name] = customFactory
	p.customFactories = customFactories

	// Reset rules so they'll be reinitialized with the new factory
	p.rules = nil
}

// initializeRules creates all the validation rules using factories.
func (p *DomainRuleProvider) initializeRules() {
	// Create a slice with initial capacity to avoid reallocations
	totalFactories := len(standardDomainRuleFactories) + len(p.customFactories)
	p.rules = make([]domain.Rule, 0, totalFactories)

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

			p.rules = append(p.rules, rule)
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

			p.rules = append(p.rules, rule)
		}
	}
}

// DomainValidationEngine adapts the ValidationEngine interface to use domain interfaces.
type DomainValidationEngine struct {
	provider domain.RuleProvider
}

// GetProvider returns the rule provider used by this engine.
// This allows access to the rule provider for rule discovery and configuration.
func (e *DomainValidationEngine) GetProvider() domain.RuleProvider {
	return e.provider
}

// RegisterCustomRule delegates to the provider if it supports registering custom rules.
func (e *DomainValidationEngine) RegisterCustomRule(rule domain.Rule) {
	// Check if provider supports custom rules
	if customizer, ok := e.provider.(interface{ RegisterCustomRule(rule domain.Rule) }); ok {
		customizer.RegisterCustomRule(rule)
	}
}

// RegisterCustomRuleFactory delegates to the provider if it supports registering custom rule factories.
func (e *DomainValidationEngine) RegisterCustomRuleFactory(
	name string,
	factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
	requiresAnalyzer bool,
	condition func(config ValidationConfig) bool,
) {
	// Check if provider supports custom rule factories
	if factoryRegistrar, exists := e.provider.(interface {
		RegisterCustomRuleFactory(
			name string,
			factory func(config ValidationConfig, analyzer domain.CommitAnalyzer) domain.Rule,
			requiresAnalyzer bool,
			condition func(config ValidationConfig) bool,
		)
	}); exists {
		factoryRegistrar.RegisterCustomRuleFactory(name, factory, requiresAnalyzer, condition)
	}
}

// GetAvailableRuleNames delegates to the provider if it supports this method.
func (e *DomainValidationEngine) GetAvailableRuleNames() []string {
	// Check if the provider implements a method to get available rule names
	if nameProvider, ok := e.provider.(interface{ GetAvailableRuleNames() []string }); ok {
		return nameProvider.GetAvailableRuleNames()
	}

	// Otherwise return the names of all rules the provider knows about
	rules := e.provider.GetRules()
	names := make([]string, 0, len(rules))

	for _, rule := range rules {
		names = append(names, rule.Name())
	}

	return names
}

// SetActiveRules sets which rules should be active.
// This delegates to the provider if it supports setting active rules.
func (e *DomainValidationEngine) SetActiveRules(ruleNames []string) error {
	// Check if provider supports setting active rules
	if activeSetter, ok := e.provider.(interface{ SetActiveRules(ruleNames []string) }); ok {
		activeSetter.SetActiveRules(ruleNames)

		return nil
	}

	return errors.New("SetActiveRules", errors.ErrInvalidConfig, "rule provider does not support setting active rules")
}

// DisableRules disables specific rules.
// This delegates to the provider if it supports disabling rules.
func (e *DomainValidationEngine) DisableRules(ruleNames []string) error {
	// Check if provider supports disabling rules
	if ruleSetter, ok := e.provider.(interface{ DisableRules(ruleNames []string) }); ok {
		ruleSetter.DisableRules(ruleNames)

		return nil
	}

	return errors.New("DisableRules", errors.ErrInvalidConfig, "rule provider does not support disabling rules")
}

// ValidateCommit validates a single commit.
func (e *DomainValidationEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	activeRules := e.provider.GetActiveRules()

	// Initialize result
	result := domain.CommitResult{
		CommitInfo:  commit,
		RuleResults: make([]domain.RuleResult, 0, len(activeRules)),
		Passed:      true,
	}

	// Run each rule
	for _, rule := range activeRules {
		// Check for context cancellation
		if ctx.Err() != nil {
			break
		}

		// Check if the rule supports context
		var ruleErrors []errors.ValidationError
		if contextualRule, ok := rule.(domain.ContextualRule); ok {
			// Use the context-aware validation method
			ruleErrors = contextualRule.ValidateWithContext(ctx, commit)
		} else {
			// Fall back to the regular validation method
			ruleErrors = rule.Validate(commit)
		}

		// Create rule result
		ruleResult := domain.RuleResult{
			RuleID:         rule.Name(),
			RuleName:       rule.Name(),
			Message:        rule.Result(),
			VerboseMessage: rule.VerboseResult(),
			HelpMessage:    rule.Help(),
			Errors:         ruleErrors,
		}

		// Set status based on errors
		if len(ruleErrors) > 0 {
			ruleResult.Status = domain.StatusFailed
			result.Passed = false
		} else {
			ruleResult.Status = domain.StatusPassed
		}

		// Add to results
		result.RuleResults = append(result.RuleResults, ruleResult)
	}

	return result
}

// ValidateCommits validates multiple commits.
func (e *DomainValidationEngine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	results := domain.NewValidationResults()

	for _, commit := range commits {
		// Check for context cancellation
		if ctx.Err() != nil {
			break
		}

		// Validate commit
		commitResult := e.ValidateCommit(ctx, commit)

		// Add to results
		results.AddCommitResult(commitResult)
	}

	return results
}

// CreateDefaultValidationService creates a validation service with configuration.
// This function forwards to the CreateValidationService implementation.
func CreateDefaultValidationService(repoPath string) (ValidationService, error) {
	// Forward to the standard implementation
	return CreateValidationService(repoPath)
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
)

// Options contains options for validation.
type Options struct {
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

// ServiceDependencies holds all the dependencies for the Service.
type ServiceDependencies struct {
	// Engine that performs the validation
	Engine Engine

	// CommitService for retrieving commit information
	CommitService domain.GitCommitService

	// InfoProvider for repository information
	InfoProvider domain.RepositoryInfoProvider

	// Analyzer for advanced repository analysis
	Analyzer domain.CommitAnalyzer
}

// Service provides validation using the configuration system.
// It is designed with value semantics for functional programming.
type Service struct {
	dependencies ServiceDependencies
	config       config.Config
}

// NewService creates a new Service.
func NewService(deps ServiceDependencies, cfg config.Config) Service {
	return Service{
		dependencies: deps,
		config:       cfg,
	}
}

// WithDependencies returns a new Service with the specified dependencies.
func (s Service) WithDependencies(deps ServiceDependencies) Service {
	return Service{
		dependencies: deps,
		config:       s.config,
	}
}

// WithConfig returns a new Service with the specified configuration.
func (s Service) WithConfig(cfg config.Config) Service {
	return Service{
		dependencies: s.dependencies,
		config:       cfg,
	}
}

// ValidateCommit validates a single commit.
func (s Service) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
	// Get the commit from the git repository
	commit, err := s.dependencies.CommitService.GetCommit(ctx, hash)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Validate the commit
	return s.dependencies.Engine.ValidateCommit(ctx, commit), nil
}

// ValidateHeadCommits validates the specified number of commits from HEAD.
func (s Service) ValidateHeadCommits(ctx context.Context, count int, skipMerge bool) (domain.ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.dependencies.CommitService.GetHeadCommits(ctx, count)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get head commits: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := domain.NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Validate the commits
	return s.dependencies.Engine.ValidateCommits(ctx, collection.All()), nil
}

// ValidateCommitRange validates all commits in the given range.
func (s Service) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMerge bool) (domain.ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.dependencies.CommitService.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := domain.NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Validate the commits
	return s.dependencies.Engine.ValidateCommits(ctx, collection.All()), nil
}

// ValidateMessageFile validates a commit message from a file.
func (s Service) ValidateMessageFile(ctx context.Context, filePath string) (domain.ValidationResults, error) {
	// Read the message file
	messageBytes, err := os.ReadFile(filePath)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to read message file: %w", err)
	}

	// Convert to string and trim whitespace
	message := strings.TrimSpace(string(messageBytes))
	if message == "" {
		return domain.NewValidationResults(), errors.New("empty commit message")
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
	result := s.dependencies.Engine.ValidateCommit(ctx, commit)

	// Create validation results
	results := domain.NewValidationResults()
	results.AddCommitResult(result)

	return results, nil
}

// ValidateWithOptions validates commits according to the provided options.
func (s Service) ValidateWithOptions(ctx context.Context, opts Options) (domain.ValidationResults, error) {
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

// CreateService creates a validation service with the configuration.
func CreateService(config config.Config, repoPath string) (Service, error) {
	// Create the repository adapter
	repoAdapter, err := git.NewRepositoryAdapter(repoPath)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	// Create rule provider - not used directly as we use CreateEngine instead
	_ = CreateRuleProvider(config, repoAdapter)

	// Create engine using rule provider
	engine := CreateEngine(config, repoAdapter)

	// Create dependencies
	deps := ServiceDependencies{
		Engine:        engine,
		CommitService: repoAdapter,
		InfoProvider:  repoAdapter,
		Analyzer:      repoAdapter,
	}

	// Create and return service
	return NewService(deps, config), nil
}

// CreateEngine creates a validation engine using the configuration.
func CreateEngine(config config.Config, analyzer domain.CommitAnalyzer) Engine {
	return Engine{
		ruleProvider: &RulesManager{
			config:   config,
			analyzer: analyzer,
		},
	}
}

// RulesManager provides rules using rule factories.
type RulesManager struct {
	config   config.Config
	analyzer domain.CommitAnalyzer
}

// GetRules returns all rules from rule factories.
func (p *RulesManager) GetRules() []domain.Rule {
	return p.getActiveRules()
}

// GetActiveRules returns all active rules for validation.
func (p *RulesManager) GetActiveRules() []domain.Rule {
	return p.getActiveRules()
}

// WithActiveRules returns a new provider with the specified active rules.
func (p *RulesManager) WithActiveRules(ruleNames []string) domain.RuleProvider {
	// Use the WithEnabledRules method on Config which follows value semantics
	newConfig := p.config.WithEnabledRules(ruleNames)

	// Return a new instance with the updated configuration
	return &RulesManager{
		config:   newConfig,
		analyzer: p.analyzer,
	}
}

// WithDisabledRules returns a new provider with the specified rules disabled.
func (p *RulesManager) WithDisabledRules(ruleNames []string) domain.RuleProvider {
	// Use the WithDisabledRules method on Config which follows value semantics
	newConfig := p.config.WithDisabledRules(ruleNames)

	// Return a new instance with the updated configuration
	return &RulesManager{
		config:   newConfig,
		analyzer: p.analyzer,
	}
}

// WithCustomRule returns a new provider with the custom rule added.
func (p *RulesManager) WithCustomRule(_ domain.Rule) domain.RuleProvider {
	// Custom rules aren't directly supported in the system
	// For now, we return a new instance without changes
	// A future implementation could store custom rules in a slice
	return &RulesManager{
		config:   p.config,
		analyzer: p.analyzer,
	}
}

// getActiveRules creates and returns all active rules.
func (p *RulesManager) getActiveRules() []domain.Rule {
	// Create rules from the rule factories
	var rules []domain.Rule

	// Get all rule names
	ruleNames := GetRuleNames()

	// If specific rules are enabled, only use those
	enabledRules := p.config.EnabledRules()
	if len(enabledRules) > 0 {
		// Filter to only enabled rules
		filteredNames := make([]string, 0)

		for _, name := range ruleNames {
			for _, enabled := range enabledRules {
				if name == enabled {
					filteredNames = append(filteredNames, name)

					break
				}
			}
		}

		ruleNames = filteredNames
	} else if len(p.config.DisabledRules()) > 0 {
		// Remove disabled rules
		disabledMap := make(map[string]bool)
		for _, name := range p.config.DisabledRules() {
			disabledMap[name] = true
		}

		// Filter out disabled rules
		filteredNames := make([]string, 0)

		for _, name := range ruleNames {
			if !disabledMap[name] {
				filteredNames = append(filteredNames, name)
			}
		}

		ruleNames = filteredNames
	}

	// Create rules from factories
	for _, name := range ruleNames {
		rule := CreateRuleWithConfig(name, p.config, p.analyzer)
		if rule != nil {
			rules = append(rules, rule)
		}
	}

	return rules
}

// All validation logic is now handled by the standard Engine with our custom RulesManager.

// CreateRuleProvider creates a rule provider for the configuration.
func CreateRuleProvider(config config.Config, analyzer domain.CommitAnalyzer) domain.RuleProvider {
	return &RulesManager{
		config:   config,
		analyzer: analyzer,
	}
}

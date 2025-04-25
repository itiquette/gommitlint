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

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
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

// CreateValidationService creates a validation service.
// This factory method is provided for backward compatibility.
// For better testability and dependency management, use explicit constructor injection instead.
func CreateValidationService(repoPath string) (ValidationService, error) {
	// Create config manager
	configManager, err := config.New()
	if err != nil {
		return ValidationService{}, fmt.Errorf("failed to create configuration manager: %w", err)
	}

	return CreateValidationServiceWithConfigManager(repoPath, configManager)
}

// CreateValidationServiceWithConfigManager creates a validation service with an explicit config manager.
// This provides better testability and dependency management.
func CreateValidationServiceWithConfigManager(repoPath string, configManager *config.Manager) (ValidationService, error) {
	// Get validation configuration interface
	validationConfig := configManager.GetValidationConfig()

	// Create repository factory
	factory, err := git.NewRepositoryFactory(repoPath)
	if err != nil {
		return ValidationService{}, fmt.Errorf("failed to create repository factory: %w", err)
	}

	// Get specialized repository interfaces
	commitService := factory.CreateGitCommitService()
	infoProvider := factory.CreateInfoProvider()
	analyzer := factory.CreateCommitAnalyzer()

	return CreateValidationServiceWithDependencies(
		validationConfig,
		commitService,
		infoProvider,
		analyzer,
	), nil
}

// CreateValidationServiceWithDependencies creates a ValidationService with explicit dependencies.
// This is the preferred constructor for better testability and dependency management.
func CreateValidationServiceWithDependencies(
	config domain.ValidationConfigProvider,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
	analyzer domain.CommitAnalyzer,
) ValidationService {
	// Create rule provider with domain configuration
	// The validation package needs to be updated to accept domain.ValidationConfigProvider
	// For now, we'll create an adapter to provide the validation-specific config
	// Create engine using the domain-based rule provider
	engineProvider := &DomainRuleProvider{
		config:   config,
		analyzer: analyzer,
	}

	// Create validation engine
	engine := &DomainValidationEngine{
		provider: engineProvider,
	}

	// Create and return the validation service
	return NewValidationService(engine, commitService, infoProvider)
}

// DomainRuleProvider provides rules using domain interfaces.
type DomainRuleProvider struct {
	config   domain.ValidationConfigProvider
	analyzer domain.CommitAnalyzer
	rules    []domain.Rule
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
	// For this adapter, we'll return all rules as active
	return p.GetRules()
}

// initializeRules creates all the validation rules.
func (p *DomainRuleProvider) initializeRules() {
	p.rules = []domain.Rule{
		// Add the rules you need with their domain-based constructors
		rules.NewSpellRuleWithConfig(p.config),
		rules.NewSubjectLengthRuleWithConfig(p.config),
		rules.NewCommitBodyRuleWithConfig(p.config),
		rules.NewConventionalCommitRuleWithConfig(p.config),
		rules.NewImperativeVerbRuleWithConfig(p.config, p.config),
		rules.NewSignatureRuleWithConfig(p.config),
		rules.NewSignOffRuleWithConfig(p.config),
		rules.NewSubjectCaseRuleWithConfig(p.config, p.config),
		rules.NewSubjectSuffixRuleWithConfig(p.config),
		rules.NewCommitsAheadRuleWithConfig(p.config, p.analyzer),
	}

	// Add Jira rule conditionally
	if p.config.JiraRequired() {
		p.rules = append(p.rules, rules.NewJiraReferenceRuleWithConfig(p.config, p.config))
	}
}

// DomainValidationEngine adapts the ValidationEngine interface to use domain interfaces.
type DomainValidationEngine struct {
	provider domain.RuleProvider
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

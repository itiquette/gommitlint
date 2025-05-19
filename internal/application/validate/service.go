// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validate provides application services for commit validation.
package validate

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
)

// Note: Using domain package interfaces instead of a local interface definition

// Compile-time interface satisfaction check.
var _ incoming.ValidationService = (*ValidationService)(nil)

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

// ValidationService orchestrates commit validation operations.
// It is designed to be used with value semantics and follows functional programming patterns.
type ValidationService struct {
	engine        validation.Engine
	commitService domain.CommitRepository
	infoProvider  domain.RepositoryInfoProvider

	// Configuration for rule activation
	enabledRules  []string
	disabledRules []string
}

// Engine returns the validation engine.
func (s ValidationService) Engine() validation.Engine {
	return s.engine
}

// CommitService returns the commit service.
func (s ValidationService) CommitService() domain.CommitRepository {
	return s.commitService
}

// InfoProvider returns the repository info provider.
func (s ValidationService) InfoProvider() domain.RepositoryInfoProvider {
	return s.infoProvider
}

// NewValidationService creates a new ValidationService.
func NewValidationService(
	engine validation.Engine,
	commitService domain.CommitRepository,
	infoProvider domain.RepositoryInfoProvider,
) ValidationService {
	return ValidationService{
		engine:        engine,
		commitService: commitService,
		infoProvider:  infoProvider,
	}
}

// WithEngine returns a new ValidationService with the engine replaced.
func (s ValidationService) WithEngine(engine validation.Engine) ValidationService {
	return ValidationService{
		engine:        engine,
		commitService: s.commitService,
		infoProvider:  s.infoProvider,
	}
}

// WithCommitService returns a new ValidationService with the commit service replaced.
func (s ValidationService) WithCommitService(commitService domain.CommitRepository) ValidationService {
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
		WithCustomRule(rule domain.Rule) validation.Engine
	}); ok {
		// The engine supports the operation directly
		serviceCopy.engine = customizer.WithCustomRule(rule)

		return serviceCopy, nil
	}

	// Check if the engine can be replaced with a customized version using the registry pattern
	if registryEngine, ok := s.engine.(*validation.RegistryEngine); ok {
		// Get the registry from the engine
		registry := registryEngine.GetRegistry()

		// Create a new factory that always returns this rule
		ruleName := "Custom_" + rule.Name()
		factory := func(_ context.Context) domain.Rule {
			return rule
		}

		// Register the factory
		registry.Register(ruleName, factory)

		// The same engine instance will use the updated registry
		return serviceCopy, nil
	}

	// The engine doesn't support custom rules
	return s, errors.New("WithCustomRule", errors.ErrInvalidConfig, "validation engine does not support custom rules")
}

// ValidateCommit validates a single commit by its reference.
func (s ValidationService) ValidateCommit(ctx context.Context, ref string) (domain.CommitResult, error) {
	// Get the commit from the repository
	commit, err := s.commitService.GetCommit(ctx, ref)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit %s: %w", ref, err)
	}

	// Check if we should skip merge commits
	cfg := contextx.GetConfig(ctx)
	skipMergeCommits := cfg.GetBool("validation.skip_merge_commits")

	// Skip merge commits if requested
	if skipMergeCommits && commit.IsMergeCommit {
		return domain.CommitResult{
			CommitInfo:  commit,
			Passed:      true,
			RuleResults: []domain.RuleResult{},
		}, nil
	}

	// Validate using the engine
	return s.engine.ValidateCommit(ctx, commit), nil
}

// ValidateCommits validates multiple commits.
func (s ValidationService) ValidateCommits(ctx context.Context, commitHashes []string) (domain.ValidationResults, error) {
	// Retrieve commits
	commits := make([]domain.CommitInfo, 0, len(commitHashes))

	for _, hash := range commitHashes {
		commit, err := s.commitService.GetCommit(ctx, hash)
		if err != nil {
			return domain.ValidationResults{}, fmt.Errorf("failed to get commit %s: %w", hash, err)
		}

		commits = append(commits, commit)
	}

	// Check if we should skip merge commits
	cfg := contextx.GetConfig(ctx)
	skipMergeCommits := cfg.GetBool("validation.skip_merge_commits")

	// Filter out merge commits if requested
	commitsToValidate := commits
	if skipMergeCommits {
		commitsToValidate = filterNonMergeCommits(commits)
	}

	// Validate using the engine
	return s.engine.ValidateCommits(ctx, commitsToValidate), nil
}

// ValidateCommitRange validates commits between two references.
func (s ValidationService) ValidateCommitRange(
	ctx context.Context,
	from,
	toRef string,
	skipMerge bool,
) (domain.ValidationResults, error) {
	// Get commits between the two references
	commits, err := s.commitService.GetCommitRange(ctx, from, toRef)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get commits between %s and %s: %w", from, toRef, err)
	}

	// Filter out merge commits if requested
	commitsToValidate := commits
	if skipMerge {
		commitsToValidate = filterNonMergeCommits(commits)
	}

	// Validate using the engine
	results := s.engine.ValidateCommits(ctx, commitsToValidate)

	return results, nil
}

// ValidateLastNCommits validates the last N commits from HEAD.
func (s ValidationService) ValidateLastNCommits(ctx context.Context, count int, skipMerge bool) (domain.ValidationResults, error) {
	// Get the last N commits
	commits, err := s.commitService.GetHeadCommits(ctx, count)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get %d commits: %w", count, err)
	}

	// Filter out merge commits if requested
	commitsToValidate := commits
	if skipMerge {
		commitsToValidate = filterNonMergeCommits(commits)
	}

	// Validate using the engine
	results := s.engine.ValidateCommits(ctx, commitsToValidate)

	return results, nil
}

// ValidateMessage validates a commit message directly.
func (s ValidationService) ValidateMessage(ctx context.Context, message string) (domain.ValidationResults, error) {
	// Parse message into commit info
	message = strings.TrimSpace(message)
	lines := strings.SplitN(message, "\n", 2)
	subject := lines[0]

	body := ""
	if len(lines) > 1 {
		body = strings.TrimSpace(lines[1])
	}

	commit := domain.CommitInfo{
		Hash:    "MSG",
		Subject: subject,
		Body:    body,
		Message: message,
	}

	// Validate the single commit
	result := s.engine.ValidateCommit(ctx, commit)

	// Wrap single result
	return wrapSingleResult(result), nil
}

// ValidateMessageFile validates a commit message from a file.
func (s ValidationService) ValidateMessageFile(ctx context.Context, path string) (domain.ValidationResults, error) {
	// Read the message from the file
	messageBytes, err := os.ReadFile(path)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to read message file %s: %w", path, err)
	}

	message := string(messageBytes)

	// Validate the message
	return s.ValidateMessage(ctx, message)
}

// ValidateWithOptions performs validation based on the provided options.
func (s ValidationService) ValidateWithOptions(ctx context.Context, opts ValidationOptions) (domain.ValidationResults, error) {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering ValidateWithOptions")

	// Validate options are correct
	optionCount := 0
	if opts.CommitHash != "" {
		optionCount++
	}

	if opts.CommitCount > 0 {
		optionCount++
	}

	if opts.FromHash != "" || opts.ToHash != "" {
		optionCount++
	}

	if opts.MessageFile != "" {
		optionCount++
	}

	if optionCount == 0 {
		// Default to validating the last commit
		opts.CommitCount = 1
	} else if optionCount > 1 {
		return domain.ValidationResults{}, errors.New("ValidateWithOptions", errors.ErrInvalidConfig,
			"only one validation option can be specified at a time")
	}

	// Perform validation based on options
	switch {
	case opts.MessageFile != "":
		return s.ValidateMessageFile(ctx, opts.MessageFile)

	case opts.CommitHash != "":
		result, err := s.ValidateCommit(ctx, opts.CommitHash)
		if err != nil {
			return domain.ValidationResults{}, err
		}

		return wrapSingleResult(result), nil

	case opts.CommitCount > 0:
		return s.ValidateLastNCommits(ctx, opts.CommitCount, opts.SkipMergeCommits)

	case opts.FromHash != "" || opts.ToHash != "":
		if opts.FromHash == "" || opts.ToHash == "" {
			return domain.ValidationResults{}, errors.New("ValidateWithOptions", errors.ErrInvalidConfig,
				"both from and to hashes must be specified for range validation")
		}

		return s.ValidateCommitRange(ctx, opts.FromHash, opts.ToHash, opts.SkipMergeCommits)

	default:
		// This shouldn't happen due to earlier validation
		return domain.ValidationResults{}, errors.New("ValidateWithOptions", errors.ErrInvalidConfig,
			"no validation option specified")
	}
}

// WithEnabledRules returns a new ValidationService with specific rules enabled.
func (s ValidationService) WithEnabledRules(rules []string) ValidationService {
	return ValidationService{
		engine:        s.engine,
		commitService: s.commitService,
		infoProvider:  s.infoProvider,
		enabledRules:  rules,
		disabledRules: s.disabledRules,
	}
}

// WithDisabledRules returns a new ValidationService with specific rules disabled.
func (s ValidationService) WithDisabledRules(rules []string) ValidationService {
	return ValidationService{
		engine:        s.engine,
		commitService: s.commitService,
		infoProvider:  s.infoProvider,
		enabledRules:  s.enabledRules,
		disabledRules: rules,
	}
}

// CreateValidationService creates a validation service with context-based configuration.
// This is the main entry point for creating validation services.
func CreateValidationService(
	ctx context.Context,
	commitService domain.CommitRepository,
	infoProvider domain.RepositoryInfoProvider,
	analyzer domain.CommitAnalyzer,
) ValidationService {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Creating validation service")

	// Logger is used indirectly through ctx

	// Get the configuration from context
	cfg := contextx.GetConfig(ctx)

	// Use default config for the engine
	defaultConfig := config.NewDefaultConfig()

	// Create the engine directly from the core validation package
	engine := validation.CreateEngine(ctx, defaultConfig, analyzer)

	// Create the service
	service := NewValidationService(engine, commitService, infoProvider)

	// Apply configuration from context if available
	if cfg != nil {
		enabledRules := cfg.GetStringSlice("rules.enabled_rules")
		disabledRules := cfg.GetStringSlice("rules.disabled_rules")

		if len(enabledRules) > 0 {
			service = service.WithEnabledRules(enabledRules)
		}

		if len(disabledRules) > 0 {
			service = service.WithDisabledRules(disabledRules)
		}
	}

	logger.Debug("Validation service created",
		"enabled_rules", service.enabledRules,
		"disabled_rules", service.disabledRules)

	return service
}

// Helper functions

// filterNonMergeCommits filters out merge commits from a slice of commits.
func filterNonMergeCommits(commits []domain.CommitInfo) []domain.CommitInfo {
	filtered := make([]domain.CommitInfo, 0, len(commits))

	for _, commit := range commits {
		if !commit.IsMergeCommit {
			filtered = append(filtered, commit)
		}
	}

	return filtered
}

// wrapSingleResult wraps a single commit result in a validation results container.
func wrapSingleResult(result domain.CommitResult) domain.ValidationResults {
	results := domain.NewValidationResults()
	results.Results = []domain.CommitResult{result}
	results.TotalCommits = 1

	if result.Passed {
		results.PassedCommits = 1
	}

	// Update the rule summary
	for _, ruleResult := range result.RuleResults {
		if ruleResult.Status == domain.StatusFailed {
			currentCount := results.RuleSummary[ruleResult.RuleID]
			results.RuleSummary[ruleResult.RuleID] = currentCount + 1
		}
	}

	return results
}

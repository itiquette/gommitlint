// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validate provides application services for commit validation.
package validate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/core/validation"
	"github.com/itiquette/gommitlint/internal/domain"
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
type ValidationEngine interface {
	ValidateCommit(ctx context.Context, commit *domain.CommitInfo) domain.CommitResult
	ValidateCommits(ctx context.Context, commits []*domain.CommitInfo) *domain.ValidationResults
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
) *ValidationService {
	return &ValidationService{
		engine:        engine,
		commitService: commitService,
		infoProvider:  infoProvider,
	}
}

// ValidateCommit validates a single commit.
func (s *ValidationService) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
	// Get the commit from the git repository
	commit, err := s.commitService.GetCommit(ctx, hash)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Validate the commit
	return s.engine.ValidateCommit(ctx, commit), nil
}

// ValidateHeadCommits validates the specified number of commits from HEAD.
func (s *ValidationService) ValidateHeadCommits(ctx context.Context, count int, skipMerge bool) (*domain.ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.commitService.GetHeadCommits(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get head commits: %w", err)
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
func (s *ValidationService) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMerge bool) (*domain.ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.commitService.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit range: %w", err)
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
func (s *ValidationService) ValidateMessageFile(ctx context.Context, filePath string) (*domain.ValidationResults, error) {
	// Read the message file
	messageBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read message file: %w", err)
	}

	// Convert to string and trim whitespace
	message := strings.TrimSpace(string(messageBytes))
	if message == "" {
		return nil, errors.New("empty commit message")
	}

	// Split into subject and body
	subject, body := domain.SplitCommitMessage(message)

	// Create a dummy commit
	commit := &domain.CommitInfo{
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
func (s *ValidationService) ValidateWithOptions(ctx context.Context, opts ValidationOptions) (*domain.ValidationResults, error) {
	// Validate commit message file
	if opts.MessageFile != "" {
		return s.ValidateMessageFile(ctx, opts.MessageFile)
	}

	// Validate specific commit
	if opts.CommitHash != "" {
		result, err := s.ValidateCommit(ctx, opts.CommitHash)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	// Create validation results
	results := domain.NewValidationResults()
	results.AddCommitResult(result)

	return results, nil
}

// CreateDefaultValidationService creates a validation service with configuration
// loaded using the default config manager (`config.New()`).
func CreateDefaultValidationService(repoPath string) (*ValidationService, error) {
	// Create repository factory
	factory, err := git.NewRepositoryFactory(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository factory: %w", err)
	}

	// Get specialized repository interfaces
	commitService := factory.CreateGitCommitService()
	infoProvider := factory.CreateInfoProvider()
	analyzer := factory.CreateCommitAnalyzer() // Create analyzer here for potential injection

	// Create config manager - uses the new config.New() which loads defaults
	// and standard paths, logging warnings internally if loading fails.
	configManager, err := config.New()
	if err != nil {
		// config.New() only returns an error if manager creation fails, not just config not found.
		return nil, fmt.Errorf("failed to create configuration manager: %w", err)
	}

	// Get validation rule configuration from the manager
	// This uses the loaded config or defaults, no error handling needed here.
	ruleConfig := configManager.GetRuleConfig()

	// Create rule provider with configuration
	ruleProvider := validation.NewDefaultRuleProvider(ruleConfig)

	// Set repository getter for CommitsAhead rule if it exists
	if commitsAheadRule, ok := ruleProvider.GetRuleByName("CommitsAhead").(*rules.CommitsAheadRule); ok && commitsAheadRule != nil {
		commitsAheadRule.SetRepositoryGetter(func() domain.CommitAnalyzer {
			return analyzer
		})
	}

	// Create validation engine
	engine := validation.NewEngine(ruleProvider)

	// Create validation service
	return NewValidationService(engine, commitService, infoProvider), nil
}

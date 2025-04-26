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

// Service orchestrates commit validation operations.
type Service struct {
	Engine        Engine
	CommitService domain.GitCommitService
	InfoProvider  domain.RepositoryInfoProvider
}

// NewService creates a new validation service.
func NewService(
	engine Engine,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
) Service {
	return Service{
		Engine:        engine,
		CommitService: commitService,
		InfoProvider:  infoProvider,
	}
}

// ValidateCommit validates a single commit.
func (s Service) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
	// Get the commit from the git repository
	commit, err := s.CommitService.GetCommit(ctx, hash)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Validate the commit
	return s.Engine.ValidateCommit(ctx, commit), nil
}

// ValidateHeadCommits validates the specified number of commits from HEAD.
func (s Service) ValidateHeadCommits(ctx context.Context, count int, skipMerge bool) (domain.ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.CommitService.GetHeadCommits(ctx, count)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get head commits: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := domain.NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Validate the commits
	return s.Engine.ValidateCommits(ctx, collection.All()), nil
}

// ValidateCommitRange validates all commits in the given range.
func (s Service) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMerge bool) (domain.ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.CommitService.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return domain.ValidationResults{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := domain.NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Validate the commits
	return s.Engine.ValidateCommits(ctx, collection.All()), nil
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
	result := s.Engine.ValidateCommit(ctx, commit)

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

// CreateServiceWithAdapter creates a ValidationService using the provided repository adapter and configuration.
// This is a simpler approach that doesn't require the factory pattern.
func CreateServiceWithAdapter(repoAdapter *git.RepositoryAdapter, config Config) Service {
	// Create rule provider with configuration AND analyzer
	ruleProvider := NewRuleProvider(config, repoAdapter)

	// Create validation engine
	engine := NewEngine(ruleProvider)

	// Create validation service
	return NewService(engine, repoAdapter, repoAdapter)
}

// CreateServiceWithDependencies creates a ValidationService with explicit dependencies.
// This is the preferred constructor for better testability.
func CreateServiceWithDependencies(
	engine Engine,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
) Service {
	return NewService(engine, commitService, infoProvider)
}

// CreateServiceWithAnalyzer creates a ValidationService with an analyzer and configuration.
// This combines the dependency and configuration approach.
func CreateServiceWithAnalyzer(
	config Config,
	analyzer domain.CommitAnalyzer,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
) Service {
	// Create rule provider with configuration and analyzer
	ruleProvider := NewRuleProvider(config, analyzer)

	// Create validation engine
	engine := NewEngine(ruleProvider)

	// Create and return validation service
	return NewService(engine, commitService, infoProvider)
}

// FactoryWithConfig creates a validation service factory function that uses
// the provided configuration. This is useful for testing or for
// programmatically configuring the validation service.
func FactoryWithConfig(config Config) func(string) (Service, error) {
	return func(repoPath string) (Service, error) {
		// Create repository adapter
		repoAdapter, err := git.NewRepositoryAdapter(repoPath)
		if err != nil {
			return Service{}, fmt.Errorf("failed to create repository adapter: %w", err)
		}

		// Create rule provider with configuration AND analyzer
		ruleProvider := NewRuleProvider(config, repoAdapter)

		// Create validation engine
		engine := NewEngine(ruleProvider)

		// Create validation service
		return NewService(engine, repoAdapter, repoAdapter), nil
	}
}

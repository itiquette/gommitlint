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

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/git"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
)

// Note: Engine implementation and methods are now defined in engine.go

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
	CommitService domain.CommitRepository

	// InfoProvider for repository information
	InfoProvider domain.RepositoryInfoProvider

	// Analyzer for advanced repository analysis
	Analyzer domain.CommitAnalyzer
}

// Service provides validation using the configuration system.
// It is designed with value semantics for functional programming.
type Service struct {
	dependencies ServiceDependencies
	config       types.Config
}

// NewService creates a new Service.
func NewService(deps ServiceDependencies, cfg types.Config) Service {
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
func (s Service) WithConfig(cfg types.Config) Service {
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
	return s.dependencies.Engine.ValidateCommits(ctx, []domain.CommitInfo(collection)), nil
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
	return s.dependencies.Engine.ValidateCommits(ctx, []domain.CommitInfo(collection)), nil
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
	results = results.WithResult(result)

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
		results = results.WithResult(result)

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

	results = results.WithResult(result)

	return results, nil
}

// CreateService creates a validation service with the configuration.
// This now uses context directly for configuration.
func CreateService(ctx context.Context, config types.Config, repoPath string) (Service, error) {
	// Create the repository adapter
	repoAdapter, err := git.NewRepositoryAdapter(ctx, repoPath)
	if err != nil {
		return Service{}, fmt.Errorf("failed to create repository adapter: %w", err)
	}

	// Create engine using rule provider and context
	engine := CreateEngine(ctx, config, repoAdapter)

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

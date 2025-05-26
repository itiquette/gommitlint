// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	internalErrors "github.com/itiquette/gommitlint/internal/errors"
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
	Engine domain.ValidationEngine

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

// ValidateCommit validates a single commit.
func (s Service) ValidateCommit(ctx context.Context, hash string, skipMergeCommits bool) (domain.CommitResult, error) {
	// Get the commit from the git repository
	commit, err := s.dependencies.CommitService.GetCommit(ctx, hash)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Skip merge commits if requested
	if skipMergeCommits && commit.IsMergeCommit {
		return domain.CommitResult{
			CommitInfo:   commit,
			RuleResults:  []domain.RuleResult{},
			RuleErrorMap: make(map[string][]internalErrors.ValidationError),
		}, nil
	}

	// Validate the commit
	return s.dependencies.Engine.ValidateCommit(ctx, commit), nil
}

// ValidateCommits validates multiple commits by their hashes.
func (s Service) ValidateCommits(ctx context.Context, commitHashes []string, skipMergeCommits bool) (domain.ValidationResults, error) {
	allResults := domain.NewValidationResults()

	for _, hash := range commitHashes {
		result, err := s.ValidateCommit(ctx, hash, skipMergeCommits)
		if err != nil {
			return allResults, err
		}

		allResults = allResults.WithResult(result)
	}

	return allResults, nil
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
	return s.dependencies.Engine.ValidateCommits(ctx, collection), nil
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
	return s.dependencies.Engine.ValidateCommits(ctx, collection), nil
}

// ValidateMessage validates a commit message string.
func (s Service) ValidateMessage(ctx context.Context, message string) (domain.ValidationResults, error) {
	// Trim whitespace
	message = strings.TrimSpace(message)
	if message == "" {
		return domain.NewValidationResults(), errors.New("empty commit message")
	}

	// Split into subject and body
	subject, body := domain.SplitCommitMessage(message)

	// Create a commit info for validation
	commit := domain.CommitInfo{
		Hash:          "0000000000000000000000000000000000000000",
		Subject:       subject,
		Body:          body,
		Message:       message,
		IsMergeCommit: false,
	}

	// Validate the commit
	result := s.dependencies.Engine.ValidateCommit(ctx, commit)

	// Create validation results using functional approach
	return domain.NewValidationResults().WithResult(result), nil
}

// ValidateWithOptions validates commits according to the provided options.
func (s Service) ValidateWithOptions(ctx context.Context, opts Options) (domain.ValidationResults, error) {
	// Message file validation should be handled by the caller
	if opts.MessageFile != "" {
		return domain.NewValidationResults(), errors.New("message file validation should use ValidateMessage after reading file content")
	}

	// Validate specific commit
	if opts.CommitHash != "" {
		result, err := s.ValidateCommit(ctx, opts.CommitHash, opts.SkipMergeCommits)
		if err != nil {
			return domain.NewValidationResults(), err
		}

		return domain.NewValidationResults().WithResult(result), nil
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
	result, err := s.ValidateCommit(ctx, "HEAD", opts.SkipMergeCommits)
	if err != nil {
		return domain.NewValidationResults(), err
	}

	return domain.NewValidationResults().WithResult(result), nil
}

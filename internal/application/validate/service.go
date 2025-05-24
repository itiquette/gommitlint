// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validate provides application services for commit validation.
package validate

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
)

// ValidationService implements the business logic for commit validation.
// It uses domain interfaces and does not depend on external layers.

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
// Rule activation is controlled via context configuration, not stored in this struct.
type ValidationService struct {
	engine        domain.ValidationEngine
	commitService domain.CommitRepository
	infoProvider  domain.RepositoryInfoProvider
}

// NewValidationService creates a new ValidationService.
func NewValidationService(
	engine domain.ValidationEngine,
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
func (s ValidationService) WithEngine(engine domain.ValidationEngine) ValidationService {
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

// ValidateCommit validates a single commit by its reference.
func (s ValidationService) ValidateCommit(ctx context.Context, ref string, skipMergeCommits bool) (domain.CommitResult, error) {
	// Get the commit from the repository
	commit, err := s.commitService.GetCommit(ctx, ref)
	if err != nil {
		return domain.CommitResult{}, fmt.Errorf("failed to get commit %s: %w", ref, err)
	}

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
func (s ValidationService) ValidateCommits(ctx context.Context, commitHashes []string, skipMergeCommits bool) (domain.ValidationResults, error) {
	// Retrieve commits
	commits := make([]domain.CommitInfo, 0, len(commitHashes))

	for _, hash := range commitHashes {
		commit, err := s.commitService.GetCommit(ctx, hash)
		if err != nil {
			return domain.ValidationResults{}, fmt.Errorf("failed to get commit %s: %w", hash, err)
		}

		commits = append(commits, commit)
	}

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

// ValidateWithOptions performs validation based on the provided options.
func (s ValidationService) ValidateWithOptions(ctx context.Context, opts ValidationOptions) (domain.ValidationResults, error) {
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
		// This should be handled by the caller reading the file
		// and calling ValidateMessage instead
		return domain.ValidationResults{}, errors.New("ValidateWithOptions", errors.ErrInvalidConfig,
			"message file validation should use ValidateMessage after reading file content")

	case opts.CommitHash != "":
		result, err := s.ValidateCommit(ctx, opts.CommitHash, opts.SkipMergeCommits)
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
	return domain.NewValidationResults().WithResult(result)
}

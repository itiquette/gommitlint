// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// Service provides validation using the configuration system.
// It is designed with value semantics for functional programming.
type Service struct {
	repo  Repository
	rules []Rule
}

// NewService creates a new Service.
func NewService(repo Repository, rules []Rule) Service {
	return Service{
		repo:  repo,
		rules: rules,
	}
}

// ValidateCommit validates a single commit.
func (s Service) ValidateCommit(ctx context.Context, hash string, skipMergeCommits bool) (CommitResult, error) {
	// Get the commit from the git repository
	commit, err := s.repo.GetCommit(ctx, hash)
	if err != nil {
		return CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Skip merge commits if requested
	if skipMergeCommits && commit.IsMergeCommit {
		return CommitResult{
			CommitInfo:  commit,
			RuleResults: []RuleResult{},
			Passed:      true,
		}, nil
	}

	// Validate the commit
	return s.validateCommitWithRules(ctx, commit, s.rules), nil
}

// ValidateCommits validates multiple commits by their hashes.
func (s Service) ValidateCommits(ctx context.Context, commitHashes []string, skipMergeCommits bool) (ValidationResults, error) {
	allResults := NewValidationResults()

	for _, hash := range commitHashes {
		result, err := s.ValidateCommit(ctx, hash, skipMergeCommits)
		if err != nil {
			return allResults, err
		}

		allResults = allResults.AddResult(result)
	}

	return allResults, nil
}

// ValidateHeadCommits validates the specified number of commits from HEAD.
func (s Service) ValidateHeadCommits(ctx context.Context, count int, skipMerge bool) (ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.repo.GetHeadCommits(ctx, count)
	if err != nil {
		return ValidationResults{}, fmt.Errorf("failed to get head commits: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Validate the commits
	return s.validateCommitsWithRules(ctx, []CommitInfo(collection)), nil
}

// ValidateCommitRange validates all commits in the given range.
func (s Service) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMerge bool) (ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.repo.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return ValidationResults{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Use CommitCollection to filter merge commits if requested
	collection := NewCommitCollection(commits)
	if skipMerge {
		collection = collection.FilterMergeCommits()
	}

	// Validate the commits
	return s.validateCommitsWithRules(ctx, []CommitInfo(collection)), nil
}

// ValidateMessage validates a commit message string.
func (s Service) ValidateMessage(ctx context.Context, message string) (ValidationResults, error) {
	// Trim whitespace
	message = strings.TrimSpace(message)
	if message == "" {
		return NewValidationResults(), errors.New("empty commit message")
	}

	// Split into subject and body
	subject, body := SplitCommitMessage(message)

	// Create a commit info for validation
	commit := CommitInfo{
		Hash:          "0000000000000000000000000000000000000000",
		Subject:       subject,
		Body:          body,
		Message:       message,
		IsMergeCommit: false,
	}

	// Validate the commit
	result := s.validateCommitWithRules(ctx, commit, s.rules)

	// Create validation results using functional approach
	return NewValidationResults().AddResult(result), nil
}

// ValidateWithOptions validates commits according to the provided options.
func (s Service) ValidateWithOptions(ctx context.Context, opts Options) (ValidationResults, error) {
	// Message file validation should be handled by the caller
	if opts.MessageFile != "" {
		return NewValidationResults(), errors.New("message file validation should use ValidateMessage after reading file content")
	}

	// Validate specific commit
	if opts.CommitHash != "" {
		result, err := s.ValidateCommit(ctx, opts.CommitHash, opts.SkipMergeCommits)
		if err != nil {
			return NewValidationResults(), err
		}

		return NewValidationResults().AddResult(result), nil
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
		return NewValidationResults(), err
	}

	return NewValidationResults().AddResult(result), nil
}

// validateCommitWithRules validates a commit against a set of rules.
func (s Service) validateCommitWithRules(ctx context.Context, commit CommitInfo, rules []Rule) CommitResult {
	// Handle no active rules case
	if len(rules) == 0 {
		return NewCommitResult(commit)
	}

	// Initialize result
	result := NewCommitResult(commit)

	// Validate against each rule
	for _, rule := range rules {
		ruleErrors := rule.Validate(ctx, commit)

		// Add result using the simplified method
		result = result.AddRuleResult(rule.Name(), ruleErrors)
	}

	return result
}

// validateCommitsWithRules validates multiple commits against all rules.
func (s Service) validateCommitsWithRules(ctx context.Context, commits []CommitInfo) ValidationResults {
	// Create results
	results := NewValidationResults()

	for _, commit := range commits {
		commitResult := s.validateCommitWithRules(ctx, commit, s.rules)
		results = results.AddResult(commitResult)
	}

	return results
}

// Logger provides logging capabilities needed by the validation service.
// This interface is defined here following dependency inversion principle.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

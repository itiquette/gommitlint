// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
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

// Service provides validation using the simplified domain model.
type Service struct {
	repo            Repository
	validator       Validator
	rules           []Rule
	commitRules     []Rule         // Rules that validate individual commits
	repositoryRules []Rule         // Rules that validate repository state
	config          *config.Config // Configuration for rules
}

// NewService creates a new Service.
func NewService(repo Repository, rules []Rule) Service {
	// Separate commit-level and repository-level rules
	commitRules, repositoryRules := SeparateRules(rules)

	return Service{
		repo:            repo,
		validator:       NewValidator(commitRules), // Only commit rules for validator
		rules:           rules,
		commitRules:     commitRules,
		repositoryRules: repositoryRules,
		config:          nil, // Will be set by WithConfig
	}
}

// WithConfig sets the configuration for the service.
func (s Service) WithConfig(cfg *config.Config) Service {
	s.config = cfg

	return s
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
			Commit:      commit,
			RuleResults: []RuleResult{},
			Passed:      true,
		}, nil
	}

	// Validate the commit (only commit-level rules)
	return s.validateCommitWithRules(ctx, commit, s.commitRules), nil
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

	// Filter merge commits if requested
	if skipMerge {
		filteredCommits := make([]Commit, 0, len(commits))

		for _, commit := range commits {
			if !commit.IsMergeCommit {
				filteredCommits = append(filteredCommits, commit)
			}
		}

		commits = filteredCommits
	}

	// Validate the commits
	return s.validateCommitsWithRules(ctx, commits), nil
}

// ValidateCommitRange validates all commits in the given range.
func (s Service) ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMerge bool) (ValidationResults, error) {
	// Get the commits from the git repository
	commits, err := s.repo.GetCommitRange(ctx, fromHash, toHash)
	if err != nil {
		return ValidationResults{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Filter merge commits if requested
	if skipMerge {
		filteredCommits := make([]Commit, 0, len(commits))

		for _, commit := range commits {
			if !commit.IsMergeCommit {
				filteredCommits = append(filteredCommits, commit)
			}
		}

		commits = filteredCommits
	}

	// Validate the commits
	return s.validateCommitsWithRules(ctx, commits), nil
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

	// Create a commit for validation
	commit := Commit{
		Hash:          "0000000000000000000000000000000000000000",
		Subject:       subject,
		Body:          body,
		Message:       message,
		IsMergeCommit: false,
	}

	// Validate the commit
	result := s.validateCommitWithRules(ctx, commit, s.commitRules)

	// Create validation results using functional approach
	return NewValidationResults().AddResult(result), nil
}

// ValidateWithOptions validates commits according to the provided options.
func (s Service) ValidateWithOptions(ctx context.Context, opts Options) (ValidationResults, error) {
	// Initialize results
	var results = NewValidationResults()

	// Check repository-level rules and collect them separately
	repoFailures := s.checkRepositoryRules(ctx)

	repoResults := make([]RuleResult, 0, len(repoFailures))

	for _, failure := range repoFailures {
		repoResult := RuleResult{
			RuleName: failure.Rule,
			Status:   StatusFailed,
			Errors:   []ValidationError{New(failure.Rule, "RULE_FAILED", failure.Message)},
		}
		repoResults = append(repoResults, repoResult)
	}

	// Message file validation should be handled by the caller
	if opts.MessageFile != "" {
		return NewValidationResults(), errors.New("message file validation should use ValidateMessage after reading file content")
	}

	// Validate specific commit
	if opts.CommitHash != "" {
		result, err := s.ValidateCommit(ctx, opts.CommitHash, opts.SkipMergeCommits)
		if err != nil {
			return results, err
		}

		results = results.AddResult(result)
	} else if opts.FromHash != "" || opts.ToHash != "" {
		// Validate commit range
		commitResults, err := s.ValidateCommitRange(ctx, opts.FromHash, opts.ToHash, opts.SkipMergeCommits)
		if err != nil {
			return results, err
		}
		// Merge commit results
		for _, result := range commitResults.Results {
			results = results.AddResult(result)
		}
	} else if opts.CommitCount > 0 {
		// Validate head commits
		commitResults, err := s.ValidateHeadCommits(ctx, opts.CommitCount, opts.SkipMergeCommits)
		if err != nil {
			return results, err
		}
		// Merge commit results
		for _, result := range commitResults.Results {
			results = results.AddResult(result)
		}
	} else {
		// Default to validating the HEAD commit
		result, err := s.ValidateCommit(ctx, "HEAD", opts.SkipMergeCommits)
		if err != nil {
			return results, err
		}

		results = results.AddResult(result)
	}

	// Add repository results at the end
	results = results.AddRepositoryResults(repoResults)

	return results, nil
}

// validateCommitWithRules validates a commit against a set of rules.
func (s Service) validateCommitWithRules(_ context.Context, commit Commit, rules []Rule) CommitResult {
	// Handle no active rules case
	if len(rules) == 0 {
		return NewCommitResult(commit)
	}

	// Initialize result
	result := NewCommitResult(commit)

	// Create validation context
	ctx := ValidationContext{
		Commit:     commit,
		Repository: s.repo,
		Config:     s.config,
	}

	// Validate against each rule using new unified interface
	for _, rule := range rules {
		// Use new unified interface - all rules now implement the same interface
		ruleFailures := rule.Validate(ctx)

		// Convert RuleFailures to ValidationErrors
		var errors []ValidationError
		for _, failure := range ruleFailures {
			errors = append(errors, New(failure.Rule, "RULE_FAILED", failure.Message))
		}

		// Add result
		result = result.AddRuleResult(rule.Name(), errors)
	}

	return result
}

// validateCommitsWithRules validates multiple commits against all rules.
func (s Service) validateCommitsWithRules(ctx context.Context, commits []Commit) ValidationResults {
	// Create results
	results := NewValidationResults()

	for _, commit := range commits {
		commitResult := s.validateCommitWithRules(ctx, commit, s.commitRules)
		results = results.AddResult(commitResult)
	}

	return results
}

// checkRepositoryRules checks repository-level rules.
func (s Service) checkRepositoryRules(_ context.Context) []RuleFailure {
	var failures []RuleFailure

	// Create validation context for repository-level rules
	validationCtx := ValidationContext{
		Commit:     Commit{}, // Empty commit for repository-level validation
		Repository: s.repo,
		Config:     s.config,
	}

	// Run repository-level rules
	for _, rule := range s.repositoryRules {
		ruleFailures := rule.Validate(validationCtx)
		failures = append(failures, ruleFailures...)
	}

	return failures
}

// Logger provides logging capabilities needed by the validation service.
// This interface is defined here following dependency inversion principle.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package application provides the application layer commands that orchestrate domain logic.
package application

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// CreateValidator creates a validator with all dependencies.
func CreateValidator(repo domain.Repository, cfg *config.Config, logger domain.Logger) domain.ValidatorWithDeps {
	rules := CreateEnabledRules(cfg)
	deps := domain.Dependencies{
		Repository: repo,
		Config:     cfg,
		Logger:     logger,
	}

	return domain.NewValidatorWithDeps(rules, deps)
}

// ValidateCommitWithValidator validates a single commit using the validator.
func ValidateCommitWithValidator(ctx context.Context, hash string, validator domain.ValidatorWithDeps) (domain.ValidationResult, error) {
	commit, err := validator.Deps.Repository.GetCommit(ctx, hash)
	if err != nil {
		return domain.ValidationResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	return validator.ValidateCommit(commit), nil
}

// ValidateMessageWithValidator validates a commit message using the validator.
func ValidateMessageWithValidator(message string, validator domain.ValidatorWithDeps) (domain.ValidationResult, error) {
	return validator.ValidateMessage(message)
}

// ValidateHeadCommitsWithValidator validates N commits from HEAD using the validator.
func ValidateHeadCommitsWithValidator(ctx context.Context, count int, skipMerge bool, validator domain.ValidatorWithDeps) ([]domain.ValidationResult, error) {
	commits, err := validator.Deps.Repository.GetHeadCommits(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get head commits: %w", err)
	}

	// Filter merge commits if requested
	commits = domain.FilterMergeCommits(commits, skipMerge)

	return validator.ValidateCommits(commits), nil
}

// ValidateCommitRangeWithValidator validates commits in a range using the validator.
func ValidateCommitRangeWithValidator(ctx context.Context, from, to string, skipMerge bool, validator domain.ValidatorWithDeps) ([]domain.ValidationResult, error) {
	commits, err := validator.Deps.Repository.GetCommitRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Filter merge commits if requested
	commits = domain.FilterMergeCommits(commits, skipMerge)

	return validator.ValidateCommits(commits), nil
}

// ValidateSingleCommit validates one commit by its hash.
func ValidateSingleCommit(ctx context.Context, hash string, repo domain.Repository, rules []domain.Rule, cfg *config.Config) (domain.ValidationResult, error) {
	commit, err := repo.GetCommit(ctx, hash)
	if err != nil {
		return domain.ValidationResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	commitRules := domain.FilterCommitRules(rules)

	return domain.ValidateCommit(commit, commitRules, repo, cfg), nil
}

// ValidateCommitRange validates all commits in the given range.
func ValidateCommitRange(ctx context.Context, from, to string, skipMerge bool, repo domain.Repository, rules []domain.Rule, cfg *config.Config) ([]domain.ValidationResult, error) {
	commits, err := repo.GetCommitRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Filter merge commits if requested
	commits = domain.FilterMergeCommits(commits, skipMerge)

	commitRules := domain.FilterCommitRules(rules)

	return domain.ValidateCommits(commits, commitRules, repo, cfg), nil
}

// ValidateHeadCommits validates N commits from HEAD.
func ValidateHeadCommits(ctx context.Context, count int, skipMerge bool, repo domain.Repository, rules []domain.Rule, cfg *config.Config) ([]domain.ValidationResult, error) {
	commits, err := repo.GetHeadCommits(ctx, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get head commits: %w", err)
	}

	// Filter merge commits if requested
	commits = domain.FilterMergeCommits(commits, skipMerge)

	commitRules := domain.FilterCommitRules(rules)

	return domain.ValidateCommits(commits, commitRules, repo, cfg), nil
}

// ValidateMessage validates a commit message string.
func ValidateMessage(message string, rules []domain.Rule, cfg *config.Config) (domain.ValidationResult, error) {
	return domain.ValidateCommitMessage(message, rules, cfg)
}

// ValidateWithRepository validates HEAD commit and includes repository checks.
func ValidateWithRepository(ctx context.Context, skipMerge bool, repo domain.Repository, rules []domain.Rule, cfg *config.Config) (domain.FullValidation, error) {
	// Get HEAD commit
	commit, err := repo.GetCommit(ctx, "HEAD")
	if err != nil {
		return domain.FullValidation{}, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Skip if it's a merge commit and skipMerge is true
	if skipMerge && commit.IsMergeCommit {
		return domain.FullValidation{
			CommitResults: []domain.ValidationResult{
				{
					Commit:   commit,
					Failures: nil,
				},
			},
			RepositoryFailures: nil,
		}, nil
	}

	// Validate commit
	commitRules := domain.FilterCommitRules(rules)
	commitResult := domain.ValidateCommit(commit, commitRules, repo, cfg)

	// Validate repository
	repoFailures := domain.ValidateRepository(rules, repo, cfg)

	return domain.FullValidation{
		CommitResults:      []domain.ValidationResult{commitResult},
		RepositoryFailures: repoFailures,
	}, nil
}

// ValidateCommitWithSkipMerge validates a single commit with merge skip option.
func ValidateCommitWithSkipMerge(ctx context.Context, hash string, skipMerge bool, repo domain.Repository, rules []domain.Rule, cfg *config.Config) (domain.ValidationResult, error) {
	commit, err := repo.GetCommit(ctx, hash)
	if err != nil {
		return domain.ValidationResult{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Skip validation if it's a merge commit and skipMerge is true
	if skipMerge && commit.IsMergeCommit {
		return domain.ValidationResult{
			Commit:   commit,
			Failures: nil,
		}, nil
	}

	commitRules := domain.FilterCommitRules(rules)

	return domain.ValidateCommit(commit, commitRules, repo, cfg), nil
}

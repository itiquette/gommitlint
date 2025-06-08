// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// ValidateTarget orchestrates validation by coordinating I/O with validation logic.
func ValidateTarget(ctx context.Context, target ValidationTarget, commitRules []domain.CommitRule,
	repoRules []domain.RepositoryRule, repo domain.Repository, cfg config.Config, logger domain.Logger) (domain.Report, error) {
	logger.Debug("Starting validation", "target_type", target.Type)

	switch target.Type {
	case "message":
		return executeMessageValidation(target.Source, commitRules, cfg, logger)
	case "commit":
		return executeCommitValidation(ctx, target.Source, commitRules, repoRules, repo, cfg, logger)
	case "range":
		return executeRangeValidation(ctx, target.Source, target.Target, commitRules, repoRules, repo, cfg, logger)
	case "count":
		return executeCountValidation(ctx, target.Source, commitRules, repoRules, repo, cfg, logger)
	default:
		return domain.Report{}, fmt.Errorf("unknown validation target type: %s", target.Type)
	}
}

// executeMessageValidation handles message file validation.
func executeMessageValidation(filePath string, rules []domain.CommitRule, cfg config.Config, logger domain.Logger) (domain.Report, error) {
	logger.Debug("Validating message from file", "path", filePath)

	// Read file
	message, err := readMessageFile(filePath)
	if err != nil {
		return domain.Report{}, err
	}

	// Validate message
	return ValidateMessageContent(message, rules, cfg)
}

// executeCommitValidation handles single commit validation.
func executeCommitValidation(ctx context.Context, ref string, commitRules []domain.CommitRule,
	repoRules []domain.RepositoryRule, repo domain.Repository, cfg config.Config, logger domain.Logger) (domain.Report, error) {
	select {
	case <-ctx.Done():
		return domain.Report{}, ctx.Err()
	default:
		logger.Debug("Validating commit", "ref", ref)
	}

	// Fetch commit from repository
	commit, err := repo.GetCommit(ctx, ref)
	if err != nil {
		return domain.Report{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Validate commit
	return ValidateSingleCommit(commit, commitRules, repoRules, repo, cfg)
}

// executeRangeValidation handles commit range validation.
func executeRangeValidation(ctx context.Context, fromRef, toRef string, commitRules []domain.CommitRule,
	repoRules []domain.RepositoryRule, repo domain.Repository, cfg config.Config, logger domain.Logger) (domain.Report, error) {
	select {
	case <-ctx.Done():
		return domain.Report{}, ctx.Err()
	default:
		logger.Debug("Validating commit range", "from", fromRef, "to", toRef)
	}

	// Fetch commits from repository
	commits, err := repo.GetCommitRange(ctx, fromRef, toRef)
	if err != nil {
		return domain.Report{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Validate commits
	return ValidateMultipleCommits(commits, commitRules, repoRules, repo, cfg)
}

// executeCountValidation handles commit count validation.
func executeCountValidation(ctx context.Context, countStr string, commitRules []domain.CommitRule,
	repoRules []domain.RepositoryRule, repo domain.Repository, cfg config.Config, logger domain.Logger) (domain.Report, error) {
	select {
	case <-ctx.Done():
		return domain.Report{}, ctx.Err()
	default:
		logger.Debug("Validating commit count", "count", countStr)
	}

	// Parse count
	count, err := parseCommitCount(countStr)
	if err != nil {
		return domain.Report{}, err
	}

	if count == 1 {
		// Single commit validation
		return executeCommitValidation(ctx, "HEAD", commitRules, repoRules, repo, cfg, logger)
	}

	// Multiple commits - delegate to range validation
	fromRef := fmt.Sprintf("HEAD~%d", count-1)

	return executeRangeValidation(ctx, fromRef, "HEAD", commitRules, repoRules, repo, cfg, logger)
}

// ValidateMessageContent validates a message string.
func ValidateMessageContent(message string, rules []domain.CommitRule, cfg config.Config) (domain.Report, error) {
	result, err := domain.ValidateMessage(message, rules, cfg)
	if err != nil {
		return domain.Report{}, fmt.Errorf("failed to validate message: %w", err)
	}

	return domain.BuildReport([]domain.ValidationResult{result}, nil, rules, nil, domain.ReportOptions{}), nil
}

// ValidateSingleCommit validates one commit.
func ValidateSingleCommit(commit domain.Commit, commitRules []domain.CommitRule, repoRules []domain.RepositoryRule,
	repo domain.Repository, cfg config.Config) (domain.Report, error) {
	// Always skip merge commits
	if commit.IsMergeCommit {
		emptyResult := domain.ValidationResult{Commit: commit, Errors: nil}

		return domain.BuildReport([]domain.ValidationResult{emptyResult}, nil, commitRules, repoRules, domain.ReportOptions{}), nil
	}

	// Validate using domain functions
	validationResult := domain.ValidateCommit(commit, commitRules, repoRules, repo, cfg)
	repoErrors := domain.ValidateRepository(repoRules, repo, cfg)

	return domain.BuildReport([]domain.ValidationResult{validationResult}, repoErrors, commitRules, repoRules, domain.ReportOptions{}), nil
}

// ValidateMultipleCommits validates multiple commits.
func ValidateMultipleCommits(commits []domain.Commit, commitRules []domain.CommitRule, repoRules []domain.RepositoryRule,
	repo domain.Repository, cfg config.Config) (domain.Report, error) {
	// Always filter out merge commits
	filteredCommits := domain.FilterMergeCommits(commits)

	// Validate using domain functions
	validationResults := domain.ValidateCommits(filteredCommits, commitRules, repoRules, repo, cfg)
	repoErrors := domain.ValidateRepository(repoRules, repo, cfg)

	return domain.BuildReport(validationResults, repoErrors, commitRules, repoRules, domain.ReportOptions{}), nil
}

// readMessageFile reads message from file or stdin.
func readMessageFile(filePath string) (string, error) {
	// Handle stdin case
	if filePath == "-" {
		message, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}

		return string(message), nil
	}

	// Handle regular file case
	message, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read message file: %w", err)
	}

	return string(message), nil
}

// parseCommitCount parses commit count string.
func parseCommitCount(countStr string) (int, error) {
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid commit count: %w", err)
	}

	if count <= 0 {
		return 0, fmt.Errorf("commit count must be positive: %d", count)
	}

	return count, nil
}

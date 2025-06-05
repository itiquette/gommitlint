// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// validateTarget validates based on target type using pure domain functions.
// This follows functional programming principles with explicit dependencies.
func validateTarget(ctx context.Context, target ValidationTarget, rules []domain.Rule,
	repo domain.Repository, cfg *config.Config, skipMergeCommits bool) (domain.Report, error) {
	switch target.Type {
	case "message":
		return validateMessageTarget(target.Source, rules, cfg)
	case "commit":
		return validateCommitTarget(ctx, target.Source, rules, repo, cfg, skipMergeCommits)
	case "range":
		return validateRangeTarget(ctx, target.Source, target.Target, rules, repo, cfg, skipMergeCommits)
	case "count":
		return validateCountTarget(ctx, target.Source, rules, repo, cfg, skipMergeCommits)
	default:
		return domain.Report{}, fmt.Errorf("unknown validation target type: %s", target.Type)
	}
}

// validateMessageTarget validates a commit message from file (pure function).
func validateMessageTarget(filePath string, rules []domain.Rule, cfg *config.Config) (domain.Report, error) {
	message, err := os.ReadFile(filePath)
	if err != nil {
		return domain.Report{}, fmt.Errorf("failed to read message file: %w", err)
	}

	result, err := domain.ValidateMessage(string(message), rules, cfg)
	if err != nil {
		return domain.Report{}, fmt.Errorf("failed to validate message: %w", err)
	}

	// Build report using simplified domain function with all rules
	commitRules := domain.FilterCommitRules(rules)

	return domain.BuildReportWithRules([]domain.ValidationResult{result}, nil, commitRules, domain.ReportOptions{}), nil
}

// validateCommitTarget validates a single commit (pure function).
func validateCommitTarget(ctx context.Context, ref string, rules []domain.Rule,
	repo domain.Repository, cfg *config.Config, skipMergeCommits bool) (domain.Report, error) {
	commit, err := repo.GetCommit(ctx, ref)
	if err != nil {
		return domain.Report{}, fmt.Errorf("failed to get commit: %w", err)
	}

	// Skip merge commits if requested
	if skipMergeCommits && commit.IsMergeCommit {
		// Create empty validation result for skipped commit
		emptyResult := domain.ValidationResult{Commit: commit, Errors: nil}
		commitRules := domain.FilterCommitRules(rules)

		return domain.BuildReportWithRules([]domain.ValidationResult{emptyResult}, nil, commitRules, domain.ReportOptions{}), nil
	}

	// Validate commit and repository using pure domain functions
	validationResult := domain.ValidateCommit(commit, rules, repo, cfg)
	repoErrors := domain.ValidateRepository(rules, repo, cfg)

	// Build report using simplified domain function with all rules
	return domain.BuildReportWithRules([]domain.ValidationResult{validationResult}, repoErrors, rules, domain.ReportOptions{}), nil
}

// validateRangeTarget validates a range of commits (pure function).
func validateRangeTarget(ctx context.Context, fromRef, toRef string, rules []domain.Rule,
	repo domain.Repository, cfg *config.Config, skipMergeCommits bool) (domain.Report, error) {
	commits, err := repo.GetCommitRange(ctx, fromRef, toRef)
	if err != nil {
		return domain.Report{}, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Filter merge commits if requested
	commits = domain.FilterMergeCommits(commits, skipMergeCommits)

	// Validate commits and repository using pure domain functions
	validationResults := domain.ValidateCommits(commits, rules, repo, cfg)
	repoErrors := domain.ValidateRepository(rules, repo, cfg)

	// Build report using simplified domain function with all rules
	return domain.BuildReportWithRules(validationResults, repoErrors, rules, domain.ReportOptions{}), nil
}

// validateCountTarget validates a count of commits from HEAD (pure function).
func validateCountTarget(ctx context.Context, countStr string, rules []domain.Rule,
	repo domain.Repository, cfg *config.Config, skipMergeCommits bool) (domain.Report, error) {
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return domain.Report{}, fmt.Errorf("invalid commit count: %w", err)
	}

	if count <= 0 {
		return domain.Report{}, fmt.Errorf("commit count must be positive: %d", count)
	}

	if count == 1 {
		// Single commit validation
		return validateCommitTarget(ctx, "HEAD", rules, repo, cfg, skipMergeCommits)
	}

	// Multiple commits - use range from HEAD~(count-1) to HEAD
	fromRef := fmt.Sprintf("HEAD~%d", count-1)

	return validateRangeTarget(ctx, fromRef, "HEAD", rules, repo, cfg, skipMergeCommits)
}

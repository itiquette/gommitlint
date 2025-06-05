// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"errors"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// ValidateCommit validates a single commit against the provided rules.
// This is a pure function that takes all dependencies as parameters.
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg *config.Config) ValidationResult {
	var errors []ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(commit, repo, cfg)...)
	}

	return ValidationResult{Commit: commit, Errors: errors}
}

// ValidateCommits validates multiple commits against the provided rules.
// Returns a slice of validation results, one for each commit.
func ValidateCommits(commits []Commit, rules []Rule, repo Repository, cfg *config.Config) []ValidationResult {
	results := make([]ValidationResult, len(commits))
	for i, commit := range commits {
		results[i] = ValidateCommit(commit, rules, repo, cfg)
	}

	return results
}

// ValidateRepository runs repository-level rules.
// Repository rules receive an empty commit as they don't validate specific commits.
func ValidateRepository(rules []Rule, repo Repository, cfg *config.Config) []ValidationError {
	repoRules := FilterRepositoryRules(rules)

	var errors []ValidationError

	// Repository rules don't need commit data
	emptyCommit := Commit{}
	for _, rule := range repoRules {
		errors = append(errors, rule.Validate(emptyCommit, repo, cfg)...)
	}

	return errors
}

// ValidateMessage validates a commit message string without repository context.
// Useful for pre-commit hooks and message file validation.
func ValidateMessage(message string, rules []Rule, cfg *config.Config) (ValidationResult, error) {
	// Trim whitespace
	message = strings.TrimSpace(message)
	if message == "" {
		return ValidationResult{}, errors.New("empty commit message")
	}

	commit := ParseCommitMessage(message)
	commitRules := FilterCommitRules(rules)

	return ValidateCommit(commit, commitRules, nil, cfg), nil
}

// FilterCommitRules returns only commit-level rules from the provided rules.
func FilterCommitRules(rules []Rule) []Rule {
	var result []Rule

	for _, rule := range rules {
		if !IsRepositoryLevelRule(rule) {
			result = append(result, rule)
		}
	}

	return result
}

// FilterRepositoryRules returns only repository-level rules from the provided rules.
func FilterRepositoryRules(rules []Rule) []Rule {
	var result []Rule

	for _, rule := range rules {
		if IsRepositoryLevelRule(rule) {
			result = append(result, rule)
		}
	}

	return result
}

// FullValidation represents both commit and repository validation results.
type FullValidation struct {
	CommitResults    []ValidationResult
	RepositoryErrors []ValidationError
}

// HasFailures returns true if there are any validation failures.
func (f FullValidation) HasFailures() bool {
	for _, result := range f.CommitResults {
		if result.HasFailures() {
			return true
		}
	}

	return len(f.RepositoryErrors) > 0
}

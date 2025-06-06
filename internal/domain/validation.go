// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"errors"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// ValidateCommit validates a single commit against both commit and repository rules.
func ValidateCommit(commit Commit, commitRules []CommitRule, repoRules []RepositoryRule, repo Repository, cfg config.Config) ValidationResult {
	var errors []ValidationError

	// Validate commit-only rules
	errors = append(errors, ValidateCommitRules(commit, commitRules, cfg)...)

	// Validate repository-dependent rules
	errors = append(errors, ValidateRepositoryRules(commit, repoRules, repo, cfg)...)

	return ValidationResult{Commit: commit, Errors: errors}
}

// ValidateCommits validates multiple commits against both rule types.
func ValidateCommits(commits []Commit, commitRules []CommitRule, repoRules []RepositoryRule, repo Repository, cfg config.Config) []ValidationResult {
	results := make([]ValidationResult, len(commits))
	for i, commit := range commits {
		results[i] = ValidateCommit(commit, commitRules, repoRules, repo, cfg)
	}

	return results
}

// ValidateRepository runs repository-level rules.
func ValidateRepository(rules []RepositoryRule, repo Repository, cfg config.Config) []ValidationError {
	var errors []ValidationError

	// Repository rules don't need commit data
	emptyCommit := Commit{}
	for _, rule := range rules {
		errors = append(errors, rule.Validate(emptyCommit, repo, cfg)...)
	}

	return errors
}

// ValidateMessage validates a commit message string without repository context.
func ValidateMessage(message string, rules []CommitRule, cfg config.Config) (ValidationResult, error) {
	// Trim whitespace
	message = strings.TrimSpace(message)
	if message == "" {
		return ValidationResult{}, errors.New("empty commit message")
	}

	commit := ParseCommitMessage(message)
	errors := ValidateCommitRules(commit, rules, cfg)

	return ValidationResult{Commit: commit, Errors: errors}, nil
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

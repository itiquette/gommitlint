// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"errors"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// Dependencies contains all dependencies needed for validation.
// This is a simple container that avoids complex dependency injection patterns.
type Dependencies struct {
	Repository Repository
	Config     *config.Config
	Logger     Logger
}

// ValidatorWithDeps contains validator and dependencies for functional validation.
// This is a simple struct that follows functional programming principles.
type ValidatorWithDeps struct {
	Validator Validator
	Deps      Dependencies
}

// NewValidatorWithDeps creates a validator with dependencies.
func NewValidatorWithDeps(rules []Rule, deps Dependencies) ValidatorWithDeps {
	return ValidatorWithDeps{
		Validator: NewValidator(rules),
		Deps:      deps,
	}
}

// ValidateCommit validates a single commit using the validator's dependencies.
func (v ValidatorWithDeps) ValidateCommit(commit Commit) ValidationResult {
	return v.Validator.ValidateCommit(commit, v.Deps.Repository, v.Deps.Config)
}

// ValidateCommits validates multiple commits using the validator's dependencies.
func (v ValidatorWithDeps) ValidateCommits(commits []Commit) []ValidationResult {
	return v.Validator.ValidateCommits(commits, v.Deps.Repository, v.Deps.Config)
}

// ValidateRepository runs repository-level rules using the validator's dependencies.
func (v ValidatorWithDeps) ValidateRepository() []RuleFailure {
	return ValidateRepository(FilterRepositoryRules(v.Validator.Rules()), v.Deps.Repository, v.Deps.Config)
}

// ValidateMessage validates a commit message string using the validator's dependencies.
func (v ValidatorWithDeps) ValidateMessage(message string) (ValidationResult, error) {
	return ValidateCommitMessage(message, v.Validator.Rules(), v.Deps.Config)
}

// ValidateCommit validates a single commit against the provided rules.
// This is a pure function that takes all dependencies as parameters.
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg *config.Config) ValidationResult {
	var failures []RuleFailure
	for _, rule := range rules {
		failures = append(failures, rule.Validate(commit, repo, cfg)...)
	}

	return ValidationResult{Commit: commit, Failures: failures}
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
func ValidateRepository(rules []Rule, repo Repository, cfg *config.Config) []RuleFailure {
	repoRules := FilterRepositoryRules(rules)

	var failures []RuleFailure

	// Repository rules don't need commit data
	emptyCommit := Commit{}
	for _, rule := range repoRules {
		failures = append(failures, rule.Validate(emptyCommit, repo, cfg)...)
	}

	return failures
}

// ValidateCommitMessage validates a commit message string without repository context.
// Useful for pre-commit hooks and message file validation.
func ValidateCommitMessage(message string, rules []Rule, cfg *config.Config) (ValidationResult, error) {
	// Trim whitespace
	message = strings.TrimSpace(message)
	if message == "" {
		return ValidationResult{}, errors.New("empty commit message")
	}

	commit := ParseCommitMessage(message)
	commitRules := FilterCommitRules(rules)

	return ValidateCommit(commit, commitRules, nil, cfg), nil
}

// FilterMergeCommits removes merge commits from the slice if skipMerge is true.
// This is a pure function that returns a new slice.
func FilterMergeCommits(commits []Commit, skipMerge bool) []Commit {
	if !skipMerge {
		return commits
	}

	var filtered []Commit

	for _, commit := range commits {
		if !commit.IsMergeCommit {
			filtered = append(filtered, commit)
		}
	}

	return filtered
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
	CommitResults      []ValidationResult
	RepositoryFailures []RuleFailure
}

// HasFailures returns true if there are any validation failures.
func (f FullValidation) HasFailures() bool {
	for _, result := range f.CommitResults {
		if result.HasFailures() {
			return true
		}
	}

	return len(f.RepositoryFailures) > 0
}

// Options contains options for validation - used by CLI layer.
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

// Logger provides logging capabilities.
// This interface is defined here following dependency inversion principle.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

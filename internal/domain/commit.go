// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package domain contains the core domain entities and interfaces.
package domain

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// SplitCommitMessage splits a commit message into subject and body.
func SplitCommitMessage(message string) (string, string) {
	var subject, body string
	// Trim whitespace from the entire message
	message = strings.TrimSpace(message)

	// Split the message by newline
	parts := strings.SplitN(message, "\n", 2)

	// The first line is the subject
	subject = strings.TrimSpace(parts[0])

	// The rest is the body (if it exists)
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}

	return subject, body
}

// HasBody returns true if the commit has a body.
func (c Commit) HasBody() bool {
	return strings.TrimSpace(c.Body) != ""
}

// IsValid returns true if the commit has basic required fields.
func (c Commit) IsValid() bool {
	return c.Hash != "" && strings.TrimSpace(c.Subject) != ""
}

// IsSigned returns true if the commit has a signature.
func (c Commit) IsSigned() bool {
	return c.Signature != ""
}

// IsValidCommitSubject checks if a commit subject follows domain rules (pure function).
func IsValidCommitSubject(subject string) bool {
	return len(strings.TrimSpace(subject)) > 0
}

// ContainsSignature checks if a commit contains a valid signature (pure function).
func ContainsSignature(commit Commit) bool {
	return commit.Signature != ""
}

// IsValidCommitMessage checks if a commit message follows domain rules (pure function).
func IsValidCommitMessage(message string) bool {
	return len(strings.TrimSpace(message)) > 0
}

// ExtractJiraTickets extracts JIRA ticket IDs from a commit message (pure function).
func ExtractJiraTickets(message string, _ string) []string {
	parts := strings.Split(message, " ")

	// Filter parts that look like JIRA tickets using functional approach
	// TODO: This needs to be updated to use domain.Filter
	// For now, implementing inline to avoid circular dependency
	result := make([]string, 0) // Initialize as empty slice, not nil

	for _, part := range parts {
		if strings.Contains(part, "-") && len(part) >= 3 {
			result = append(result, part)
		}
	}

	return result
}

// Commit represents a Git commit for validation.
// This is a pure domain entity with value semantics.
type Commit struct {
	// Hash is the commit hash.
	Hash string

	// Subject is the first line of the commit message.
	Subject string

	// Body is the rest of the commit message.
	Body string

	// Message is the full commit message (subject + body).
	Message string

	// Author is the name of the commit author.
	Author string

	// AuthorEmail is the email of the commit author.
	AuthorEmail string

	// CommitDate is the date of the commit in ISO format.
	CommitDate string

	// Signature is the signature attached to the commit, if any.
	Signature string

	// IsMergeCommit indicates whether this is a merge commit.
	IsMergeCommit bool
}

// ValidationResult represents the validation outcome for a single commit.
type ValidationResult struct {
	Commit   Commit
	Failures []RuleFailure
}

// RuleFailure represents a single rule violation.
type RuleFailure struct {
	Rule    string
	Message string
	Help    string // Optional
}

// HasFailures returns true if there are any validation failures.
func (v ValidationResult) HasFailures() bool {
	return len(v.Failures) > 0
}

// Passed returns true if validation passed (no failures).
func (v ValidationResult) Passed() bool {
	return len(v.Failures) == 0
}

// ValidationService is the domain service interface.
type ValidationService interface {
	ValidateCommit(commit Commit) ValidationResult
	ValidateCommits(commits []Commit) []ValidationResult
}

// NewCommit creates a Commit from its components.
// Pure function that constructs a properly initialized Commit.
func NewCommit(hash, message, author, authorEmail, commitDate, signature string, isMerge bool) Commit {
	subject, body := SplitCommitMessage(message)

	return Commit{
		Hash:          hash,
		Subject:       subject,
		Body:          body,
		Message:       message,
		Author:        author,
		AuthorEmail:   authorEmail,
		CommitDate:    commitDate,
		Signature:     signature,
		IsMergeCommit: isMerge,
	}
}

// ParseCommitMessage creates a Commit from a message string.
func ParseCommitMessage(message string) Commit {
	return NewCommit("", message, "", "", "", "", false)
}

// Validator implements ValidationService using functional composition.
type Validator struct {
	rules []Rule // Injected rules, immutable after creation
}

// NewValidator creates a validator with the given rules.
func NewValidator(rules []Rule) Validator {
	return Validator{rules: rules}
}

// ValidateCommit validates a single commit against all rules.
// Repository can be nil for validation without repository context.
func (v Validator) ValidateCommit(commit Commit, repo Repository, cfg *config.Config) ValidationResult {
	var failures []RuleFailure

	ctx := ValidationContext{
		Commit:     commit,
		Repository: repo,
		Config:     cfg,
	}

	for _, rule := range v.rules {
		if ruleFailures := rule.Validate(ctx); len(ruleFailures) > 0 {
			failures = append(failures, ruleFailures...)
		}
	}

	return ValidationResult{
		Commit:   commit,
		Failures: failures,
	}
}

// ValidateCommits validates multiple commits concurrently.
func (v Validator) ValidateCommits(commits []Commit, repo Repository, cfg *config.Config) []ValidationResult {
	results := make([]ValidationResult, len(commits))

	type indexed struct {
		index  int
		result ValidationResult
	}

	resultChannel := make(chan indexed, len(commits))

	for i, commit := range commits {
		go func(idx int, c Commit) {
			resultChannel <- indexed{idx, v.ValidateCommit(c, repo, cfg)}
		}(i, commit)
	}

	for range commits {
		ir := <-resultChannel
		results[ir.index] = ir.result
	}

	return results
}

// Note: CommitReader, CommitHistoryReader, CommitAnalyzer, and
// RepositoryInfoProvider interfaces are defined in commitinterfaces.go

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package domain contains the core domain entities and interfaces.
package domain

import (
	"context"
	"strings"
)

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

// CommitCollection provides functional operations on commit slices.
type CommitCollection []Commit

// NewCommitCollection creates a new CommitCollection from a slice of commits.
func NewCommitCollection(commits []Commit) CommitCollection {
	result := make(CommitCollection, len(commits))
	copy(result, commits)

	return result
}

// Filter returns a new collection with commits matching the predicate.
func (c CommitCollection) Filter(predicate func(Commit) bool) CommitCollection {
	result := make(CommitCollection, 0)

	for _, commit := range c {
		if predicate(commit) {
			result = append(result, commit)
		}
	}

	return result
}

// FilterMergeCommits returns a new collection with merge commits filtered out.
func (c CommitCollection) FilterMergeCommits() CommitCollection {
	return c.Filter(func(commit Commit) bool {
		return !commit.IsMergeCommit
	})
}

// FilterByAuthor returns a new collection with commits by the specified author.
func (c CommitCollection) FilterByAuthor(authorNameOrEmail string) CommitCollection {
	return c.Filter(func(commit Commit) bool {
		return commit.Author == authorNameOrEmail || commit.AuthorEmail == authorNameOrEmail
	})
}

// Map transforms commits using the provided function.
func (c CommitCollection) Map(fn func(Commit) Commit) CommitCollection {
	result := make(CommitCollection, len(c))
	for i, commit := range c {
		result[i] = fn(commit)
	}

	return result
}

// Any returns true if any commit matches the predicate.
func (c CommitCollection) Any(predicate func(Commit) bool) bool {
	for _, commit := range c {
		if predicate(commit) {
			return true
		}
	}

	return false
}

// All returns true if all commits match the predicate.
func (c CommitCollection) All(predicate func(Commit) bool) bool {
	for _, commit := range c {
		if !predicate(commit) {
			return false
		}
	}

	return true
}

// First returns the first commit or empty Commit if collection is empty.
func (c CommitCollection) First() Commit {
	if len(c) == 0 {
		return Commit{}
	}

	return c[0]
}

// Last returns the last commit or empty Commit if collection is empty.
func (c CommitCollection) Last() Commit {
	if len(c) == 0 {
		return Commit{}
	}

	return c[len(c)-1]
}

// Count returns the number of commits in the collection.
func (c CommitCollection) Count() int {
	return len(c)
}

// IsEmpty returns true if the collection is empty.
func (c CommitCollection) IsEmpty() bool {
	return len(c) == 0
}

// Contains returns true if the collection contains a commit with the specified hash.
func (c CommitCollection) Contains(hash string) bool {
	return c.Any(func(commit Commit) bool {
		return commit.Hash == hash
	})
}

// With returns a new collection with the commit added.
func (c CommitCollection) With(commit Commit) CommitCollection {
	result := make(CommitCollection, len(c), len(c)+1)
	copy(result, c)

	return append(result, commit)
}

// WithAll returns a new collection with all commits from other added.
func (c CommitCollection) WithAll(other CommitCollection) CommitCollection {
	result := make(CommitCollection, len(c), len(c)+len(other))
	copy(result, c)

	return append(result, other...)
}

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

// FilterMergeCommits returns a new slice with merge commits filtered out.
// This is a convenience function for working with plain slices.
func FilterMergeCommits(commits []Commit, skipMerge bool) []Commit {
	if !skipMerge {
		return commits
	}

	return NewCommitCollection(commits).FilterMergeCommits()
}

// Repository defines the contract for accessing Git repository data.
// This is a port in hexagonal architecture - domain defines what it needs.
type Repository interface {
	// GetCommit retrieves a single commit by reference.
	GetCommit(ctx context.Context, ref string) (Commit, error)

	// GetCommitRange retrieves commits in a range.
	GetCommitRange(ctx context.Context, from, to string) ([]Commit, error)

	// GetHeadCommits retrieves N commits from HEAD.
	GetHeadCommits(ctx context.Context, count int) ([]Commit, error)

	// GetCommitsAheadCount returns how many commits the current branch is ahead of the reference.
	GetCommitsAheadCount(ctx context.Context, referenceBranch string) (int, error)
}

// ValidationResult represents the validation outcome for a single commit.
type ValidationResult struct {
	Commit Commit
	Errors []ValidationError
}

// HasFailures returns true if there are any validation failures.
func (v ValidationResult) HasFailures() bool {
	return len(v.Errors) > 0
}

// Passed returns true if validation passed (no failures).
func (v ValidationResult) Passed() bool {
	return len(v.Errors) == 0
}

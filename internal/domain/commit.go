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
type Commit struct {
	// Hash is the Git commit SHA.
	Hash string

	// Subject is the first line of the commit message.
	Subject string

	// Body is the rest of the commit message after the subject.
	Body string

	// Message is the complete commit message including subject and body.
	Message string

	// Author is the name of the commit author.
	Author string

	// AuthorEmail is the email address of the commit author.
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

// SplitCommitMessage splits a commit message into subject and body following Git conventions.
// Git convention: subject + blank line + body. Without blank line, everything is subject.
func SplitCommitMessage(message string) (string, string) {
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return "", ""
	}

	subject := strings.TrimSpace(lines[0])

	// Check for proper Git structure: subject + blank line + body
	if len(lines) >= 3 && strings.TrimSpace(lines[1]) == "" {
		// Valid structure: join lines 2+ as body
		bodyLines := lines[2:]
		body := strings.TrimSpace(strings.Join(bodyLines, "\n"))

		return subject, body
	}

	// No proper structure: only subject, no body
	return subject, ""
}

// NewCommit creates a Commit from its components.
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
func FilterMergeCommits(commits []Commit) []Commit {
	result := make([]Commit, 0)

	for _, commit := range commits {
		if !commit.IsMergeCommit {
			result = append(result, commit)
		}
	}

	return result
}

// Repository defines the contract for accessing Git repository data.
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

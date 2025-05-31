// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package domain contains the core domain entities and interfaces.
package domain

import (
	"strings"
)

// CommitInfo represents information about a Git commit.
// This is a pure domain entity with value semantics.
type CommitInfo struct {
	// Hash is the commit hash.
	Hash string

	// Subject is the first line of the commit message.
	Subject string

	// Body is the rest of the commit message.
	Body string

	// Message is the full commit message (subject + body).
	Message string

	// Signature is the signature attached to the commit, if any.
	Signature string

	// IsMergeCommit indicates whether this is a merge commit.
	IsMergeCommit bool

	// AuthorName is the name of the commit author.
	AuthorName string

	// AuthorEmail is the email of the commit author.
	AuthorEmail string

	// CommitDate is the date of the commit in ISO format.
	CommitDate string
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

// HasBody returns true if the commit has a body.
func (c CommitInfo) HasBody() bool {
	return strings.TrimSpace(c.Body) != ""
}

// IsValid returns true if the commit has basic required fields.
func (c CommitInfo) IsValid() bool {
	return c.Hash != "" && strings.TrimSpace(c.Subject) != ""
}

// IsSigned returns true if the commit has a signature.
func (c CommitInfo) IsSigned() bool {
	return c.Signature != ""
}

// IsValidCommitSubject checks if a commit subject follows domain rules (pure function).
func IsValidCommitSubject(subject string) bool {
	return len(strings.TrimSpace(subject)) > 0
}

// ContainsSignature checks if a commit contains a valid signature (pure function).
func ContainsSignature(commit CommitInfo) bool {
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

// Note: CommitReader, CommitHistoryReader, CommitAnalyzer, and
// RepositoryInfoProvider interfaces are defined in commitinterfaces.go

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package domain contains the core domain entities and interfaces.
package domain

import (
	"strings"
)

// CommitInfo represents information about a Git commit.
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

	// RawCommit contains the raw commit object from the git library.
	// This is used by infrastructure code and should not be accessed by domain logic.
	RawCommit interface{}
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

// CommitService provides domain operations for commits.
type CommitService interface {
	// IsValidCommitSubject checks if a commit subject follows domain rules.
	IsValidCommitSubject(subject string) bool

	// ContainsSignature checks if a commit contains a valid signature.
	ContainsSignature(commit *CommitInfo) bool

	// IsValidCommitMessage checks if a commit message follows domain rules.
	IsValidCommitMessage(message string) bool

	// ExtractJiraTickets extracts JIRA ticket IDs from a commit message.
	ExtractJiraTickets(message string, pattern string) []string
}

// Note: CommitReader, CommitHistoryReader, CommitAnalyzer, and
// RepositoryInfoProvider interfaces are defined in git_interfaces.go

// DefaultCommitService provides domain services for commits.
type DefaultCommitService struct{}

// NewCommitService creates a new DefaultCommitService.
func NewCommitService() *DefaultCommitService {
	return &DefaultCommitService{}
}

// IsValidCommitSubject checks if a commit subject follows domain rules.
func (s *DefaultCommitService) IsValidCommitSubject(subject string) bool {
	// A valid subject must not be empty
	return len(strings.TrimSpace(subject)) > 0
}

// ContainsSignature checks if a commit contains a valid signature.
func (s *DefaultCommitService) ContainsSignature(commit *CommitInfo) bool {
	return commit.Signature != ""
}

// IsValidCommitMessage checks if a commit message follows domain rules.
func (s *DefaultCommitService) IsValidCommitMessage(message string) bool {
	// A valid message must not be empty
	return len(strings.TrimSpace(message)) > 0
}

// ExtractJiraTickets extracts JIRA ticket IDs from a commit message.
func (s *DefaultCommitService) ExtractJiraTickets(message string, _ string) []string {
	// This is a simplified implementation
	// In a real application, you would use regex matching
	var tickets []string

	parts := strings.Split(message, " ")
	for _, part := range parts {
		// Simple check for JIRA-123 format
		if strings.Contains(part, "-") && len(part) >= 3 {
			tickets = append(tickets, part)
		}
	}

	return tickets
}

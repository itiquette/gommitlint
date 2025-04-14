// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testdata provides test fixtures and builders for domain tests.
// All test helpers use value semantics and functional options.
package testdata

import (
	"context"
	"strings"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
)

// CommitOption configures a commit during creation.
type CommitOption func(domain.Commit) domain.Commit

// Commit creates a test commit with the given message and options.
func Commit(message string, opts ...CommitOption) domain.Commit {
	commit := domain.Commit{
		Hash:          "abc123def456",
		Subject:       message,
		Message:       message,
		Author:        "Test User",
		AuthorEmail:   "test@example.com",
		CommitDate:    time.Now().Format(time.RFC3339),
		IsMergeCommit: false,
	}

	// Parse subject and body from message
	lines := strings.Split(message, "\n")
	if len(lines) > 0 {
		commit.Subject = lines[0]
	}

	// Parse body - everything after the subject line
	// If there's an empty line, body starts after it
	// Otherwise, body is everything after the first line
	if len(lines) > 1 {
		// Check if there's an empty line separator
		bodyStart := 1

		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				bodyStart = i + 1

				break
			}
		}

		// Set body as everything after subject (or after empty line if found)
		if bodyStart < len(lines) {
			commit.Body = strings.Join(lines[bodyStart:], "\n")
		} else if bodyStart == 1 {
			// No empty line found, body is everything after first line
			commit.Body = strings.Join(lines[1:], "\n")
		}
	}

	// Apply all options
	for _, opt := range opts {
		commit = opt(commit)
	}

	return commit
}

// WithSignature adds a signature to the commit.
func WithSignature(sig string) CommitOption {
	return func(c domain.Commit) domain.Commit {
		c.Signature = sig
		c.Hash = "signed123def"

		return c
	}
}

// WithMerge marks the commit as a merge commit.
func WithMerge() CommitOption {
	return func(c domain.Commit) domain.Commit {
		c.IsMergeCommit = true
		c.Hash = "merge123def"

		return c
	}
}

// AsInvalid creates an invalid commit that should fail validation.
func AsInvalid() CommitOption {
	return func(commit domain.Commit) domain.Commit {
		commit.Hash = "def456abc123"
		commit.Subject = "bad commit message without conventional format and way too long to pass length validation"
		commit.Message = "bad commit message without conventional format and way too long to pass length validation"
		commit.Body = ""

		return commit
	}
}

// MockRule implements domain.Rule for testing.
type MockRule struct {
	name   string
	errors []domain.ValidationError
}

// NewMockRule creates a rule that returns the specified errors.
func NewMockRule(name string, errs ...domain.ValidationError) *MockRule {
	return &MockRule{name: name, errors: errs}
}

// Name returns the rule name.
func (r *MockRule) Name() string {
	return r.name
}

// Validate returns the configured errors.
func (r *MockRule) Validate(_ context.Context, _ domain.Commit) []domain.ValidationError {
	return r.errors
}
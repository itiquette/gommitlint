// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
)

// CommitBuilder provides a functional builder for creating test commit objects.
// It uses value semantics to ensure immutability during the build process.
type CommitBuilder struct {
	commit domain.CommitInfo
}

// NewCommitBuilder creates a new CommitBuilder with default values.
func NewCommitBuilder() CommitBuilder {
	return CommitBuilder{
		commit: domain.CommitInfo{
			// Set sensible defaults
			Hash:          "abc123def456",
			Subject:       "Default commit subject",
			Body:          "",
			Message:       "Default commit subject",
			Signature:     "",
			IsMergeCommit: false,
			AuthorName:    "Test User",
			AuthorEmail:   "test@example.com",
			CommitDate:    time.Now().Format(time.RFC3339),
		},
	}
}

// WithHash sets the commit hash.
func (b CommitBuilder) WithHash(hash string) CommitBuilder {
	b.commit.Hash = hash

	return b
}

// WithSubject sets the commit subject.
func (b CommitBuilder) WithSubject(subject string) CommitBuilder {
	b.commit.Subject = subject

	return b
}

// WithBody sets the commit body.
func (b CommitBuilder) WithBody(body string) CommitBuilder {
	b.commit.Body = body

	return b
}

// WithAuthorName sets the commit author name.
func (b CommitBuilder) WithAuthorName(name string) CommitBuilder {
	b.commit.AuthorName = name

	return b
}

// WithAuthorEmail sets the commit author email.
func (b CommitBuilder) WithAuthorEmail(email string) CommitBuilder {
	b.commit.AuthorEmail = email

	return b
}

// WithSignature sets the commit signature.
func (b CommitBuilder) WithSignature(signature string) CommitBuilder {
	b.commit.Signature = signature

	return b
}

// WithIsMergeCommit sets whether this is a merge commit.
func (b CommitBuilder) WithIsMergeCommit(isMerge bool) CommitBuilder {
	b.commit.IsMergeCommit = isMerge

	return b
}

// WithMessage combines subject and body into a message.
func (b CommitBuilder) WithMessage(message string) CommitBuilder {
	b.commit.Message = message

	return b
}

// WithCommitDate sets the commit date.
func (b CommitBuilder) WithCommitDate(date string) CommitBuilder {
	b.commit.CommitDate = date

	return b
}

// WithTimestamp sets the commit date using a time.Time.
func (b CommitBuilder) WithTimestamp(timestamp time.Time) CommitBuilder {
	b.commit.CommitDate = timestamp.Format(time.RFC3339)

	return b
}

// Build returns the constructed CommitInfo.
func (b CommitBuilder) Build() domain.CommitInfo {
	return b.commit
}

// Default returns a new builder with default test configuration.
func Default() CommitBuilder {
	return NewCommitBuilder()
}

// Valid returns a builder configured with a typical valid commit.
func Valid() CommitBuilder {
	return NewCommitBuilder().
		WithSubject("Add new feature").
		WithBody("This commit adds a new feature to the application.")
}

// Invalid returns a builder configured with an invalid commit.
func Invalid() CommitBuilder {
	return NewCommitBuilder().
		WithSubject("") // Empty subject is invalid
}

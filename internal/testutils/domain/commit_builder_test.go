// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	testdomain "github.com/itiquette/gommitlint/internal/testutils/domain"
	"github.com/stretchr/testify/require"
)

func TestCommitBuilder(t *testing.T) {
	tests := []struct {
		name         string
		buildFunc    func() domain.CommitInfo
		expectedHash string
		subject      string
		body         string
		authorName   string
		authorEmail  string
	}{
		{
			name: "default commit",
			buildFunc: func() domain.CommitInfo {
				return testdomain.NewCommitBuilder().Build()
			},
			expectedHash: "abc123def456",
			subject:      "Default commit subject",
			body:         "",
			authorName:   "Test User",
			authorEmail:  "test@example.com",
		},
		{
			name: "custom commit",
			buildFunc: func() domain.CommitInfo {
				return testdomain.NewCommitBuilder().
					WithHash("xyz789").
					WithSubject("Custom subject").
					WithBody("Custom body").
					WithAuthorName("Custom Author").
					WithAuthorEmail("custom@example.com").
					Build()
			},
			expectedHash: "xyz789",
			subject:      "Custom subject",
			body:         "Custom body",
			authorName:   "Custom Author",
			authorEmail:  "custom@example.com",
		},
		{
			name: "valid preset",
			buildFunc: func() domain.CommitInfo {
				return testdomain.Valid().Build()
			},
			expectedHash: "abc123def456",
			subject:      "Add new feature",
			body:         "This commit adds a new feature to the application.",
			authorName:   "Test User",
			authorEmail:  "test@example.com",
		},
		{
			name: "invalid preset",
			buildFunc: func() domain.CommitInfo {
				return testdomain.Invalid().Build()
			},
			expectedHash: "abc123def456",
			subject:      "",
			body:         "",
			authorName:   "Test User",
			authorEmail:  "test@example.com",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			commit := testCase.buildFunc()

			require.Equal(t, testCase.expectedHash, commit.Hash)
			require.Equal(t, testCase.subject, commit.Subject)
			require.Equal(t, testCase.body, commit.Body)
			require.Equal(t, testCase.authorName, commit.AuthorName)
			require.Equal(t, testCase.authorEmail, commit.AuthorEmail)
			require.NotEmpty(t, commit.CommitDate)
		})
	}
}

func TestCommitBuilderImmutability(t *testing.T) {
	// Ensure builder methods don't modify the original
	original := testdomain.NewCommitBuilder()
	modified := original.WithSubject("Modified subject")

	originalCommit := original.Build()
	modifiedCommit := modified.Build()

	require.NotEqual(t, originalCommit.Subject, modifiedCommit.Subject)
	require.Equal(t, "Default commit subject", originalCommit.Subject)
	require.Equal(t, "Modified subject", modifiedCommit.Subject)
}

func TestCommitBuilderTimestamp(t *testing.T) {
	customTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	commit := testdomain.NewCommitBuilder().
		WithTimestamp(customTime).
		Build()

	require.Equal(t, customTime.Format(time.RFC3339), commit.CommitDate)
}

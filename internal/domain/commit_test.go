// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestCommitInfoMethods(t *testing.T) {
	tests := []struct {
		name     string
		commit   domain.CommitInfo
		testFunc func(domain.CommitInfo) bool
		expected bool
	}{
		{
			name: "HasBody returns true for commit with body",
			commit: domain.CommitInfo{
				Body: "This is a commit body",
			},
			testFunc: domain.CommitInfo.HasBody,
			expected: true,
		},
		{
			name: "HasBody returns false for commit without body",
			commit: domain.CommitInfo{
				Body: "",
			},
			testFunc: domain.CommitInfo.HasBody,
			expected: false,
		},
		{
			name: "HasBody returns false for whitespace-only body",
			commit: domain.CommitInfo{
				Body: "   \n\t  ",
			},
			testFunc: domain.CommitInfo.HasBody,
			expected: false,
		},
		{
			name: "IsValid returns true for valid commit",
			commit: domain.CommitInfo{
				Hash:    "abc123",
				Subject: "feat: add feature",
			},
			testFunc: domain.CommitInfo.IsValid,
			expected: true,
		},
		{
			name: "IsValid returns false for commit without hash",
			commit: domain.CommitInfo{
				Subject: "feat: add feature",
			},
			testFunc: domain.CommitInfo.IsValid,
			expected: false,
		},
		{
			name: "IsValid returns false for commit without subject",
			commit: domain.CommitInfo{
				Hash: "abc123",
			},
			testFunc: domain.CommitInfo.IsValid,
			expected: false,
		},
		{
			name: "IsSigned returns true for signed commit",
			commit: domain.CommitInfo{
				Signature: "-----BEGIN PGP SIGNATURE-----",
			},
			testFunc: domain.CommitInfo.IsSigned,
			expected: true,
		},
		{
			name: "IsSigned returns false for unsigned commit",
			commit: domain.CommitInfo{
				Signature: "",
			},
			testFunc: domain.CommitInfo.IsSigned,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.testFunc(tc.commit)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidCommitSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected bool
	}{
		{
			name:     "valid subject",
			subject:  "feat: add new feature",
			expected: true,
		},
		{
			name:     "empty subject",
			subject:  "",
			expected: false,
		},
		{
			name:     "whitespace only subject",
			subject:  "   ",
			expected: false,
		},
		{
			name:     "subject with leading/trailing spaces",
			subject:  "  valid subject  ",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.IsValidCommitSubject(tc.subject)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestContainsSignature(t *testing.T) {
	tests := []struct {
		name     string
		commit   domain.CommitInfo
		expected bool
	}{
		{
			name: "commit with signature",
			commit: domain.CommitInfo{
				Signature: "-----BEGIN PGP SIGNATURE-----\n...\n-----END PGP SIGNATURE-----",
			},
			expected: true,
		},
		{
			name: "commit without signature",
			commit: domain.CommitInfo{
				Signature: "",
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.ContainsSignature(tc.commit)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractJiraTickets(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		pattern  string
		expected []string
	}{
		{
			name:     "message with JIRA tickets",
			message:  "feat: add PROJ-123 and PROJ-456",
			expected: []string{"PROJ-123", "PROJ-456"},
		},
		{
			name:     "message without JIRA tickets",
			message:  "feat: add new feature",
			expected: []string{},
		},
		{
			name:     "message with hyphenated words",
			message:  "add feature-flag FOO-123",
			expected: []string{"feature-flag", "FOO-123"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.ExtractJiraTickets(tc.message, tc.pattern)
			require.Equal(t, tc.expected, result)
		})
	}
}

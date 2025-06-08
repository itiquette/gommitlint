// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConventionalCommit(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected ConventionalCommitFormat
	}{
		{
			name:    "simple feat commit",
			subject: "feat: add new feature",
			expected: ConventionalCommitFormat{
				Type:        "feat",
				Scopes:      nil,
				Breaking:    false,
				Description: "add new feature",
				RawScope:    "",
				IsValid:     true,
			},
		},
		{
			name:    "feat with scope",
			subject: "feat(auth): add login functionality",
			expected: ConventionalCommitFormat{
				Type:        "feat",
				Scopes:      []string{"auth"},
				Breaking:    false,
				Description: "add login functionality",
				RawScope:    "auth",
				IsValid:     true,
			},
		},
		{
			name:    "breaking change",
			subject: "feat(api)!: change auth endpoint structure",
			expected: ConventionalCommitFormat{
				Type:        "feat",
				Scopes:      []string{"api"},
				Breaking:    true,
				Description: "change auth endpoint structure",
				RawScope:    "api",
				IsValid:     true,
			},
		},
		{
			name:    "multiple scopes",
			subject: "feat(ui,api): add login functionality",
			expected: ConventionalCommitFormat{
				Type:        "feat",
				Scopes:      []string{"ui", "api"},
				Breaking:    false,
				Description: "add login functionality",
				RawScope:    "ui,api",
				IsValid:     true,
			},
		},
		{
			name:    "not conventional - missing colon",
			subject: "feat add new feature",
			expected: ConventionalCommitFormat{
				IsValid: false,
			},
		},
		{
			name:    "not conventional - no type",
			subject: "add new feature",
			expected: ConventionalCommitFormat{
				IsValid: false,
			},
		},
		{
			name:    "not conventional - random text",
			subject: "not conventional commit",
			expected: ConventionalCommitFormat{
				IsValid: false,
			},
		},
		{
			name:    "malformed conventional - missing closing paren",
			subject: "feat(scope: missing closing paren",
			expected: ConventionalCommitFormat{
				IsValid: false,
			},
		},
		{
			name:    "valid structure - no space after colon",
			subject: "feat:",
			expected: ConventionalCommitFormat{
				Type:        "feat",
				Scopes:      nil,
				Breaking:    false,
				Description: "",
				RawScope:    "",
				IsValid:     true,
			},
		},
		{
			name:    "valid structure - empty description",
			subject: "feat: ",
			expected: ConventionalCommitFormat{
				Type:        "feat",
				Scopes:      nil,
				Breaking:    false,
				Description: "",
				RawScope:    "",
				IsValid:     true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := ParseConventionalCommit(testCase.subject)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestIsConventionalCommit(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected bool
	}{
		{
			name:     "valid conventional commit",
			subject:  "feat: add new feature",
			expected: true,
		},
		{
			name:     "valid with scope",
			subject:  "feat(auth): add login",
			expected: true,
		},
		{
			name:     "valid breaking change",
			subject:  "feat!: breaking change",
			expected: true,
		},
		{
			name:     "not conventional",
			subject:  "not conventional commit",
			expected: false,
		},
		{
			name:     "empty string",
			subject:  "",
			expected: false,
		},
		{
			name:     "malformed conventional - missing closing paren",
			subject:  "feat(scope: missing closing paren",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := IsConventionalCommit(testCase.subject)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestIsConventionalCommitLike(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected bool
	}{
		{
			name:     "valid conventional commit",
			subject:  "feat: add new feature",
			expected: true,
		},
		{
			name:     "valid with scope",
			subject:  "feat(auth): add login",
			expected: true,
		},
		{
			name:     "malformed - missing closing paren",
			subject:  "feat(scope: missing closing paren",
			expected: true,
		},
		{
			name:     "malformed - missing colon",
			subject:  "feat add new feature",
			expected: true,
		},
		{
			name:     "not conventional at all",
			subject:  "random commit message",
			expected: false,
		},
		{
			name:     "not conventional - starts with add",
			subject:  "add new feature",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := IsConventionalCommitLike(testCase.subject)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestExtractDescriptionFromConventional(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected string
	}{
		{
			name:     "conventional commit",
			subject:  "feat: add new feature",
			expected: "add new feature",
		},
		{
			name:     "conventional with scope",
			subject:  "feat(auth): add login",
			expected: "add login",
		},
		{
			name:     "not conventional - return full subject",
			subject:  "not conventional commit",
			expected: "not conventional commit",
		},
		{
			name:     "valid structure but empty description",
			subject:  "feat:",
			expected: "feat:",
		},
		{
			name:     "valid structure but empty description",
			subject:  "feat: ",
			expected: "feat: ",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := ExtractDescriptionFromConventional(testCase.subject)
			require.Equal(t, testCase.expected, result)
		})
	}
}

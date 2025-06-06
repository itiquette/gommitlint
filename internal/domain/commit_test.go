// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestCommitMethods(t *testing.T) {
	tests := []struct {
		name     string
		commit   domain.Commit
		testFunc func(domain.Commit) bool
		expected bool
	}{
		{
			name: "HasBody returns true for commit with body",
			commit: domain.Commit{
				Body: "This is a commit body",
			},
			testFunc: domain.Commit.HasBody,
			expected: true,
		},
		{
			name: "HasBody returns false for commit without body",
			commit: domain.Commit{
				Body: "",
			},
			testFunc: domain.Commit.HasBody,
			expected: false,
		},
		{
			name: "HasBody returns false for whitespace-only body",
			commit: domain.Commit{
				Body: "   \n\t  ",
			},
			testFunc: domain.Commit.HasBody,
			expected: false,
		},
		{
			name: "IsValid returns true for valid commit",
			commit: domain.Commit{
				Hash:    "abc123",
				Subject: "feat: add feature",
			},
			testFunc: domain.Commit.IsValid,
			expected: true,
		},
		{
			name: "IsValid returns false for commit without hash",
			commit: domain.Commit{
				Subject: "feat: add feature",
			},
			testFunc: domain.Commit.IsValid,
			expected: false,
		},
		{
			name: "IsValid returns false for commit without subject",
			commit: domain.Commit{
				Hash: "abc123",
			},
			testFunc: domain.Commit.IsValid,
			expected: false,
		},
		{
			name: "IsSigned returns true for signed commit",
			commit: domain.Commit{
				Signature: "-----BEGIN PGP SIGNATURE-----",
			},
			testFunc: domain.Commit.IsSigned,
			expected: true,
		},
		{
			name: "IsSigned returns false for unsigned commit",
			commit: domain.Commit{
				Signature: "",
			},
			testFunc: domain.Commit.IsSigned,
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

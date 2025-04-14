// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestIdentity(t *testing.T) {
	tests := []struct {
		name       string
		inputName  string
		inputEmail string
		wantString string
		wantEmpty  bool
	}{
		{
			name:       "Both name and email",
			inputName:  "John Doe",
			inputEmail: "john@example.com",
			wantString: "John Doe <john@example.com>",
			wantEmpty:  false,
		},
		{
			name:       "Email only",
			inputName:  "",
			inputEmail: "john@example.com",
			wantString: "john@example.com",
			wantEmpty:  false,
		},
		{
			name:       "Name only",
			inputName:  "John Doe",
			inputEmail: "",
			wantString: "John Doe <>",
			wantEmpty:  false,
		},
		{
			name:       "Empty identity",
			inputName:  "",
			inputEmail: "",
			wantString: "",
			wantEmpty:  true,
		},
		{
			name:       "Trim whitespace",
			inputName:  "  John Doe  ",
			inputEmail: "  john@example.com  ",
			wantString: "John Doe <john@example.com>",
			wantEmpty:  false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			identity := domain.NewIdentity(testCase.inputName, testCase.inputEmail)

			// For the whitespace test, we expect the trimmed version
			if testCase.name == "Trim whitespace" {
				require.Equal(t, strings.TrimSpace(testCase.inputName), identity.Name())
			} else {
				require.Equal(t, testCase.inputName, identity.Name())
			}
			// For the whitespace test, we expect the trimmed version
			if testCase.name == "Trim whitespace" {
				require.Equal(t, strings.TrimSpace(testCase.inputEmail), identity.Email())
			} else {
				require.Equal(t, testCase.inputEmail, identity.Email())
			}

			require.Equal(t, testCase.wantString, identity.String())
			require.Equal(t, testCase.wantEmpty, identity.IsEmpty())
		})
	}
}

func TestIdentityFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantEmail string
	}{
		{
			name:      "Standard format",
			input:     "John Doe <john@example.com>",
			wantName:  "John Doe",
			wantEmail: "john@example.com",
		},
		{
			name:      "Email only with angle brackets",
			input:     "<john@example.com>",
			wantName:  "",
			wantEmail: "john@example.com",
		},
		{
			name:      "Email only without brackets",
			input:     "john@example.com",
			wantName:  "",
			wantEmail: "john@example.com",
		},
		{
			name:      "Name only",
			input:     "John Doe",
			wantName:  "John Doe",
			wantEmail: "",
		},
		{
			name:      "With extra whitespace",
			input:     "  John Doe   <  john@example.com  >  ",
			wantName:  "John Doe",
			wantEmail: "john@example.com",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			identity := domain.NewIdentityFromString(testCase.input)

			require.Equal(t, testCase.wantName, identity.Name())
			require.Equal(t, testCase.wantEmail, identity.Email())
		})
	}
}

func TestIdentityMatches(t *testing.T) {
	identity1 := domain.NewIdentity("John Doe", "john@example.com")
	identity2 := domain.NewIdentity("John Smith", "john@example.com")
	identity3 := domain.NewIdentity("Jane Doe", "jane@example.com")
	emptyIdentity := domain.NewIdentity("", "")

	require.True(t, identity1.Matches(identity2), "Identities with same email should match")
	require.False(t, identity1.Matches(identity3), "Identities with different emails should not match")
	require.False(t, identity1.Matches(emptyIdentity), "Identity should not match empty identity")
	require.False(t, emptyIdentity.Matches(identity1), "Empty identity should not match any identity")
}

func TestIdentityMatchesAny(t *testing.T) {
	identity1 := domain.NewIdentity("John Doe", "john@example.com")
	identities := []domain.Identity{
		domain.NewIdentity("Alice", "alice@example.com"),
		domain.NewIdentity("Bob", "bob@example.com"),
		domain.NewIdentity("John Smith", "john@example.com"),
	}

	require.True(t, identity1.MatchesAny(identities), "Should match one of the identities")

	noMatchIdentities := []domain.Identity{
		domain.NewIdentity("Alice", "alice@example.com"),
		domain.NewIdentity("Bob", "bob@example.com"),
	}

	require.False(t, identity1.MatchesAny(noMatchIdentities), "Should not match any identity")
}

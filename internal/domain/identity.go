// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"fmt"
	"strings"
)

// Identity represents a committer or author identity with name and email.
type Identity struct {
	name  string
	email string
}

// NewIdentity creates a new identity with separate name and email.
func NewIdentity(name, email string) Identity {
	return Identity{
		name:  strings.TrimSpace(name),
		email: strings.TrimSpace(email),
	}
}

// NewIdentityFromString parses an identity string in the format "Name <email@example.com>".
func NewIdentityFromString(identity string) Identity {
	name, email := parseIdentityString(identity)

	return NewIdentity(name, email)
}

// Name returns the identity's name.
func (i Identity) Name() string {
	return i.name
}

// Email returns the identity's email.
func (i Identity) Email() string {
	return i.email
}

// String returns the identity in standard "Name <email>" format.
func (i Identity) String() string {
	// If both name and email are empty, return empty string
	if i.name == "" && i.email == "" {
		return ""
	}

	// If only name is empty, return just the email
	if i.name == "" {
		return i.email
	}

	// Normal case with both name and email
	return fmt.Sprintf("%s <%s>", i.name, i.email)
}

// Matches checks if this identity matches another one.
// Two identities match if they have the same email address.
func (i Identity) Matches(other Identity) bool {
	return i.email != "" && strings.EqualFold(i.email, other.email)
}

// MatchesAny checks if this identity matches any in the provided slice.
func (i Identity) MatchesAny(identities []Identity) bool {
	for _, other := range identities {
		if i.Matches(other) {
			return true
		}
	}

	return false
}

// IsEmpty returns true if both name and email are empty.
func (i Identity) IsEmpty() bool {
	return i.name == "" && i.email == ""
}

// parseIdentityString extracts name and email from a string like "Name <email@example.com>".
func parseIdentityString(identity string) (string, string) {
	identity = strings.TrimSpace(identity)

	// Check if the identity is in standard format
	start := strings.LastIndex(identity, "<")
	end := strings.LastIndex(identity, ">")

	if start != -1 && end != -1 && start < end {
		userName := strings.TrimSpace(identity[:start])
		userEmail := strings.TrimSpace(identity[start+1 : end])

		return userName, userEmail
	}

	// If not in standard format but contains @, treat as email
	if strings.Contains(identity, "@") {
		return "", identity
	}

	// Otherwise, treat as name
	return identity, ""
}

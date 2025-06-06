// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// IdentityRule validates that commit signatures match the committer identity.
type IdentityRule struct {
	allowedSigners []string
}

// NewIdentityRule creates a new rule for validating signature identity from config.
func NewIdentityRule(cfg config.Config) IdentityRule {
	return IdentityRule{
		allowedSigners: cfg.Signing.AllowedSigners,
	}
}

// Validate validates that commit authors are in the allowed signers list.
func (r IdentityRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// If no allowed signers configured, allow all authors
	if len(r.allowedSigners) == 0 {
		return nil
	}

	// Check if commit author is in allowed signers list
	authorString := commit.Author + " <" + commit.AuthorEmail + ">"

	for _, allowedSigner := range r.allowedSigners {
		// Exact match for full format
		if allowedSigner == authorString {
			return nil // Author is allowed
		}
		// Case-insensitive email-only match
		if strings.EqualFold(allowedSigner, commit.AuthorEmail) {
			return nil // Author email is allowed
		}
		// Extract email from "Name <email>" format and compare case-insensitively
		emailRegex := regexp.MustCompile(`<([^>]+)>`)
		if matches := emailRegex.FindStringSubmatch(allowedSigner); len(matches) > 1 {
			allowedEmail := matches[1]
			if strings.EqualFold(allowedEmail, commit.AuthorEmail) {
				return nil // Email extracted from allowed signer matches
			}
		}
	}

	return []domain.ValidationError{
		domain.New(r.Name(), domain.ErrKeyNotTrusted, "Author not in allowed signers list").
			WithHelp("Use an authorized identity or add this author to the allowed signers list"),
	}
}

// Name returns the rule name.
func (r IdentityRule) Name() string {
	return "SignedIdentity"
}

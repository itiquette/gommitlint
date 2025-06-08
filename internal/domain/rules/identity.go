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

// IdentityRule validates that commit authors are in the allowed list.
type IdentityRule struct {
	allowedAuthors []string
}

// NewIdentityRule creates a new rule for validating author identity from config.
func NewIdentityRule(cfg config.Config) IdentityRule {
	return IdentityRule{
		allowedAuthors: cfg.Identity.AllowedAuthors,
	}
}

// Validate validates that commit authors are in the allowed authors list.
func (r IdentityRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// If no allowed authors configured, allow all authors
	if len(r.allowedAuthors) == 0 {
		return nil
	}

	return r.validateAuthorIdentity(commit)
}

// Name returns the rule name.
func (r IdentityRule) Name() string {
	return "Identity"
}

// validateAuthorIdentity checks if commit author is in allowed list.
func (r IdentityRule) validateAuthorIdentity(commit domain.Commit) []domain.ValidationError {
	authorString := commit.Author + " <" + commit.AuthorEmail + ">"

	for _, allowedAuthor := range r.allowedAuthors {
		if r.matchesAuthor(allowedAuthor, authorString, commit.AuthorEmail) {
			return nil
		}
	}

	return []domain.ValidationError{
		domain.New(r.Name(), domain.ErrKeyNotTrusted, "Author not authorized").
			WithContextMap(map[string]string{
				"actual":   commit.AuthorEmail,
				"expected": strings.Join(r.allowedAuthors, ", "),
			}).
			WithHelp("Use an authorized identity or add this author to the allowed authors list"),
	}
}

// matchesAuthor checks if an allowed author matches the commit author.
func (r IdentityRule) matchesAuthor(allowedAuthor, authorString, authorEmail string) bool {
	// Exact match for full format
	if allowedAuthor == authorString {
		return true
	}

	// Case-insensitive email-only match
	if strings.EqualFold(allowedAuthor, authorEmail) {
		return true
	}

	// Extract email from "Name <email>" format and compare case-insensitively
	emailRegex := regexp.MustCompile(`<([^>]+)>`)
	if matches := emailRegex.FindStringSubmatch(allowedAuthor); len(matches) > 1 {
		allowedEmail := matches[1]
		if strings.EqualFold(allowedEmail, authorEmail) {
			return true
		}
	}

	return false
}

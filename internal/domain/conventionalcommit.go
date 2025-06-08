// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"regexp"
	"strings"
)

// ConventionalCommitFormat represents the parsed components of a conventional commit.
type ConventionalCommitFormat struct {
	Type        string   // e.g., "feat", "fix", "docs"
	Scopes      []string // e.g., ["ui", "api"] from "feat(ui,api):"
	Breaking    bool     // true if "!" is present
	Description string   // the text after the colon
	RawScope    string   // original scope string for error reporting
	IsValid     bool     // whether the subject matches conventional format
}

var (
	// Format: <type>[optional scope][optional !]: <description>.
	conventionalCommitRegex = regexp.MustCompile(
		`^(?P<type>[a-zA-Z]+)(?:\((?P<scope>[a-z0-9/,-]+)\))?(?P<breaking>!)?:(?P<space>\s?)(?P<description>.*)`,
	)

	// partialConventionalRegex matches strings that look like they're trying to be conventional
	// but may have formatting issues.
	partialConventionalRegex = regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:`)
)

// ParseConventionalCommit parses a commit subject and returns its conventional commit components.
// This is a pure function that can be used by multiple rules.
// Uses the same parsing logic as the ConventionalCommit rule for consistency.
func ParseConventionalCommit(subject string) ConventionalCommitFormat {
	// Don't trim the subject as spacing is part of the conventional format validation
	match := conventionalCommitRegex.FindStringSubmatch(subject)
	if match == nil {
		return ConventionalCommitFormat{
			IsValid: false,
		}
	}

	// Extract named groups using the same approach as ConventionalCommit rule
	groups := make(map[string]string)

	for i, name := range conventionalCommitRegex.SubexpNames() {
		if i != 0 && name != "" && i < len(match) {
			groups[name] = match[i]
		}
	}

	// Parse scopes (support multiple scopes like "ui,api")
	var scopes []string

	rawScope := groups["scope"]
	if rawScope != "" {
		if strings.Contains(rawScope, ",") {
			// Split by comma and trim whitespace
			for _, scope := range strings.Split(rawScope, ",") {
				if trimmed := strings.TrimSpace(scope); trimmed != "" {
					scopes = append(scopes, trimmed)
				}
			}
		} else {
			scopes = []string{rawScope}
		}
	}

	description := strings.TrimSpace(groups["description"])

	// A valid conventional commit structure just requires the basic format
	// Individual rules handle spacing and content validation
	isValidFormat := true

	return ConventionalCommitFormat{
		Type:        groups["type"],
		Scopes:      scopes,
		Breaking:    groups["breaking"] == "!",
		Description: description,
		RawScope:    rawScope,
		IsValid:     isValidFormat,
	}
}

// IsConventionalCommit returns true if the subject matches conventional commit format.
// This is a convenience function for simple boolean checks.
func IsConventionalCommit(subject string) bool {
	return ParseConventionalCommit(subject).IsValid
}

// IsConventionalCommitLike returns true if the subject looks like it's trying to be
// a conventional commit but may have formatting issues.
func IsConventionalCommitLike(subject string) bool {
	subject = strings.TrimSpace(subject)

	// Check if it matches partial conventional patterns
	if partialConventionalRegex.MatchString(subject) {
		return true
	}

	// Check for malformed conventional patterns:
	// 1. Starts with word and opening parenthesis but no closing paren or colon
	malformedScopePattern := regexp.MustCompile(`^[a-zA-Z]+\([^)]*$`)
	if malformedScopePattern.MatchString(subject) {
		return true
	}

	// 2. Common conventional types followed by space but no colon
	missingColonPattern := regexp.MustCompile(`^(?:feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)\s+`)

	return missingColonPattern.MatchString(subject)
}

// ExtractDescriptionFromConventional extracts just the description part from a conventional commit.
// Returns the full subject if it's not a conventional commit.
func ExtractDescriptionFromConventional(subject string) string {
	parsed := ParseConventionalCommit(subject)
	if parsed.IsValid && parsed.Description != "" {
		return parsed.Description
	}

	return subject
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// CommitBodyRule validates commit message bodies.
type CommitBodyRule struct {
	minLength   int
	minLines    int
	signOffOnly bool
}

// NewCommitBodyRule creates a new CommitBodyRule from config.
func NewCommitBodyRule(cfg config.Config) CommitBodyRule {
	return CommitBodyRule{
		minLength:   cfg.Message.Body.MinLength,
		minLines:    cfg.Message.Body.MinLines,
		signOffOnly: cfg.Message.Body.AllowSignoffOnly,
	}
}

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return "CommitBody"
}

// Validate checks if a commit's body meets the required criteria.
func (r CommitBodyRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	trimmedBody := strings.TrimSpace(commit.Body)
	hasOnlySignOff := hasOnlySignOffLines(trimmedBody)
	bodyLength := len(trimmedBody)

	var failures []domain.ValidationError

	// Check missing body
	if r.minLength > 0 && trimmedBody == "" {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrMissingBody, "commit body is missing").
				WithHelp("Add a blank line after the subject, followed by a detailed description"))
	}

	// Check sign-off only
	if !r.signOffOnly && hasOnlySignOff && trimmedBody != "" {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrInvalidBody, "commit body cannot contain only sign-off line").
				WithHelp("Add a detailed description before the sign-off line"))
	}

	// Check minimum length
	if r.minLength > 0 && bodyLength < r.minLength && trimmedBody != "" {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrBodyTooShort,
				fmt.Sprintf("body too short (minimum: %d characters, actual: %d)", r.minLength, bodyLength)).
				WithHelp(fmt.Sprintf("Provide at least %d characters of detail", r.minLength)))
	}

	// Check minimum lines
	if r.minLines > 0 && trimmedBody != "" && !(hasOnlySignOff && r.signOffOnly) {
		lines := strings.Split(trimmedBody, "\n")
		actualLines := len(lines)

		if actualLines < r.minLines {
			failures = append(failures,
				domain.New(r.Name(), domain.ErrBodyTooShort,
					fmt.Sprintf("body has too few lines (minimum: %d, actual: %d)", r.minLines, actualLines)).
					WithHelp(fmt.Sprintf("Provide at least %d lines of detail", r.minLines)))
		}
	}

	return failures
}

// hasOnlySignOffLines checks if a commit body contains only sign-off lines
// like "Signed-off-by: Name <email>".
func hasOnlySignOffLines(body string) bool {
	if body == "" {
		return false
	}

	// Pattern for sign-off lines (multiple formats supported)
	signOffPattern := regexp.MustCompile(`(?m)^(Signed-off-by|Co-authored-by|Reviewed-by|Acked-by|Tested-by):.*<.*@.*>.*$`)

	// Split the body into lines and remove empty lines
	var contentLines []string

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			contentLines = append(contentLines, trimmed)
		}
	}

	// If no content lines, return false
	if len(contentLines) == 0 {
		return false
	}

	// Check if all non-empty lines are sign-off lines
	for _, line := range contentLines {
		if !signOffPattern.MatchString(line) {
			return false
		}
	}

	return true
}

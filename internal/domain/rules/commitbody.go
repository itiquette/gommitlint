// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"regexp"
	"strconv"
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
			domain.New(r.Name(), domain.ErrMissingBody, fmt.Sprintf("Missing body (requires %d+ characters)", r.minLength)).
				WithContextMap(map[string]string{
					"required_length": strconv.Itoa(r.minLength),
					"current_state":   "No body present",
				}).
				WithHelp("Add a blank line after the subject, followed by a detailed description"))
	}

	// Check sign-off only
	if !r.signOffOnly && hasOnlySignOff && trimmedBody != "" {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrInvalidBody, "Cannot have only sign-off lines").
				WithHelp("Add a detailed description before the sign-off line"))
	}

	// Check minimum length
	if r.minLength > 0 && bodyLength < r.minLength && trimmedBody != "" {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrBodyTooShort,
				fmt.Sprintf("Too short (%d/%d characters)", bodyLength, r.minLength)).
				WithContextMap(map[string]string{
					"current_length":  strconv.Itoa(bodyLength),
					"required_length": strconv.Itoa(r.minLength),
					"deficit":         strconv.Itoa(r.minLength - bodyLength),
				}).
				WithHelp(fmt.Sprintf("Provide at least %d characters of detail explaining the change", r.minLength)))
	}

	// Check minimum lines
	if r.minLines > 0 && trimmedBody != "" && !(hasOnlySignOff && r.signOffOnly) {
		lines := strings.Split(trimmedBody, "\n")
		actualLines := len(lines)

		if actualLines < r.minLines {
			failures = append(failures,
				domain.New(r.Name(), domain.ErrBodyTooShort,
					fmt.Sprintf("Too few lines (%d/%d required)", actualLines, r.minLines)).
					WithHelp(fmt.Sprintf("Provide at least %d lines of detailed explanation", r.minLines)))
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

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/functional"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// CommitBodyRule validates commit message bodies based on configured requirements.
type CommitBodyRule struct {
	name        string
	minLength   int
	minLines    int  // Minimum number of lines in the body
	signOffOnly bool // Whether to allow commits with only sign-off lines
}

// CommitBodyOption configures a CommitBodyRule.
type CommitBodyOption func(CommitBodyRule) CommitBodyRule

// WithMinLength sets the minimum body length.
func WithMinLength(length int) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		if length > 0 {
			r.minLength = length
		}

		return r
	}
}

// WithMinLines sets the minimum number of lines in the body.
func WithMinLines(lines int) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		if lines > 0 {
			r.minLines = lines
		}

		return r
	}
}

// WithAllowSignOffOnly sets whether to allow commits with only sign-off lines.
func WithAllowSignOffOnly(allow bool) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		r.signOffOnly = allow

		return r
	}
}

// NewCommitBodyRule creates a new CommitBodyRule.
func NewCommitBodyRule(options ...CommitBodyOption) CommitBodyRule {
	rule := CommitBodyRule{
		name:        "CommitBody",
		minLength:   0,    // Default to no minimum
		minLines:    0,    // Default to no minimum
		signOffOnly: true, // Default to allow sign-off only
	}

	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return r.name
}

// Validate checks if a commit's body meets the required criteria.
// Rule enabling/disabling is handled by the rule registry, so we don't check it here.
func (r CommitBodyRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Pre-compute values used in multiple checks
	trimmedBody := strings.TrimSpace(commit.Body)
	hasOnlySignOff := hasOnlySignOffLines(trimmedBody)
	bodyLength := len(trimmedBody)

	// Define all validation checks as pure functions
	checkMissingBody := func() []appErrors.ValidationError {
		if r.minLength > 0 && trimmedBody == "" {
			return functional.SingleError(appErrors.NewBodyError(
				appErrors.ErrMissingBody,
				r.Name(),
				"Commit body is missing",
				"is empty",
				"Add a blank line after the subject, followed by a detailed description of your changes",
			).WithContextMap(map[string]string{"commit": commit.Hash}))
		}

		return nil
	}

	checkSignOffOnly := func() []appErrors.ValidationError {
		if !r.signOffOnly && hasOnlySignOff && trimmedBody != "" {
			return functional.SingleError(appErrors.NewBodyError(
				appErrors.ErrMissingBody,
				r.Name(),
				"Commit body cannot contain only sign-off line",
				"contains only sign-off",
				"Add a detailed description of your changes before the sign-off line",
			).WithContextMap(map[string]string{"commit": commit.Hash}))
		}

		return nil
	}

	checkMinLength := func() []appErrors.ValidationError {
		if r.minLength > 0 && bodyLength < r.minLength && trimmedBody != "" {
			return functional.SingleError(appErrors.NewBodyError(
				appErrors.ErrBodyTooShort,
				r.Name(),
				fmt.Sprintf("Commit body must be at least %d characters", r.minLength),
				fmt.Sprintf("has %d characters", bodyLength),
				fmt.Sprintf("Provide at least %d characters of detail in your commit body", r.minLength),
			).WithContextMap(map[string]string{
				"commit":        commit.Hash,
				"min_length":    strconv.Itoa(r.minLength),
				"actual_length": strconv.Itoa(bodyLength),
			}))
		}

		return nil
	}

	checkMinLines := func() []appErrors.ValidationError {
		// Skip if body is allowed to have only sign-off lines and it does
		if hasOnlySignOff && r.signOffOnly {
			return nil
		}

		if r.minLines > 0 && trimmedBody != "" {
			lines := strings.Split(trimmedBody, "\n")
			actualLines := len(lines)

			if actualLines < r.minLines {
				return functional.SingleError(appErrors.NewBodyError(
					appErrors.ErrBodyTooShort,
					r.Name(),
					fmt.Sprintf("Commit body must have at least %d lines", r.minLines),
					fmt.Sprintf("has %d lines", actualLines),
					fmt.Sprintf("Provide at least %d lines of detail in your commit body", r.minLines),
				).WithContextMap(map[string]string{
					"commit":       commit.Hash,
					"min_lines":    strconv.Itoa(r.minLines),
					"actual_lines": strconv.Itoa(actualLines),
				}))
			}
		}

		return nil
	}

	// Combine all validation checks functionally
	return functional.AllErrors(
		checkMissingBody,
		checkSignOffOnly,
		checkMinLength,
		checkMinLines,
	)
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

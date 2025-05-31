// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
)

// signOffRegex returns a regex for matching sign-off lines.
func signOffRegex() *regexp.Regexp {
	return regexp.MustCompile(`^Signed-off-by:\s+.+\s+<.+>$`)
}

// SignOffRule validates that commit messages include a sign-off line.
type SignOffRule struct {
	name            string
	requireSignOff  bool
	acceptAltFormat bool
}

// NewSignOffRule creates a new rule for validating commit sign-offs from config.
func NewSignOffRule(cfg config.Config) SignOffRule {
	return SignOffRule{
		name:            "SignOff",
		requireSignOff:  cfg.Message.Body.RequireSignoff,
		acceptAltFormat: cfg.Signing.RequireMultiSignoff,
	}
}

// Validate checks for the presence and format of a Developer Certificate of Origin sign-off.
func (r SignOffRule) Validate(_ context.Context, commit domain.CommitInfo) []domain.ValidationError {
	// Check if sign-off is required
	if !r.requireSignOff {
		return nil
	}

	// Use the commit message text for validation
	messageText := commit.Message
	bodyText := commit.Body

	// If message is empty but body isn't, use the body as the message
	// This is a valid production scenario where only the body field might be populated
	if messageText == "" && bodyText != "" {
		messageText = bodyText
	}

	// Handle empty message cases separately - only if both Message and Body are empty
	if messageText == "" && bodyText == "" {
		return []domain.ValidationError{
			domain.New(
				"SignOff",
				domain.ErrMissingSignoff,
				"Missing sign-off",
			).WithHelp("Add 'Signed-off-by: Your Name <email@example.com>'"),
		}
	}

	// For signoff checking, focus on the body text
	textToCheck := bodyText
	if bodyText == "" {
		// If no dedicated body, check the whole message
		textToCheck = messageText
	}

	// Check for sign-off in the text
	hasSignOff := hasSignOffLine(textToCheck, r.acceptAltFormat)
	if !hasSignOff {
		return []domain.ValidationError{
			domain.New(
				"SignOff",
				domain.ErrMissingSignoff,
				"Missing sign-off",
			).WithHelp("Add 'Signed-off-by: Your Name <email@example.com>'").WithContextMap(map[string]string{
				"author": fmt.Sprintf("%s <%s>", commit.AuthorName, commit.AuthorEmail),
			}),
		}
	}

	return nil
}

// Name returns the rule name.
func (r SignOffRule) Name() string {
	return r.name
}

// hasSignOffLine checks if a commit body contains a sign-off line.
func hasSignOffLine(body string, acceptAltFormat bool) bool {
	lines := strings.Split(body, "\n")
	regex := signOffRegex()

	// Find where sign-offs start by iterating from the end backwards
	signOffStartIdx := -1
	foundNonSignOff := false

	for idx := len(lines) - 1; idx >= 0; idx-- {
		trimmedLine := strings.TrimSpace(lines[idx])

		// Skip empty lines at the end
		if trimmedLine == "" {
			continue
		}

		// Check if this line is a sign-off
		isSignOff := false
		if regex.MatchString(trimmedLine) {
			isSignOff = true
		} else if acceptAltFormat && isSignOffLine(trimmedLine) {
			isSignOff = true
		}

		if isSignOff {
			if foundNonSignOff {
				// Found sign-off after non-sign-off content, invalid placement
				return false
			}

			signOffStartIdx = idx
		} else {
			// This is not a sign-off line
			foundNonSignOff = true
		}
	}

	// Return true only if we found at least one sign-off
	return signOffStartIdx != -1
}

// isSignOffLine checks if a line is a sign-off line.
func isSignOffLine(line string) bool {
	// Line should already be trimmed by the caller
	prefixes := []string{
		"Signed-off-by:",
		"Co-authored-by:",
		"Reviewed-by:",
		"Tested-by:",
		"Acked-by:",
		"Cc:",
		"Reported-by:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return false
}

// CountBodyLines counts the number of non-empty lines in a body.
func CountBodyLines(body string) int {
	// If body is empty, return 0
	if strings.TrimSpace(body) == "" {
		return 0
	}

	lines := strings.Split(body, "\n")
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			count++
		}
	}

	return count
}

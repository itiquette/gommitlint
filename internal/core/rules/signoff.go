// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SignOffRegex is a regex for matching sign-off lines.
var SignOffRegex = regexp.MustCompile(`(?i)^Signed-off-by:\s+.+\s+<.+>$`)

// SignOffRule validates that commit messages include a sign-off line.
type SignOffRule struct {
	name            string
	requireSignOff  bool
	acceptAltFormat bool
}

// SignOffOption configures a SignOffRule.
type SignOffOption func(SignOffRule) SignOffRule

// WithRequireSignOff configures whether sign-offs are required.
func WithRequireSignOff(required bool) SignOffOption {
	return func(r SignOffRule) SignOffRule {
		result := r
		result.requireSignOff = required

		return result
	}
}

// WithMultipleSignoffs configures whether multiple sign-offs are allowed.
func WithMultipleSignoffs(allow bool) SignOffOption {
	return func(r SignOffRule) SignOffRule {
		result := r
		result.acceptAltFormat = allow // Reuse this flag

		return result
	}
}

// WithCustomSignOffRegex configures a custom regex for sign-off validation.
func WithCustomSignOffRegex(_ *regexp.Regexp) SignOffOption {
	return func(r SignOffRule) SignOffRule {
		// For future implementation - in a fully featured solution, this would store the custom regex
		result := r

		return result
	}
}

// NewSignOffRule creates a new rule for validating commit sign-offs.
func NewSignOffRule(options ...SignOffOption) SignOffRule {
	// Create a rule with default values
	rule := SignOffRule{
		name:            "SignOff",
		requireSignOff:  true,  // Default to requiring sign-off
		acceptAltFormat: false, // Default to strict format
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// WithContext implements the ConfigurableRule interface for SignOffRule.
// It returns a new rule with configuration from the provided context.
func (r SignOffRule) WithContext(ctx context.Context) domain.Rule {
	// Get config from common interface
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return r
	}

	// Default values - use the rule's current values
	requireSignOff := r.requireSignOff
	acceptAltFormat := r.acceptAltFormat

	// Try to get security settings from config
	if cfg.GetBool("message.body.require_signoff") {
		requireSignOff = true
	}

	if cfg.GetBool("signing.allow_multiple_signoffs") {
		acceptAltFormat = true
	}

	// Log configuration at debug level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Sign-off rule configuration from context",
		"require_sign_off", requireSignOff,
		"accept_alternative_format", acceptAltFormat)

	// Create a copy of the rule
	result := r

	// Update settings from context
	result.requireSignOff = requireSignOff
	result.acceptAltFormat = acceptAltFormat

	return result
}

// Validate checks for the presence and format of a Developer Certificate of Origin sign-off.
func (r SignOffRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating sign-off",
		"rule", r.Name(),
		"commit_hash", commit.Hash)

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
		return []appErrors.ValidationError{
			appErrors.NewSignOffError(
				appErrors.ErrMissingSignoff,
				"SignOff",
				"Missing sign-off",
				"Add 'Signed-off-by: Your Name <email@example.com>'",
			),
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
		return []appErrors.ValidationError{
			appErrors.NewSignOffError(
				appErrors.ErrMissingSignoff,
				"SignOff",
				"Missing sign-off",
				"Add 'Signed-off-by: Your Name <email@example.com>'",
			).WithContextMap(map[string]string{
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
	// Standard format: "Signed-off-by: Name <email@example.com>"
	standardRegex := regexp.MustCompile(`(?m)^Signed-off-by:\s+.+\s+<.+>$`)
	if standardRegex.MatchString(body) {
		return true
	}

	// Alternative formats if accepted
	if acceptAltFormat {
		// Check if any line is a sign-off line as defined in isSignOffLine
		for _, line := range strings.Split(body, "\n") {
			if isSignOffLine(strings.TrimSpace(line)) {
				return true
			}
		}
	}

	return false
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

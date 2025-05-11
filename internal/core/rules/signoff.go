// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// SignOffRegex is a regex for matching sign-off lines.
var SignOffRegex = regexp.MustCompile(`(?i)^Signed-off-by:\s+.+\s+<.+>$`)

// SignOffRule validates that commit messages include a sign-off line.
type SignOffRule struct {
	baseRule        BaseRule
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

// WithAllowMultipleSignOffs configures whether multiple sign-offs are allowed.
func WithAllowMultipleSignOffs(allow bool) SignOffOption {
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
		baseRule:        NewBaseRule("SignOff"),
		requireSignOff:  true,  // Default to requiring sign-off
		acceptAltFormat: false, // Default to strict format
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate checks for the presence and format of a Developer Certificate of Origin sign-off
// using configuration from context.
func (r SignOffRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating sign-off using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the existing validation logic
	errors, _ := validateSignOffWithState(rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r SignOffRule) withContextConfig(ctx context.Context) SignOffRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Extract configuration values
	requireSignOff := cfg.Security.SignOffRequired
	acceptAltFormat := cfg.Security.AllowMultipleSignOffs // Reuse this flag for alternative formats

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Bool("require_sign_off", requireSignOff).
		Bool("accept_alternative_format", acceptAltFormat).
		Msg("Sign-off rule configuration from context")

	// Create a copy of the rule
	result := r

	// Update settings from context
	result.requireSignOff = requireSignOff
	result.acceptAltFormat = acceptAltFormat

	return result
}

// validateSignOffWithState validates the commit sign-off and returns both the errors and an updated rule.
func validateSignOffWithState(rule SignOffRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SignOffRule) {
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Check if sign-off is required
	if !rule.requireSignOff {
		return result.baseRule.Errors(), result
	}

	// If no body, there can't be a sign-off
	if commit.Body == "" {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrMissingSignoff,
			"commit is missing a sign-off line",
		)
		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// Check for sign-off in body
	hasSignOff := hasSignOffLine(commit.Body, rule.acceptAltFormat)
	if !hasSignOff {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrMissingSignoff,
			"commit is missing a sign-off line",
		).WithContext("author", commit.AuthorName+" <"+commit.AuthorEmail+">")
		result.baseRule = result.baseRule.WithError(validationErr)
	}

	return result.baseRule.Errors(), result
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

// Name returns the rule name.
func (r SignOffRule) Name() string {
	return r.baseRule.Name()
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r SignOffRule) SetErrors(errors []appErrors.ValidationError) SignOffRule {
	result := r
	result.baseRule = r.baseRule.WithClearedErrors()

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)
	}

	return result
}

// Errors returns all validation errors found by this rule.
func (r SignOffRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SignOffRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// SetSignOffInfo sets sign-off information for detailed output and returns an updated rule.
func (r SignOffRule) SetSignOffInfo(_ bool, _ string) SignOffRule {
	// This method isn't used in the new implementation but kept for compatibility
	return r
}

// Result returns a concise validation result.
func (r SignOffRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return "❌ Missing sign-off"
	}

	return "✓ Properly signed-off"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SignOffRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		var author string

		for _, err := range errors {
			if err.Context != nil && err.Context["author"] != "" {
				author = err.Context["author"]

				break
			}
		}

		if author != "" {
			return "❌ Commit is missing a sign-off line for author: " + author
		}

		return "❌ Commit is missing a sign-off line"
	}

	return "✓ Commit is properly signed-off"
}

// Help returns guidance for fixing rule violations.
func (r SignOffRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	helpText := "Your commit must include a sign-off line indicating that you agree to the Developer Certificate of Origin (DCO).\n\n"

	helpText += "Add a sign-off to your commit using one of these methods:\n"
	helpText += "1. Use the `-s` or `--signoff` flag with git commit:\n"
	helpText += "   `git commit -s -m \"Your commit message\"`\n\n"

	helpText += "2. Manually add a line to the end of your commit message:\n"
	helpText += "   ```\n"
	helpText += "   Signed-off-by: Your Name <your.email@example.com>\n"
	helpText += "   ```\n\n"

	helpText += "The sign-off certifies that you have the right to submit your contribution under the project's license."

	return helpText
}

// IsSignOffOnly checks if a body contains only sign-off lines.
func IsSignOffOnly(body string) bool {
	// If body is empty, it doesn't contain only sign-off lines
	if strings.TrimSpace(body) == "" {
		return false
	}

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// If any non-empty line is not a sign-off line, return false
		if !isSignOffLine(trimmed) {
			return false
		}
	}

	// All non-empty lines are sign-off lines
	return true
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

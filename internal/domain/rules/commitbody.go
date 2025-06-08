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
	required         bool
	minLength        int
	allowSignOffOnly bool
}

// NewCommitBodyRule creates a new CommitBodyRule from config.
func NewCommitBodyRule(cfg config.Config) CommitBodyRule {
	return CommitBodyRule{
		required:         cfg.Message.Body.Required,
		minLength:        cfg.Message.Body.MinLength,
		allowSignOffOnly: cfg.Message.Body.AllowSignoffOnly,
	}
}

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return "CommitBody"
}

// Validate checks if a commit's body meets the required criteria.
func (r CommitBodyRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Skip merge commits
	if commit.IsMergeCommit {
		return nil
	}

	trimmedBody := strings.TrimSpace(commit.Body)

	// Functional composition of validations
	var errors []domain.ValidationError

	errors = append(errors, r.validateStructure(commit)...)
	errors = append(errors, r.validateLength(trimmedBody)...)
	errors = append(errors, r.validateSignOffRules(trimmedBody)...)

	return errors
}

// validateStructure validates Git message structure (subject + blank line + body).
func (r CommitBodyRule) validateStructure(commit domain.Commit) []domain.ValidationError {
	lines := strings.Split(commit.Message, "\n")

	// If requiring body but parsed body is empty (no proper structure)
	if r.required && strings.TrimSpace(commit.Body) == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrInvalidStructure, "Invalid commit message structure").
				WithContextMap(map[string]string{
					"actual":   "subject only",
					"expected": "subject + blank line + body",
				}).
				WithHelp("Use format: subject line, blank line, then detailed body"),
		}
	}

	// Always validate blank line separation when there's content after subject
	if len(lines) >= 2 && strings.TrimSpace(lines[1]) != "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingBlankLine, "Missing blank line between subject and body").
				WithContextMap(map[string]string{
					"actual":   "subject + body (no blank line)",
					"expected": "subject + blank line + body",
				}).
				WithHelp("Git convention requires a blank line between subject and body"),
		}
	}

	return nil
}

// validateLength validates minimum body length requirement.
func (r CommitBodyRule) validateLength(trimmedBody string) []domain.ValidationError {
	if !r.required || r.minLength == 0 {
		return nil
	}

	if trimmedBody == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingBody, fmt.Sprintf("Missing body (requires %d+ characters)", r.minLength)).
				WithContextMap(map[string]string{
					"actual":   "no body",
					"expected": fmt.Sprintf("min %d characters", r.minLength),
				}).
				WithHelp("Add a blank line after the subject, followed by a detailed description"),
		}
	}

	bodyLength := len(trimmedBody)
	if bodyLength < r.minLength {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrBodyTooShort,
				fmt.Sprintf("Too short (%d/%d characters)", bodyLength, r.minLength)).
				WithContextMap(map[string]string{
					"actual":   strconv.Itoa(bodyLength),
					"expected": fmt.Sprintf("min %d", r.minLength),
				}).
				WithHelp(fmt.Sprintf("Provide at least %d characters of detail explaining the change", r.minLength)),
		}
	}

	return nil
}

// validateLines is no longer used - removed minLines configuration.

// validateSignOffRules validates sign-off positioning and content rules.
func (r CommitBodyRule) validateSignOffRules(trimmedBody string) []domain.ValidationError {
	if trimmedBody == "" {
		return nil
	}

	var errors []domain.ValidationError

	// Check positioning (always validate)
	errors = append(errors, r.validateSignOffPositioning(trimmedBody)...)

	// Check content rules only if required and sign-off-only is not allowed
	if r.required && !r.allowSignOffOnly {
		// Show only the most relevant error to avoid confusion
		errors = append(errors, r.validateSignOffContent(trimmedBody)...)
	}

	return errors
}

// validateSignOffPositioning ensures sign-offs are at the end with trailers allowed after.
func (r CommitBodyRule) validateSignOffPositioning(body string) []domain.ValidationError {
	if body == "" {
		return nil
	}

	lines := strings.Split(body, "\n")
	firstSignOffIndex := r.findFirstSignOffIndex(lines)

	if firstSignOffIndex == -1 {
		return nil
	}

	// Check content after first sign-off - only trailers allowed
	trailerPattern := regexp.MustCompile(`^(Signed-off-by|Co-authored-by|Reviewed-by|Acked-by|Tested-by|Reported-by):.*<.*@.*>.*$|^(Fixes|Closes|Resolves):\s*(#\d+|https?://[^\s]+).*$`)

	for i := firstSignOffIndex + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}

		if !trailerPattern.MatchString(trimmed) {
			return []domain.ValidationError{
				domain.New(r.Name(), domain.ErrMisplacedSignoff, "Content found after sign-off lines").
					WithContextMap(map[string]string{
						"actual":   trimmed,
						"expected": "sign-off lines at end",
					}).
					WithHelp("Move all sign-off lines to the very end of the commit body"),
			}
		}
	}

	return nil
}

// validateSignOffContent checks if body has only sign-offs (when not allowed).
func (r CommitBodyRule) validateSignOffContent(body string) []domain.ValidationError {
	if body == "" {
		return nil
	}

	// Check if body only contains sign-offs
	if r.hasOnlySignOffLines(body) {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrInvalidBody, "Body contains only sign-off lines").
				WithContextMap(map[string]string{
					"actual":   "only sign-off lines",
					"expected": "descriptive content + sign-offs",
				}).
				WithHelp("Add a detailed description before the sign-off line"),
		}
	}

	// Check if body starts with sign-off (less ideal but still needs descriptive content first)
	firstLine := strings.TrimSpace(strings.Split(body, "\n")[0])
	signOffPattern := regexp.MustCompile(`^Signed-off-by:.*<.*@.*>.*$`)

	if signOffPattern.MatchString(firstLine) {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrInvalidBody, "Body should start with descriptive content").
				WithContextMap(map[string]string{
					"actual":   "starts with sign-off",
					"expected": "descriptive content first",
				}).
				WithHelp("Start with actual content explaining your changes, then add sign-off lines at the end"),
		}
	}

	return nil
}

// hasOnlySignOffLines checks if body contains only sign-off lines.
func (r CommitBodyRule) hasOnlySignOffLines(body string) bool {
	if body == "" {
		return false
	}

	lines := r.getNonEmptyLines(body)
	if len(lines) == 0 {
		return false
	}

	signOffPattern := regexp.MustCompile(`^Signed-off-by:.*<.*@.*>.*$`)
	for _, line := range lines {
		if !signOffPattern.MatchString(line) {
			return false
		}
	}

	return true
}

// getNonEmptyLines extracts non-empty lines from text.
func (r CommitBodyRule) getNonEmptyLines(text string) []string {
	var lines []string

	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}

	return lines
}

// findFirstSignOffIndex finds the index of the first actual sign-off line.
func (r CommitBodyRule) findFirstSignOffIndex(lines []string) int {
	signOffPattern := regexp.MustCompile(`^Signed-off-by:.*<.*@.*>.*$`)
	for i, line := range lines {
		if signOffPattern.MatchString(strings.TrimSpace(line)) {
			return i
		}
	}

	return -1
}

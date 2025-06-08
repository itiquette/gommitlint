// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ValidationTarget represents what should be validated.
// This is a focused value type with single responsibility.
type ValidationTarget struct {
	Type   string // "message", "commit", "range", "count"
	Source string // file path, commit ref, or count
	Target string // end ref for ranges, empty otherwise
}

// NewValidationTarget creates a ValidationTarget from CLI parameters.
// It uses precedence-based logic to determine validation target.
func NewValidationTarget(messageFile, gitReference, commitRange, baseBranch string, commitCount int) (ValidationTarget, error) {
	// Validate all inputs first
	if err := validateInputs(messageFile, gitReference, commitRange, baseBranch, commitCount); err != nil {
		return ValidationTarget{}, err
	}

	// Apply validation source with precedence order
	if messageFile != "" {
		// 1. Message from file (highest priority)
		return ValidationTarget{
			Type:   "message",
			Source: filepath.Clean(messageFile),
			Target: "",
		}, nil
	}

	if baseBranch != "" {
		// 2. Base branch comparison (convenience wrapper: <base-branch>..HEAD)
		return ValidationTarget{
			Type:   "range",
			Source: baseBranch,
			Target: "HEAD",
		}, nil
	}

	if commitRange != "" {
		// 3. Explicit range (full control)
		parts := parseRevisionRange(commitRange)
		if len(parts) == 2 {
			return ValidationTarget{
				Type:   "range",
				Source: parts[0],
				Target: parts[1],
			}, nil
		}

		return ValidationTarget{}, fmt.Errorf("invalid range format: %s (expected format: from..to)", commitRange)
	}

	if gitReference != "" {
		// 4. Single git reference
		return ValidationTarget{
			Type:   "commit",
			Source: gitReference,
			Target: "",
		}, nil
	}

	if commitCount > 1 {
		// 5. Commit count (only if explicitly set to > 1)
		return ValidationTarget{
			Type:   "count",
			Source: strconv.Itoa(commitCount),
			Target: "",
		}, nil
	}

	// Default to HEAD (when commit count is 1 or no options provided)
	return ValidationTarget{
		Type:   "commit",
		Source: "HEAD",
		Target: "",
	}, nil
}

// validateInputs validates all inputs.
func validateInputs(messageFile, gitReference, commitRange, baseBranch string, commitCount int) error {
	if err := validateFilePath(messageFile); err != nil {
		return fmt.Errorf("invalid message file: %w", err)
	}

	if err := validateGitReference(gitReference); err != nil {
		return fmt.Errorf("invalid git reference: %w", err)
	}

	if err := validateGitReference(baseBranch); err != nil {
		return fmt.Errorf("invalid base branch: %w", err)
	}

	if err := validateCommitCount(commitCount); err != nil {
		return fmt.Errorf("invalid commit count: %w", err)
	}

	if commitRange != "" {
		if err := validateParameterLength("Range", commitRange, MaxRefLength); err != nil {
			return err
		}

		// Parse and validate range parts
		parts := parseRevisionRange(commitRange)
		if parts == nil {
			return errors.New("invalid commit range format")
		}

		if len(parts) == 2 {
			if err := validateGitReference(parts[0]); err != nil {
				return fmt.Errorf("invalid range start: %w", err)
			}

			if err := validateGitReference(parts[1]); err != nil {
				return fmt.Errorf("invalid range end: %w", err)
			}
		}
	}

	return nil
}

// IsMessageFile returns true if target is a message file.
func (t ValidationTarget) IsMessageFile() bool {
	return t.Type == "message"
}

// IsCommit returns true if target is a single commit.
func (t ValidationTarget) IsCommit() bool {
	return t.Type == "commit"
}

// IsRange returns true if target is a commit range.
func (t ValidationTarget) IsRange() bool {
	return t.Type == "range"
}

// IsCount returns true if target is a commit count.
func (t ValidationTarget) IsCount() bool {
	return t.Type == "count"
}

// Input validation constraints.
const (
	// MaxPathLength is the maximum allowed length for file paths.
	MaxPathLength = 4096 // Linux PATH_MAX

	// MaxRefLength is the maximum allowed length for git references.
	MaxRefLength = 255 // Git ref name limit

	// MaxCommitCount is the maximum number of commits we'll validate at once.
	MaxCommitCount = 1000
)

// validateFilePath checks if a file path is valid and safe.
func validateFilePath(path string) error {
	if path == "" {
		return nil // Empty path is valid (not used)
	}

	// Special case: "-" means stdin
	if path == "-" {
		return nil
	}

	// Check path length
	if len(path) > MaxPathLength {
		return errors.New("path too long")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return errors.New("path contains null bytes")
	}

	// Ensure it's not trying to escape using ../
	if strings.Contains(path, "..") {
		return errors.New("path cannot contain '..'")
	}

	return nil
}

// validateGitReference checks if a git reference is valid.
func validateGitReference(ref string) error {
	if ref == "" {
		return nil // Empty ref is valid (not used)
	}

	// Check length
	if len(ref) > MaxRefLength {
		return errors.New("reference too long")
	}

	// Check for null bytes
	if strings.Contains(ref, "\x00") {
		return errors.New("reference contains null bytes")
	}

	// Basic git ref validation
	// Git refs cannot start with . or contain ..
	if strings.HasPrefix(ref, ".") || strings.Contains(ref, "..") {
		return errors.New("invalid git reference format")
	}

	// Check for shell metacharacters that could be dangerous
	dangerous := regexp.MustCompile(`[;&|<>$` + "`" + `\\]`)
	if dangerous.MatchString(ref) {
		return errors.New("reference contains invalid characters")
	}

	return nil
}

// validateParameterLength checks if a parameter length is within bounds.
func validateParameterLength(name, value string, maxLength int) error {
	if len(value) > maxLength {
		return fmt.Errorf("%s exceeds maximum length of %d characters", name, maxLength)
	}

	return nil
}

// validateCommitCount checks if a commit count is reasonable.
func validateCommitCount(count int) error {
	if count < 0 {
		return errors.New("commit count cannot be negative")
	}

	if count > MaxCommitCount {
		return fmt.Errorf("commit count exceeds maximum allowed value (%d)", MaxCommitCount)
	}

	return nil
}

// parseRevisionRange parses a commit range string (format: from..to).
// Returns nil if the range format is invalid.
func parseRevisionRange(revRange string) []string {
	// Reject ranges with invalid format patterns
	if strings.HasPrefix(revRange, "..") || strings.HasSuffix(revRange, "..") {
		return nil // Invalid format
	}

	if strings.Contains(revRange, "....") { // 4 or more consecutive dots
		return nil // Invalid format
	}

	// Try ... (symmetric difference) first to avoid false matches with ..
	parts := strings.Split(revRange, "...")
	if len(parts) == 2 {
		return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
	}

	// Split on .. (standard git range format)
	parts = strings.Split(revRange, "..")
	if len(parts) == 2 {
		return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
	}

	// Reject malformed ranges with multiple separators
	if len(parts) > 2 {
		return nil // Invalid format
	}

	return []string{revRange}
}

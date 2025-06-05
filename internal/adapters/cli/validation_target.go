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

// ValidationTargetBuilder helps construct ValidationTarget from CLI parameters.
type ValidationTargetBuilder struct {
	messageFile   string
	gitReference  string
	commitCount   int
	revisionRange string
	baseBranch    string
}

// NewValidationTargetBuilder creates a new builder.
func NewValidationTargetBuilder() *ValidationTargetBuilder {
	return &ValidationTargetBuilder{}
}

// WithMessageFile sets the message file source.
func (b *ValidationTargetBuilder) WithMessageFile(file string) *ValidationTargetBuilder {
	b.messageFile = file

	return b
}

// WithGitReference sets the git reference.
func (b *ValidationTargetBuilder) WithGitReference(ref string) *ValidationTargetBuilder {
	b.gitReference = ref

	return b
}

// WithCommitCount sets the commit count.
func (b *ValidationTargetBuilder) WithCommitCount(count int) *ValidationTargetBuilder {
	b.commitCount = count

	return b
}

// WithRevisionRange sets the revision range.
func (b *ValidationTargetBuilder) WithRevisionRange(revRange string) *ValidationTargetBuilder {
	b.revisionRange = revRange

	return b
}

// WithBaseBranch sets the base branch.
func (b *ValidationTargetBuilder) WithBaseBranch(branch string) *ValidationTargetBuilder {
	b.baseBranch = branch

	return b
}

// Build creates the ValidationTarget with precedence-based logic.
func (b *ValidationTargetBuilder) Build() (ValidationTarget, error) {
	// Validate all inputs first
	if err := b.validateInputs(); err != nil {
		return ValidationTarget{}, err
	}

	// Apply validation source with precedence order
	if b.messageFile != "" {
		// 1. Message from file (highest priority)
		return ValidationTarget{
			Type:   "message",
			Source: filepath.Clean(b.messageFile),
			Target: "",
		}, nil
	}

	if b.baseBranch != "" {
		// 2. Base branch comparison
		return ValidationTarget{
			Type:   "range",
			Source: b.baseBranch,
			Target: "HEAD",
		}, nil
	}

	if b.revisionRange != "" {
		// 3. Revision range
		parts := parseRevisionRange(b.revisionRange)
		if len(parts) == 2 {
			return ValidationTarget{
				Type:   "range",
				Source: parts[0],
				Target: parts[1],
			}, nil
		}

		return ValidationTarget{}, fmt.Errorf("invalid revision range format: %s (expected format: from..to)", b.revisionRange)
	}

	if b.gitReference != "" {
		// 4. Single git reference
		return ValidationTarget{
			Type:   "commit",
			Source: b.gitReference,
			Target: "",
		}, nil
	}

	if b.commitCount > 1 {
		// 5. Commit count (only if explicitly set to > 1)
		return ValidationTarget{
			Type:   "count",
			Source: strconv.Itoa(b.commitCount),
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

// validateInputs validates all builder inputs.
func (b *ValidationTargetBuilder) validateInputs() error {
	if err := validateFilePath(b.messageFile); err != nil {
		return fmt.Errorf("invalid message file: %w", err)
	}

	if err := validateGitReference(b.gitReference); err != nil {
		return fmt.Errorf("invalid git reference: %w", err)
	}

	if err := validateGitReference(b.baseBranch); err != nil {
		return fmt.Errorf("invalid base branch: %w", err)
	}

	if err := validateCommitCount(b.commitCount); err != nil {
		return fmt.Errorf("invalid commit count: %w", err)
	}

	if b.revisionRange != "" {
		if err := validateParameterLength("Revision range", b.revisionRange, MaxRefLength); err != nil {
			return err
		}

		// Parse and validate range parts
		parts := parseRevisionRange(b.revisionRange)
		if len(parts) == 2 {
			if err := validateGitReference(parts[0]); err != nil {
				return fmt.Errorf("invalid revision range start: %w", err)
			}

			if err := validateGitReference(parts[1]); err != nil {
				return fmt.Errorf("invalid revision range end: %w", err)
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

// Validation helper functions (moved from validateparams.go).

// validateFilePath checks if a file path is valid and safe.
func validateFilePath(path string) error {
	if path == "" {
		return nil // Empty path is valid (not used)
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
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
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

// parseRevisionRange parses a revision range string (format: from..to).
func parseRevisionRange(revRange string) []string {
	// Split on .. (standard git range format)
	parts := strings.Split(revRange, "..")
	if len(parts) == 2 {
		return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
	}

	// Try ... (symmetric difference)
	parts = strings.Split(revRange, "...")
	if len(parts) == 2 {
		return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
	}

	return []string{revRange}
}

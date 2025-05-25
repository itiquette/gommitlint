// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/output"
	"github.com/itiquette/gommitlint/internal/ports"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
	"github.com/spf13/cobra"
)

type ValidationParameters struct {
	// No context is stored directly in the struct - we'll pass it as a parameter
	OutWriter io.Writer

	// Validation options
	MessageFile      string
	GitReference     string
	CommitCount      int
	RevisionRange    string
	BaseBranch       string
	RepoPath         string
	SkipMergeCommits bool

	// Report options
	Format       string
	Verbose      bool
	ExtraVerbose bool
	RuleHelp     string
	LightMode    bool
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

// NewValidateParams creates a ValidationParameters from command flags.
func NewValidateParams(cmd *cobra.Command) ValidationParameters {
	params := ValidationParameters{
		OutWriter: cmd.OutOrStdout(),
	}

	// Get flags from command - these should already have defaults from flag definitions
	params.MessageFile, _ = cmd.Flags().GetString("message-file")
	params.GitReference, _ = cmd.Flags().GetString("git-reference")
	params.CommitCount, _ = cmd.Flags().GetInt("commit-count")
	params.RevisionRange, _ = cmd.Flags().GetString("revision-range")
	params.BaseBranch, _ = cmd.Flags().GetString("base-branch")
	params.RepoPath, _ = cmd.Flags().GetString("repo-path")
	params.SkipMergeCommits, _ = cmd.Flags().GetBool("skip-merge-commits")

	// Report options
	params.Format, _ = cmd.Flags().GetString("format")
	params.Verbose, _ = cmd.Flags().GetBool("verbose")
	params.ExtraVerbose, _ = cmd.Flags().GetBool("extra-verbose")
	params.RuleHelp, _ = cmd.Flags().GetString("rulehelp")
	params.LightMode, _ = cmd.Flags().GetBool("light-mode")

	return params
}

// GetRepoPath returns the repository path, with empty string indicating current directory.
func (p ValidationParameters) GetRepoPath() string {
	return p.RepoPath
}

// ToReportOptions converts ValidationParameters to ports.ReportOptions.
func (p ValidationParameters) ToReportOptions() ports.ReportOptions {
	return ports.ReportOptions{
		Format:         p.Format,
		Verbose:        p.Verbose,
		ExtraVerbose:   p.ExtraVerbose,
		ShowHelp:       p.ExtraVerbose || p.RuleHelp != "",
		RuleToShowHelp: p.RuleHelp,
		LightMode:      p.LightMode,
		Writer:         p.OutWriter,
	}
}

// CreateFormatter creates a formatter based on the parameters.
func (p ValidationParameters) CreateFormatter() outgoing.ResultFormatter {
	switch p.Format {
	case "json":
		return output.NewJSONFormatter()
	case "github":
		return output.NewGitHubFormatter().
			WithVerbose(p.Verbose).
			WithShowHelp(p.ExtraVerbose || p.RuleHelp != "")
	case "gitlab":
		return output.NewGitLabFormatter().
			WithVerbose(p.Verbose).
			WithShowHelp(p.ExtraVerbose || p.RuleHelp != "")
	case "text":
		fallthrough
	default:
		return output.NewTextFormatter().
			WithVerbose(p.Verbose).
			WithShowHelp(p.ExtraVerbose || p.RuleHelp != "").
			WithLightMode(p.LightMode)
	}
}

// GetValidationTarget determines what to validate based on parameters.
// Returns (targetType, target1, target2, error).
// targetType can be: "commit", "range", "message", "count".
func (p ValidationParameters) GetValidationTarget() (string, string, string, error) {
	// Validate all input parameters for security
	if err := validateFilePath(p.MessageFile); err != nil {
		return "", "", "", fmt.Errorf("invalid message file: %w", err)
	}

	if err := validateGitReference(p.GitReference); err != nil {
		return "", "", "", fmt.Errorf("invalid git reference: %w", err)
	}

	if err := validateGitReference(p.BaseBranch); err != nil {
		return "", "", "", fmt.Errorf("invalid base branch: %w", err)
	}

	if err := validateCommitCount(p.CommitCount); err != nil {
		return "", "", "", fmt.Errorf("invalid commit count: %w", err)
	}

	// Apply validation source with precedence order
	if p.MessageFile != "" {
		// 1. Message from file (highest priority)
		return "message", filepath.Clean(p.MessageFile), "", nil
	} else if p.BaseBranch != "" {
		// 2. Base branch comparison
		return "range", p.BaseBranch, "HEAD", nil
	} else if p.RevisionRange != "" {
		// 3. Revision range
		// Validate revision range
		if err := validateParameterLength("Revision range", p.RevisionRange, MaxRefLength); err != nil {
			return "", "", "", err
		}

		// Parse revision range (format: from..to)
		parts := parseRevisionRange(p.RevisionRange)
		if len(parts) == 2 {
			// Validate both parts
			if err := validateGitReference(parts[0]); err != nil {
				return "", "", "", fmt.Errorf("invalid revision range start: %w", err)
			}

			if err := validateGitReference(parts[1]); err != nil {
				return "", "", "", fmt.Errorf("invalid revision range end: %w", err)
			}

			return "range", parts[0], parts[1], nil
		}

		return "", "", "", fmt.Errorf("invalid revision range format: %s (expected format: from..to)", p.RevisionRange)
	} else if p.GitReference != "" {
		// 4. Single git reference
		return "commit", p.GitReference, "", nil
	} else if p.CommitCount > 1 {
		// 5. Commit count (only if explicitly set to > 1)
		return "count", strconv.Itoa(p.CommitCount), "", nil
	}

	// Default to HEAD (when commit count is 1 or no options provided)
	return "commit", "HEAD", "", nil
}

// Validation helper functions.

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

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/output"
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
	SkipMergeCommits bool
	RepoPath         string

	// Reporting options
	Verbose      bool
	ExtraVerbose bool
	LightMode    bool
	RuleHelp     string
	Format       string

	// Dependencies
	Dependencies *AppDependencies
}

// NewValidationParameters creates ValidationParameters from a cobra command.
func NewValidationParameters(cmd *cobra.Command) ValidationParameters {
	// Extract all parameters from the command
	messageFile, _ := cmd.Flags().GetString("message-file")
	gitReference, _ := cmd.Flags().GetString("git-reference")
	commitCount, _ := cmd.Flags().GetInt("commit-count")
	revisionRange, _ := cmd.Flags().GetString("revision-range")
	baseBranch, _ := cmd.Flags().GetString("base-branch")
	verbose, _ := cmd.Flags().GetBool("verbose")
	extraVerbose, _ := cmd.Flags().GetBool("extra-verbose")
	lightMode, _ := cmd.Flags().GetBool("light-mode")
	ruleHelp, _ := cmd.Flags().GetString("rulehelp")
	format, _ := cmd.Flags().GetString("format")
	skipMergeCommits, _ := cmd.Flags().GetBool("skip-merge-commits")
	repoPath, _ := cmd.Flags().GetString("repo-path")

	// Get dependencies from context if available
	var deps *AppDependencies

	if cmd.Context() != nil {
		if d, ok := cmd.Context().Value(dependenciesKey).(*AppDependencies); ok {
			deps = d
		}
	}

	return ValidationParameters{
		OutWriter:        cmd.OutOrStdout(),
		MessageFile:      messageFile,
		GitReference:     gitReference,
		CommitCount:      commitCount,
		RevisionRange:    revisionRange,
		BaseBranch:       baseBranch,
		Verbose:          verbose,
		ExtraVerbose:     extraVerbose,
		LightMode:        lightMode,
		RuleHelp:         ruleHelp,
		Format:           format,
		SkipMergeCommits: skipMergeCommits,
		RepoPath:         repoPath,
		Dependencies:     deps,
	}
}

// Maximum allowed values for CLI parameters.
const (
	MaxCommitCount = 1000 // Maximum reasonable number of commits to validate
	MaxPathLength  = 4096 // Maximum path length
	MaxRefLength   = 256  // Maximum Git reference length
)

// validateParameterLength checks if a parameter value exceeds the maximum safe length.
func validateParameterLength(name, value string, maxLength int) error {
	if len(value) > maxLength {
		return fmt.Errorf("%s exceeds maximum allowed length (%d)", name, maxLength)
	}

	return nil
}

// validateGitReference checks if a Git reference value is safe.
func validateGitReference(ref string) error {
	if ref == "" {
		return nil
	}

	// Check length
	if err := validateParameterLength("Git reference", ref, MaxRefLength); err != nil {
		return err
	}

	// Validate git reference format - allowing common git reference formats
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-./~^{}[\]@]+$`)
	if !validPattern.MatchString(ref) {
		return errors.New("git reference contains invalid characters")
	}

	return nil
}

// validateFilePath checks if a file path is safe.
func validateFilePath(path string) error {
	if path == "" {
		return nil
	}

	// Check length
	if err := validateParameterLength("File path", path, MaxPathLength); err != nil {
		return err
	}

	// Validate basic path safety
	// This doesn't check if the path exists, just if it's a valid path format
	path = filepath.Clean(path)
	if strings.Contains(path, "..") {
		return errors.New("path traversal detected in file path")
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

// ToValidationOptions converts ValidationParameters to validate.ValidationOptions.
func (p ValidationParameters) ToValidationOptions() (validate.ValidationOptions, error) {
	// Validate all input parameters for security
	if err := validateFilePath(p.MessageFile); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid message file: %w", err)
	}

	if err := validateGitReference(p.GitReference); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid git reference: %w", err)
	}

	if err := validateGitReference(p.BaseBranch); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid base branch: %w", err)
	}

	if err := validateCommitCount(p.CommitCount); err != nil {
		return validate.ValidationOptions{}, fmt.Errorf("invalid commit count: %w", err)
	}

	// Process validation flags with precedence
	opts := validate.ValidationOptions{
		SkipMergeCommits: p.SkipMergeCommits,
	}

	// Default to validating HEAD if no other option is provided
	if p.GitReference == "" && p.RevisionRange == "" && p.BaseBranch == "" && p.MessageFile == "" && p.CommitCount <= 0 {
		// Default behavior uses HEAD
		opts.CommitHash = "HEAD"
	}

	// Apply validation source with precedence order
	if p.MessageFile != "" {
		// 1. Message from file (highest priority)
		opts.MessageFile = filepath.Clean(p.MessageFile)
	} else if p.BaseBranch != "" {
		// 2. Base branch comparison
		opts.FromHash = p.BaseBranch
		opts.ToHash = "HEAD"
	} else if p.RevisionRange != "" {
		// 3. Revision range
		// Validate revision range
		if err := validateParameterLength("Revision range", p.RevisionRange, MaxRefLength); err != nil {
			return validate.ValidationOptions{}, err
		}

		// Parse revision range (format: from..to)
		parts := parseRevisionRange(p.RevisionRange)
		if len(parts) == 2 {
			// Validate both parts
			if err := validateGitReference(parts[0]); err != nil {
				return validate.ValidationOptions{}, fmt.Errorf("invalid revision range start: %w", err)
			}

			if err := validateGitReference(parts[1]); err != nil {
				return validate.ValidationOptions{}, fmt.Errorf("invalid revision range end: %w", err)
			}

			opts.FromHash = parts[0]
			opts.ToHash = parts[1]
		} else {
			return validate.ValidationOptions{}, fmt.Errorf("invalid revision range format: %s (expected format: from..to)", p.RevisionRange)
		}
	} else if p.GitReference != "" {
		// 4. Single git reference
		opts.CommitHash = p.GitReference
	} else if p.CommitCount > 0 {
		// 5. Commit count
		opts.CommitCount = p.CommitCount
	}

	return opts, nil
}

// ToReportOptions converts ValidationParameters to report.Options.
func (p ValidationParameters) ToReportOptions() report.Options {
	return report.Options{
		Format:         getReportFormat(p.Format),
		Verbose:        p.Verbose,
		ShowHelp:       p.ExtraVerbose || p.RuleHelp != "",
		RuleToShowHelp: p.RuleHelp,
		LightMode:      p.LightMode,
		Writer:         p.OutWriter,
	}
}

// getReportFormat converts a string format to the report.Format type.
func getReportFormat(format string) report.Format {
	switch format {
	case "json":
		return report.FormatJSON
	case "github":
		return report.FormatGitHubActions
	case "gitlab":
		return report.FormatGitLabCI
	default:
		return report.FormatText
	}
}

// CreateFormatter creates a formatter based on the parameters.
func (p ValidationParameters) CreateFormatter() domain.ResultFormatter {
	reportOptions := p.ToReportOptions()

	switch reportOptions.Format {
	case report.FormatJSON:
		return output.NewJSONFormatter()
	case report.FormatGitHubActions:
		return output.NewGitHubFormatter().
			WithVerbose(reportOptions.Verbose).
			WithShowHelp(reportOptions.ShowHelp)
	case report.FormatGitLabCI:
		return output.NewGitLabFormatter().
			WithVerbose(reportOptions.Verbose).
			WithShowHelp(reportOptions.ShowHelp)
	case report.FormatText:
		fallthrough
	default:
		return output.NewTextFormatter().
			WithVerbose(reportOptions.Verbose).
			WithShowHelp(reportOptions.ShowHelp).
			WithLightMode(reportOptions.LightMode)
	}
}

// CreateValidationService creates a validation service based on parameters.
func (p ValidationParameters) CreateValidationService(ctx context.Context) (validate.ValidationService, error) {
	if err := validateFilePath(p.RepoPath); err != nil {
		return validate.ValidationService{}, fmt.Errorf("invalid repository path: %w", err)
	}

	return constructValidationService(ctx, p.Dependencies, p.RepoPath)
}

// parseRevisionRange parses a revision range string (e.g., "main..HEAD") into a slice of parts.
func parseRevisionRange(revisionRange string) []string {
	// Split by ".." to get from and to
	parts := []string{}

	// Handle both ".." and "..." formats
	if idx := indexOf(revisionRange, "..."); idx >= 0 {
		parts = append(parts, revisionRange[:idx], revisionRange[idx+3:])
	} else if idx := indexOf(revisionRange, ".."); idx >= 0 {
		parts = append(parts, revisionRange[:idx], revisionRange[idx+2:])
	}

	return parts
}

// indexOf returns the index of the first occurrence of substring in s, or -1 if not found.
func indexOf(s, substring string) int {
	for i := 0; i <= len(s)-len(substring); i++ {
		if s[i:i+len(substring)] == substring {
			return i
		}
	}

	return -1
}

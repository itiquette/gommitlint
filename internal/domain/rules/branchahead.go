// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// BranchAheadRule validates that the current branch is not too many commits ahead
// of the reference branch.
type BranchAheadRule struct {
	maxCommitsAhead int
	reference       string
}

// NewBranchAheadRule creates a new rule for checking commits ahead of a reference branch from config.
func NewBranchAheadRule(cfg config.Config) BranchAheadRule {
	maxCommitsAhead := cfg.Repo.MaxCommitsAhead
	if maxCommitsAhead <= 0 {
		maxCommitsAhead = 10 // Simple default
	}

	reference := cfg.Repo.ReferenceBranch
	if reference == "" {
		reference = "main" // Simple default
	}

	return BranchAheadRule{
		maxCommitsAhead: maxCommitsAhead,
		reference:       reference,
	}
}

// validateConfig validates the rule configuration for potential issues.
func (r BranchAheadRule) validateConfig() []domain.ValidationError {
	var errors []domain.ValidationError

	// Validate maxCommitsAhead is positive
	if r.maxCommitsAhead < 0 {
		errors = append(errors,
			domain.New(r.Name(), domain.ErrInvalidConfig,
				fmt.Sprintf("maxCommitsAhead cannot be negative: %d", r.maxCommitsAhead)).
				WithContextMap(map[string]string{
					"actual":   strconv.Itoa(r.maxCommitsAhead),
					"expected": "positive integer",
				}).
				WithHelp("Set maxCommitsAhead to a positive number or 0 to disable the check"))
	}

	// Validate reference branch is not empty
	if r.reference == "" {
		errors = append(errors,
			domain.New(r.Name(), domain.ErrMissingReference,
				"Reference branch cannot be empty").
				WithContextMap(map[string]string{
					"actual":   "empty",
					"expected": "non-empty branch name",
				}).
				WithHelp("Set a valid reference branch name (e.g., 'main', 'master', 'develop')"))
	}

	return errors
}

// Validate checks that the current branch is not too many commits ahead of the reference branch.
// This rule requires repository access, so it checks if repository is available.
func (r BranchAheadRule) Validate(_ domain.Commit, repo domain.Repository, cfg config.Config) []domain.ValidationError {
	// First validate configuration
	if configErrors := r.validateConfig(); len(configErrors) > 0 {
		return configErrors
	}

	// Skip if no repository is provided
	if repo == nil {
		return nil
	}

	// Skip validation if explicitly disabled in original config (set to 0)
	if cfg.Repo.MaxCommitsAhead == 0 {
		return nil
	}

	ctx := context.Background()

	// Get the number of commits ahead with enhanced error handling
	commitsAhead, err := repo.GetCommitsAheadCount(ctx, r.reference)
	if err != nil {
		// Handle reference not found as 0 commits ahead (original behavior)
		if isReferenceNotFoundError(err.Error()) {
			// Reference doesn't exist - treat as 0 commits ahead (no validation error)
			// This is the original behavior: new repos, first commits, etc. pass validation
			return nil
		}

		// Other errors are actual problems
		return r.handleRepositoryError(err)
	}

	// Validate against the maximum
	if commitsAhead > r.maxCommitsAhead {
		return []domain.ValidationError{
			r.createValidationError(commitsAhead),
		}
	}

	return nil
}

// handleRepositoryError provides enhanced error handling for different types of repository errors.
// Note: Reference not found errors are handled separately in Validate() as non-errors.
func (r BranchAheadRule) handleRepositoryError(err error) []domain.ValidationError {
	errMsg := err.Error()

	// Different error handling based on error type
	switch {
	case isRepositoryAccessError(errMsg):
		// Repository access issues - return error to inform user
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrInvalidRepo,
				"Cannot access repository for branch comparison").
				WithContextMap(map[string]string{
					"actual":   "access denied",
					"expected": "accessible repository",
				}).
				WithHelp("Ensure you're in a valid Git repository with proper access permissions"),
		}
	default:
		// Other git operation errors - provide general error
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrGitOperationFailed,
				fmt.Sprintf("Failed to check commits ahead of '%s'", r.reference)).
				WithContextMap(map[string]string{
					"actual":   "operation failed",
					"expected": "successful git operation",
				}).
				WithHelp("Check your Git repository status and network connectivity"),
		}
	}
}

// createValidationError creates an enhanced validation error with actionable guidance.
func (r BranchAheadRule) createValidationError(commitsAhead int) domain.ValidationError {
	excess := commitsAhead - r.maxCommitsAhead

	// Enhanced error message with clear context
	message := fmt.Sprintf("Current branch is %d commits ahead of '%s' (maximum allowed: %d)",
		commitsAhead, r.reference, r.maxCommitsAhead)

	// Enhanced help text with specific guidance based on number of excess commits
	var helpText string

	switch {
	case excess <= 3:
		helpText = fmt.Sprintf("Consider squashing the %d excess commits into fewer logical commits, or rebase onto the latest '%s'", excess, r.reference)
	case excess <= 10:
		helpText = fmt.Sprintf("You have %d excess commits. Consider rebasing onto '%s' and squashing related commits together", excess, r.reference)
	default:
		helpText = fmt.Sprintf("You have %d excess commits. Consider breaking this into smaller pull requests or rebasing onto '%s'", excess, r.reference)
	}

	return domain.New(r.Name(), domain.ErrTooManyCommits, message).
		WithContextMap(map[string]string{
			"actual":   strconv.Itoa(commitsAhead),
			"expected": "max " + strconv.Itoa(r.maxCommitsAhead),
		}).
		WithHelp(helpText)
}

// isReferenceNotFoundError checks if the error indicates the reference branch was not found.
func isReferenceNotFoundError(errMsg string) bool {
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "unknown revision") ||
		strings.Contains(errMsg, "bad revision")
}

// isRepositoryAccessError checks if the error indicates repository access issues.
func isRepositoryAccessError(errMsg string) bool {
	return strings.Contains(errMsg, "not a git repository") ||
		strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "access denied") ||
		strings.Contains(errMsg, "authentication failed")
}

// Name returns the rule name.
func (r BranchAheadRule) Name() string {
	return "BranchAhead"
}

// ensureFullReference converts short reference names to full Git reference paths.
// This ensures unambiguous reference resolution in Git commands, supporting both
// packed and loose reference storage formats.
func ensureFullReference(ref string) string {
	// Already a full reference path
	if strings.HasPrefix(ref, "refs/") {
		return ref
	}

	// Handle remote branch references (e.g., "origin/main" -> "refs/remotes/origin/main")
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		remote := parts[0]
		branch := parts[1]

		return fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	}

	// Convert local branch name to full reference (e.g., "main" -> "refs/heads/main")
	return "refs/heads/" + ref
}

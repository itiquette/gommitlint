// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// CommitsAheadRule enforces a maximum number of commits ahead of a reference.
// This rule helps teams maintain clean git histories by preventing branches
// from diverging too far from a baseline reference branch, reducing the
// complexity of eventual merges.
type CommitsAheadRule struct {
	*BaseRule
	maxCommitsAhead  int
	ref              string
	repositoryGetter func() domain.CommitAnalyzer
	ahead            int // stores number of commits ahead for result reporting
}

// CommitsAheadOption is a function that configures a CommitsAheadRule.
type CommitsAheadOption func(*CommitsAheadRule)

// WithMaxCommitsAhead sets the maximum number of commits allowed ahead of the reference.
func WithMaxCommitsAhead(max int) CommitsAheadOption {
	return func(r *CommitsAheadRule) {
		r.maxCommitsAhead = max
	}
}

// WithReference sets the reference branch to compare against.
func WithReference(ref string) CommitsAheadOption {
	return func(r *CommitsAheadRule) {
		r.ref = ref
	}
}

// WithRepositoryGetter sets the function that provides the Git repository.
func WithRepositoryGetter(getter func() domain.CommitAnalyzer) CommitsAheadOption {
	return func(r *CommitsAheadRule) {
		r.repositoryGetter = getter
	}
}

// NewCommitsAheadRule creates a new CommitsAheadRule with the given options.
func NewCommitsAheadRule(options ...CommitsAheadOption) *CommitsAheadRule {
	rule := &CommitsAheadRule{
		BaseRule:        NewBaseRule("CommitsAhead"),
		ref:             "main", // Default reference branch
		maxCommitsAhead: 5,      // Default maximum commits ahead
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// The Name method is provided by BaseRule.

// Validate checks if the current HEAD exceeds the maximum allowed commits ahead of the reference branch.
func (r *CommitsAheadRule) Validate(commitInfo *domain.CommitInfo) []appErrors.ValidationError {
	ctx := context.Background()

	return r.ValidateWithContext(ctx, commitInfo)
}

// ValidateWithContext checks if the current HEAD exceeds the maximum allowed commits ahead of the reference branch.
func (r *CommitsAheadRule) ValidateWithContext(ctx context.Context, _ *domain.CommitInfo) []appErrors.ValidationError {
	// Reset errors
	r.ClearErrors()
	r.MarkAsRun()

	// Skip validation if we can't get the repository
	if r.repositoryGetter == nil {
		r.addError(
			"nil_repo",
			"Repository object is nil",
			nil,
		)

		return r.Errors()
	}

	// Get repository
	analyzer := r.repositoryGetter()

	// Skip validation if analyzer is nil
	if analyzer == nil {
		r.addError(
			"nil_repo",
			"Repository object is nil",
			nil,
		)

		return r.Errors()
	}

	// Check for empty reference
	if r.ref == "" {
		r.addError(
			"empty_ref",
			"reference branch name cannot be empty",
			nil,
		)

		return r.Errors()
	}

	// Check context cancellation
	if ctx.Err() != nil {
		r.addError(
			"context_cancelled",
			"operation was cancelled: "+ctx.Err().Error(),
			map[string]string{
				"error_details": ctx.Err().Error(),
			},
		)

		return r.Errors()
	}

	// Get count of commits ahead
	ahead, err := analyzer.GetCommitsAhead(ctx, r.ref)

	// Handle git error
	if err != nil {
		r.addError(
			"git_error",
			"failed to get commits ahead: "+err.Error(),
			map[string]string{
				"error_details": err.Error(),
				"reference":     r.ref,
			},
		)

		return r.Errors()
	}

	// Store the ahead count for use in results
	r.ahead = ahead

	// Return error if ahead count exceeds maximum
	if ahead > r.maxCommitsAhead {
		r.addError(
			"too_many_commits",
			fmt.Sprintf("HEAD is %d commit(s) ahead of %s (maximum allowed: %d)", ahead, r.ref, r.maxCommitsAhead),
			map[string]string{
				"ahead":     strconv.Itoa(ahead),
				"max":       strconv.Itoa(r.maxCommitsAhead),
				"reference": r.ref,
			},
		)
	}

	return r.Errors()
}

// addError uses the new error handling pattern.
func (r *CommitsAheadRule) addError(code, message string, context map[string]string) {
	var appCode appErrors.ValidationErrorCode

	// Map internal code to appErrors code
	switch code {
	case "nil_repo":
		appCode = appErrors.ErrInvalidRepo
	case "empty_ref":
		appCode = appErrors.ErrInvalidConfig
	case "context_cancelled":
		appCode = appErrors.ErrContextCancelled
	case "git_error":
		appCode = appErrors.ErrGitOperationFailed
	case "too_many_commits":
		appCode = appErrors.ErrTooManyCommits
	default:
		appCode = appErrors.ErrUnknown
	}

	// Add the error with context if provided
	if context != nil {
		r.AddErrorWithContext(appCode, message, context)
	} else {
		r.AddErrorWithCode(appCode, message)
	}
}

// Result returns a concise result message.
func (r *CommitsAheadRule) Result() string {
	if r.HasErrors() {
		return "Too many commits ahead of reference branch"
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", r.ahead, r.ref)
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *CommitsAheadRule) VerboseResult() string {
	if r.HasErrors() {
		// Return a more detailed error message in verbose mode
		errors := r.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		switch appErrors.ValidationErrorCode(validationErr.Code) {
		case appErrors.ErrInvalidRepo:
		default:
			return "Repository object is nil. Cannot validate commits ahead."
		case appErrors.ErrInvalidConfig:
			return "Reference branch name is empty. Cannot validate commits ahead."
		case appErrors.ErrCancelled:
			return "Operation was cancelled by context. Cannot validate commits ahead."
		case appErrors.ErrTooManyCommits:
			return fmt.Sprintf(
				"HEAD is %d commit(s) ahead of %s (maximum allowed: %d). Consider merging or rebasing with %s.",
				r.ahead, r.ref, r.maxCommitsAhead, r.ref)
		case appErrors.ErrGitOperationFailed:
			// Get the error details from the context map directly
			errorDetails := ""

			for k, v := range validationErr.Context {
				if k == "error_details" {
					errorDetails = v

					break
				}
			}

			return "Failed to get commits ahead: " + errorDetails
		}

		// Fallback
		return validationErr.Message
	}

	// Success message with details
	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s (within limit of %d)",
		r.ahead, r.ref, r.maxCommitsAhead)
}

// Help returns guidance on how to fix the rule violation.
func (r *CommitsAheadRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix"
	}

	// Check error code
	errors := r.Errors()
	if len(errors) > 0 {
		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		switch appErrors.ValidationErrorCode(validationErr.Code) {
		case appErrors.ErrTooManyCommits:
		default:
			return fmt.Sprintf(`Your branch is too far ahead of %s. To fix this, either:

1. Merge %s into your branch:
   git fetch
   git merge %s

2. Rebase your branch onto the latest %s:
   git fetch
   git rebase %s

3. Squash some commits to reduce the total count:
   git rebase -i HEAD~%d

The maximum allowed commits ahead is %d, but your branch is %d commits ahead.`,
				r.ref, r.ref, r.ref, r.ref, r.ref, r.ahead, r.maxCommitsAhead, r.ahead)
		case appErrors.ErrInvalidRepo:
			return "Configure the repository correctly to enable commit validation."
		case appErrors.ErrInvalidConfig:
			return "Specify a valid reference branch name in the configuration."
		case appErrors.ErrGitOperationFailed:
			return "Ensure your repository is valid and accessible, then try again."
		}
	}

	// Default help
	return fmt.Sprintf(`Ensure your branch is not more than %d commits ahead of %s by regularly merging or rebasing.`, r.maxCommitsAhead, r.ref)
}

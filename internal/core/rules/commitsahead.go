// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"strconv"

	"github.com/itiquette/gommitlint/internal/domain"
)

// CommitsAheadRule enforces a maximum number of commits ahead of a reference.
// This rule helps teams maintain clean git histories by preventing branches
// from diverging too far from a baseline reference branch, reducing the
// complexity of eventual merges.
type CommitsAheadRule struct {
	ref              string
	ahead            int
	maxCommitsAhead  int
	errors           []*domain.ValidationError
	repositoryGetter func() domain.CommitAnalyzer
}

// CommitsAheadOption configures a CommitsAheadRule.
type CommitsAheadOption func(*CommitsAheadRule)

// WithMaxCommitsAhead sets the maximum allowed commits ahead.
func WithMaxCommitsAhead(maxCommitsAhead int) CommitsAheadOption {
	return func(r *CommitsAheadRule) {
		if maxCommitsAhead >= 0 {
			r.maxCommitsAhead = maxCommitsAhead
		}
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
		ref:             "main", // Default reference branch
		maxCommitsAhead: 5,      // Default maximum commits ahead
		errors:          []*domain.ValidationError{},
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// Name returns the rule identifier.
func (r *CommitsAheadRule) Name() string {
	return "CommitsAhead"
}

// Validate checks if the current HEAD exceeds the maximum allowed commits ahead of the reference branch.
func (r *CommitsAheadRule) Validate(_ *domain.CommitInfo) []*domain.ValidationError {
	// Reset errors
	r.errors = []*domain.ValidationError{}

	// Skip validation if we can't get the repository
	if r.repositoryGetter == nil {
		r.addError(
			"nil_repo",
			"repository retrieval function is not set",
			nil,
		)

		return r.errors
	}

	// Get repository
	analyzer := r.repositoryGetter()

	// Skip validation if analyzer is nil
	if analyzer == nil {
		r.addError(
			"nil_repo",
			"repository cannot be nil",
			nil,
		)

		return r.errors
	}

	// Skip validation if reference is empty
	if r.ref == "" {
		r.addError(
			"empty_ref",
			"reference cannot be empty",
			nil,
		)

		return r.errors
	}

	// Get commits ahead count
	commitsAhead, err := analyzer.GetCommitsAhead(r.ref)
	if err != nil {
		r.addError(
			"git_error",
			"failed to count commits ahead",
			map[string]string{"error_details": err.Error()},
		)

		return r.errors
	}

	// Store the value for results reporting
	r.ahead = commitsAhead

	// Check if too many commits ahead
	if commitsAhead > r.maxCommitsAhead {
		r.addError(
			"too_many_commits",
			fmt.Sprintf("too many commits ahead of %s: %d (max: %d)",
				r.ref, commitsAhead, r.maxCommitsAhead),
			map[string]string{
				"reference": r.ref,
				"count":     strconv.Itoa(commitsAhead),
				"max":       strconv.Itoa(r.maxCommitsAhead),
			},
		)
	}

	return r.errors
}

// Result returns a concise string representation of the rule's status.
func (r *CommitsAheadRule) Result() string {
	if len(r.errors) > 0 {
		// Provide concise error
		return "Too many commits ahead of " + r.ref
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", r.ahead, r.ref)
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *CommitsAheadRule) VerboseResult() string {
	if len(r.errors) > 0 {
		// Return a more detailed error message in verbose mode
		switch r.errors[0].Code {
		case "nil_repo":
			return "Repository object is nil. Cannot validate commits ahead."
		case "empty_ref":
			return "Reference branch name is empty. Cannot validate commits ahead."
		case "too_many_commits":
			return fmt.Sprintf(
				"HEAD is %d commit(s) ahead of %s (maximum allowed: %d). Consider merging or rebasing with %s.",
				r.ahead, r.ref, r.maxCommitsAhead, r.ref)
		case "git_error":
			// Get the error details from the context map directly
			errorDetails := ""

			for k, v := range r.errors[0].Context {
				if k == "error_details" {
					errorDetails = v

					break
				}
			}

			return "Git error when counting commits ahead: " + errorDetails
		default:
			return r.errors[0].Error()
		}
	}

	// Provide more context in verbose mode
	if r.ahead == 0 {
		return fmt.Sprintf("HEAD is up-to-date with %s (0 commits ahead)", r.ref)
	}

	if r.ahead == 1 {
		return fmt.Sprintf("HEAD is 1 commit ahead of %s (within limit of %d)", r.ref, r.maxCommitsAhead)
	}

	return fmt.Sprintf("HEAD is %d commits ahead of %s (within limit of %d)", r.ahead, r.ref, r.maxCommitsAhead)
}

// Help returns a description of how to fix the rule violation.
func (r *CommitsAheadRule) Help() string {
	if len(r.errors) == 0 {
		return "No errors to fix"
	}

	// Check for specific error codes
	if len(r.errors) > 0 {
		switch r.errors[0].Code {
		case "nil_repo":
			return "You must provide a valid git repository to check commits ahead."

		case "empty_ref":
			return "You must provide a valid reference branch name to compare against."

		case "too_many_commits":
			return fmt.Sprintf(
				"Your current branch has too many commits ahead of %s.\n"+
					"Consider one of the following solutions:\n"+
					"1. Merge or rebase your branch with %s\n"+
					"2. Consider splitting your changes into smaller, more manageable pull requests\n"+
					"3. If this limit is too restrictive, you can configure a higher limit with WithMaxCommitsAhead option",
				r.ref, r.ref)

		case "git_error":
			return "There was an error accessing the git repository. Please check permissions and ensure the repository is valid."
		}
	}

	// Default help message
	return fmt.Sprintf(
		"Your current branch has too many commits ahead of %s.\n"+
			"Consider one of the following solutions:\n"+
			"1. Merge or rebase your branch with %s\n"+
			"2. Consider splitting your changes into smaller, more manageable pull requests\n"+
			"3. If this limit is too restrictive, you can configure a higher limit with WithMaxCommitsAhead option",
		r.ref, r.ref)
}

// Errors returns any violations detected by the rule.
func (r *CommitsAheadRule) Errors() []*domain.ValidationError {
	return r.errors
}

// addError adds a structured validation error.
func (r *CommitsAheadRule) addError(code, message string, context map[string]string) {
	err := domain.NewValidationError("CommitsAhead", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	r.errors = append(r.errors, err)
}

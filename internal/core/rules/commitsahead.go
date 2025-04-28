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
type CommitsAheadRule struct {
	baseRule         BaseRule
	maxCommitsAhead  int
	ref              string
	repositoryGetter func() domain.CommitAnalyzer
	ahead            int // stores number of commits ahead for result reporting
	errors           []appErrors.ValidationError
}

// CommitsAheadOption is a function that configures a CommitsAheadRule.
type CommitsAheadOption func(CommitsAheadRule) CommitsAheadRule

// WithMaxCommitsAhead sets the maximum number of commits allowed ahead of the reference.
func WithMaxCommitsAhead(maxCommits int) CommitsAheadOption {
	return func(r CommitsAheadRule) CommitsAheadRule {
		result := r
		result.maxCommitsAhead = maxCommits

		return result
	}
}

// WithReference sets the reference branch to compare against.
func WithReference(ref string) CommitsAheadOption {
	return func(r CommitsAheadRule) CommitsAheadRule {
		result := r
		result.ref = ref

		return result
	}
}

// WithRepositoryGetter sets the function that provides the Git repository.
func WithRepositoryGetter(getter func() domain.CommitAnalyzer) CommitsAheadOption {
	return func(r CommitsAheadRule) CommitsAheadRule {
		result := r
		result.repositoryGetter = getter

		return result
	}
}

// NewCommitsAheadRule creates a new CommitsAheadRule with the given options.
func NewCommitsAheadRule(options ...CommitsAheadOption) CommitsAheadRule {
	rule := CommitsAheadRule{
		baseRule:        NewBaseRule("CommitsAhead"),
		ref:             "main", // Default reference branch
		maxCommitsAhead: 5,      // Default maximum commits ahead
		errors:          make([]appErrors.ValidationError, 0),
	}

	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// NewCommitsAheadRuleWithConfig creates a new CommitsAheadRule using domain configuration interfaces.
func NewCommitsAheadRuleWithConfig(config domain.RepositoryConfigProvider, analyzer domain.CommitAnalyzer) CommitsAheadRule {
	rule := CommitsAheadRule{
		baseRule:        NewBaseRule("CommitsAhead"),
		ref:             "main", // Default reference branch
		maxCommitsAhead: 5,      // Default maximum commits ahead
		errors:          make([]appErrors.ValidationError, 0),
	}

	// Apply configuration if provided
	if config != nil {
		// Set reference branch if configured
		if ref := config.ReferenceBranch(); ref != "" {
			rule.ref = ref
		}

		// Set max commits ahead if configured
		if maxCommits := config.MaxCommitsAhead(); maxCommits > 0 {
			rule.maxCommitsAhead = maxCommits
		}
	}

	// Set repository getter if analyzer is provided
	if analyzer != nil {
		rule.repositoryGetter = func() domain.CommitAnalyzer {
			return analyzer
		}
	}

	return rule
}

// validateCommitsAheadWithState validates commits and returns errors along with updated rule state.
func validateCommitsAheadWithState(ctx context.Context, rule CommitsAheadRule, _ domain.CommitInfo) ([]appErrors.ValidationError, CommitsAheadRule) {
	errors := make([]appErrors.ValidationError, 0)
	updatedRule := rule

	// Skip validation if we can't get the repository
	if rule.repositoryGetter == nil {
		err := createCommitsAheadError(
			rule.Name(),
			appErrors.ErrInvalidRepo,
			"Repository object is nil",
			nil,
		)
		errors = append(errors, err)
		updatedRule.errors = errors

		return errors, updatedRule
	}

	analyzer := rule.repositoryGetter()
	if analyzer == nil {
		err := createCommitsAheadError(
			rule.Name(),
			appErrors.ErrInvalidRepo,
			"Repository object is nil - Git repository not accessible",
			nil,
		)
		errors = append(errors, err)
		updatedRule.errors = errors

		return errors, updatedRule
	}

	// Get commits ahead
	ahead, err := analyzer.GetCommitsAhead(ctx, rule.ref)
	if err != nil {
		// Create validation error for repository error
		validationErr := createCommitsAheadError(
			rule.Name(),
			"repo_access_error",
			"Error accessing repository: "+err.Error(),
			map[string]string{
				"error": err.Error(),
			},
		)
		errors = append(errors, validationErr)
		updatedRule.errors = errors

		return errors, updatedRule
	}

	// Check if we exceed the maximum
	if rule.maxCommitsAhead > 0 && ahead > rule.maxCommitsAhead {
		// Create validation error for too many commits ahead
		err := createCommitsAheadError(
			rule.Name(),
			appErrors.ErrTooManyCommits,
			fmt.Sprintf("HEAD is %d commits ahead of %s (maximum allowed: %d)",
				ahead, rule.ref, rule.maxCommitsAhead),
			map[string]string{
				"commits_ahead":     strconv.Itoa(ahead),
				"max_allowed":       strconv.Itoa(rule.maxCommitsAhead),
				"reference":         rule.ref,
				"exceeds_by":        strconv.Itoa(ahead - rule.maxCommitsAhead),
				"is_feature_branch": "false", // Simplified as we don't have branch information
			},
		)
		errors = append(errors, err)
		updatedRule.errors = errors

		return errors, updatedRule
	}

	// Store commits ahead in the rule state for the result message
	updatedRule.ahead = ahead

	return errors, updatedRule
}

// createCommitsAheadError creates a validation error for CommitsAheadRule.
func createCommitsAheadError(ruleName string, code appErrors.ValidationErrorCode, message string, ctx map[string]string) appErrors.ValidationError {
	err := appErrors.New(
		ruleName,
		code,
		message,
	).WithContext("validation_rule", ruleName)

	for k, v := range ctx {
		err = err.WithContext(k, v)
	}

	return err
}

// ValidateWithContext checks if the current HEAD exceeds the maximum allowed commits ahead of the reference branch.
func (r CommitsAheadRule) ValidateWithContext(ctx context.Context, commitInfo domain.CommitInfo) []appErrors.ValidationError {
	errors, _ := validateCommitsAheadWithState(ctx, r, commitInfo)

	return errors
}

// ValidateCommitsAheadWithState validates a commit and returns errors along with an updated rule state.
// Exported for testing purposes.
func ValidateCommitsAheadWithState(ctx context.Context, rule CommitsAheadRule, commitInfo domain.CommitInfo) ([]appErrors.ValidationError, CommitsAheadRule) {
	return validateCommitsAheadWithState(ctx, rule, commitInfo)
}

// Validate checks if the current HEAD exceeds the maximum allowed commits ahead of the reference branch.
func (r CommitsAheadRule) Validate(commitInfo domain.CommitInfo) []appErrors.ValidationError {
	ctx := context.Background()

	return r.ValidateWithContext(ctx, commitInfo)
}

// SetErrors sets the validation errors and returns a new rule.
func (r CommitsAheadRule) SetErrors(errors []appErrors.ValidationError) CommitsAheadRule {
	result := r
	result.errors = errors

	// Also update baseRule for consistency
	baseRule := r.baseRule
	for _, err := range errors {
		baseRule = baseRule.WithError(err)
	}

	result.baseRule = baseRule

	return result
}

// Errors returns the validation errors for the rule.
func (r CommitsAheadRule) Errors() []appErrors.ValidationError {
	return r.errors
}

// HasErrors returns true if there are any validation errors.
func (r CommitsAheadRule) HasErrors() bool {
	return len(r.errors) > 0
}

// Result returns a concise result message.
func (r CommitsAheadRule) Result() string {
	if r.HasErrors() {
		errors := r.Errors()
		if len(errors) > 0 {
			validationErr := errors[0]

			switch appErrors.ValidationErrorCode(validationErr.Code) { //nolint:exhaustive
			case appErrors.ErrInvalidRepo:
				return "Git repository not accessible"
			case appErrors.ErrTooManyCommits:
				// Extract ahead count from context if possible
				if aheadStr, exists := validationErr.Context["commits_ahead"]; exists {
					if ahead, err := strconv.Atoi(aheadStr); err == nil {
						return fmt.Sprintf("Too many commits ahead of %s (%d)", r.ref, ahead)
					}
				}

				// Fallback to stored ahead count in rule's state
				if r.ahead > 0 {
					return fmt.Sprintf("Too many commits ahead of %s (%d)", r.ref, r.ahead)
				}

				// Generic fallback message
				return "Too many commits ahead of reference branch"
			case appErrors.ErrInvalidConfig:
				return "Invalid branch configuration"
			default:
				return "Validation error occurred"
			}
		}

		return "Too many commits ahead of reference branch"
	}

	// Only for success case
	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", r.ahead, r.ref)
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r CommitsAheadRule) VerboseResult() string {
	if r.HasErrors() {
		// Return a more detailed error message in verbose mode
		errors := r.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		validationErr := errors[0]

		// Extract ahead count from context if possible
		ahead := r.ahead

		if aheadStr, exists := validationErr.Context["ahead"]; exists {
			if parsedAhead, err := strconv.Atoi(aheadStr); err == nil {
				ahead = parsedAhead
			}
		}

		switch appErrors.ValidationErrorCode(validationErr.Code) { //nolint:exhaustive
		case appErrors.ErrInvalidRepo:
			return "Repository object is nil. Cannot validate commits ahead."
		case appErrors.ErrInvalidConfig:
			return "Reference branch name is empty. Cannot validate commits ahead."
		case appErrors.ErrContextCancelled:
			return "Operation was cancelled by context. Cannot validate commits ahead."
		case appErrors.ErrTooManyCommits:
			return fmt.Sprintf(
				"HEAD is %d commit(s) ahead of %s (maximum allowed: %d). Consider merging or rebasing with %s.",
				ahead, r.ref, r.maxCommitsAhead, r.ref)
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
		default:
			return validationErr.Message
		}
	}

	// Success message with details
	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s (within limit of %d)",
		r.ahead, r.ref, r.maxCommitsAhead)
}

// Help returns guidance on how to fix the rule violation.
func (r CommitsAheadRule) Help() string {
	// First check if the rule has errors - this should be the primary check
	if !r.HasErrors() {
		return "No errors to fix"
	}

	// Now check error code for specific guidance
	errors := r.Errors()
	if len(errors) > 0 {
		validationErr := errors[0]

		// Extract ahead count from context if possible
		aheadValue := r.ahead

		if aheadStr, exists := validationErr.Context["commits_ahead"]; exists {
			if parsedAhead, err := strconv.Atoi(aheadStr); err == nil {
				aheadValue = parsedAhead
			}
		}

		switch appErrors.ValidationErrorCode(validationErr.Code) { //nolint:exhaustive
		case appErrors.ErrTooManyCommits:
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
				r.ref, r.ref, r.ref, r.ref, r.ref, aheadValue, r.maxCommitsAhead, aheadValue)
		case appErrors.ErrInvalidRepo:
			return "The Git repository is not accessible. Ensure you are in a valid Git repository and have appropriate permissions."
		case appErrors.ErrInvalidConfig:
			return "Specify a valid reference branch name in the configuration."
		case appErrors.ErrGitOperationFailed:
			return "Ensure your repository is valid and accessible, then try again."
		}

		// Default help for any other error cases
		return fmt.Sprintf(`Ensure your branch is not more than %d commits ahead of %s by regularly merging or rebasing.
For better git hygiene, consider using smaller, more focused commits.`, r.maxCommitsAhead, r.ref)
	}

	// If we somehow have HasErrors() true but no specific errors, provide a default message
	return fmt.Sprintf("Ensure your branch is not more than %d commits ahead of %s", r.maxCommitsAhead, r.ref)
}

// Name returns the rule name.
func (r CommitsAheadRule) Name() string {
	return r.baseRule.Name()
}

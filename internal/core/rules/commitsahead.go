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

// Helper function to extract ahead count from errors.
func extractAheadCount(errors []appErrors.ValidationError) int {
	for _, err := range errors {
		if countStr, exists := err.Context["ahead_count"]; exists {
			if count, parseErr := strconv.Atoi(countStr); parseErr == nil {
				return count
			}
		}
		// Fallback to other keys
		for _, key := range []string{"commitsahead_count", "actual_ahead_count", "commits_ahead"} {
			if countStr, exists := err.Context[key]; exists {
				if count, parseErr := strconv.Atoi(countStr); parseErr == nil {
					return count
				}
			}
		}
	}

	return 0 // Default if not found
}

// ValidateCommitsAheadWithState validates commits and returns errors along with updated rule state.
func ValidateCommitsAheadWithState(ctx context.Context, rule CommitsAheadRule, _ domain.CommitInfo) ([]appErrors.ValidationError, CommitsAheadRule) {
	errors := make([]appErrors.ValidationError, 0)
	result := rule // Create a copy to return

	// Skip validation if we can't get the repository
	if rule.repositoryGetter == nil {
		// Create error context with rich information
		ctx := appErrors.NewContext()
		helpMessage := `Repository access error: Unable to access the Git repository.
The CommitsAhead rule requires access to the Git repository to check the number of commits
ahead of the reference branch. However, the repository object is nil, which means the
system cannot access Git information.
This could be caused by:
- The current directory is not a Git repository
- The Git repository is corrupt or missing
- Insufficient permissions to access the repository
- The Git binary is not installed or not in PATH
- Internal configuration issues in the application
To fix this issue:
1. Ensure you are running the command from within a valid Git repository
2. Verify that Git is properly installed and accessible
3. Check repository permissions and integrity with 'git status'
4. If running in CI/CD, ensure Git is available in the build environment`
		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrInvalidRepo,
			"Repository object is nil",
			helpMessage,
			ctx,
		)
		errors = append(errors, err)
		result.errors = errors

		return errors, result
	}

	analyzer := rule.repositoryGetter()
	if analyzer == nil {
		// Create error context with rich information
		ctx := appErrors.NewContext()
		helpMessage := `Repository access error: The Git repository analyzer is not available.
The CommitsAhead rule successfully obtained a repository getter function, but the function
returned nil when called. This indicates a problem accessing or initializing the Git repository analyzer.
This could be caused by:
- The repository was accessible during initialization but became unavailable
- Permission issues with the Git repository
- A corrupted Git repository state
- Internal application errors in the Git integration
To fix this issue:
1. Verify that the repository is still accessible with 'git status'
2. Check if Git hooks or configurations have been modified unexpectedly
3. Ensure that no other processes are locking the Git repository
4. Try running 'git gc' to clean up the repository if it's corrupted`
		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrInvalidRepo,
			"Repository object is nil - Git repository not accessible",
			helpMessage,
			ctx,
		)
		errors = append(errors, err)
		result.errors = errors

		return errors, result
	}

	// Get commits ahead
	ahead, err := analyzer.GetCommitsAhead(ctx, rule.ref)
	if err != nil {
		// Create error context with rich information
		errorCtx := appErrors.NewContext()
		helpMessage := `Repository Operation Error: Failed to calculate commits ahead.
The CommitsAhead rule encountered an error while trying to determine how many commits
your current branch is ahead of the reference branch '${rule.ref}'.
Error details: ${err.Error()}
This could be caused by:
- The reference branch '${rule.ref}' doesn't exist
- Network issues when trying to access remote repository
- Git configuration problems
- Git repository corruption
- Permission issues with the repository
To fix this issue:
1. Verify that the reference branch exists:
   git branch -a | grep ${rule.ref}
2. Make sure your local repository is up-to-date:
   git fetch
3. Check for any Git configuration issues:
   git config --list
4. If the reference branch is remote, ensure you have access to it:
   git ls-remote origin ${rule.ref}
5. Try running Git garbage collection if repository seems corrupted:
   git gc`
		validationErr := appErrors.CreateRichError(
			rule.Name(),
			"repo_access_error",
			"Error accessing repository: "+err.Error(),
			helpMessage,
			errorCtx,
		)
		// Add additional context
		validationErr = validationErr.WithContext("error", err.Error())
		validationErr = validationErr.WithContext("reference", rule.ref)
		errors = append(errors, validationErr)
		result.errors = errors

		return errors, result
	}

	// Create a status info error with ahead count
	statusErr := appErrors.CreateBasicError(
		rule.Name(),
		"status_info",
		fmt.Sprintf("HEAD is %d commit(s) ahead of %s", ahead, rule.ref),
	).WithContext("ahead_count", strconv.Itoa(ahead))
	errors = append(errors, statusErr)

	// Check if we exceed the maximum
	if rule.maxCommitsAhead > 0 && ahead > rule.maxCommitsAhead {
		exceedCount := ahead - rule.maxCommitsAhead
		helpMessage := fmt.Sprintf(`Too Many Commits Error: Your branch has too many commits ahead of '%s'.
Your branch is currently %d commits ahead of the reference branch '%s', which exceeds
the maximum allowed limit of %d commits. This means your branch has diverged significantly
from the reference branch and should be synchronized.
✅ RECOMMENDED ACTIONS:
1. Merge the reference branch into your branch to stay up-to-date:
   git fetch
   git merge %s
2. Rebase your branch onto the latest reference branch version:
   git fetch
   git rebase %s
3. Squash some of your commits to reduce the total count:
   git rebase -i HEAD~%d
   (in the editor, change some "pick" to "squash" or "fixup")
4. Consider creating a pull request/merge request if your feature is complete
WHY THIS MATTERS:
- Keeping branches in sync reduces future merge conflicts
- Smaller, more focused branches are easier to review and merge
- Long-lived branches can accumulate technical debt
- Following branch hygiene helps maintain project velocity
Current commits ahead:  %d
Maximum allowed:        %d
Exceeds maximum by:     %d commits`,
			rule.ref, ahead, rule.ref, rule.maxCommitsAhead,
			rule.ref, rule.ref, ahead, ahead, rule.maxCommitsAhead, exceedCount)
		// Create the rich error
		err := appErrors.CreateBasicError(
			rule.Name(),
			appErrors.ErrTooManyCommits,
			fmt.Sprintf("HEAD is %d commits ahead of %s (maximum allowed: %d)",
				ahead, rule.ref, rule.maxCommitsAhead),
		)
		// Add additional context using WithContext method from ValidationError
		err = err.WithContext("ahead_count", strconv.Itoa(ahead))
		err = err.WithContext("commits_ahead", strconv.Itoa(ahead))
		err = err.WithContext("max_allowed", strconv.Itoa(rule.maxCommitsAhead))
		err = err.WithContext("reference", rule.ref)
		err = err.WithContext("exceeds_by", strconv.Itoa(exceedCount))
		err = err.WithContext("is_feature_branch", "false") // Simplified as we don't have branch information
		err = err.WithContext("help", helpMessage)
		errors = append(errors, err)
	}

	result.errors = errors

	return errors, result
}

func (r CommitsAheadRule) Validate(ctx context.Context, commitInfo domain.CommitInfo) []appErrors.ValidationError {
	errors, _ := ValidateCommitsAheadWithState(ctx, r, commitInfo)

	return errors
}

// SetErrors sets the validation errors and returns a new rule.
func (r CommitsAheadRule) SetErrors(errors []appErrors.ValidationError) CommitsAheadRule {
	result := r
	result.errors = errors

	// Also update baseRule for consistency
	baseRule := r.baseRule

	for _, err := range errors {
		// Skip status info "errors" that aren't real validation errors
		if code := appErrors.ValidationErrorCode(err.Code); code != "status_info" {
			baseRule = baseRule.WithError(err)
		}
	}

	result.baseRule = baseRule

	return result
}

// Errors returns the validation errors for the rule.
func (r CommitsAheadRule) Errors() []appErrors.ValidationError {
	// Filter out status info "errors" that aren't real validation errors
	realErrors := make([]appErrors.ValidationError, 0)

	for _, err := range r.errors {
		if code := appErrors.ValidationErrorCode(err.Code); code != "status_info" {
			realErrors = append(realErrors, err)
		}
	}

	return realErrors
}

// HasErrors returns true if there are any validation errors.
func (r CommitsAheadRule) HasErrors() bool {
	// Check for real errors, not status info
	for _, err := range r.errors {
		if code := appErrors.ValidationErrorCode(err.Code); code != "status_info" {
			return true
		}
	}

	return false
}

// Result returns a concise result message.
func (r CommitsAheadRule) Result(errors []appErrors.ValidationError) string {
	// Extract ahead count from errors
	aheadCount := extractAheadCount(errors)

	// Check for validation errors
	for _, err := range errors {
		// Skip status info errors
		if code := appErrors.ValidationErrorCode(err.Code); code == "status_info" {
			continue
		}

		// Format appropriate message based on error code
		if code := appErrors.ValidationErrorCode(err.Code); code == appErrors.ErrTooManyCommits {
			return fmt.Sprintf("HEAD is %d commits ahead of %s (exceeds limit of %d)",
				aheadCount, r.ref, r.maxCommitsAhead)
		} else if code == appErrors.ErrInvalidRepo || code == appErrors.ErrGitOperationFailed {
			return "Repository access error - " + err.Message
		}
	}

	// Use the extracted ahead count for result reporting
	if aheadCount > r.maxCommitsAhead {
		return fmt.Sprintf("HEAD is %d commits ahead of %s (exceeds limit of %d)",
			aheadCount, r.ref, r.maxCommitsAhead)
	} else if aheadCount == 0 {
		return "HEAD is at same commit as " + r.ref
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s (within limit)", aheadCount, r.ref)
}

// AheadCount returns the number of commits ahead of the reference branch.
func (r CommitsAheadRule) AheadCount(errors []appErrors.ValidationError) int {
	return extractAheadCount(errors)
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r CommitsAheadRule) VerboseResult(errors []appErrors.ValidationError) string {
	// Extract ahead count from errors
	aheadCount := extractAheadCount(errors)

	// Check for validation errors
	for _, err := range errors {
		// Skip status info errors
		if code := appErrors.ValidationErrorCode(err.Code); code == "status_info" {
			continue
		}

		// Format appropriate message based on error code
		if code := appErrors.ValidationErrorCode(err.Code); code == appErrors.ErrTooManyCommits {
			return fmt.Sprintf(
				"HEAD is currently %d commits ahead of %s (maximum allowed: %d). Consider merging or rebasing with %s.",
				aheadCount, r.ref, r.maxCommitsAhead, r.ref)
		} else if code == appErrors.ErrInvalidRepo || code == appErrors.ErrGitOperationFailed {
			return "Repository access error - " + err.Message
		}
	}

	// Use the extracted ahead count for result reporting
	if aheadCount > r.maxCommitsAhead {
		return fmt.Sprintf(
			"HEAD is currently %d commits ahead of %s (maximum allowed: %d). Consider merging or rebasing with %s.",
			aheadCount, r.ref, r.maxCommitsAhead, r.ref)
	} else if aheadCount == 0 {
		return fmt.Sprintf("HEAD is currently at same commit as %s (limit is %d)",
			r.ref, r.maxCommitsAhead)
	}

	return fmt.Sprintf("HEAD is currently %d commit(s) ahead of %s (within limit of %d)",
		aheadCount, r.ref, r.maxCommitsAhead)
}

// Help returns guidance on how to fix the rule violation.
func (r CommitsAheadRule) Help(errors []appErrors.ValidationError) string {
	// Filter out status info errors
	realErrors := make([]appErrors.ValidationError, 0)

	for _, err := range errors {
		if code := appErrors.ValidationErrorCode(err.Code); code != "status_info" {
			realErrors = append(realErrors, err)
		}
	}

	// Return early if no real errors, including if errors slice is empty
	if len(realErrors) == 0 {
		// Check if the first error has a help message
		for _, err := range errors {
			if help, ok := err.Context["help"]; ok && help != "" {
				return help
			}
		}

		// For test cases that expect "No errors to fix"
		if len(errors) == 0 {
			return "No errors to fix"
		}

		// For status errors, still provide a useful message
		return fmt.Sprintf("Your branch is in sync with %s. No action required.", r.ref)
	}

	// Extract ahead count from errors
	aheadCount := extractAheadCount(errors)

	// Find the first real error
	validationErr := realErrors[0]
	code := appErrors.ValidationErrorCode(validationErr.Code)

	// Check if there's a help context in the error
	if help, ok := validationErr.Context["help"]; ok && help != "" {
		return help
	}

	if code == appErrors.ErrTooManyCommits {
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
			r.ref, r.ref, r.ref, r.ref, r.ref, aheadCount, r.maxCommitsAhead, aheadCount)
	}

	if code == appErrors.ErrInvalidRepo {
		return "The Git repository is not accessible. Ensure you are in a valid Git repository and have appropriate permissions."
	}

	if code == appErrors.ErrInvalidConfig {
		return "Specify a valid reference branch name in the configuration."
	}

	if code == appErrors.ErrGitOperationFailed {
		return "Ensure your repository is valid and accessible, then try again."
	}

	// Default help for any other error cases
	return fmt.Sprintf(`Ensure your branch is not more than %d commits ahead of %s by regularly merging or rebasing.
For better git hygiene, consider using smaller, more focused commits.`, r.maxCommitsAhead, r.ref)
}

// Name returns the rule name.
func (r CommitsAheadRule) Name() string {
	return r.baseRule.Name()
}

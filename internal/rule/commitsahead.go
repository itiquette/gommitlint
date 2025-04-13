// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/model"
)

// CommitsAheadConfig provides configuration for the commitsAhead.
type CommitsAheadConfig struct {
	// MaxCommitsAhead defines the maximum allowed commits ahead of reference
	MaxCommitsAhead int
}

// DefaultCommitsAheadConfig returns the default configuration.
func DefaultCommitsAheadConfig() CommitsAheadConfig {
	return CommitsAheadConfig{
		MaxCommitsAhead: 20,
	}
}

// CommitsAhead enforces a maximum number of commits ahead of a reference.
// This rule helps teams maintain clean git histories by preventing branches
// from diverging too far from a baseline reference branch, reducing the
// complexity of eventual merges.
//
// The rule validates that the current HEAD is not too far ahead of a specified
// reference branch (typically main/master). When the number of commits exceeds
// the configured maximum, the rule fails and recommends actions to resolve the issue.
//
// Example usage:
//
//	rule := ValidateNumberOfCommits(repo, "main", WithMaxCommitsAhead(15))
//	if len(rule.Errors()) > 0 {
//	    fmt.Println(rule.Help())
//	}
type CommitsAhead struct {
	ref    string
	Ahead  int
	config CommitsAheadConfig
	errors []*model.ValidationError
}

// Name returns the rule identifier.
func (c *CommitsAhead) Name() string {
	return "CommitsAhead"
}

// Result returns a concise string representation of the rule's status.
func (c *CommitsAhead) Result() string {
	if len(c.errors) > 0 {
		// Provide concise error
		return "Too many commits ahead of " + c.ref
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", c.Ahead, c.ref)
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (c *CommitsAhead) VerboseResult() string {
	if len(c.errors) > 0 {
		// Return a more detailed error message in verbose mode
		switch c.errors[0].Code {
		case "nil_repo":
			return "Repository object is nil. Cannot validate commits ahead."
		case "empty_ref":
			return "Reference branch name is empty. Cannot validate commits ahead."
		case "too_many_commits":
			return fmt.Sprintf(
				"HEAD is %d commit(s) ahead of %s (maximum allowed: %d). Consider merging or rebasing with %s.",
				c.Ahead, c.ref, c.config.MaxCommitsAhead, c.ref)
		case "git_error":
			// Get the error details from the context map directly
			errorDetails := ""

			for k, v := range c.errors[0].Context {
				if k == "error_details" {
					errorDetails = v

					break
				}
			}

			return "Git error when counting commits ahead: " + errorDetails
		default:
			return c.errors[0].Error()
		}
	}

	// Provide more context in verbose mode
	if c.Ahead == 0 {
		return fmt.Sprintf("HEAD is up-to-date with %s (0 commits ahead)", c.ref)
	}

	if c.Ahead == 1 {
		return fmt.Sprintf("HEAD is 1 commit ahead of %s (within limit of %d)", c.ref, c.config.MaxCommitsAhead)
	}

	return fmt.Sprintf("HEAD is %d commits ahead of %s (within limit of %d)", c.Ahead, c.ref, c.config.MaxCommitsAhead)
}

// Errors returns any violations detected by the rule.
func (c *CommitsAhead) Errors() []*model.ValidationError {
	return c.errors
}

// Help returns a description of how to fix the rule violation.
func (c *CommitsAhead) Help() string {
	if len(c.errors) == 0 {
		return "No errors to fix"
	}

	// Check for specific error codes
	if len(c.errors) > 0 {
		switch c.errors[0].Code {
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
				c.ref, c.ref)

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
		c.ref, c.ref)
}

// Option configures a CommitsAheadConfig.
type Option func(*CommitsAheadConfig)

// WithMaxCommitsAhead sets the maximum allowed commits ahead.
func WithMaxCommitsAhead(maxCommitsAhead int) Option {
	return func(c *CommitsAheadConfig) {
		if maxCommitsAhead >= 0 {
			c.MaxCommitsAhead = maxCommitsAhead
		}
	}
}

// addError adds a structured validation error.
func (c *CommitsAhead) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("CommitsAhead", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	c.errors = append(c.errors, err)
}

// ValidateNumberOfCommits checks if the current HEAD exceeds the maximum
// allowed commits ahead of a reference branch.
//
// Parameters:
//   - repo: Repository to check
//   - ref: Reference branch name (can be a short name like "main" or full ref path)
//   - opts: Optional configuration settings
//
// Returns:
//   - A commitsAhead instance with validation results
func ValidateNumberOfCommits(
	repo *model.Repository,
	ref string,
	opts ...Option,
) *CommitsAhead {
	// Apply configuration options
	config := DefaultCommitsAheadConfig()
	for _, opt := range opts {
		opt(&config)
	}

	rule := &CommitsAhead{
		ref:    ref,
		config: config,
	}

	if repo == nil {
		rule.addError(
			"nil_repo",
			"repository cannot be nil",
			nil,
		)

		return rule
	}

	if ref == "" {
		rule.addError(
			"empty_ref",
			"reference cannot be empty",
			nil,
		)

		return rule
	}

	// Ensure reference has proper format
	fullRef := ensureFullReference(ref)

	// Count commits ahead
	ahead, err := countCommitsAhead(repo, fullRef)
	if err != nil {
		rule.addError(
			"git_error",
			fmt.Sprintf("failed to count commits: %v", err),
			map[string]string{
				"error_details": err.Error(),
			},
		)

		return rule
	}

	rule.Ahead = ahead

	// Check if exceeds maximum allowed
	if ahead > config.MaxCommitsAhead {
		rule.addError(
			"too_many_commits",
			fmt.Sprintf("HEAD is %d commit(s) ahead of %s (max: %d)",
				ahead, ref, config.MaxCommitsAhead),
			map[string]string{
				"actual":    strconv.Itoa(ahead),
				"maximum":   strconv.Itoa(config.MaxCommitsAhead),
				"reference": ref,
			},
		)
	}

	return rule
}

// ensureFullReference ensures the reference has the correct format.
// Git references should be in the format refs/heads/branch-name.
func ensureFullReference(ref string) string {
	// Already a full reference path
	if strings.HasPrefix(ref, "refs/") {
		return ref
	}

	// Handle common cases like "origin/main"
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		remote := parts[0]
		branch := parts[1]

		return fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	}

	// Convert short branch name to full reference
	return "refs/heads/" + ref
}

// countCommitsAhead safely counts commits ahead of the reference.
// Uses panic recovery to handle unexpected issues with the git package.
func countCommitsAhead(repo *model.Repository, ref string) (int, error) {
	var err error

	// Use a closure with recover to handle potential panics from the git package
	ahead, err := func() (int, error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic while checking ahead count: %v", r)
			}
		}()

		return git.IsAhead(repo, ref)
	}()

	// Special cases handling
	if err != nil {
		// Reference not found is treated as 0 commits ahead
		if strings.Contains(err.Error(), "reference not found") {
			return 0, nil
		}

		// Handle permission errors separately
		if strings.Contains(err.Error(), "permission denied") {
			return 0, fmt.Errorf("permission denied accessing git repository: %w", err)
		}
	}

	return ahead, err
}

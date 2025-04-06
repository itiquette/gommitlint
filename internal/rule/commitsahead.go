// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
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
	errors []error
}

// Name returns the rule identifier.
func (c *CommitsAhead) Name() string {
	return "CommitsAhead"
}

// Result returns a string representation of the rule's status.
func (c *CommitsAhead) Result() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", c.Ahead, c.ref)
}

// Errors returns any violations detected by the rule.
func (c *CommitsAhead) Errors() []error {
	return c.errors
}

// Help returns a description of how to fix the rule violation.
func (c *CommitsAhead) Help() string {
	if len(c.errors) == 0 {
		return "No errors to fix"
	}

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

// addErrorf adds an error to the rule's errors slice.
func (c *CommitsAhead) addErrorf(format string, args ...interface{}) {
	c.errors = append(c.errors, fmt.Errorf(format, args...))
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
		rule.addErrorf("repository cannot be nil")

		return rule
	}

	if ref == "" {
		rule.addErrorf("reference cannot be empty")

		return rule
	}

	// Ensure reference has proper format
	fullRef := ensureFullReference(ref)

	// Count commits ahead
	ahead, err := countCommitsAhead(repo, fullRef)
	if err != nil {
		rule.addErrorf("failed to count commits: %v", err)

		return rule
	}

	rule.Ahead = ahead

	// Check if exceeds maximum allowed
	if ahead > config.MaxCommitsAhead {
		rule.addErrorf("HEAD is %d commit(s) ahead of %s (max: %d)",
			ahead, ref, config.MaxCommitsAhead)
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

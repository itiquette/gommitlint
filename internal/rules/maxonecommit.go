// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rules

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// CommitsAheadConfig provides configuration for commits ahead validation.
type CommitsAheadConfig struct {
	// MaxCommitsAhead defines the maximum number of commits allowed ahead of the reference
	MaxCommitsAhead int
	// IgnoreBranches allows excluding specific branches from the check
	IgnoreBranches []string
	// EnforceOnBranches limits the check to specific branches
	EnforceOnBranches []string
}

// DefaultCommitsAheadConfig returns a default configuration.
func DefaultCommitsAheadConfig() CommitsAheadConfig {
	return CommitsAheadConfig{
		MaxCommitsAhead:   1,
		IgnoreBranches:    []string{"main", "master"},
		EnforceOnBranches: nil, // enforce on all branches by default
	}
}

// NumberOfCommitsCheck enforces a maximum number of commits ahead of a reference.
type NumberOfCommitsCheck struct {
	ref    string
	ahead  int
	config CommitsAheadConfig
	errors []error
}

// Status returns the name of the check.
func (c *NumberOfCommitsCheck) Status() string {
	return "Number of Commits"
}

// Message returns the check message.
func (c *NumberOfCommitsCheck) Message() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", c.ahead, c.ref)
}

// Errors returns any violations of the check.
func (c *NumberOfCommitsCheck) Errors() []error {
	return c.errors
}
func ValidateNumberOfCommits(
	gitClient *git.Git,
	ref string,
	opts ...func(*CommitsAheadConfig),
) interfaces.Check {
	// Start with default configuration
	config := DefaultCommitsAheadConfig()

	// Apply any provided configuration options
	for _, opt := range opts {
		opt(&config)
	}

	check := &NumberOfCommitsCheck{
		ref:    ref,
		config: config,
	}

	// Validate git client
	if gitClient == nil {
		check.errors = append(check.errors,
			errors.New("git client is nil"))

		return check
	}

	// Get current branch name
	head, err := gitClient.Repo.Head()
	if err != nil {
		check.errors = append(check.errors,
			fmt.Errorf("failed to get current branch: %w", err))

		return check
	}

	currentBranch := head.Name().Short()

	// Check if branch should be ignored
	if contains(config.IgnoreBranches, currentBranch) {
		return check
	}

	// Check if enforcement is limited to specific branches
	if len(config.EnforceOnBranches) > 0 &&
		!contains(config.EnforceOnBranches, currentBranch) {
		return check
	}

	// Ensure the ref is a full reference name
	if !strings.HasPrefix(ref, "refs/") {
		ref = "refs/heads/" + ref
	}

	// Gracefully handle reference checking
	ahead, _, err := func() (int, int, error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic while checking ahead/behind: %v", r)
			}
		}()

		return gitClient.AheadBehind(ref)
	}()

	if err != nil {
		// If reference not found, treat as 0 commits ahead
		if strings.Contains(err.Error(), "reference not found") {
			ahead = 0
		} else {
			check.errors = append(check.errors,
				fmt.Errorf("failed to check ahead/behind status: %w", err))

			return check
		}
	}

	check.ahead = ahead

	// Validate number of commits
	if ahead > config.MaxCommitsAhead {
		check.errors = append(check.errors,
			fmt.Errorf("HEAD is %d commit(s) ahead of %s (max: %d)",
				ahead, ref, config.MaxCommitsAhead))
	}

	return check
}

func WithMaxCommitsAhead(maxnr int) func(*CommitsAheadConfig) {
	return func(c *CommitsAheadConfig) {
		if maxnr >= 0 {
			c.MaxCommitsAhead = maxnr
		}
	}
}

func WithIgnoreBranches(branches ...string) func(*CommitsAheadConfig) {
	return func(c *CommitsAheadConfig) {
		c.IgnoreBranches = append(c.IgnoreBranches, branches...)
	}
}

func WithEnforceOnBranches(branches ...string) func(*CommitsAheadConfig) {
	return func(c *CommitsAheadConfig) {
		c.EnforceOnBranches = branches
	}
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

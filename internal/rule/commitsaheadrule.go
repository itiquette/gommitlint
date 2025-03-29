// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/model"
)

// CommitsAheadConfig provides configuration for the CommitsAheadRule.
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

// CommitsAheadRule enforces a maximum number of commits ahead of a reference.
type CommitsAheadRule struct {
	ref    string
	Ahead  int
	config CommitsAheadConfig
	errors []error
}

// Name returns the rule identifier.
func (c *CommitsAheadRule) Name() string {
	return "CommitsAheadRule"
}

// Result returns a string representation of the rule's status.
func (c *CommitsAheadRule) Result() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", c.Ahead, c.ref)
}

// Errors returns any violations detected by the rule.
func (c *CommitsAheadRule) Errors() []error {
	return c.errors
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

// ValidateNumberOfCommits checks if the current HEAD exceeds the maximum
// allowed commits ahead of a reference branch.
func ValidateNumberOfCommits(
	repo *model.Repository,
	ref string,
	opts ...Option,
) *CommitsAheadRule {
	// Apply configuration options
	config := DefaultCommitsAheadConfig()
	for _, opt := range opts {
		opt(&config)
	}

	rule := &CommitsAheadRule{
		ref:    ref,
		config: config,
	}

	// Ensure reference has proper format
	fullRef := ensureFullReference(ref)

	// Count commits ahead
	ahead, err := countCommitsAhead(repo, fullRef)
	if err != nil {
		rule.errors = append(rule.errors, err)

		return rule
	}

	rule.Ahead = ahead

	// Check if exceeds maximum allowed
	if ahead > config.MaxCommitsAhead {
		rule.errors = append(rule.errors,
			fmt.Errorf("HEAD is %d commit(s) ahead of %s (max: %d)",
				ahead, ref, config.MaxCommitsAhead))
	}

	return rule
}

// ensureFullReference ensures the reference has the correct format.
func ensureFullReference(ref string) string {
	if !strings.HasPrefix(ref, "refs/") {
		return "refs/heads/" + ref
	}

	return ref
}

// countCommitsAhead safely counts commits ahead of the reference.
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

	// Handle special case for missing reference
	if err != nil && strings.Contains(err.Error(), "reference not found") {
		return 0, nil
	}

	return ahead, err
}

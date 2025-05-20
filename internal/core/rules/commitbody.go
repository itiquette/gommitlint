// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// CommitBodyRule validates commit message bodies based on configured requirements.
type CommitBodyRule struct {
	name      string
	verbosity string
}

// NewCommitBodyRule creates a new CommitBodyRule.
func NewCommitBodyRule() CommitBodyRule {
	return CommitBodyRule{
		name: "CommitBody",
	}
}

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return r.name
}

// Validate checks if a commit's body meets the required criteria from configuration.
func (r CommitBodyRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Log validation at trace level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating commit body",
		"rule", r.Name(),
		"commit_hash", commit.Hash)

	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		// If no config is available, return nil (no validation)
		return nil
	}

	// Check if this rule is enabled
	if !IsRuleEnabled(ctx, r.Name()) {
		logger.Debug("Body rule is disabled, skipping validation")

		return nil
	}

	errors := []appErrors.ValidationError{}

	// Check if body is empty
	if strings.TrimSpace(commit.Body) == "" {
		err := appErrors.NewBodyError(
			appErrors.ErrMissingBody,
			r.Name(),
			"Commit message body is missing",
			"is empty",
			"Add a blank line after the subject, followed by a detailed description of your changes",
		).WithContext("commit", commit.Hash)
		errors = append(errors, err)
	}

	// Check minimum body length
	minLength := cfg.GetInt("body.min_length")
	if minLength > 0 && len(strings.TrimSpace(commit.Body)) < minLength {
		actualLength := len(strings.TrimSpace(commit.Body))
		err := appErrors.NewBodyError(
			appErrors.ErrBodyTooShort,
			r.Name(),
			fmt.Sprintf("Commit body is too short (%d chars, minimum: %d)", actualLength, minLength),
			fmt.Sprintf("has %d characters", actualLength),
			fmt.Sprintf("Provide at least %d characters of detail in your commit body", minLength),
		).WithContextMap(map[string]string{
			"commit":        commit.Hash,
			"min_length":    strconv.Itoa(minLength),
			"actual_length": strconv.Itoa(actualLength),
		})
		errors = append(errors, err)
	}

	logger.Debug("Commit body validation complete",
		"error_count", len(errors),
		"min_length", minLength)

	return errors
}

// WithVerbosity returns a new CommitBodyRule with the specified verbosity level.
func (r CommitBodyRule) WithVerbosity(verbosity string) CommitBodyRule {
	newRule := r
	newRule.verbosity = verbosity

	return newRule
}

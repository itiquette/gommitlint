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
	minLength int
}

// NewCommitBodyRule creates a new CommitBodyRule.
func NewCommitBodyRule() CommitBodyRule {
	return CommitBodyRule{
		name:      "CommitBody",
		minLength: 0, // Default to no minimum
	}
}

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return r.name
}

// WithContext implements the ConfigurableRule interface for CommitBodyRule.
// It returns a new rule with configuration from the provided context.
func (r CommitBodyRule) WithContext(ctx context.Context) domain.Rule {
	// Get configuration directly from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return r
	}

	// Create a copy of the rule
	result := r

	// Get min length from configuration
	minLength := cfg.GetInt("message.body.min_length")
	if minLength > 0 {
		result.minLength = minLength
	}

	return result
}

// Validate checks if a commit's body meets the required criteria.
func (r CommitBodyRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Check if this rule is enabled using domain priority service
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return nil
	}

	enabledRules := cfg.GetStringSlice("rules.enabled")
	disabledRules := cfg.GetStringSlice("rules.disabled")
	priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())

	if !priorityService.IsRuleEnabled(ctx, r.Name(), enabledRules, disabledRules) {
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
	if r.minLength > 0 && len(strings.TrimSpace(commit.Body)) < r.minLength {
		actualLength := len(strings.TrimSpace(commit.Body))
		err := appErrors.NewBodyError(
			appErrors.ErrBodyTooShort,
			r.Name(),
			fmt.Sprintf("Commit body is too short (%d chars, minimum: %d)", actualLength, r.minLength),
			fmt.Sprintf("has %d characters", actualLength),
			fmt.Sprintf("Provide at least %d characters of detail in your commit body", r.minLength),
		).WithContextMap(map[string]string{
			"commit":        commit.Hash,
			"min_length":    strconv.Itoa(r.minLength),
			"actual_length": strconv.Itoa(actualLength),
		})
		errors = append(errors, err)
	}

	return errors
}

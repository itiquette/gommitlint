// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"

	"github.com/itiquette/gommitlint/internal/errors"
)

// SeverityLevel represents the severity of a rule violation.
type SeverityLevel string

const (
	// SeverityError indicates a rule violation that should fail validation.
	SeverityError SeverityLevel = "error"

	// SeverityWarning indicates a rule violation that should warn but not fail validation.
	SeverityWarning SeverityLevel = "warning"

	// SeverityInfo indicates informational output that isn't a violation.
	SeverityInfo SeverityLevel = "info"
)

// RuleMetadata provides information about a validation rule.
type RuleMetadata struct {
	// ID is the unique identifier for the rule.
	ID string

	// Name is the human-readable name of the rule.
	Name string

	// Description is a detailed description of what the rule validates.
	Description string

	// Severity indicates how severe violations of this rule are.
	Severity SeverityLevel
}

// Rule defines the interface for all validation rules.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation against a commit.
	Validate(ctx context.Context, commit CommitInfo) []errors.ValidationError

	// Result returns a concise result message.
	Result(errors []errors.ValidationError) string

	// VerboseResult returns a detailed result message.
	VerboseResult(errors []errors.ValidationError) string

	// Help returns guidance on how to fix rule violations.
	Help(errors []errors.ValidationError) string

	// Errors returns all validation errors found by this rule.
	Errors() []errors.ValidationError
}

// RuleProvider defines an interface for retrieving validation rules.
type RuleProvider interface {
	// GetRules returns all available validation rules.
	GetRules(ctx context.Context) []Rule

	// GetActiveRules returns all active validation rules based on configuration.
	GetActiveRules(ctx context.Context) []Rule

	// WithActiveRules returns a new provider with the specified active rules.
	WithActiveRules(ruleNames []string) RuleProvider

	// WithDisabledRules returns a new provider with the specified rules disabled.
	WithDisabledRules(ruleNames []string) RuleProvider

	// WithCustomRule returns a new provider with the custom rule added.
	WithCustomRule(rule Rule) RuleProvider
}

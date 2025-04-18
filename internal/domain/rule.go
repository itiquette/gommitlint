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

// Note: All validation error handling has been moved to the errors package.
// Use errors.ValidationErrorCode and errors.ValidationError instead.

// Rule defines the interface for all validation rules.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation against a commit.
	Validate(commit *CommitInfo) []errors.ValidationError

	// Result returns a concise result message.
	Result() string

	// VerboseResult returns a detailed result message.
	VerboseResult() string

	// Help returns guidance on how to fix rule violations.
	Help() string

	// Errors returns all validation errors found by this rule.
	Errors() []errors.ValidationError
}

// ContextualRule extends Rule with context-aware methods.
type ContextualRule interface {
	Rule

	// ValidateWithContext performs validation with context.
	ValidateWithContext(ctx context.Context, commit *CommitInfo) []errors.ValidationError
}

// RuleProvider defines an interface for retrieving validation rules.
type RuleProvider interface {
	// GetRules returns all available validation rules.
	GetRules() []Rule

	// GetActiveRules returns all active validation rules based on configuration.
	GetActiveRules() []Rule
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// ValidationStatus represents the result status of a validation rule.
type ValidationStatus string

const (
	// StatusPassed indicates the rule passed validation.
	StatusPassed ValidationStatus = "passed"

	// StatusFailed indicates the rule failed validation.
	StatusFailed ValidationStatus = "failed"

	// StatusSkipped indicates the rule was skipped for some reason.
	StatusSkipped ValidationStatus = "skipped"
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

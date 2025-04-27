// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package domain contains the core domain entities and interfaces for gommitlint.
package domain

// RuleConfigProvider defines the interface for configuring validation rules.
// This interface is separate from RuleProvider in rule.go, which deals with
// retrieving rule instances.
type RuleConfigProvider interface {
	// GetAvailableRules returns a list of all available rule names.
	GetAvailableRules() []string

	// GetActiveRules returns a list of currently active rule names.
	GetActiveRules() []string

	// EnabledRules returns the list of explicitly enabled rule names from configuration.
	EnabledRules() []string

	// DisabledRules returns the list of explicitly disabled rule names from configuration.
	DisabledRules() []string

	// SetEnabledRules sets the list of enabled rule names.
	SetEnabledRules(rules []string)

	// SetDisabledRules sets the list of disabled rule names.
	SetDisabledRules(rules []string)
}

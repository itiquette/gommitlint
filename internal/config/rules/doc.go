// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package rules provides rule-specific configuration utilities.
//
// This package contains helper functions for managing validation rules,
// including priority logic for determining which rules are enabled or
// disabled based on configuration.
//
// The priority system follows these principles:
//  1. Explicitly disabled rules are always disabled (highest priority)
//  2. Explicitly enabled rules are enabled (unless explicitly disabled)
//  3. Default-disabled rules are disabled (unless explicitly enabled)
//  4. All other rules are enabled by default
//
// Functions:
//   - CleanRuleName: Standardizes rule names by removing quotes and whitespace
//   - RulePriority: Determines if a rule should be active
//   - IsRuleEnabled: Checks if a rule is enabled based on configuration
//   - IsExplicitlyEnabled: Checks if a rule is explicitly enabled
//   - RemoveExplicitlyEnabledFromDisabled: Filters disabled rules
//   - MergeEnabledRules: Combines default and configured enabled rules
package rules

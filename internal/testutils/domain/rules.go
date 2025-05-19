// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package domain contains test utilities for the domain package.
// This package is intended for testing purposes only.
//
// # Context Keys
//
// This package defines context keys specifically used for overriding behavior
// in rule tests. These keys allow tests to control rule behavior without
// modifying production code paths.
package domain

// ContextKey is a type used for context keys in tests.
// These keys are used to signal test-specific behavior overrides.
type ContextKey string

// Predefined context keys for various rule tests.
// These constants allow tests to control rule behavior without affecting production code.
const (
	// JiraTestOverrideKey signals the Jira rule to use test-specific logic.
	JiraTestOverrideKey ContextKey = "jira_test_override"

	// SignatureTestOverrideKey signals the Signature rule to use test-specific logic.
	SignatureTestOverrideKey ContextKey = "signature_test_override"
)

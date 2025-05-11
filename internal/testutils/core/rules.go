// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package core contains test utilities for the core package.
// This package is intended for testing purposes only.
package core

// ContextKey is a type used for context keys in tests.
type ContextKey string

// Predefined context keys for various rule tests.
const (
	JiraTestOverrideKey      ContextKey = "jira_test_override"
	SignatureTestOverrideKey ContextKey = "signature_test_override"
)

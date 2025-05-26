// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation contains test utilities for validation components.
// This package is intended for testing purposes only.
package validation

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// CreateTestEngine creates an engine instance for testing with a custom rule registry.
// This method now returns a TestEngine which is a test-specific wrapper around domain.ValidationEngine.
func CreateTestEngine(registry domain.RuleRegistry) *TestEngine {
	return &TestEngine{
		Registry: registry,
	}
}

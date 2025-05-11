// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// CreateTestEngine creates an engine instance for testing with a custom rule provider.
func CreateTestEngine(provider domain.RuleProvider) Engine {
	return Engine{
		ruleProvider: provider,
	}
}

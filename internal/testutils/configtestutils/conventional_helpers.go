// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithConventionalRequired returns a new ConventionalConfig with the updated required setting.
func (t *TestUtils) WithConventionalRequired(c types.ConventionalConfig, required bool) types.ConventionalConfig {
	result := c
	result.Required = required

	return result
}

// WithRequireScope returns a new ConventionalConfig with the updated scope requirement.
func (t *TestUtils) WithRequireScope(c types.ConventionalConfig, require bool) types.ConventionalConfig {
	result := c
	result.RequireScope = require

	return result
}

// WithTypes returns a new ConventionalConfig with the updated allowed types.
func (t *TestUtils) WithTypes(c types.ConventionalConfig, types []string) types.ConventionalConfig {
	result := c
	result.Types = make([]string, len(types))
	copy(result.Types, types)

	return result
}

// WithScopes returns a new ConventionalConfig with the updated allowed scopes.
func (t *TestUtils) WithScopes(c types.ConventionalConfig, scopes []string) types.ConventionalConfig {
	result := c
	result.Scopes = make([]string, len(scopes))
	copy(result.Scopes, scopes)

	return result
}

// WithAllowBreakingChanges returns a new ConventionalConfig with the updated breaking changes setting.
func (t *TestUtils) WithAllowBreakingChanges(c types.ConventionalConfig, allow bool) types.ConventionalConfig {
	result := c
	result.AllowBreakingChanges = allow

	return result
}

// WithMaxDescriptionLength returns a new ConventionalConfig with the updated max description length.
func (t *TestUtils) WithMaxDescriptionLength(c types.ConventionalConfig, maxLength int) types.ConventionalConfig {
	result := c
	result.MaxDescriptionLength = maxLength

	return result
}

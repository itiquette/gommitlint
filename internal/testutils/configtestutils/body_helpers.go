// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithRequired returns a new BodyConfig with the updated required setting.
func (t *TestUtils) WithRequired(c types.BodyConfig, required bool) types.BodyConfig {
	result := c
	result.Required = required

	return result
}

// WithMinLength returns a new BodyConfig with the updated minimum length setting.
func (t *TestUtils) WithMinLength(c types.BodyConfig, minLength int) types.BodyConfig {
	result := c
	result.MinLength = minLength

	return result
}

// WithMinimumLines returns a new BodyConfig with the updated minimum lines setting.
func (t *TestUtils) WithMinimumLines(c types.BodyConfig, minLines int) types.BodyConfig {
	result := c
	result.MinimumLines = minLines

	return result
}

// WithAllowSignOffOnly returns a new BodyConfig with the updated sign-off only setting.
func (t *TestUtils) WithAllowSignOffOnly(c types.BodyConfig, allow bool) types.BodyConfig {
	result := c
	result.AllowSignOffOnly = allow

	return result
}

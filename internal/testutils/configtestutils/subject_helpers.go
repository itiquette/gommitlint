// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithMaxLength returns a new SubjectConfig with the updated max length.
func (t *TestUtils) WithMaxLength(c types.SubjectConfig, maxLength int) types.SubjectConfig {
	result := c
	result.MaxLength = maxLength

	return result
}

// WithCase returns a new SubjectConfig with the updated case.
func (t *TestUtils) WithCase(c types.SubjectConfig, caseStyle string) types.SubjectConfig {
	result := c
	result.Case = caseStyle

	return result
}

// WithRequireImperative returns a new SubjectConfig with the updated imperative requirement.
func (t *TestUtils) WithRequireImperative(c types.SubjectConfig, require bool) types.SubjectConfig {
	result := c
	result.RequireImperative = require

	return result
}

// WithDisallowedSuffixes returns a new SubjectConfig with the updated disallowed suffixes.
func (t *TestUtils) WithDisallowedSuffixes(c types.SubjectConfig, suffixes []string) types.SubjectConfig {
	result := c
	result.DisallowedSuffixes = suffixes

	return result
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithFormat returns a new OutputConfig with the updated format.
func (t *TestUtils) WithFormat(c types.OutputConfig, format string) types.OutputConfig {
	result := c
	result.Format = format

	return result
}

// WithVerbose returns a new OutputConfig with the updated verbose setting.
func (t *TestUtils) WithVerbose(c types.OutputConfig, verbose bool) types.OutputConfig {
	result := c
	result.Verbose = verbose

	return result
}

// WithJSON returns a new OutputConfig with format set to JSON.
func (t *TestUtils) WithJSON(c types.OutputConfig) types.OutputConfig {
	return t.WithFormat(c, "json")
}

// WithText returns a new OutputConfig with format set to text.
func (t *TestUtils) WithText(c types.OutputConfig) types.OutputConfig {
	return t.WithFormat(c, "text")
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithGPGRequired returns a new SecurityConfig with the updated GPG required setting.
func (t *TestUtils) WithGPGRequired(c types.SecurityConfig, required bool) types.SecurityConfig {
	result := c
	result.GPGRequired = required

	return result
}

// WithSignOffRequired returns a new SecurityConfig with the updated sign-off required setting.
func (t *TestUtils) WithSignOffRequired(c types.SecurityConfig, required bool) types.SecurityConfig {
	result := c
	result.SignOffRequired = required

	return result
}

// WithKeyDirectory returns a new SecurityConfig with the updated key directory.
func (t *TestUtils) WithKeyDirectory(c types.SecurityConfig, keyDir string) types.SecurityConfig {
	result := c
	result.KeyDirectory = keyDir

	return result
}

// WithAllowedSignatureTypes returns a new SecurityConfig with the updated allowed signature types.
func (t *TestUtils) WithAllowedSignatureTypes(c types.SecurityConfig, types []string) types.SecurityConfig {
	result := c
	result.AllowedSignatureTypes = types

	return result
}

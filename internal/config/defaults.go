// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/config/types"
)

// NewDefaultConfig creates a default configuration.
// This is a convenience function that forwards to the config service.
func NewDefaultConfig() types.Config {
	return config.NewDefaultConfig()
}

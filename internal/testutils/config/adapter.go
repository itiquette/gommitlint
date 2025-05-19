// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

// THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE

import (
	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/config/types"
)

// Adapter implements the common.config.Config interface for testing.
type Adapter struct {
	*infraConfig.Adapter
}

// NewAdapter creates a new Adapter from a types.Config.
func NewAdapter(cfg types.Config) *Adapter {
	return &Adapter{
		Adapter: infraConfig.NewAdapter(cfg),
	}
}

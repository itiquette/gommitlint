// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config contains test utilities for configuration components.
// This package is intended for testing purposes only.
package config

// THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE

import (
	"context"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WrapAndInjectConfig adds a wrapped internal Config to the context.
func WrapAndInjectConfig(ctx context.Context, cfg types.Config) context.Context {
	adapter := NewAdapter(cfg)

	return contextx.WithConfig(ctx, adapter.Adapter)
}

// WrappedConfig takes an internal config and returns it as a common interface.
func WrappedConfig(cfg types.Config) interface{} {
	return NewAdapter(cfg).Adapter
}

// NewContextWithConfig creates a new context with the given configuration.
func NewContextWithConfig(ctx context.Context, cfg types.Config) context.Context {
	return contextx.WithConfig(ctx, NewAdapter(cfg).Adapter)
}

// CreateTestConfig creates a new test config with sensible defaults.
func CreateTestConfig() types.Config {
	return config.NewDefaultConfig()
}

// CreateMinimalTestConfig creates a minimal test config with most rules disabled.
func CreateMinimalTestConfig() types.Config {
	return Minimal().Build()
}

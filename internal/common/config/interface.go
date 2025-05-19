// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides common interfaces and utilities for configuration.
// This package has minimal dependencies and is designed to be
// imported by any package without creating import cycles.
//
// This package uses private context keys to enforce encapsulation and proper
// access patterns. Configuration must be accessed through the GetConfig()
// function rather than direct context access, following hexagonal architecture
// principles and ensuring domain logic doesn't depend on implementation details.
package config

import "context"

// Config is a minimal interface for accessing configuration.
// Each component can define its own specific interface that extends this one.
type Config interface {
	// Get returns a configuration value for a given key.
	Get(key string) interface{}
	// GetString returns a string configuration value for a given key.
	GetString(key string) string
	// GetBool returns a boolean configuration value for a given key.
	GetBool(key string) bool
	// GetInt returns an integer configuration value for a given key.
	GetInt(key string) int
	// GetStringSlice returns a string slice configuration value for a given key.
	GetStringSlice(key string) []string
}

// Provider is an interface for providing configuration.
type Provider interface {
	// GetConfig returns the current configuration.
	GetConfig() Config
}

// WithConfig adds configuration to the context.
func WithConfig(ctx context.Context, cfg Config) context.Context {
	return context.WithValue(ctx, configKey{}, cfg)
}

// GetConfig retrieves configuration from the context.
func GetConfig(ctx context.Context) Config {
	if ctx == nil {
		return &emptyConfig{}
	}

	if cfg, ok := ctx.Value(configKey{}).(Config); ok {
		return cfg
	}

	return &emptyConfig{}
}

// WithProvider adds a configuration provider to the context.
func WithProvider(ctx context.Context, provider Provider) context.Context {
	return context.WithValue(ctx, providerKey{}, provider)
}

// GetProvider retrieves a configuration provider from the context.
func GetProvider(ctx context.Context) Provider {
	if ctx == nil {
		return &emptyProvider{}
	}

	if provider, ok := ctx.Value(providerKey{}).(Provider); ok {
		return provider
	}

	return &emptyProvider{}
}

// Key returns the key used to store configuration in the context.
// This is exported to allow other packages to access the same key.
func Key() interface{} {
	return configKey{}
}

// ProviderKey returns the key used to store the provider in the context.
// This is exported to allow other packages to access the same key.
func ProviderKey() interface{} {
	return providerKey{}
}

// configKey is a private type for context keys to avoid collisions.
type configKey struct{}

// providerKey is a private type for context keys to avoid collisions.
type providerKey struct{}

// emptyConfig is a Config implementation that returns empty values.
type emptyConfig struct{}

func (c *emptyConfig) Get(_ string) interface{}         { return nil }
func (c *emptyConfig) GetString(_ string) string        { return "" }
func (c *emptyConfig) GetBool(_ string) bool            { return false }
func (c *emptyConfig) GetInt(_ string) int              { return 0 }
func (c *emptyConfig) GetStringSlice(_ string) []string { return []string{} }

// emptyProvider is a Provider implementation that returns an empty config.
type emptyProvider struct{}

func (p *emptyProvider) GetConfig() Config { return &emptyConfig{} }

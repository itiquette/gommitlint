// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigInContext(t *testing.T) {
	// Create a context with configuration
	cfg := DefaultConfig().
		WithSubject(DefaultConfig().Subject.WithMaxLength(100))

	ctx := context.Background()
	ctx = WithConfig(ctx, cfg)

	// Retrieve configuration from context
	retrievedConfig := GetConfig(ctx)
	require.Equal(t, 100, retrievedConfig.Subject.MaxLength)

	// Update configuration in context
	ctx = UpdateConfig(ctx, func(c Config) Config {
		return c.WithBody(c.Body.WithRequired(false))
	})

	// Verify update
	updatedConfig := GetConfig(ctx)
	require.Equal(t, 100, updatedConfig.Subject.MaxLength)
	require.False(t, updatedConfig.Body.Required)
}

func TestNewProviderWithConfig(t *testing.T) {
	// Create a custom configuration
	cfg := DefaultConfig().
		WithSubject(DefaultConfig().Subject.WithMaxLength(100))

	// Create a provider with the configuration
	provider := NewProviderWithConfig(cfg)

	// Verify provider configuration
	require.Equal(t, 100, provider.GetConfig().Subject.MaxLength)

	// Update provider configuration
	provider.UpdateConfig(func(c Config) Config {
		return c.WithBody(c.Body.WithRequired(false))
	})

	// Verify update
	updatedConfig := provider.GetConfig()
	require.Equal(t, 100, updatedConfig.Subject.MaxLength)
	require.False(t, updatedConfig.Body.Required)
}

func TestProviderInContext(t *testing.T) {
	// Create a provider with custom configuration
	cfg := DefaultConfig().
		WithSubject(DefaultConfig().Subject.WithMaxLength(100))

	provider := NewProviderWithConfig(cfg)

	// Add provider to context
	ctx := context.Background()
	ctx = WithProviderInContext(ctx, provider)

	// Retrieve provider from context
	retrievedProvider, err := GetProviderFromContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, retrievedProvider)
	require.Equal(t, 100, retrievedProvider.GetConfig().Subject.MaxLength)

	// Update provider configuration
	retrievedProvider.UpdateConfig(func(c Config) Config {
		return c.WithBody(c.Body.WithRequired(false))
	})

	// Verify update through context
	updatedConfig := GetConfig(ctx)
	require.Equal(t, 100, updatedConfig.Subject.MaxLength)
	require.False(t, updatedConfig.Body.Required)

	// Verify original provider was updated
	require.False(t, provider.GetConfig().Body.Required)
}

func TestWithGitRepository(t *testing.T) {
	// Create a provider
	provider := NewProviderWithConfig(DefaultConfig())

	// Set repository path
	provider.WithGitRepository("/path/to/repo")

	// Verify repository path was set
	cfg := provider.GetConfig()
	require.Equal(t, "/path/to/repo", cfg.Repository.Path)
}

func TestUpdateConfig(t *testing.T) {
	// Create a context without a provider
	ctx := context.Background()

	// Update configuration
	ctx = UpdateConfig(ctx, func(c Config) Config {
		return c.WithSubject(c.Subject.WithMaxLength(100))
	})

	// Verify a provider was created and configuration was updated
	cfg := GetConfig(ctx)
	require.Equal(t, 100, cfg.Subject.MaxLength)

	// Add another update
	ctx = UpdateConfig(ctx, func(c Config) Config {
		return c.WithBody(c.Body.WithRequired(false))
	})

	// Verify both updates were applied
	updatedCfg := GetConfig(ctx)
	require.Equal(t, 100, updatedCfg.Subject.MaxLength)
	require.False(t, updatedCfg.Body.Required)
}

func TestDefaultConfigFromContext(t *testing.T) {
	// Create a context without configuration
	ctx := context.Background()

	// Get configuration (should be default)
	cfg := GetConfig(ctx)

	// Verify default values
	require.Equal(t, 72, cfg.Subject.MaxLength)
	require.Equal(t, "sentence", cfg.Subject.Case)
	require.True(t, cfg.Body.Required)
}

func TestNilContext(t *testing.T) {
	// Get configuration from empty context (should be default)
	cfg := GetConfig(context.Background())

	// Verify default values
	require.Equal(t, 72, cfg.Subject.MaxLength)
	require.Equal(t, "sentence", cfg.Subject.Case)
	require.True(t, cfg.Body.Required)

	// Get provider from empty context (should return error)
	provider, err := GetProviderFromContext(context.Background())
	require.Error(t, err)
	require.Nil(t, provider)
}

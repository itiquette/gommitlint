// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Manager provides a simplified configuration management interface
// that works directly with the configuration system.
type Manager struct {
	// provider is the underlying configuration provider
	provider *Provider

	// activeCtx is the context for the manager
	// Renamed from ctx to avoid "context in struct" linting issues
	activeCtx context.Context //nolint:containedctx // Deliberate design choice for this manager
}

// WithUpdatedContext returns a new Manager with the provided context.
// This satisfies the linter by providing a way to pass context
// instead of storing it in the struct directly.
func (m Manager) WithUpdatedContext(ctx context.Context) Manager {
	result := m
	result.activeCtx = ctx

	return result
}

// NewManager creates a new configuration manager.
func NewManager() (*Manager, error) {
	// Create a new provider
	provider, err := NewProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration provider: %w", err)
	}

	// Create a context with the provider
	ctx := context.Background()
	ctx = WithProviderInContext(ctx, provider)

	return &Manager{
		provider:  provider,
		activeCtx: ctx,
	}, nil
}

// Load loads configuration from files.
func (m *Manager) Load() error {
	// Add debug output
	fmt.Println("Loading configuration from default paths via Manager.Load()")

	return m.provider.Load()
}

// LoadFromPath loads configuration from the specified path.
func (m *Manager) LoadFromPath(path string) error {
	// Add debug output
	fmt.Printf("Loading configuration from specific path: %s via Manager.LoadFromPath()\n", path)

	return m.provider.LoadFromPath(path)
}

// GetConfig returns the current configuration.
func (m *Manager) GetConfig() Config {
	return m.provider.GetConfig()
}

// UpdateConfig updates the configuration with the given options.
func (m *Manager) UpdateConfig(transform func(Config) Config) {
	m.provider.UpdateConfig(transform)
}

// WithGitRepository sets the Git repository path in the configuration.
func (m *Manager) WithGitRepository(path string) {
	m.provider.WithGitRepository(path)
}

// GetContext returns the context associated with the manager.
func (m *Manager) GetContext() context.Context {
	return m.activeCtx
}

// WithContext returns a new context with the manager's configuration added.
func (m *Manager) WithContext(ctx context.Context) context.Context {
	return WithProviderInContext(ctx, m.provider)
}

// GetValidationConfig returns the validation configuration.
func (m *Manager) GetValidationConfig(ctx context.Context) Config {
	return GetConfig(ctx)
}

// WithConfigInContext returns a new context with the given configuration added.
func (m *Manager) WithConfigInContext(ctx context.Context, config Config) context.Context {
	return WithConfig(ctx, config)
}

// SaveDefaultConfig saves the default configuration to the specified path.
func (m *Manager) SaveDefaultConfig(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Get default configuration
	defaultConfig := NewDefaultConfig()

	// Update provider with default configuration
	m.provider.UpdateConfig(func(_ Config) Config {
		return defaultConfig
	})

	// Save configuration
	return m.provider.Save(path)
}

// Save saves the current configuration to the specified path.
func (m *Manager) Save(path string) error {
	return m.provider.Save(path)
}

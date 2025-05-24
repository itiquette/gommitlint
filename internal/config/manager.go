// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"context"
	"fmt"

	infra "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
)

// Manager provides a simplified configuration management interface
// that works directly with the configuration service.
type Manager struct {
	// service is the underlying configuration service
	service *infra.Service
}

// NewManager creates a new configuration manager.
func NewManager(ctx context.Context) (*Manager, error) {
	// Use the logger from the context
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering config.NewManager")

	// Create a new service
	service, err := infra.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration service: %w", err)
	}

	// Load configuration
	loadedService, err := service.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &Manager{
		service: &loadedService,
	}, nil
}

// Load loads configuration from files.
func (m *Manager) Load() error {
	service, err := m.service.Load()
	if err != nil {
		return err
	}

	m.service = &service

	return nil
}

// LoadFromPath loads configuration from the specified path.
func (m *Manager) LoadFromPath(path string) error {
	service, err := m.service.LoadFromPath(path)
	if err != nil {
		return err
	}

	m.service = &service

	return nil
}

// GetConfig returns the current configuration.
func (m Manager) GetConfig() types.Config {
	return m.service.GetConfig()
}

// UpdateConfig returns a new Manager with updated configuration.
func (m Manager) UpdateConfig(transform func(types.Config) types.Config) Manager {
	newService := m.service.UpdateConfig(transform)
	newServicePtr := &newService

	return Manager{service: newServicePtr}
}

// WithGitRepository returns a new Manager with the Git repository path set.
func (m Manager) WithGitRepository(path string) Manager {
	return m.UpdateConfig(func(cfg types.Config) types.Config {
		cfg.Repo.Path = path

		return cfg
	})
}

// WithContext returns a new context with the manager's configuration added.
func (m Manager) WithContext(ctx context.Context) context.Context {
	adapter := m.service.GetAdapter()

	return contextx.WithConfig(ctx, adapter)
}

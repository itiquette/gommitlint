// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"context"
	"fmt"

	infra "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/config/types"
)

// Loader provides a simplified configuration loading interface
// that works directly with the configuration service.
type Loader struct {
	// service is the underlying configuration service
	service *infra.Service
}

// NewLoader creates a new configuration loader.
func NewLoader(ctx context.Context) (*Loader, error) {
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

	return &Loader{
		service: &loadedService,
	}, nil
}

// Load loads configuration from files.
func (l *Loader) Load() error {
	service, err := l.service.Load()
	if err != nil {
		return err
	}

	l.service = &service

	return nil
}

// LoadFromPath loads configuration from the specified path.
func (l *Loader) LoadFromPath(path string) error {
	service, err := l.service.LoadFromPath(path)
	if err != nil {
		return err
	}

	l.service = &service

	return nil
}

// GetConfig returns the current configuration.
func (l Loader) GetConfig() types.Config {
	return l.service.GetConfig()
}

// UpdateConfig returns a new Loader with updated configuration.
func (l Loader) UpdateConfig(transform func(types.Config) types.Config) Loader {
	newService := l.service.UpdateConfig(transform)
	newServicePtr := &newService

	return Loader{service: newServicePtr}
}

// WithGitRepository returns a new Loader with the Git repository path set.
func (l Loader) WithGitRepository(path string) Loader {
	return l.UpdateConfig(func(cfg types.Config) types.Config {
		cfg.Repo.Path = path

		return cfg
	})
}


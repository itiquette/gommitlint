// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package loader_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/loader"
	configTypes "github.com/itiquette/gommitlint/internal/config"
)

func TestService_Creation(t *testing.T) {
	// Test creating a new service
	service, err := loader.NewService()
	require.NoError(t, err)
	require.NotNil(t, service)

	// Verify default config is loaded
	cfg := service.GetConfig()
	require.NotNil(t, cfg)
	require.Equal(t, "sentence", cfg.Message.Subject.Case)
	require.Equal(t, 72, cfg.Message.Subject.MaxLength)
}

func TestService_UpdateConfig(t *testing.T) {
	// Create service
	service, err := loader.NewService()
	require.NoError(t, err)

	// Update config
	newService := service.UpdateConfig(func(cfg configTypes.Config) configTypes.Config {
		cfg.Message.Subject.MaxLength = 100

		return cfg
	})

	// Verify the update
	require.Equal(t, 100, newService.GetConfig().Message.Subject.MaxLength)
	// Verify immutability
	require.Equal(t, 72, service.GetConfig().Message.Subject.MaxLength)
}

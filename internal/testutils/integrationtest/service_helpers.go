// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"context"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/composition"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// CreateValidationService creates a validation service for tests using the composition root.
func CreateValidationService(ctx context.Context, logger outgoing.Logger, repoPath string) (incoming.ValidationService, error) {
	// Create config service
	configService, err := config.NewService()
	if err != nil {
		return nil, err
	}

	// Load configuration
	configService, err = configService.Load()
	if err != nil {
		return nil, err
	}

	// Create dependency container
	container := composition.NewContainer(logger, configService.GetAdapter().GetConfig())

	// Create validation service
	return container.CreateValidationService(ctx, repoPath)
}

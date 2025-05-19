// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/adapters/incoming/cli"
	infra "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// Root represents the composition root for the application.
// It wires up all the dependencies according to the Dependency Inversion Principle.
type Root struct {
	// Core components
	config types.Config

	// Adapters
	configAdapter *infra.Adapter
	gitRepository domain.CommitRepository

	// Application services
	validationService incoming.ValidationService
	ruleRegistry      *domain.RuleRegistry

	// Factories
	outgoingFactory *OutgoingAdapterFactory
	incomingFactory *IncomingAdapterFactory
}

// NewRoot creates a new composition root.
func NewRoot() *Root {
	return &Root{}
}

// Initialize initializes all components and wires them together.
func (r *Root) Initialize(ctx context.Context) error {
	// Initialize core components
	if err := r.initializeCore(ctx); err != nil {
		return fmt.Errorf("core initialization failed: %w", err)
	}

	// Create factories
	r.createFactories(ctx)

	// Initialize adapters
	if err := r.initializeAdapters(ctx); err != nil {
		return fmt.Errorf("adapter initialization failed: %w", err)
	}

	// Initialize application services
	r.initializeApplicationServices(ctx)

	return nil
}

// Getters for dependencies.

func (r *Root) GetConfig() types.Config {
	return r.config
}

func (r *Root) GetLogger() outgoing.Logger {
	// Logger should be retrieved from context
	// This method is a placeholder for backward compatibility
	return nil
}

func (r *Root) GetGitRepository() domain.CommitRepository {
	return r.gitRepository
}

func (r *Root) GetValidationService() incoming.ValidationService {
	return r.validationService
}

func (r *Root) GetRuleRegistry() *domain.RuleRegistry {
	return r.ruleRegistry
}

// GetCLIDependencies returns the dependencies needed by the CLI.
func (r *Root) GetCLIDependencies() *cli.Dependencies {
	return r.incomingFactory.CreateCLIDependencies(
		r.validationService,
		r.gitRepository,
	)
}

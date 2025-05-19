// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"

	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/ports/incoming"
)

// Dependencies holds all the dependencies required by the CLI.
type Dependencies struct {
	// ValidationService provides validation of Git commits
	ValidationService incoming.ValidationService

	// GitRepository provides access to Git commits
	GitRepository domain.CommitRepository

	// Config provides application configuration
	Config types.Config
}

// contextKey is a custom type for context keys to avoid collisions.
type dependenciesKey struct{}

// WithDependencies adds dependencies to the context.
func WithDependencies(ctx context.Context, deps interface{}) context.Context {
	if dependencies, ok := deps.(*Dependencies); ok {
		return context.WithValue(ctx, dependenciesKey{}, dependencies)
	}

	return ctx
}

// GetDependencies retrieves dependencies from the context.
func GetDependencies(ctx context.Context) *Dependencies {
	if ctx == nil {
		return nil
	}

	if deps, ok := ctx.Value(dependenciesKey{}).(*Dependencies); ok {
		return deps
	}

	return nil
}

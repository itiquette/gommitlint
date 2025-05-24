// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package rules provides test helpers for testing rule components.
package rules

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
)

// CreateTestContext creates an isolated test context.
// Each test should use a fresh context to prevent state leakage between tests.
func CreateTestContext(t *testing.T) context.Context {
	t.Helper()

	return context.Background()
}

// CreateTestConfigAdapter creates a test configuration adapter with the specified configuration.
// This follows the recommended pattern of accessing configuration directly
// rather than embedding it in context.
func CreateTestConfigAdapter(t *testing.T, configModifier func(types.Config) types.Config) *config.Adapter {
	t.Helper()

	// Create base configuration
	baseConfig := types.Config{}

	// Apply modifier if provided
	if configModifier != nil {
		baseConfig = configModifier(baseConfig)
	}

	// Create and return the adapter
	return config.NewAdapter(baseConfig)
}

// NOTE: CreateTestContextWithConfig has been removed.
// Use CreateTestConfigAdapter instead and pass the configuration directly
// to functions that need it, rather than embedding in context.

// CreateTestCommit creates a commit for testing with specified attributes.
func CreateTestCommit(hash, signature string) domain.CommitInfo {
	return domain.CommitInfo{
		Hash:      hash,
		Signature: signature,
	}
}

// StandardSignatures provides test signature examples for common scenarios.
var StandardSignatures = struct {
	GPG string
	SSH string
}{
	GPG: `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2

iQEcBAABCAAGBQJkglfUAQoJECXBz7a2zOr65JcIAK+ghUvxzS4DZBnY1t2+8JLk
xDrX6MT7MHuJcOd05yVgGJdyj2o73Qlxc/lGlQe/i+3R9cwTAqxY+GOKVhzM+QtT
3qfyQlKUwlwN3fPlgwZZbCgXUtQ+/71zLFEMmU7QtgRBUYKi4LVD5TuWUxQOLHbm
I5AXZIZOUmODdOVEwQT13mDV7FP/Mt+9ZKidq7YP3jUvlf7i9OCqzGsXQGa6FKs/
zC5FfCXR0dNrHz2V9IPhJsV8bxNTSZ5dRELaFa3mc0ew0mYriZz1LgaGu6Km3mJv
0mZlH6y9MWy9lx5FhAOA8b8EFWIDdHaDu4F0ZMUJtZx9/G0QEY1U6P/iuR8=
=QLiW
-----END PGP SIGNATURE-----`,

	SSH: `-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAgm5CqhCBSQDCbXJc94GFxy+9EYhYt
0zn/S58cPKoFNRMAAAAHZ2l0LXNzaAAAAACzAAAAMAAAAAtzdWJqZWN0CUFkZCBmZWF0dXJl
LwAAACRrZXktMS11c2VyQHVuaXQuZXhhbXBsZQAAAAAAAAAAAAAAAA==
-----END SSH SIGNATURE-----`,
}

// NOTE: The WithTestKeyDirectory function has been moved to identity_test_helpers.go
// to provide a cleaner implementation using the new WithTest* methods.

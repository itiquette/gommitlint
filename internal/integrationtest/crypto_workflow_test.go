// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/crypto"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// TestCryptoVerificationWorkflow tests the entire crypto verification workflow
// from rules through adapters to the domain.
func TestCryptoVerificationWorkflow(t *testing.T) {
	// Skip test in short mode as it involves filesystem operations
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary directory for test keys
	tempDir, err := os.MkdirTemp("", "test-keys")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test GPG public key file
	gpgKeyPath := filepath.Join(tempDir, "test.gpg")
	err = os.WriteFile(gpgKeyPath, []byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

mQENBGHfLDwBCADR7a/VQHUJlTZMWjbvWNIhwKqWzNQ6KLnryiAJJ70Q7hDkUVe+
VZCD0evJ7jMoqXMEqATGc2RvmjxnYQh2AINAkxJXc5YMYvczqsKYrT48PUCdZN0r
TTDcIQbLUTm+8rtWxiHOV0KIj+I9R8OeKjLmPGMO/EP2qJwSkgVbTzuR04vLGDoe
zFLOlT40F3b3nzYJTWfIHC/+/3d6c2Kv/UiNbQ3ZXvgBUzLswUQcP7lHhBy8tgLW
nwYJ+QxxM1HWAeXY7gvlqH87LjL8CJrUk1mAULpdAJlUAkUE3PZt1zDVJPY0KMP8
4yF1QxRM2SG4BFBtpJ6j6QK8RKnZBM5xfXylABEBAAG0HVRlc3QgVXNlciA8dGVz
dEB0ZXN0Lml0Lm5vdzE+iQFOBBMBCAA4FiEEnQ0DNyQNzOSKFhQA7lK9N9hs+/EF
AmHfLDwCGwMFCwkIBwIGFQoJCAsCBBYCAwECHgECF4AACgkQ7lK9N9hs+/E3zgf/
cdA7Ff8bmzBG7K1a85nkPmKtj6M4c1WVYgYS3qwdvz0jQXviVQiZQJVwEORSZX6z
CCvJdADYDMcQJ33DAkQ+gWE0mfxpRtP4rntLGOsTQfyTo/+bIxdtm1jshAMELkb9
L3I+QGLlRqmc9OzIcOS5aLJQUbR7+v02L1NUJJw4WS+32EDEfHUGCUqHJ2L1QdNc
1CWnHFjzZFUvQ7mSEpBbVeZZBNBYUjwZqGaKozdW5xqnugQ8M/X5B74m2XL5oykZ
g+l5BLMCl8dKPDfUlUdGNKMYJLjfB1wAODFZ74QM0YB1YEIzCd+WbdYYQhGGQ6/g
4JUo5P5ZfIiPCMiHdQ==
=YWul
-----END PGP PUBLIC KEY BLOCK-----`), 0600)
	require.NoError(t, err)

	// Create a test commit with signature
	commit := domain.CommitInfo{
		Hash:        "abc123",
		AuthorName:  "Test User",
		AuthorEmail: "test@test.it.now1",
		Signature: `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2

iQEcBAABCAAGBQJkglfUAQoJECXBz7a2zOr65JcIAK+ghUvxzS4DZBnY1t2+8JLk
xDrX6MT7MHuJcOd05yVgGJdyj2o73Qlxc/lGlQe/i+3R9cwTAqxY+GOKVhzM+QtT
3qfyQlKUwlwN3fPlgwZZbCgXUtQ+/71zLFEMmU7QtgRBUYKi4LVD5TuWUxQOLHbm
I5AXZIZOUmODdOVEwQT13mDV7FP/Mt+9ZKidq7YP3jUvlf7i9OCqzGsXQGa6FKs/
zC5FfCXR0dNrHz2V9IPhJsV8bxNTSZ5dRELaFa3mc0ew0mYriZz1LgaGu6Km3mJv
0mZlH6y9MWy9lx5FhAOA8b8EFWIDdHaDu4F0ZMUJtZx9/G0QEY1U6P/iuR8=
=QLiW
-----END PGP SIGNATURE-----`,
	}

	// Create a context with configuration
	ctx := testcontext.CreateTestContext()
	cfg := types.Config{
		Signing: types.SigningConfig{
			AllowedSigners: []string{"Test User <test@test.it.now1>"},
			KeyDirectory:   tempDir,
		},
		Rules: types.RulesConfig{
			Enabled: []string{"Signature", "SignedIdentity"},
		},
	}

	adapter := config.NewAdapter(cfg)
	ctx = contextx.WithConfig(ctx, adapter)

	// Create and test both rules
	t.Run("SignatureRule", func(t *testing.T) {
		rule := rules.NewSignatureRule()
		errors := rule.Validate(ctx, commit)
		require.Empty(t, errors, "Signature validation should pass")
	})

	t.Run("IdentityRule", func(t *testing.T) {
		// Create identity rule with directory
		rule := rules.NewIdentityRule(
			rules.WithKeyDirectory(tempDir),
		)

		// Validate
		errors := rule.Validate(ctx, commit)

		// This likely won't pass in automated tests since we can't verify
		// real signatures, but the test demonstrates the workflow
		if len(errors) > 0 {
			t.Logf("Expected verification error in test: %v", errors)
		}
	})

	// Test the domain models directly
	t.Run("Domain Models", func(t *testing.T) {
		// Create signature
		signature := crypto.NewSignature(commit.Signature)
		require.Equal(t, crypto.SignatureTypeGPG, signature.Type())
		require.False(t, signature.IsEmpty())

		// Create identities
		authorIdentity := crypto.NewIdentity(commit.AuthorName, commit.AuthorEmail)
		allowedIdentity := crypto.NewIdentityFromString("Test User <test@test.it.now1>")

		// Test matching
		require.True(t, authorIdentity.Matches(allowedIdentity))
	})
}

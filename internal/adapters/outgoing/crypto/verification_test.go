// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	domainCrypto "github.com/itiquette/gommitlint/internal/domain/crypto"
	testCrypto "github.com/itiquette/gommitlint/internal/testutils/crypto"
	"github.com/stretchr/testify/require"
)

// TestVerificationAdapter tests the crypto verification adapter.
func TestVerificationAdapter(t *testing.T) {
	// Create a repository and an adapter
	keyRepo := crypto.NewFileSystemKeyRepository("/tmp")
	adapter := crypto.NewVerificationAdapter(keyRepo)

	// Test with a commit with no signature
	commit := domain.CommitInfo{
		Hash:        "1234567890abcdef",
		Message:     "Test commit",
		AuthorName:  "Test Author",
		AuthorEmail: "author@example.com",
		Signature:   "",
	}

	// Create a context without config
	ctx := context.Background()
	result, err := adapter.VerifyCommit(ctx, commit)
	require.NoError(t, err)
	require.Equal(t, domainCrypto.VerificationStatusFailed, result.Status())
	require.Equal(t, "missing_signature", result.ErrorCode())
	require.Contains(t, result.ErrorMessage(), "no signature")

	// Test with invalid GPG signature
	commit.Signature = "-----BEGIN PGP SIGNATURE-----\nInvalid signature\n-----END PGP SIGNATURE-----"
	result, err = adapter.VerifyCommit(ctx, commit)
	require.NoError(t, err)
	// The status could be either "failed" or "no_key" depending on environment
	require.Contains(t, []domainCrypto.VerificationStatus{
		domainCrypto.VerificationStatusFailed,
		domainCrypto.VerificationStatusNoKey,
	}, result.Status(), "Expected either failed or no_key status")

	// Test with context that has config
	cfg := types.Config{
		Signing: types.SigningConfig{
			KeyDirectory: "/custom/keys",
		},
	}
	// Create config adapter
	configAdapter := config.NewAdapter(cfg)
	ctxWithConfig := contextx.WithConfig(ctx, configAdapter)
	result, err = adapter.VerifyCommit(ctxWithConfig, commit)
	require.NoError(t, err)
	// Status could vary depending on environment
	require.Contains(t, []domainCrypto.VerificationStatus{
		domainCrypto.VerificationStatusFailed,
		domainCrypto.VerificationStatusNoKey,
		domainCrypto.VerificationStatusUnsupported,
	}, result.Status(), "Expected a valid verification status")
}

func TestGetCommitBytes(t *testing.T) {
	// This is a placeholder test since the function is also a placeholder
	// In a real implementation, this would test with actual git commits
	result, err := testCrypto.GetCommitBytes(nil)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "nil commit")
}

func TestWithConfiguration(t *testing.T) {
	// Create a repository and an adapter
	keyRepo := crypto.NewFileSystemKeyRepository("/tmp")
	adapter := crypto.NewVerificationAdapter(keyRepo)

	// Create a commit with no signature (just for testing context usage)
	commit := domain.CommitInfo{
		Hash:        "1234567890abcdef",
		Message:     "Test commit",
		AuthorName:  "Test Author",
		AuthorEmail: "author@example.com",
		Signature:   "",
	}

	// Create a configuration with different key directories
	configs := []struct {
		name          string
		keyDirectory  string
		hasConfigPath bool
	}{
		{
			name:          "Empty config",
			keyDirectory:  "",
			hasConfigPath: false,
		},
		{
			name:          "Custom directory",
			keyDirectory:  "/custom/keys",
			hasConfigPath: true,
		},
		{
			name:          "Absolute path",
			keyDirectory:  "/tmp/keys",
			hasConfigPath: true,
		},
		{
			name:          "Relative path",
			keyDirectory:  "./keys",
			hasConfigPath: true,
		},
	}

	// Test each configuration
	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			// Create a configuration
			configObj := types.Config{}

			if cfg.hasConfigPath {
				configObj.Signing.KeyDirectory = cfg.keyDirectory
			}

			// Create config adapter
			configAdapter := config.NewAdapter(configObj)
			ctx := contextx.WithConfig(context.Background(), configAdapter)

			// We can't easily verify which directory was used internally,
			// but we can at least ensure the function completes without error
			_, err := adapter.VerifyCommit(ctx, commit)
			require.NoError(t, err)
		})
	}
}

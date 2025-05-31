// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

// TestVerifier tests the crypto verification adapter.
func TestVerifier(t *testing.T) {
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
	require.Equal(t, domain.VerificationStatusFailed, result.Status())
	require.Equal(t, "missing_signature", result.ErrorCode())
	require.Contains(t, result.ErrorMessage(), "no signature")

	// Test with invalid GPG signature
	commit.Signature = "-----BEGIN PGP SIGNATURE-----\nInvalid signature\n-----END PGP SIGNATURE-----"
	result, err = adapter.VerifyCommit(ctx, commit)
	require.NoError(t, err)
	// The status could be either "failed" or "no_key" depending on environment
	require.Contains(t, []domain.VerificationStatus{
		domain.VerificationStatusFailed,
		domain.VerificationStatusNoKey,
	}, result.Status(), "Expected either failed or no_key status")
}

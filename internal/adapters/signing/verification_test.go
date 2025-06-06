// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signing_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

// TestVerifier tests the simplified crypto verification adapter.
func TestVerifier(t *testing.T) {
	// Create a simplified adapter - no complex dependencies
	adapter := signing.NewVerificationAdapter()

	// Test with a commit with no signature
	commit := domain.Commit{
		Hash:        "1234567890abcdef",
		Message:     "Test commit",
		Author:      "Test Author",
		AuthorEmail: "author@example.com",
		Signature:   "",
	}

	// Create a context and test with explicit keyDir parameter
	ctx := context.Background()
	keyDir := "/tmp"
	result := adapter.VerifyCommit(ctx, commit, keyDir)
	require.Equal(t, domain.VerificationStatusFailed, result.Status())
	require.Equal(t, "missing_signature", result.ErrorCode())
	require.Contains(t, result.ErrorMessage(), "no signature")

	// Test with invalid GPG signature
	commit.Signature = "-----BEGIN PGP SIGNATURE-----\nInvalid signature\n-----END PGP SIGNATURE-----"
	result = adapter.VerifyCommit(ctx, commit, keyDir)
	// The status could be either "failed" or "no_key" depending on environment
	require.Contains(t, []domain.VerificationStatus{
		domain.VerificationStatusFailed,
		domain.VerificationStatusNoKey,
	}, result.Status(), "Expected either failed or no_key status")
}

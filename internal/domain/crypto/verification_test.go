// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain/crypto"
	"github.com/stretchr/testify/require"
)

func TestVerificationResult(t *testing.T) {
	identity := crypto.NewIdentity("John Doe", "john@example.com")
	signature := crypto.NewSignature("test-signature")

	// Test successful verification
	result := crypto.NewVerificationResult(
		crypto.VerificationStatusVerified,
		identity,
		signature,
	)

	require.Equal(t, crypto.VerificationStatusVerified, result.Status())
	require.Equal(t, identity, result.Identity())
	require.Equal(t, signature, result.Signature())
	require.True(t, result.IsVerified())
	require.False(t, result.HasError())

	// Test failed verification with error
	failedResult := result.WithError("test_error", "Test error message")

	require.Equal(t, crypto.VerificationStatusVerified, failedResult.Status(), "Status should not change")
	require.Equal(t, "test_error", failedResult.ErrorCode())
	require.Equal(t, "Test error message", failedResult.ErrorMessage())
	require.True(t, failedResult.HasError())

	// Test different statuses
	failedVerification := crypto.NewVerificationResult(
		crypto.VerificationStatusFailed,
		identity,
		signature,
	)
	require.Equal(t, crypto.VerificationStatusFailed, failedVerification.Status())
	require.False(t, failedVerification.IsVerified())

	noKeyVerification := crypto.NewVerificationResult(
		crypto.VerificationStatusNoKey,
		identity,
		signature,
	)
	require.Equal(t, crypto.VerificationStatusNoKey, noKeyVerification.Status())
	require.False(t, noKeyVerification.IsVerified())

	unsupportedVerification := crypto.NewVerificationResult(
		crypto.VerificationStatusUnsupported,
		identity,
		signature,
	)
	require.Equal(t, crypto.VerificationStatusUnsupported, unsupportedVerification.Status())
	require.False(t, unsupportedVerification.IsVerified())
}

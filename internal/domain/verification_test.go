// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestVerificationResult(t *testing.T) {
	identity := domain.NewIdentity("John Doe", "john@example.com")
	signature := domain.NewSignature("test-signature")

	// Test successful verification
	result := domain.NewVerificationResult(
		domain.VerificationStatusVerified,
		identity,
		signature,
	)

	require.Equal(t, domain.VerificationStatusVerified, result.Status())
	require.Equal(t, identity, result.Identity())
	require.Equal(t, signature, result.Signature())
	require.True(t, result.IsVerified())
	require.False(t, result.HasError())

	// Test failed verification with error
	failedResult := result.WithError("test_error", "Test error message")

	require.Equal(t, domain.VerificationStatusVerified, failedResult.Status(), "Status should not change")
	require.Equal(t, "test_error", failedResult.ErrorCode())
	require.Equal(t, "Test error message", failedResult.ErrorMessage())
	require.True(t, failedResult.HasError())

	// Test different statuses
	failedVerification := domain.NewVerificationResult(
		domain.VerificationStatusFailed,
		identity,
		signature,
	)
	require.Equal(t, domain.VerificationStatusFailed, failedVerification.Status())
	require.False(t, failedVerification.IsVerified())

	noKeyVerification := domain.NewVerificationResult(
		domain.VerificationStatusNoKey,
		identity,
		signature,
	)
	require.Equal(t, domain.VerificationStatusNoKey, noKeyVerification.Status())
	require.False(t, noKeyVerification.IsVerified())

	unsupportedVerification := domain.NewVerificationResult(
		domain.VerificationStatusUnsupported,
		identity,
		signature,
	)
	require.Equal(t, domain.VerificationStatusUnsupported, unsupportedVerification.Status())
	require.False(t, unsupportedVerification.IsVerified())
}

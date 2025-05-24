// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto

import (
	"context"
	"fmt"

	coreCrypto "github.com/itiquette/gommitlint/internal/core/crypto"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/crypto"
)

// VerificationAdapter provides crypto verification services for the application.
type VerificationAdapter struct {
	service    coreCrypto.VerificationService
	repository domain.CryptoKeyRepository
	keyDir     string
}

// NewVerificationAdapter creates a new adapter with the default verifiers.
// For more flexibility, use NewVerificationAdapterWithOptions instead.
func NewVerificationAdapter(keyRepository domain.CryptoKeyRepository) *VerificationAdapter {
	return NewVerificationAdapterWithOptions(WithKeyRepository(keyRepository))
}

// VerifyCommit verifies the signature on a commit.
func (a VerificationAdapter) VerifyCommit(_ context.Context, commit domain.CommitInfo) (crypto.VerificationResult, error) {
	// Create a signature from the commit
	signature := crypto.NewSignature(commit.Signature)

	// Skip if there's no signature
	if signature.IsEmpty() {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusFailed,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("missing_signature", "Commit has no signature"), nil
	}

	// Prepare commit data for verification
	// In a real implementation, this would extract the actual commit data
	// For now, we'll just use a placeholder
	commitData := []byte(fmt.Sprintf("commit %s by %s", commit.Hash, commit.AuthorEmail))

	// Verify the signature
	result := a.service.Verify(signature, commitData, a.keyDir)

	return result, nil
}

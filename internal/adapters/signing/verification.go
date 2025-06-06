// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signing

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
)

// VerificationAdapter provides crypto verification services.
// It implements the domain.SignatureVerifier interface directly without complex patterns.
type VerificationAdapter struct{}

// Ensure VerificationAdapter implements SignatureVerifier interface.
var _ domain.SignatureVerifier = (*VerificationAdapter)(nil)

// NewVerificationAdapter creates a new verification adapter.
func NewVerificationAdapter() *VerificationAdapter {
	return &VerificationAdapter{}
}

// VerifyCommit implements the domain.SignatureVerifier interface.
func (a *VerificationAdapter) VerifyCommit(ctx context.Context, commit domain.Commit, keyDir string) domain.VerificationResult {
	return VerifyCommit(ctx, commit, keyDir)
}

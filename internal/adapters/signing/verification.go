// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
)

// VerificationAdapter provides crypto verification services using simplified interface.
// It implements the domain.SignatureVerifier interface directly without complex factory patterns.
type VerificationAdapter struct {
	verifier *SimpleVerifier
}

// Ensure VerificationAdapter implements SignatureVerifier interface.
var _ domain.SignatureVerifier = (*VerificationAdapter)(nil)

// NewVerificationAdapter creates a new simplified verification adapter.
// No complex options or dependencies - pure functional approach.
func NewVerificationAdapter() *VerificationAdapter {
	return &VerificationAdapter{
		verifier: NewSimpleVerifier(),
	}
}

// VerifyCommit implements the domain.SignatureVerifier interface.
// All dependencies are passed as explicit parameters following functional principles.
func (a *VerificationAdapter) VerifyCommit(ctx context.Context, commit domain.Commit, keyDir string) domain.VerificationResult {
	return a.verifier.VerifyCommit(ctx, commit, keyDir)
}

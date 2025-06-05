// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
)

// SignatureVerifier defines the single cohesive interface for signature verification.
// This replaces the previous fragmented CryptoVerifier, CryptoKeyRepository, and Verifier interfaces.
// Following functional hexagonal principles, all dependencies are passed as explicit parameters.
type SignatureVerifier interface {
	// VerifyCommit verifies a commit's signature and returns verification result.
	// keyDir parameter is passed explicitly instead of being injected via constructor.
	// This follows the functional pattern of explicit dependencies.
	VerifyCommit(ctx context.Context, commit Commit, keyDir string) VerificationResult
}

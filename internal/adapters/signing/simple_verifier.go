// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// internalVerifier defines the interface that internal verifiers must implement.
// This is kept internal to the crypto package and not exposed as a domain interface.
type internalVerifier interface {
	CanVerify(signature domain.Signature) bool
	Verify(signature domain.Signature, data []byte, keyDir string) domain.VerificationResult
}

// SimpleVerifier implements the simplified SignatureVerifier interface.
// It replaces the complex factory pattern with direct functional composition.
type SimpleVerifier struct {
	gpgVerifier internalVerifier
	sshVerifier internalVerifier
}

// Ensure SimpleVerifier implements SignatureVerifier interface.
var _ domain.SignatureVerifier = (*SimpleVerifier)(nil)

// NewSimpleVerifier creates a new simplified signature verifier.
// Uses functional composition instead of complex factory patterns.
func NewSimpleVerifier() *SimpleVerifier {
	return &SimpleVerifier{
		gpgVerifier: NewDefaultGPGVerificationService(),
		sshVerifier: NewDefaultSSHVerificationService(),
	}
}

// VerifyCommit implements the domain.SignatureVerifier interface.
// All dependencies are passed as explicit parameters following functional principles.
func (v *SimpleVerifier) VerifyCommit(_ context.Context, commit domain.Commit, keyDir string) domain.VerificationResult {
	// Create signature from commit
	signature := domain.NewSignature(commit.Signature)

	// Skip if there's no signature
	if signature.IsEmpty() {
		return domain.NewVerificationResult(
			domain.VerificationStatusFailed,
			domain.NewIdentity("", ""),
			signature,
		).WithError("missing_signature", "Commit has no signature")
	}

	// Prepare commit data for verification
	commitData := []byte(fmt.Sprintf("commit %s by %s", commit.Hash, commit.AuthorEmail))

	// Try GPG verification first
	if v.gpgVerifier.CanVerify(signature) {
		return v.gpgVerifier.Verify(signature, commitData, keyDir)
	}

	// Try SSH verification second
	if v.sshVerifier.CanVerify(signature) {
		return v.sshVerifier.Verify(signature, commitData, keyDir)
	}

	// No suitable verifier found
	return domain.NewVerificationResult(
		domain.VerificationStatusUnsupported,
		domain.NewIdentity("", ""),
		signature,
	).WithError("unsupported_signature", "No suitable verifier found for signature type")
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package crypto provides implementations of crypto domain interfaces
// for verifying cryptographic signatures in commits.
package crypto

import (
	"github.com/itiquette/gommitlint/internal/domain/crypto"
)

// VerifierFactory creates the appropriate verifier for a signature.
type VerifierFactory struct {
	verifiers []crypto.Verifier
}

// NewVerifierFactory creates a new verifier factory with the provided verifiers.
func NewVerifierFactory(verifiers ...crypto.Verifier) VerifierFactory {
	return VerifierFactory{
		verifiers: verifiers,
	}
}

// GetVerifier returns the appropriate verifier for the given signature.
// If no appropriate verifier is found, it returns nil.
func (f VerifierFactory) GetVerifier(signature crypto.Signature) crypto.Verifier {
	for _, verifier := range f.verifiers {
		if verifier.CanVerify(signature) {
			return verifier
		}
	}

	return nil
}

// VerificationService provides high-level signature verification.
type VerificationService struct {
	factory VerifierFactory
}

// NewVerificationService creates a new verification service with the provided verifiers.
func NewVerificationService(verifiers ...crypto.Verifier) VerificationService {
	return VerificationService{
		factory: NewVerifierFactory(verifiers...),
	}
}

// Verify verifies a signature with the appropriate verifier.
func (s VerificationService) Verify(signature crypto.Signature, data []byte, keyDir string) crypto.VerificationResult {
	verifier := s.factory.GetVerifier(signature)

	if verifier == nil {
		// No suitable verifier found
		return crypto.NewVerificationResult(
			crypto.VerificationStatusUnsupported,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("unsupported_signature", "No suitable verifier found for signature type")
	}

	return verifier.Verify(signature, data, keyDir)
}

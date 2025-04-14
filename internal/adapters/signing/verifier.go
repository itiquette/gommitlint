// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signing

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// VerifyCommit implements signature verification for commit messages.
func VerifyCommit(_ context.Context, commit domain.Commit, keyDir string) domain.VerificationResult {
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
	if CanVerifyGPG(signature) {
		return VerifyGPGSignature(signature, commitData, keyDir, DefaultGPGSecuritySettings())
	}

	// Try SSH verification second
	if CanVerifySSH(signature) {
		return VerifySSHSignature(signature, commitData, keyDir, DefaultSSHSecuritySettings())
	}

	// No suitable verifier found
	return domain.NewVerificationResult(
		domain.VerificationStatusUnsupported,
		domain.NewIdentity("", ""),
		signature,
	).WithError("unsupported_signature", "No suitable verifier found for signature type")
}

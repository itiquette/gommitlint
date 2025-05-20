// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	coreCrypto "github.com/itiquette/gommitlint/internal/core/crypto"
	"github.com/itiquette/gommitlint/internal/core/crypto/gpg"
	"github.com/itiquette/gommitlint/internal/core/crypto/ssh"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/crypto"
)

// VerificationAdapter provides crypto verification services for the application.
type VerificationAdapter struct {
	service    coreCrypto.VerificationService
	repository KeyRepository
	defaultDir string
}

// NewVerificationAdapter creates a new adapter with the default verifiers.
func NewVerificationAdapter(keyRepository KeyRepository) *VerificationAdapter {
	// Create verifiers with default settings
	gpgVerifier := gpg.NewDefaultVerifier()
	sshVerifier := ssh.NewDefaultVerifier()

	// Create verification service with all verifiers
	service := coreCrypto.NewVerificationService(gpgVerifier, sshVerifier)

	return &VerificationAdapter{
		service:    service,
		repository: keyRepository,
		defaultDir: keyRepository.GetKeyDirectory(),
	}
}

// VerifyCommit verifies the signature on a commit.
func (a *VerificationAdapter) VerifyCommit(ctx context.Context, commit domain.CommitInfo) (crypto.VerificationResult, error) {
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

	// Get key directory from context or use default
	keyDir := a.getKeyDirectoryFromContext(ctx)

	// Prepare commit data for verification
	// In a real implementation, this would extract the actual commit data
	// For now, we'll just use a placeholder
	commitData := []byte(fmt.Sprintf("commit %s by %s", commit.Hash, commit.AuthorEmail))

	// Verify the signature
	result := a.service.Verify(signature, commitData, keyDir)

	return result, nil
}

// getKeyDirectoryFromContext gets the key directory from the context configuration.
func (a *VerificationAdapter) getKeyDirectoryFromContext(ctx context.Context) string {
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return a.defaultDir
	}

	// Try to get key directory from config
	keyDir := cfg.GetString("signing.key_directory")
	if keyDir == "" {
		return a.defaultDir
	}

	return keyDir
}

// GetCommitBytes extracts the byte representation of a commit for signature verification.
// This is typically needed as the input for signature verification.
func GetCommitBytes(commit *object.Commit) ([]byte, error) {
	if commit == nil {
		return nil, errors.New("nil commit")
	}

	// In a real implementation, you would:
	// 1. Create a memory object
	// 2. Encode the commit without its signature
	// 3. Read the bytes back

	// For now, we'll return a placeholder
	return []byte(fmt.Sprintf("commit %s", commit.Hash)), nil
}

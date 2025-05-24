// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"

	domainCrypto "github.com/itiquette/gommitlint/internal/domain/crypto"
)

// CryptoVerifier defines the interface for signature verification operations.
type CryptoVerifier interface {
	VerifyCommit(ctx context.Context, commit CommitInfo) (domainCrypto.VerificationResult, error)
}

// CryptoKeyRepository defines the interface for key management operations.
type CryptoKeyRepository interface {
	GetKeyDirectory() string
	FindKeyFiles(extensions []string) ([]string, error)
	ReadKeyFile(path string) ([]byte, error)
}

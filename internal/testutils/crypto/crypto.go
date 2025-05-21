// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package crypto provides test utilities for crypto operations
package crypto

import (
	"os"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/common/security"
)

// WithRelaxedPermissions creates a FileSystemKeyRepository with relaxed
// security settings for testing purposes. This should only be used in tests.
func WithRelaxedPermissions(keyDir string) *crypto.FileSystemKeyRepository {
	// Create a test security checker that bypasses security checks
	securityChecker := &security.TestSecurityChecker{}

	// Use type assertion to convert to the required type
	return crypto.NewFileSystemKeyRepositoryWithOptions(
		keyDir,
		crypto.WithSecurityChecker(securityChecker),
	)
}

// TestKeyFilepath is a test helper to create key file paths safely.
func TestKeyFilepath(keyDir, username, keyType string) string {
	if keyDir == "" {
		tempDir, _ := os.MkdirTemp("", "gommitlint-test-keys")
		keyDir = tempDir
	}

	ext := ".pub"
	if keyType == "private" {
		ext = ""
	}

	return keyDir + "/" + username + ext
}

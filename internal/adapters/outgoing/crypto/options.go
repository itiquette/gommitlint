// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto

import (
	"github.com/itiquette/gommitlint/internal/common/config"
	coreCrypto "github.com/itiquette/gommitlint/internal/core/crypto"
	"github.com/itiquette/gommitlint/internal/core/crypto/gpg"
	"github.com/itiquette/gommitlint/internal/core/crypto/ssh"
)

// Option is a function that configures a VerificationAdapter.
type Option func(*VerificationAdapter)

// WithVerificationService sets a custom verification service.
func WithVerificationService(service coreCrypto.VerificationService) Option {
	return func(a *VerificationAdapter) {
		a.service = service
	}
}

// WithKeyRepository sets a custom key repository.
func WithKeyRepository(repository KeyRepository) Option {
	return func(a *VerificationAdapter) {
		a.repository = repository
		a.defaultDir = repository.GetKeyDirectory()
	}
}

// WithConfiguration applies configuration to the adapter.
// It makes it easy to configure the adapter from a config object.
func WithConfiguration(cfg config.Config) Option {
	return func(adapter *VerificationAdapter) {
		// Get key directory from configuration
		keyDir := config.ResolvePath(cfg, "signing.key_directory", adapter.defaultDir)

		// Create a new repository with the configured key directory
		adapter.repository = NewFileSystemKeyRepository(keyDir)
		adapter.defaultDir = keyDir
	}
}

// NewVerificationAdapterWithOptions creates a new adapter with the given options.
// This is the primary entry point for creating verification adapters.
func NewVerificationAdapterWithOptions(options ...Option) *VerificationAdapter {
	// Create default verifiers
	gpgVerifier := gpg.NewDefaultVerifier()
	sshVerifier := ssh.NewDefaultVerifier()

	// Create verification service with all verifiers
	service := coreCrypto.NewVerificationService(gpgVerifier, sshVerifier)

	// Create default key repository
	repository := NewFileSystemKeyRepository("")

	// Create adapter with defaults
	adapter := &VerificationAdapter{
		service:    service,
		repository: repository,
		defaultDir: repository.GetKeyDirectory(),
	}

	// Apply options
	for _, option := range options {
		option(adapter)
	}

	return adapter
}

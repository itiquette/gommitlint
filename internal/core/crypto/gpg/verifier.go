// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package gpg provides GPG-specific signature verification.
package gpg

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/itiquette/gommitlint/internal/common/fsutils"
	"github.com/itiquette/gommitlint/internal/domain/crypto"
)

// SecuritySettings defines security requirements for GPG keys.
type SecuritySettings struct {
	MinimumRSABits uint16
	MinimumECBits  uint16
}

// DefaultSecuritySettings provides reasonable default security settings.
func DefaultSecuritySettings() SecuritySettings {
	return SecuritySettings{
		MinimumRSABits: 2048,
		MinimumECBits:  256,
	}
}

// Verifier implements the crypto.Verifier interface for GPG signatures.
type Verifier struct {
	settings SecuritySettings
}

// NewVerifier creates a new GPG verifier with the given security settings.
func NewVerifier(settings SecuritySettings) Verifier {
	return Verifier{
		settings: settings,
	}
}

// NewDefaultVerifier creates a new GPG verifier with default security settings.
func NewDefaultVerifier() Verifier {
	return NewVerifier(DefaultSecuritySettings())
}

// CanVerify checks if this verifier can handle the given signature.
func (v Verifier) CanVerify(signature crypto.Signature) bool {
	return signature.Type() == crypto.SignatureTypeGPG
}

// Verify checks if the GPG signature is valid for the given data.
func (v Verifier) Verify(signature crypto.Signature, data []byte, keyDir string) crypto.VerificationResult {
	if signature.IsEmpty() {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusFailed,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("empty_signature", "GPG signature is empty")
	}

	// Sanitize key directory path
	sanitizedKeyDir, err := fsutils.SanitizePath(keyDir)
	if err != nil {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusFailed,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("invalid_key_dir", fmt.Sprintf("Invalid key directory: %s", err))
	}

	// Find GPG key files
	keyFiles, err := fsutils.FindFilesWithExtensions(sanitizedKeyDir, []string{".gpg", ".asc", ".pub"})
	if err != nil {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusFailed,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("key_dir_error", fmt.Sprintf("Failed to find GPG keys: %s", err))
	}

	if len(keyFiles) == 0 {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusNoKey,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("no_keys", "No GPG key files found in "+keyDir)
	}

	// Try each key file
	for _, keyFile := range keyFiles {
		entities, err := loadGPGKey(keyFile)
		if err != nil {
			continue // Skip invalid keys
		}

		// Try each key in the file
		for _, entity := range entities {
			// Skip invalid keys
			if isKeyRevoked(entity) || isKeyExpired(entity, time.Now()) || !hasMinimumKeyStrength(entity, v.settings) {
				continue
			}

			// Verify signature
			dataReader := strings.NewReader(string(data))
			sigReader := strings.NewReader(signature.Data())

			verifiedEntity, err := openpgp.CheckArmoredDetachedSignature(
				openpgp.EntityList{entity},
				dataReader,
				sigReader,
				nil,
			)

			if err == nil && verifiedEntity != nil {
				// Found a matching key
				identity := extractIdentity(verifiedEntity)

				return crypto.NewVerificationResult(
					crypto.VerificationStatusVerified,
					identity,
					signature,
				)
			}
		}
	}

	// If we get here, no keys matched
	return crypto.NewVerificationResult(
		crypto.VerificationStatusFailed,
		crypto.NewIdentity("", ""),
		signature,
	).WithError("verification_failed", "GPG signature not verified with any trusted key")
}

// loadGPGKey loads a GPG key from a file.
func loadGPGKey(path string) ([]*openpgp.Entity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read GPG key file: %w", err)
	}

	// Try armored format first
	entities, err := openpgp.ReadArmoredKeyRing(strings.NewReader(string(data)))
	if err == nil {
		return entities, nil
	}

	// Fall back to binary format
	return openpgp.ReadKeyRing(strings.NewReader(string(data)))
}

// isKeyRevoked checks if a GPG key has been revoked.
func isKeyRevoked(entity *openpgp.Entity) bool {
	// Check direct key revocations
	for _, sig := range entity.Revocations {
		if sig.RevocationReason != nil {
			return true
		}
	}

	// Check identity revocations
	for _, id := range entity.Identities {
		for _, sig := range id.Signatures {
			if sig.RevocationReason != nil {
				return true
			}
		}
	}

	return false
}

// isKeyExpired checks if a GPG key has expired.
func isKeyExpired(entity *openpgp.Entity, now time.Time) bool {
	// Check primary key expiration first
	for _, ident := range entity.Identities {
		if ident.SelfSignature != nil && ident.SelfSignature.KeyLifetimeSecs != nil {
			expiry := ident.SelfSignature.CreationTime.Add(time.Duration(*ident.SelfSignature.KeyLifetimeSecs) * time.Second)
			if now.After(expiry) {
				return true // Primary key is expired
			}

			break
		}
	}

	// If we're checking a signature, we need at least one valid signing key
	// Check for an unexpired signing subkey
	for _, subkey := range entity.Subkeys {
		// Only check subkeys that can sign
		if subkey.Sig != nil && subkey.Sig.FlagsValid && subkey.Sig.FlagSign {
			if subkey.Sig.KeyLifetimeSecs != nil {
				expiry := subkey.Sig.CreationTime.Add(time.Duration(*subkey.Sig.KeyLifetimeSecs) * time.Second)
				if !now.After(expiry) {
					return false // Found a valid signing subkey
				}
			} else {
				return false // Found a signing subkey with no expiration
			}
		}
	}

	// No valid signing subkeys found, but primary key is valid
	// This is fine if the primary key can sign
	return false
}

// hasMinimumKeyStrength checks if a GPG key meets minimum strength requirements.
func hasMinimumKeyStrength(entity *openpgp.Entity, settings SecuritySettings) bool {
	// Check RSA keys against minimum bit length
	if entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoRSA ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoRSAEncryptOnly ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoRSASignOnly {
		bitLength, err := entity.PrimaryKey.BitLength()
		if err != nil {
			return false // If we can't determine bit length, reject for safety
		}

		return bitLength >= settings.MinimumRSABits
	}

	// For EC keys
	if entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoECDSA ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoEdDSA ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoECDH {
		// Try to get bit length directly
		bitLength, err := entity.PrimaryKey.BitLength()
		if err == nil {
			return bitLength >= settings.MinimumECBits
		}

		// If BitLength() failed, fall back to algorithm-specific checks
		if entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoEdDSA {
			return 256 >= settings.MinimumECBits // Ed25519 is always 256 bits
		}

		// For other EC types without bit length info, assume minimum standards
		// This is a conservative approach
		return false
	}

	// Reject any other algorithms
	return false
}

// extractIdentity extracts a crypto.Identity from an OpenPGP entity.
func extractIdentity(entity *openpgp.Entity) crypto.Identity {
	if entity == nil {
		return crypto.NewIdentity("", "")
	}

	// Get the first identity from the entity
	for name := range entity.Identities {
		// Usually in format "Name <email@example.com>"
		return crypto.NewIdentityFromString(name)
	}

	// Fallback: Use key ID if no identities
	keyID := entity.PrimaryKey.KeyIdString()

	return crypto.NewIdentity("GPG Key "+keyID, "")
}

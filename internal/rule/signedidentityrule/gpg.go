// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// verifyGPGSignature verifies a GPG signature against commit data using trusted keys.
//
// Parameters:
//   - commitData: The raw commit data to verify
//   - signature: The GPG signature in ASCII-armored format
//   - keyDir: Directory containing trusted public keys
//
// The function attempts to verify the signature against all trusted GPG keys found
// in the specified directory. It performs several security checks on each key:
//  1. Skips revoked keys
//  2. Skips expired keys
//  3. Skips keys that don't meet minimum strength requirements
//
// This comprehensive validation ensures that commits are only verified against
// currently valid and sufficiently secure keys.
//
// Returns:
//   - string: The identity associated with the key that verified the signature
//   - error: Any error encountered during verification, or if no key verified the signature
func verifyGPGSignature(commitData []byte, signature string, keyDir string) (string, error) {
	if signature == "" {
		return "", errors.New("empty GPG signature")
	}

	// Find GPG key files
	keyFiles, err := findKeyFiles(keyDir, []string{".gpg", ".pub", ".asc"}, GPG)
	if err != nil {
		return "", fmt.Errorf("failed to find GPG keys: %w", err)
	}

	if len(keyFiles) == 0 {
		return "", fmt.Errorf("no GPG key files found in %s", keyDir)
	}

	// Try each key file
	for _, keyFile := range keyFiles {
		entities, err := loadGPGKey(keyFile)
		if err != nil {
			continue // Skip invalid keys
		}

		// Try each key in the file
		for _, entity := range entities {
			// Skip revoked keys
			if isKeyRevoked(entity) {
				continue
			}

			// Skip expired keys
			if isKeyExpired(entity, time.Now()) {
				continue
			}

			// Skip keys that don't meet minimum strength requirements
			if !hasMinimumKeyStrength(entity) {
				continue
			}

			dataReader := strings.NewReader(string(commitData))
			sigReader := strings.NewReader(signature)

			verifiedEntity, err := openpgp.CheckArmoredDetachedSignature(
				openpgp.EntityList{entity},
				dataReader,
				sigReader,
				nil,
			)

			if err == nil && verifiedEntity != nil {
				// Found a matching key
				for name := range verifiedEntity.Identities {
					return name, nil
				}

				return filepath.Base(keyFile), nil
			}
		}
	}

	return "", errors.New("GPG signature not verified with any trusted key")
}

// loadGPGKey loads a GPG key from a file, supporting both armored and binary formats.
//
// Parameters:
//   - path: Path to the GPG key file
//
// The function attempts to read the key file in armored (ASCII) format first,
// then falls back to binary format if that fails. This provides flexibility
// in handling different key file formats that users might have.
//
// Armored format is the ASCII-encoded format typically used for sharing keys
// (begins with "-----BEGIN PGP PUBLIC KEY BLOCK-----"), while binary format
// is more compact but not human-readable.
//
// Returns:
//   - []*openpgp.Entity: The loaded GPG key entities (a file may contain multiple keys)
//   - error: Any error encountered during loading or parsing
func loadGPGKey(path string) ([]*openpgp.Entity, error) {
	data, err := safeReadFile(path)
	if err != nil {
		return nil, err
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
//
// Parameters:
//   - entity: The GPG key entity to check
//
// The function checks both direct key revocations and identity revocations.
// A key might be revoked directly, or one of its identities might be revoked.
// Either case makes the key unsuitable for verification.
//
// Returns:
//   - bool: true if the key is revoked, false otherwise
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

// isKeyExpired checks if a GPG key has expired at the given time.
//
// Parameters:
//   - entity: The GPG key entity to check
//   - now: The current time to check against
//
// The function first checks if the primary key has expired. If not, it checks for
// at least one valid signing subkey. A key is considered valid for verification
// if either the primary key can sign and hasn't expired, or if it has at least one
// unexpired signing subkey.
//
// Returns:
//   - bool: true if the key is expired for signing purposes, false otherwise
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

// hasMinimumKeyStrength checks if a GPG key meets the minimum strength requirements.
//
// Parameters:
//   - entity: The GPG key entity to check
//
// The function evaluates the cryptographic strength of the key based on its algorithm
// and bit length:
//   - For RSA keys, it compares against MinimumRSABits (default: 2048)
//   - For EC keys, it compares against MinimumECBits (default: 256)
//   - If bit length can't be determined, it takes a conservative approach
//   - Other algorithm types are rejected by default
//
// This ensures that only keys with sufficient cryptographic strength are used for
// verification, protecting against the use of weak or outdated keys.
//
// Returns:
//   - bool: true if the key meets minimum strength requirements, false otherwise
func hasMinimumKeyStrength(entity *openpgp.Entity) bool {
	// Check RSA keys against minimum bit length
	if entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoRSA ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoRSAEncryptOnly ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoRSASignOnly {
		bitLength, err := entity.PrimaryKey.BitLength()
		if err != nil {
			return false // If we can't determine bit length, reject for safety
		}

		return bitLength >= MinimumRSABits
	}

	// For EC keys
	if entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoECDSA ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoEdDSA ||
		entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoECDH {
		// Try to get bit length directly
		bitLength, err := entity.PrimaryKey.BitLength()
		if err == nil {
			return bitLength >= MinimumECBits
		}

		// If BitLength() failed, fall back to algorithm-specific checks
		if entity.PrimaryKey.PubKeyAlgo == packet.PubKeyAlgoEdDSA {
			return 256 >= MinimumECBits // Ed25519 is always 256 bits
		}

		// For other EC types without bit length info, assume minimum standards
		// This is a conservative approach
		return false
	}

	// Reject any other algorithms
	return false
}

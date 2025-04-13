// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// parseSSHSignature parses an SSH signature string into its components.
//
// Parameters:
//   - signature: The SSH signature string in the format "format:blob"
//
// The function splits the signature string into its format and blob components.
// SSH signatures are typically structured as "algorithm:base64-encoded-data".
// Common formats include "ssh-rsa", "ssh-ed25519", and "ecdsa-sha2-nistp256".
//
// Returns:
//   - format: The signature algorithm format (e.g., "ssh-rsa", "ssh-ed25519")
//   - blob: The decoded signature data as a byte slice
//   - error: Any error encountered during parsing
func parseSSHSignature(signature string) (string, []byte, error) {
	parts := strings.SplitN(signature, ":", 2)
	if len(parts) != 2 {
		return "", nil, errors.New("invalid SSH signature format, expected 'format:blob'")
	}

	format := parts[0]
	blobStr := parts[1]

	// Directly try to decode with standard base64 encoding
	blob, err := base64.StdEncoding.DecodeString(blobStr)
	if err != nil {
		return "", nil, fmt.Errorf("invalid SSH signature blob: %w", err)
	}

	if len(blob) == 0 {
		return "", nil, errors.New("empty SSH signature blob after decoding")
	}

	return format, blob, nil
}

// verifySSHSignature verifies an SSH signature against commit data using trusted keys.
//
// Parameters:
//   - commitData: The raw commit data to verify
//   - format: The signature algorithm format (e.g., "ssh-rsa")
//   - blob: The signature blob data
//   - keyDir: Directory containing trusted public keys
//
// The function attempts to verify the signature against all trusted SSH keys found
// in the specified directory. It checks that keys meet minimum strength requirements
// before attempting verification.
//
// Returns:
//   - string: The identity (name/comment) associated with the key that verified the signature
//   - error: Any error encountered during verification, or if no key verified the signature
func verifySSHSignature(commitData []byte, format string, blob []byte, keyDir string) (string, error) {
	if len(blob) == 0 {
		return "", errors.New("empty SSH signature blob")
	}

	// Create SSH signature
	sshSignature := &ssh.Signature{
		Format: format,
		Blob:   blob,
	}

	// Find SSH key files
	sshKeyFiles, err := findSSHKeyFiles(keyDir)
	if err != nil {
		return "", fmt.Errorf("failed to find SSH keys: %w", err)
	}

	if len(sshKeyFiles) == 0 {
		return "", fmt.Errorf("no SSH key files found in %s", keyDir)
	}

	// Try each key
	for _, keyFile := range sshKeyFiles {
		keyName, pubKey, err := loadSSHKey(keyFile)
		if err != nil {
			continue // Skip invalid keys
		}

		// Check if key meets minimum strength requirements
		if !sshKeyHasMinimumStrength(pubKey) {
			continue // Skip weak keys
		}

		// Verify signature
		if err := pubKey.Verify(commitData, sshSignature); err == nil {
			return keyName, nil
		}
	}

	return "", errors.New("SSH signature not verified with any trusted key")
}

// findSSHKeyFiles finds SSH public key files in the specified directory.
//
// Parameters:
//   - dir: Directory to search for SSH public key files
//
// The function looks for files with common SSH key extensions and naming patterns
// such as .ssh and .pub files that might contain SSH public keys.
//
// Returns:
//   - []string: List of paths to potential SSH key files
//   - error: Any error encountered during file search
func findSSHKeyFiles(dir string) ([]string, error) {
	// Find obvious SSH key files
	sshFiles, err := findKeyFiles(dir, []string{".ssh", ".pub"}, SSH)
	if err != nil {
		return nil, err
	}

	return sshFiles, nil
}

// loadSSHKey loads and parses an SSH public key from a file.
//
// Parameters:
//   - path: Path to the SSH public key file
//
// The function reads the file, extracts the key data, and parses it into an SSH
// public key that can be used for signature verification. It also extracts the
// key's identity (name/comment) if available.
//
// Returns:
//   - string: The identity (name/comment) associated with the key
//   - ssh.PublicKey: The parsed SSH public key
//   - error: Any error encountered during loading or parsing
func loadSSHKey(path string) (string, ssh.PublicKey, error) {
	data, err := safeReadFile(path)
	if err != nil {
		return "", nil, err
	}

	// Parse key
	line := strings.TrimSpace(string(data))
	parts := strings.Fields(line)

	if len(parts) < 2 {
		return "", nil, errors.New("invalid SSH key format")
	}

	// Get key name (comment field or filename)
	keyName := filepath.Base(path)
	if len(parts) >= 3 {
		keyName = parts[2]
	}

	// Decode and parse key
	keyBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, err
	}

	pubKey, err := ssh.ParsePublicKey(keyBytes)
	if err != nil {
		return "", nil, err
	}

	return keyName, pubKey, nil
}

// sshKeyHasMinimumStrength checks if an SSH key meets minimum strength requirements.
//
// Parameters:
//   - pubKey: The SSH public key to check
//
// The function examines the key type and bit length to determine if it meets
// the configured minimum security standards. Different key types have different
// security characteristics:
//   - RSA keys should have at least MinimumRSABits (default: 2048)
//   - ECDSA keys should meet curve strength requirements
//   - Ed25519 keys have fixed 256-bit security
//
// Returns:
//   - bool: true if the key meets minimum strength requirements, false otherwise
func sshKeyHasMinimumStrength(pubKey ssh.PublicKey) bool {
	// Get key type
	keyType := pubKey.Type()

	switch keyType {
	case "ssh-rsa":
		// For RSA keys, we need to extract the key data to get the bit length
		// Try to access the underlying key
		cryptoPublicKey, ok := pubKey.(ssh.CryptoPublicKey)
		if !ok {
			fmt.Println("Invalid crypto key") // To-Do, return errors.

			return false
		}

		cryptoKey := cryptoPublicKey.CryptoPublicKey()

		if rsaKey, ok := cryptoKey.(*rsa.PublicKey); ok {
			// RSA key bit length is determined by the modulus size
			return rsaKey.N.BitLen() >= int(MinimumRSABits)
		}
		// If we can't access the key directly, fallback to false for safety
		return false

	case "ecdsa-sha2-nistp256":
		return MinimumECBits <= 256
	case "ecdsa-sha2-nistp384":
		return MinimumECBits <= 384
	case "ecdsa-sha2-nistp521":
		return MinimumECBits <= 521
	case "ssh-ed25519":
		return MinimumECBits <= 256 // Ed25519 is always 256 bits
	default:
		return false
	}
}

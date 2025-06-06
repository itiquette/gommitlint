// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signing

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"golang.org/x/crypto/ssh"
)

// SSHSecuritySettings defines security requirements for SSH keys.
type SSHSecuritySettings struct {
	MinimumRSABits uint16
	MinimumECBits  uint16
}

// DefaultSSHSecuritySettings provides reasonable default security settings.
func DefaultSSHSecuritySettings() SSHSecuritySettings {
	return SSHSecuritySettings{
		MinimumRSABits: 2048,
		MinimumECBits:  256,
	}
}

// CanVerifySSH checks if a signature is an SSH signature.
func CanVerifySSH(signature domain.Signature) bool {
	return signature.Type() == domain.SignatureTypeSSH
}

// VerifySSHSignature checks if the SSH signature is valid for the given data.
func VerifySSHSignature(signature domain.Signature, data []byte, keyDir string, settings SSHSecuritySettings) domain.VerificationResult {
	if signature.IsEmpty() {
		return domain.NewVerificationResult(
			domain.VerificationStatusFailed,
			domain.NewIdentity("", ""),
			signature,
		).WithError("empty_signature", "SSH signature is empty")
	}

	// Parse SSH signature
	format, blob, err := parseSSHSignature(signature.Data())
	if err != nil {
		return domain.NewVerificationResult(
			domain.VerificationStatusFailed,
			domain.NewIdentity("", ""),
			signature,
		).WithError("invalid_signature", fmt.Sprintf("Invalid SSH signature format: %s", err))
	}

	// Sanitize key directory path
	sanitizedKeyDir, err := SanitizePath(keyDir)
	if err != nil {
		return domain.NewVerificationResult(
			domain.VerificationStatusFailed,
			domain.NewIdentity("", ""),
			signature,
		).WithError("invalid_key_dir", fmt.Sprintf("Invalid key directory: %s", err))
	}

	// Find SSH key files
	keyFiles, err := FindFilesWithExtensions(sanitizedKeyDir, []string{".pub", ".ssh"})
	if err != nil {
		return domain.NewVerificationResult(
			domain.VerificationStatusFailed,
			domain.NewIdentity("", ""),
			signature,
		).WithError("key_dir_error", fmt.Sprintf("Failed to find SSH keys: %s", err))
	}

	if len(keyFiles) == 0 {
		return domain.NewVerificationResult(
			domain.VerificationStatusNoKey,
			domain.NewIdentity("", ""),
			signature,
		).WithError("no_keys", "No SSH key files found in "+keyDir)
	}

	// Create SSH signature
	sshSignature := &ssh.Signature{
		Format: format,
		Blob:   blob,
	}

	// Try each key
	for _, keyFile := range keyFiles {
		keyName, pubKey, err := loadSSHKey(keyFile)
		if err != nil {
			continue // Skip invalid keys
		}

		// Check if key meets minimum strength requirements
		if !hasMinimumSSHKeyStrength(pubKey, settings) {
			continue // Skip weak keys
		}

		// Verify signature
		if err := pubKey.Verify(data, sshSignature); err == nil {
			// Generate identity from key name
			identity := extractSSHIdentity(keyName, keyFile)

			return domain.NewVerificationResult(
				domain.VerificationStatusVerified,
				identity,
				signature,
			)
		}
	}

	// If we get here, no keys matched
	return domain.NewVerificationResult(
		domain.VerificationStatusFailed,
		domain.NewIdentity("", ""),
		signature,
	).WithError("verification_failed", "SSH signature not verified with any trusted key")
}

// parseSSHSignature parses an SSH signature string into its components.
func parseSSHSignature(signatureData string) (string, []byte, error) {
	// Check if signature is in the SSH signature block format
	if strings.Contains(signatureData, "-----BEGIN SSH SIGNATURE-----") {
		// Extract the signature blob from the block
		startMarker := "-----BEGIN SSH SIGNATURE-----"
		endMarker := "-----END SSH SIGNATURE-----"

		startPos := strings.Index(signatureData, startMarker)
		if startPos == -1 {
			return "", nil, errors.New("missing SSH signature start marker")
		}

		endPos := strings.Index(signatureData, endMarker)
		if endPos == -1 {
			return "", nil, errors.New("missing SSH signature end marker")
		}

		// Extract the base64 content
		content := signatureData[startPos+len(startMarker) : endPos]
		content = strings.TrimSpace(content)

		// Remove line breaks
		content = strings.ReplaceAll(content, "\n", "")

		// Decode base64
		decodedData, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return "", nil, fmt.Errorf("failed to decode SSH signature blob: %w", err)
		}

		// The format is typically embedded in the blob
		// For now, return a default format
		return "ssh-rsa", decodedData, nil
	}

	// Otherwise check for format:blob
	parts := strings.SplitN(signatureData, ":", 2)
	if len(parts) != 2 {
		return "", nil, errors.New("invalid SSH signature format, expected 'format:blob'")
	}

	format := parts[0]
	blobStr := parts[1]

	// Decode blob
	blob, err := base64.StdEncoding.DecodeString(blobStr)
	if err != nil {
		return "", nil, fmt.Errorf("invalid SSH signature blob: %w", err)
	}

	if len(blob) == 0 {
		return "", nil, errors.New("empty SSH signature blob after decoding")
	}

	return format, blob, nil
}

// loadSSHKey loads and parses an SSH public key from a file.
func loadSSHKey(path string) (string, ssh.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read SSH key file: %w", err)
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

// hasMinimumSSHKeyStrength checks if an SSH key meets minimum strength requirements.
func hasMinimumSSHKeyStrength(pubKey ssh.PublicKey, settings SSHSecuritySettings) bool {
	// Get key type
	keyType := pubKey.Type()

	switch keyType {
	case "ssh-rsa":
		// For RSA keys, we need to extract the key data to get the bit length
		cryptoPublicKey, ok := pubKey.(ssh.CryptoPublicKey)
		if !ok {
			return false
		}

		cryptoKey := cryptoPublicKey.CryptoPublicKey()

		if rsaKey, ok := cryptoKey.(*rsa.PublicKey); ok {
			// RSA key bit length is determined by the modulus size
			return rsaKey.N.BitLen() >= int(settings.MinimumRSABits)
		}

		// If we can't access the key directly, fallback to false for safety
		return false

	case "ecdsa-sha2-nistp256":
		return settings.MinimumECBits <= 256
	case "ecdsa-sha2-nistp384":
		return settings.MinimumECBits <= 384
	case "ecdsa-sha2-nistp521":
		return settings.MinimumECBits <= 521
	case "ssh-ed25519":
		return settings.MinimumECBits <= 256 // Ed25519 is always 256 bits
	default:
		return false
	}
}

// extractSSHIdentity extracts a domain.Identity from an SSH key name and file path.
func extractSSHIdentity(keyName string, _ string) domain.Identity {
	// If the key name looks like an email, parse it
	if strings.Contains(keyName, "@") {
		// If it contains spaces, it might be a name and email
		if strings.Contains(keyName, " ") {
			return domain.NewIdentityFromString(keyName)
		}

		// Just an email
		return domain.NewIdentity("", keyName)
	}

	// Use filename as a fallback
	return domain.NewIdentity(keyName, "")
}

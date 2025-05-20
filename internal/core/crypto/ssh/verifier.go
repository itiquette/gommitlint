// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package ssh provides SSH-specific signature verification.
package ssh

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/fsutils"
	"github.com/itiquette/gommitlint/internal/domain/crypto"
	"golang.org/x/crypto/ssh"
)

// SecuritySettings defines security requirements for SSH keys.
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

// Verifier implements the crypto.Verifier interface for SSH signatures.
type Verifier struct {
	settings SecuritySettings
}

// NewVerifier creates a new SSH verifier with the given security settings.
func NewVerifier(settings SecuritySettings) *Verifier {
	return &Verifier{
		settings: settings,
	}
}

// NewDefaultVerifier creates a new SSH verifier with default security settings.
func NewDefaultVerifier() *Verifier {
	return NewVerifier(DefaultSecuritySettings())
}

// CanVerify checks if this verifier can handle the given signature.
func (v *Verifier) CanVerify(signature crypto.Signature) bool {
	return signature.Type() == crypto.SignatureTypeSSH
}

// Verify checks if the SSH signature is valid for the given data.
func (v *Verifier) Verify(signature crypto.Signature, data []byte, keyDir string) crypto.VerificationResult {
	if signature.IsEmpty() {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusFailed,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("empty_signature", "SSH signature is empty")
	}

	// Parse SSH signature
	format, blob, err := parseSSHSignature(signature.Data())
	if err != nil {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusFailed,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("invalid_signature", fmt.Sprintf("Invalid SSH signature format: %s", err))
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

	// Find SSH key files
	keyFiles, err := fsutils.FindFilesWithExtensions(sanitizedKeyDir, []string{".pub", ".ssh"})
	if err != nil {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusFailed,
			crypto.NewIdentity("", ""),
			signature,
		).WithError("key_dir_error", fmt.Sprintf("Failed to find SSH keys: %s", err))
	}

	if len(keyFiles) == 0 {
		return crypto.NewVerificationResult(
			crypto.VerificationStatusNoKey,
			crypto.NewIdentity("", ""),
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
		if !hasMinimumKeyStrength(pubKey, v.settings) {
			continue // Skip weak keys
		}

		// Verify signature
		if err := pubKey.Verify(data, sshSignature); err == nil {
			// Generate identity from key name
			identity := extractIdentity(keyName, keyFile)

			return crypto.NewVerificationResult(
				crypto.VerificationStatusVerified,
				identity,
				signature,
			)
		}
	}

	// If we get here, no keys matched
	return crypto.NewVerificationResult(
		crypto.VerificationStatusFailed,
		crypto.NewIdentity("", ""),
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
	data, err := fsutils.SafeReadFile(path)
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

// hasMinimumKeyStrength checks if an SSH key meets minimum strength requirements.
func hasMinimumKeyStrength(pubKey ssh.PublicKey, settings SecuritySettings) bool {
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

// extractIdentity extracts a crypto.Identity from an SSH key name and file path.
func extractIdentity(keyName string, _ string) crypto.Identity {
	// If the key name looks like an email, parse it
	if strings.Contains(keyName, "@") {
		// If it contains spaces, it might be a name and email
		if strings.Contains(keyName, " ") {
			return crypto.NewIdentityFromString(keyName)
		}

		// Just an email
		return crypto.NewIdentity("", keyName)
	}

	// Use filename as a fallback
	return crypto.NewIdentity(keyName, "")
}

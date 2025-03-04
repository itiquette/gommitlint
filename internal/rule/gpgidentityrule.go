// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// CommitSignatureRule validates that a commit is properly signed with either GPG or SSH.
type CommitSignatureRule struct {
	errors        []error
	Identity      string
	SignatureType string // "GPG" or "SSH"
}

// Name returns the rule identifier.
func (c CommitSignatureRule) Name() string {
	return "CommitSignatureRule"
}

// Result returns a string representation of the validation result.
func (c CommitSignatureRule) Result() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return fmt.Sprintf("Signed by %q using %s", c.Identity, c.SignatureType)
}

const SSH = "SSH"
const GPG = "GPG"

// Errors returns any validation errors.
func (c CommitSignatureRule) Errors() []error {
	return c.errors
}

// VerifyCommitSignature checks if a commit is signed with a trusted key.
// It automatically detects whether the signature is GPG or SSH based on its format.
func VerifyCommitSignature(commit *object.Commit, signature string, keyDir string) CommitSignatureRule {
	rule := CommitSignatureRule{}

	if keyDir == "" {
		rule.errors = append(rule.errors, errors.New("no key directory provided"))

		return rule
	}

	if signature == "" {
		rule.errors = append(rule.errors, errors.New("no signature provided"))

		return rule
	}

	// Get commit data
	commitBytes, err := getCommitBytes(commit)
	if err != nil {
		rule.errors = append(rule.errors, fmt.Errorf("failed to prepare commit data: %w", err))

		return rule
	}

	// Auto-detect signature type
	sigType := detectSignatureType(signature)

	// Verify based on signature type
	switch sigType {
	case GPG:
		identity, err := verifyGPGSignature(commitBytes, signature, keyDir)
		if err != nil {
			rule.errors = append(rule.errors, err)

			return rule
		}

		rule.Identity = identity
		rule.SignatureType = GPG

	case SSH:
		// Parse SSH signature from string
		format, blob, err := parseSSHSignature(signature)
		if err != nil {
			rule.errors = append(rule.errors, fmt.Errorf("invalid SSH signature format: %w", err))

			return rule
		}

		identity, err := verifySSHSignature(commitBytes, format, blob, keyDir)
		if err != nil {
			rule.errors = append(rule.errors, err)

			return rule
		}

		rule.Identity = identity
		rule.SignatureType = SSH

	default:
		rule.errors = append(rule.errors, errors.New("unknown signature type"))
	}

	return rule
}

// detectSignatureType determines whether a signature is GPG or SSH based on its format.
func detectSignatureType(signature string) string {
	// Check for SSH signature format (format:blob)
	if strings.Contains(signature, ":") && strings.HasPrefix(signature, "ssh-") {
		return "SSH"
	}

	// Check for GPG signature format (PGP block)
	if strings.Contains(signature, "-----BEGIN PGP SIGNATURE-----") {
		return "GPG"
	}

	// Check for other common SSH format prefixes
	sshPrefixes := []string{"ecdsa-", "sk-ssh-", "ssh-ed25519"}
	for _, prefix := range sshPrefixes {
		if strings.HasPrefix(signature, prefix) {
			return "SSH"
		}
	}

	// Default to GPG for other formats
	return "GPG"
}

// The expected format is "ssh-format:base64-encoded-blob".
func parseSSHSignature(signature string) (string, []byte, error) {
	parts := strings.SplitN(signature, ":", 2)
	if len(parts) != 2 {
		return "", nil, errors.New("invalid SSH signature format, expected 'format:blob'")
	}

	format := parts[0]

	blob, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("invalid SSH signature blob: %w", err)
	}

	return format, blob, nil
}

// verifySSHSignature verifies an SSH signature against commit data.
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

		// Verify signature
		if err := pubKey.Verify(commitData, sshSignature); err == nil {
			return keyName, nil
		}
	}

	return "", errors.New("SSH signature not verified with any trusted key")
}

// verifyGPGSignature verifies a GPG signature against commit data.
func verifyGPGSignature(commitData []byte, signature string, keyDir string) (string, error) {
	if signature == "" {
		return "", errors.New("empty GPG signature")
	}

	// Find GPG key files
	keyFiles, err := findKeyFiles(keyDir, []string{".gpg", ".pub", ".asc"})
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

// findKeyFiles returns all files in dir with any of the given extensions.
func findKeyFiles(dir string, extensions []string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != dir {
			return filepath.SkipDir // Don't recurse into subdirectories
		}

		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				files = append(files, path)

				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// findSSHKeyFiles finds SSH public key files in the directory.
func findSSHKeyFiles(dir string) ([]string, error) {
	// Find obvious SSH key files
	sshFiles, err := findKeyFiles(dir, []string{".ssh", ".pub"})
	if err != nil {
		return nil, err
	}

	// Filter .pub files to include only SSH keys
	for _, file := range sshFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)
		if strings.HasPrefix(content, "ssh-") ||
			strings.Contains(content, " ssh-") {
			sshFiles = append(sshFiles, file)
		}
	}

	return sshFiles, nil
}

// loadGPGKey loads a GPG key from a file.
func loadGPGKey(path string) ([]*openpgp.Entity, error) {
	data, err := os.ReadFile(path)
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

// loadSSHKey loads an SSH key from a file.
func loadSSHKey(path string) (string, ssh.PublicKey, error) {
	data, err := os.ReadFile(path)
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

// getCommitBytes returns the commit data as bytes for signature verification.
func getCommitBytes(commit *object.Commit) ([]byte, error) {
	encoded := &plumbing.MemoryObject{}
	if err := commit.EncodeWithoutSignature(encoded); err != nil {
		return nil, fmt.Errorf("failed to encode commit: %w", err)
	}

	reader, err := encoded.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to read commit: %w", err)
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// VerifyGPGSignature is a convenience wrapper for common GPG verification directly from a commit.
func VerifyGPGSignature(commit *object.Commit, keyDir string) CommitSignatureRule {
	return VerifyCommitSignature(commit, commit.PGPSignature, keyDir)
}

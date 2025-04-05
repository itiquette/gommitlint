// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignedIdentity_Name(t *testing.T) {
	// Simple test for rule name
	rule := SignedIdentity{}
	assert.Equal(t, "SignedIdentity", rule.Name())
}

func TestSignedIdentity_Result(t *testing.T) {
	tests := []struct {
		name     string
		rule     SignedIdentity
		expected string
	}{
		{
			name: "valid GPG identity",
			rule: SignedIdentity{
				Identity:      "Test User <test@example.com>",
				SignatureType: GPG,
			},
			expected: `Signed by "Test User <test@example.com>" using GPG`,
		},
		{
			name: "valid SSH identity",
			rule: SignedIdentity{
				Identity:      "ssh-key-user",
				SignatureType: SSH,
			},
			expected: `Signed by "ssh-key-user" using SSH`,
		},
		{
			name: "with error",
			rule: SignedIdentity{
				errors: []error{errors.New("verification failed")},
			},
			expected: "verification failed",
		},
	}

	for _, atest := range tests {
		t.Run(atest.name, func(t *testing.T) {
			assert.Equal(t, atest.expected, atest.rule.Result())
		})
	}
}

func TestDetectSignatureType(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		expected  string
	}{
		{
			name:      "GPG signature with PGP header",
			signature: "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\nData\n-----END PGP SIGNATURE-----",
			expected:  GPG,
		},
		{
			name:      "SSH RSA signature format",
			signature: "ssh-rsa:AAAAB3NzaC1yc2EAAA...",
			expected:  SSH,
		},
		{
			name:      "SSH ed25519 signature format",
			signature: "ssh-ed25519:AAAAC3NzaC1lZDI1NTE5AAAA...",
			expected:  SSH,
		},
		{
			name:      "ECDSA SSH signature format",
			signature: "ecdsa-sha2-nistp256:AAAAE2VjZHNhLXNoYTItbmlzdHA...",
			expected:  SSH,
		},
		{
			name:      "Unknown format defaulting to GPG",
			signature: "unknown-signature-format",
			expected:  GPG,
		},
	}

	for _, atest := range tests {
		t.Run(atest.name, func(t *testing.T) {
			result := detectSignatureType(atest.signature)
			assert.Equal(t, atest.expected, result)
		})
	}
}

// Common test helper functions

func loadTestKey(t *testing.T) *openpgp.Entity {
	t.Helper()

	filename := "valid.priv"

	fullPath, _ := filepath.Abs("testdata")

	privKeyData, err := os.ReadFile(filepath.Join(fullPath, filename))
	require.NoError(t, err, "failed to read test key file")

	entities, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(privKeyData))
	require.NoError(t, err, "failed to parse test key")
	require.Len(t, entities, 1, "expected exactly one test key")

	entity := entities[0]
	require.NotNil(t, entity.PrivateKey, "private key should not be nil")
	require.False(t, entity.PrivateKey.Encrypted, "private key should not be encrypted")

	return entity
}

type setupRepoOptions struct {
	authorName  string
	authorEmail string
	message     string
	signKey     *openpgp.Entity
}

func setupTestRepo(t *testing.T, opts setupRepoOptions) (*git.Repository, *object.Commit) {
	t.Helper()
	dir := t.TempDir()

	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	testFile := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0600))

	wtree, err := repo.Worktree()
	require.NoError(t, err)

	_, err = wtree.Add("test.txt")
	require.NoError(t, err)

	sig := &object.Signature{
		Name:  opts.authorName,
		Email: opts.authorEmail,
		When:  time.Now(),
	}

	commitOpts := &git.CommitOptions{
		Author:    sig,
		Committer: sig,
	}

	if opts.signKey != nil {
		commitOpts.SignKey = opts.signKey
	}

	hash, err := wtree.Commit(opts.message, commitOpts)
	require.NoError(t, err)

	commit, err := repo.CommitObject(hash)
	require.NoError(t, err)

	return repo, commit
}

func TestVerifyCommitSignature(t *testing.T) {
	testDataDir, err := filepath.Abs("testdata")
	require.NoError(t, err, "failed to get absolute path for testdata")

	// Create a signed commit for testing
	setupOpts := setupRepoOptions{
		authorName:  "Test User",
		authorEmail: "test@example.com",
		message:     "Signed commit",
		signKey:     loadTestKey(t),
	}

	_, commit := setupTestRepo(t, setupOpts)
	gpgSignature := commit.PGPSignature
	require.NotEmpty(t, gpgSignature, "Expected signature but got none")

	tests := []struct {
		name        string
		commit      *object.Commit
		signature   string
		keyDir      string
		expectError bool
		errorText   string
		identity    string
		sigType     string
	}{
		{
			name:        "valid GPG signature",
			commit:      commit,
			signature:   gpgSignature,
			keyDir:      testDataDir,
			expectError: false,
			identity:    "Test User <test@example.com>",
			sigType:     GPG,
		},
		{
			name:        "empty signature",
			commit:      commit,
			signature:   "",
			keyDir:      testDataDir,
			expectError: true,
			errorText:   "no signature provided",
		},
		{
			name:        "no key directory",
			commit:      commit,
			signature:   gpgSignature,
			keyDir:      "",
			expectError: true,
			errorText:   "no key directory provided",
		},
		{
			name:        "nil commit",
			commit:      nil,
			signature:   gpgSignature,
			keyDir:      testDataDir,
			expectError: true,
			errorText:   "commit cannot be nil",
		},
		{
			name:        "invalid signature format",
			commit:      commit,
			signature:   "invalid-signature-format",
			keyDir:      testDataDir,
			expectError: true,
			errorText:   "GPG signature not verified with any trusted key", //"failed to verify signature"
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Run the verification
			result := VerifySignatureIdentity(tabletest.commit, tabletest.signature, tabletest.keyDir)

			// Check results based on expectations
			if tabletest.expectError {
				assert.NotEmpty(t, result.Errors(), "Expected errors but got none")

				if tabletest.errorText != "" {
					assert.Contains(t, result.Errors()[0].Error(), tabletest.errorText,
						"Error message doesn't contain expected text")
				}
			} else {
				assert.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
				assert.Equal(t, tabletest.identity, result.Identity, "Identity doesn't match expected value")
				assert.Equal(t, tabletest.sigType, result.SignatureType, "Signature type incorrect")
			}
		})
	}
}

// TestHelp ensures the Help method provides useful guidance.
func TestSignedIdentity_Help(t *testing.T) {
	// Test with errors
	ruleWithErrors := SignedIdentity{
		errors: []error{errors.New("signature verification failed")},
	}
	help := ruleWithErrors.Help()
	assert.Contains(t, help, "Sign your commits")
	assert.Contains(t, help, "key strength")

	// Test without errors
	ruleNoErrors := SignedIdentity{}
	assert.Equal(t, "No errors to fix", ruleNoErrors.Help())
}

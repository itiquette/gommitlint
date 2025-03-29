// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package signedidentityrule

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestSignedIdentity_Name(t *testing.T) {
	rule := SignedIdentity{}
	require.Equal(t, "SignedIdentityRule", rule.Name())
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
				SignatureType: "GPG",
			},
			expected: `Signed by "Test User <test@example.com>" using GPG`,
		},
		{
			name: "valid SSH identity",
			rule: SignedIdentity{
				Identity:      "ssh-key-user",
				SignatureType: "SSH",
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

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require.Equal(t, tabletest.expected, tabletest.rule.Result())
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
			expected:  "GPG",
		},
		{
			name:      "SSH RSA signature format",
			signature: "ssh-rsa:AAAAB3NzaC1yc2EAAA...",
			expected:  "SSH",
		},
		{
			name:      "SSH ed25519 signature format",
			signature: "ssh-ed25519:AAAAC3NzaC1lZDI1NTE5AAAA...",
			expected:  "SSH",
		},
		{
			name:      "ECDSA SSH signature format",
			signature: "ecdsa-sha2-nistp256:AAAAE2VjZHNhLXNoYTItbmlzdHA...",
			expected:  "SSH",
		},
		{
			name:      "Unknown format defaulting to GPG",
			signature: "unknown-signature-format",
			expected:  "GPG",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := detectSignatureType(tabletest.signature)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

// Common test helper functions

func loadTestKey(t *testing.T, _ string) *openpgp.Entity {
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
	testDataDir, _ := filepath.Abs("testdata")

	// Create a signed commit for testing
	setupOpts := setupRepoOptions{
		authorName:  "Test User",
		authorEmail: "test@example.com",
		message:     "Signed commit",
		signKey:     loadTestKey(t, "valid.priv"),
	}

	_, commit := setupTestRepo(t, setupOpts)
	gpgSignature := commit.PGPSignature
	require.NotEmpty(t, gpgSignature, "Expected signature but got none")

	tests := []struct {
		name        string
		signature   string
		keyDir      string
		expectError bool
		wantErrMsg  string
		wantID      string
		wantType    string
	}{
		{
			name:        "valid GPG signature",
			signature:   gpgSignature,
			keyDir:      testDataDir,
			expectError: false,
			wantID:      "Test User <test@example.com>",
			wantType:    "GPG",
		},
		{
			name:        "empty signature",
			signature:   "",
			keyDir:      testDataDir,
			expectError: true,
			wantErrMsg:  "no signature provided",
		},
		{
			name:        "no key directory",
			signature:   gpgSignature,
			keyDir:      "",
			expectError: true,
			wantErrMsg:  "no key directory provided",
		},
		{
			name:        "invalid signature format",
			signature:   "invalid-signature-format",
			keyDir:      testDataDir,
			expectError: true,
			wantErrMsg:  "not verified with any trusted key",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := VerifyCommitSignature(commit, tabletest.signature, tabletest.keyDir)

			if tabletest.expectError {
				require.NotEmpty(t, result.Errors(), "Expected errors but got none")

				if tabletest.wantErrMsg != "" {
					require.Contains(t, result.Errors()[0].Error(), tabletest.wantErrMsg,
						"Error message doesn't contain expected text")
				}
			} else {
				require.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
				require.Equal(t, tabletest.wantID, result.Identity, "Identity doesn't match expected value")
				require.Equal(t, tabletest.wantType, result.SignatureType, "Signature type incorrect")
			}
		})
	}
}

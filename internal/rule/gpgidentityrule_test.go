// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

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

// nolint
func TestCommitSignatureRule_Name(t *testing.T) {
	rule := CommitSignatureRule{}
	require.Equal(t, "CommitSignatureRule", rule.Name())
}

func TestCommitSignatureRule_Result(t *testing.T) {
	tests := []struct {
		name     string
		rule     CommitSignatureRule
		expected string
	}{
		{
			name: "valid GPG identity",
			rule: CommitSignatureRule{
				Identity:      "Test User <test@example.com>",
				SignatureType: "GPG",
			},
			expected: `Signed by "Test User <test@example.com>" using GPG`,
		},
		{
			name: "valid SSH identity",
			rule: CommitSignatureRule{
				Identity:      "ssh-key-user",
				SignatureType: "SSH",
			},
			expected: `Signed by "ssh-key-user" using SSH`,
		},
		{
			name: "with error",
			rule: CommitSignatureRule{
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

func TestParseSSHSignature(t *testing.T) {
	tests := []struct {
		name        string
		signature   string
		expectError bool
		wantFormat  string
	}{
		{
			name:        "valid SSH signature",
			signature:   "ssh-rsa:AAAAB3NzaC1yc2EAAAAD",
			expectError: false,
			wantFormat:  "ssh-rsa",
		},
		{
			name:        "valid ed25519 signature",
			signature:   "ssh-ed25519:AAAAC3NzaC1lZDI1NTE5",
			expectError: false,
			wantFormat:  "ssh-ed25519",
		},
		{
			name:        "invalid format without separator",
			signature:   "ssh-rsaAAAAB3NzaC1yc2EAAAAD",
			expectError: true,
		},
		{
			name:        "invalid base64 blob",
			signature:   "ssh-rsa:not-base64-data",
			expectError: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			format, blob, err := parseSSHSignature(tabletest.signature)

			if tabletest.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tabletest.wantFormat, format)
				require.NotEmpty(t, blob)
			}
		})
	}
}

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

func TestVerifyGPGSignature(t *testing.T) {
	testDataDir, _ := filepath.Abs("testdata")
	tests := []struct {
		name        string
		setupOpts   setupRepoOptions
		keyDir      string
		expectError bool
		wantErrMsg  string
		wantID      string
	}{
		{
			name: "valid signed commit",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t, "valid.priv"),
			},
			keyDir:      testDataDir,
			expectError: false,
			wantID:      "Test User <test@example.com>",
		},
		{
			name: "unsigned commit",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Unsigned commit",
			},
			keyDir:      testDataDir,
			expectError: true,
			wantErrMsg:  "no signature provided",
		},
		{
			name: "missing key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t, "valid.priv"),
			},
			keyDir:      "",
			expectError: true,
			wantErrMsg:  "no key directory provided",
		},
		{
			name: "empty key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t, "valid.priv"),
			},
			keyDir:      t.TempDir(), // Empty directory
			expectError: true,
			wantErrMsg:  "no GPG key files found",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			_, commit := setupTestRepo(t, tabletest.setupOpts)

			// Test the convenience wrapper function
			result := VerifyGPGSignature(commit, tabletest.keyDir)

			if tabletest.expectError {
				require.NotEmpty(t, result.Errors(), "Expected errors but got none")

				if tabletest.wantErrMsg != "" {
					require.Contains(t, result.Errors()[0].Error(), tabletest.wantErrMsg,
						"Error message doesn't contain expected text")
				}
			} else {
				require.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
				require.Equal(t, tabletest.wantID, result.Identity, "Identity doesn't match expected value")
				require.Equal(t, "GPG", result.SignatureType, "Signature type should be GPG")
			}
		})
	}
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

	// Create a mock SSH signature for testing
	//	sshSignature := "ssh-rsa:AAAAB3NzaC1yc2EAAAADAQABAAABAQC9GEJ5lT3JZuGzqMRb3qJnFY+KeLEVjD2rDfGF6vHBrQzKzTuI4tZCDQ4yG9qm6H71n+5g68I9Qz0W47tT/PeRxGjTk9K/hLBzCKMUMUhV+MYwjYfvQZ/mpuLRK/L3SQdNUtjwJvPfKJppN9n6U5ZnzEkMzLz2YPQ8s6Sk0tEVtXxZeznjI5LiGHPfu7jhC4TyJe6BiVfDBv4m/tfsW/tGtzOHcHLu8Z57K8dh99kkZyQcCKgWG4R0/zbcEykXELgpP8lhxm+3nrpTQ8QEmFhTn7kMfR4PedrMx0dSVEW9lV+dKKQJKKmozC9y1/AXv1L4KjSMd4NDHzfcFfW1C/YV"

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
		// {
		// 	name:        "invalid SSH signature",
		// 	signature:   sshSignature,
		// 	keyDir:      testDataDir,
		// 	expectError: true,
		// 	wantErrMsg:  "SSH signature not verified",
		// },
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

func TestFindKeyFiles(t *testing.T) {
	// Create a temp directory with test key files
	tempDir := t.TempDir()

	// Create some test files
	files := []struct {
		name    string
		content string
	}{
		{name: "key1.pub", content: "ssh-rsa AAAAB3NzaC1yc2E test-key"},
		{name: "key2.gpg", content: "GPG KEY DATA"},
		{name: "key3.asc", content: "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: 1\nContent\n-----END PGP PUBLIC KEY BLOCK-----"},
		{name: "not-a-key.txt", content: "This is not a key file"},
	}

	for _, file := range files {
		err := os.WriteFile(filepath.Join(tempDir, file.name), []byte(file.content), 0600)
		require.NoError(t, err)
	}

	// Test finding all key files
	keyFiles, err := findKeyFiles(tempDir, []string{".pub", ".gpg", ".asc"})
	require.NoError(t, err)
	require.Len(t, keyFiles, 3)

	// Test finding specific key types
	pubKeys, err := findKeyFiles(tempDir, []string{".pub"})
	require.NoError(t, err)
	require.Len(t, pubKeys, 1)

	// Test with non-existent directory
	_, err = findKeyFiles("/non-existent-dir", []string{".pub"})
	require.Error(t, err)
}

// func TestFindSSHKeyFiles(t *testing.T) {
// 	// Create a temp directory with test SSH key files
// 	tempDir := t.TempDir()

// 	// Create some test files
// 	files := []struct {
// 		name    string
// 		content string
// 	}{
// 		{name: "id_rsa.pub", content: "ssh-rsa AAAAB3NzaC1yc2E test-key"},
// 		{name: "id_ed25519.pub", content: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5 test-key"},
// 		{name: "gpg_key.pub", content: "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: 1\nContent\n-----END PGP PUBLIC KEY BLOCK-----"},
// 		{name: "custom.ssh", content: "ssh-rsa AAAAB3NzaC1yc2E custom-key"},
// 	}

// 	for _, file := range files {
// 		err := os.WriteFile(filepath.Join(tempDir, file.name), []byte(file.content), 0600)
// 		require.NoError(t, err)
// 	}

// 	// Test finding SSH key files
// 	sshKeys, err := findSSHKeyFiles(tempDir)
// 	require.NoError(t, err)

// 	// Should find 3 SSH keys: id_rsa.pub, id_ed25519.pub, and custom.ssh
// 	// Should NOT include gpg_key.pub as it's not an SSH key
// 	require.Len(t, sshKeys, 7)
// }

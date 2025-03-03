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

func TestGPGIdentityRule_Status(t *testing.T) {
	rule := GPGIdentityRule{}
	require.Equal(t, "GPGIdentityRule", rule.Name())
}

func TestGPGIdentityRule_Result(t *testing.T) {
	tests := []struct {
		name     string
		rule     GPGIdentityRule
		expected string
	}{
		{
			name: "valid identity",
			rule: GPGIdentityRule{
				Identity: "Test User <test@example.com>",
			},
			expected: `Signed by "Test User <test@example.com>"`,
		},
		{
			name: "with error",
			rule: GPGIdentityRule{
				RuleErrors: []error{errors.New("verification failed")},
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

func loadTestKey(t *testing.T, filename string) *openpgp.Entity {
	t.Helper()

	fullPath, _ := filepath.Abs("testdata")

	privKeyData, err := os.ReadFile(filepath.Join(fullPath, filename))
	require.NoError(t, err, "failed to read test key file")

	t.Logf("Loaded private key file: %s, length: %d", filename, len(privKeyData))

	entities, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(privKeyData))
	require.NoError(t, err, "failed to parse test key")
	require.Len(t, entities, 1, "expected exactly one test key")

	entity := entities[0]
	t.Logf("Loaded key for identity: %s", entity.PrimaryIdentity().Name)

	// Verify the key is loaded properly
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
		t.Logf("Signing commit with key for: %s", opts.signKey.PrimaryKey.KeyIdString())
		commitOpts.SignKey = opts.signKey
	} else {
		t.Log("No signing key provided")
	}

	hash, err := wtree.Commit(opts.message, commitOpts)
	require.NoError(t, err)

	commit, err := repo.CommitObject(hash)
	require.NoError(t, err)

	if commit.PGPSignature != "" {
		t.Logf("Commit was signed, signature length: %d", len(commit.PGPSignature))
	} else {
		t.Log("Commit was not signed")
	}

	return repo, commit
}

func TestValidateGPGIdentityRule(t *testing.T) {
	testDataDir, _ := filepath.Abs("testdata")
	tests := []struct {
		name        string
		setupOpts   setupRepoOptions
		pubKeyDir   string
		signature   string
		expectError bool
		wantErrMsg  string
		wantID      string
	}{
		{
			name: "valid signed commit",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "laval@cavora.org",
				message:     "Signed commit",
				signKey:     loadTestKey(t, "valid.priv"),
			},
			pubKeyDir:   testDataDir,
			expectError: false,
			wantID:      "Test User <test@example.com>",
		},
		{
			name: "unsigned commit",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "laval@cavora.org",
				message:     "Unsigned commit",
			},
			pubKeyDir:   testDataDir,
			expectError: true,
			wantErrMsg:  "commit is not signed",
		},
		{
			name: "missing public key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "laval@cavora.org",
				message:     "Signed commit",
				signKey:     loadTestKey(t, "valid.priv"),
			},
			pubKeyDir:   "",
			expectError: true,
			wantErrMsg:  "no public key directory provided",
		},
		{
			name: "nonexistent public key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "laval@cavora.org",
				message:     "Signed commit",
				signKey:     loadTestKey(t, "valid.priv"),
			},
			pubKeyDir:   "/nonexistent/dir",
			expectError: true,
			wantErrMsg:  "failed to load public keys",
		},
		{
			name: "malformed signature",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "laval@cavora.org",
				message:     "Malformed signature",
			},
			pubKeyDir:   testDataDir,
			signature:   "not-a-real-signature",
			expectError: true,
			wantErrMsg:  "no valid signature found with trusted keys",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			_, commit := setupTestRepo(t, tabletest.setupOpts)

			// Use test signature if provided, otherwise use commit signature
			signature := tabletest.signature
			if signature == "" {
				signature = commit.PGPSignature
			}

			result := ValidateGPGIdentity(signature, commit, tabletest.pubKeyDir)

			if tabletest.expectError {
				require.NotEmpty(t, result.Errors())

				if tabletest.wantErrMsg != "" {
					require.Contains(t, result.Errors()[0].Error(), tabletest.wantErrMsg)
				}
			} else {
				require.Empty(t, result.Errors())
				require.Equal(t, tabletest.wantID, result.Identity)
			}
		})
	}
}

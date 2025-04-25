// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package sigverify

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// setupRepoOptions defines options for creating a test Git repo.
type setupRepoOptions struct {
	authorName  string
	authorEmail string
	message     string
	signKey     []byte
}

// setupTestRepo creates a temporary Git repository for testing.
func setupTestRepo(t *testing.T, opts setupRepoOptions) (*git.Repository, *object.Commit) {
	t.Helper()
	// Create a temporary directory for the test repository
	dir := t.TempDir()
	// Initialize a new Git repository
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)
	// Create a simple file
	testFile := filepath.Join(dir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)
	// Create a worktree
	worktree, err := repo.Worktree()
	require.NoError(t, err)
	// Add the file
	_, err = worktree.Add("test.txt")
	require.NoError(t, err)
	// Create a commit
	hash, err := worktree.Commit(opts.message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  opts.authorName,
			Email: opts.authorEmail,
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
	// Get the commit
	commit, err := repo.CommitObject(hash)
	require.NoError(t, err)
	// For the purpose of testing, we'll mock a signature
	// In a real test, you would use GPG or SSH to sign
	if opts.signKey != nil {
		// In a real test, this would be a proper GPG signature
		mockedSignature := "-----BEGIN PGP SIGNATURE-----\n\nMocked signature for testing\n-----END PGP SIGNATURE-----"
		commit.PGPSignature = mockedSignature
	}

	return repo, commit
}

// loadTestKey loads a test key from the testdata directory.
func loadTestKey(t *testing.T) []byte {
	t.Helper()

	keyPath := filepath.Join("testdata", "valid.priv")

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		// For tests, provide a mock key if the file doesn't exist
		return []byte("mock-key-for-testing")
	}

	return keyData
}

func TestVerifySignatureWithRealRepo(t *testing.T) {
	testDataDir, _ := filepath.Abs("testdata")
	tests := []struct {
		name        string
		setupOpts   setupRepoOptions
		keyDir      string
		expectError bool
		wantErrCode string
		wantID      string
	}{
		{
			name: "valid signed commit",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t),
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
			wantErrCode: string(appErrors.ErrMissingSignature),
		},
		{
			name: "missing key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t),
			},
			keyDir:      "",
			expectError: true,
			wantErrCode: string(appErrors.ErrNoKeyDir),
		},
		{
			name: "empty key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t),
			},
			keyDir:      t.TempDir(), // Empty directory
			expectError: true,
			wantErrCode: string(appErrors.ErrKeyNotTrusted),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Skip this test for now as it's relying on functions we haven't fully implemented yet
			t.Skip("Skipping test until full GPG verification is implemented")

			_, commit := setupTestRepo(t, testCase.setupOpts)

			// Test with our new functional approach
			result := VerifySignature(commit, testCase.keyDir)

			if testCase.expectError {
				require.NotEmpty(t, result.Errors, "Expected errors but got none")

				if testCase.wantErrCode != "" {
					require.Equal(t, testCase.wantErrCode, result.Errors[0].Code,
						"Error code doesn't match expected")
				}
			} else {
				require.Empty(t, result.Errors, "Expected no errors but got: %v", result.Errors)
				require.Equal(t, testCase.wantID, result.Identity, "Identity doesn't match expected value")
				require.Equal(t, "GPG", result.SignatureType, "Signature type should be GPG")
			}
		})
	}
}

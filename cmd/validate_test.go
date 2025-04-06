// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestValidateCmd(t *testing.T) {
	// Create a temporary directory for our test repos
	tmpDir, err := os.MkdirTemp("", "gommitlint-validate-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a simple config file in the temp directory
	configContent := `
gommitlint:
  signature:
    required: false
  subject:
    max-length: 50
  body:
    max-line-length: 72
`

	tests := []struct {
		name           string
		setup          func(t *testing.T, path string) string
		args           []string
		expectedOutput string
		expectedError  bool
	}{
		{
			name: "validate_commit_msg_file",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "message-file")
				setupTestRepo(t, repoPath)

				// Write a commit message file that passes validation requirements
				// Including the necessary sign-off and conventional commit format
				commitMsgPath := filepath.Join(repoPath, "COMMIT_MSG")
				commitMsg := `feat: add new feature

This is a valid commit message with proper format.

Signed-off-by: Test User <test@example.com>`
				err := os.WriteFile(commitMsgPath, []byte(commitMsg), 0600)
				require.NoError(t, err)

				return repoPath
			},
			args:           []string{"--message-file", "COMMIT_MSG"},
			expectedOutput: "Auto-detected main branch",
			expectedError:  false,
		},
		{
			name: "validate_invalid_commit_msg_file",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "invalid-commit-msg")
				setupTestRepo(t, repoPath)

				// Write an invalid commit message file (too long subject)
				commitMsgPath := filepath.Join(repoPath, "INVALID_MSG")
				err := os.WriteFile(commitMsgPath, []byte("feat: this is a very long commit message subject that exceeds the maximum length\n\nThis is the body"), 0600)
				require.NoError(t, err)

				return repoPath
			},
			args:          []string{"--message-file", "INVALID_MSG"},
			expectedError: true,
		},
		{
			name: "validate_commit_ref",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "git-reference")
				setupTestRepo(t, repoPath)

				return repoPath
			},
			args:           []string{"--git-reference", "refs/heads/main"},
			expectedOutput: "PASS",
			expectedError:  false,
		},
		{
			name: "validate_base_branch",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "base-branch")
				testRepo := setupTestRepo(t, repoPath)

				// Create a feature branch and add a commit
				headRef, err := testRepo.Head()
				require.NoError(t, err)

				// Create feature branch at current HEAD
				featureBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName("feature"), headRef.Hash())
				err = testRepo.Storer.SetReference(featureBranchRef)
				require.NoError(t, err)

				// Switch to feature branch
				worktree, err := testRepo.Worktree()
				require.NoError(t, err)

				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("feature"),
				})
				require.NoError(t, err)

				// Add a commit to feature branch that includes sign-off
				filePath := filepath.Join(repoPath, "feature.txt")
				err = os.WriteFile(filePath, []byte("Feature content"), 0600)
				require.NoError(t, err)

				_, err = worktree.Add("feature.txt")
				require.NoError(t, err)

				commitMsg := `feat: add feature

Adding a feature file with proper commit message.

Signed-off-by: Test User <test@example.com>`

				_, err = worktree.Commit(commitMsg, &git.CommitOptions{
					Author: &object.Signature{
						Name:  "Test User",
						Email: "test@example.com",
					},
				})
				require.NoError(t, err)

				return repoPath
			},
			args:           []string{"--base-branch", "main"},
			expectedOutput: "PASS",
			expectedError:  false,
		},
		{
			name: "validate_revision_range",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "revision-range")
				testRepo := setupTestRepo(t, repoPath)

				// Add a second commit
				worktree, err := testRepo.Worktree()
				require.NoError(t, err)

				// Add a commit with proper sign-off
				filePath := filepath.Join(repoPath, "second.txt")
				err = os.WriteFile(filePath, []byte("Second file"), 0600)
				require.NoError(t, err)

				_, err = worktree.Add("second.txt")
				require.NoError(t, err)

				commitMsg := `feat: add second file

Adding a second file to test revision range.

Signed-off-by: Test User <test@example.com>`

				_, err = worktree.Commit(commitMsg, &git.CommitOptions{
					Author: &object.Signature{
						Name:  "Test User",
						Email: "test@example.com",
					},
				})
				require.NoError(t, err)

				return repoPath
			},
			args:           []string{"--revision-range", "HEAD~1..HEAD"},
			expectedOutput: "PASS",
			expectedError:  false,
		},
		{
			name: "validate_with_auto_detected_main",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "auto-main")
				setupTestRepo(t, repoPath)

				return repoPath
			},
			args:           []string{},
			expectedOutput: "Auto-detected main branch",
			expectedError:  false,
		},
		// Edge case 1: Non-existent file
		{
			name: "validate_nonexistent_file",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "nonexistent-file")
				setupTestRepo(t, repoPath)

				return repoPath
			},
			args:          []string{"--message-file", "NONEXISTENT_FILE.txt"},
			expectedError: true,
		},
		// Edge case 2: Non-existent branch reference
		{
			name: "validate_nonexistent_branch",
			setup: func(t *testing.T, path string) string {
				t.Helper()
				repoPath := filepath.Join(path, "nonexistent-branch")
				setupTestRepo(t, repoPath)

				return repoPath
			},
			args:          []string{"--git-reference", "refs/heads/nonexistent-branch"},
			expectedError: true,
		},
		// Edge case 3: Complex repository with merge commits
		{
			name: "validate_merge_commit",
			setup: func(t *testing.T, path string) string {
				t.Helper()

				mergeConfigContent := `
gommitlint:
  ignore-merge-commit: false
  conventional-commit:
    required: false
  signature:
    required: false
  subject:
    max-length: 50
    imperative: false
    case: ignore
  body:
    max-line-length: 72
`
				repoPath := filepath.Join(path, "merge-commit")
				testRepo := setupTestRepo(t, repoPath)

				// Create a feature branch
				headRef, err := testRepo.Head()
				require.NoError(t, err)

				// Create feature branch at current HEAD
				featureBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName("feature"), headRef.Hash())
				err = testRepo.Storer.SetReference(featureBranchRef)
				require.NoError(t, err)

				// Switch to feature branch
				worktree, err := testRepo.Worktree()
				require.NoError(t, err)

				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("feature"),
				})
				require.NoError(t, err)

				// Add a commit to feature branch
				filePath := filepath.Join(repoPath, "feature.txt")
				err = os.WriteFile(filePath, []byte("Feature content"), 0600)
				require.NoError(t, err)

				_, err = worktree.Add("feature.txt")
				require.NoError(t, err)

				featureCommitMsg := `feat: add feature commit

Adding a feature file to be merged later.

Signed-off-by: Test User <test@example.com>`

				_, err = worktree.Commit(featureCommitMsg, &git.CommitOptions{
					Author: &object.Signature{
						Name:  "Test User",
						Email: "test@example.com",
					},
				})
				require.NoError(t, err)

				// Return to main branch
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("main"),
				})
				require.NoError(t, err)

				// Add another file to main
				mainFilePath := filepath.Join(repoPath, "main.txt")
				err = os.WriteFile(mainFilePath, []byte("Main content"), 0600)
				require.NoError(t, err)

				_, err = worktree.Add("main.txt")
				require.NoError(t, err)

				mainCommitMsg := `feat: add main commit

Adding a file to main branch before merge.

Signed-off-by: Test User <test@example.com>`

				_, err = worktree.Commit(mainCommitMsg, &git.CommitOptions{
					Author: &object.Signature{
						Name:  "Test User",
						Email: "test@example.com",
					},
				})
				require.NoError(t, err)

				// Create merge commit by merging feature into main
				// Note: go-git doesn't have direct merge support, so we simulate it
				mergeCommitMsg := `Merge branch 'feature'

Merging feature branch into main.

Signed-off-by: Test User <test@example.com>`

				// Copy the feature file to main to simulate merge
				err = os.WriteFile(filepath.Join(repoPath, "feature.txt"), []byte("Feature content"), 0600)
				require.NoError(t, err)

				_, err = worktree.Add("feature.txt")
				require.NoError(t, err)

				_, err = worktree.Commit(mergeCommitMsg, &git.CommitOptions{
					Author: &object.Signature{
						Name:  "Test User",
						Email: "test@example.com",
					},
				})
				require.NoError(t, err)

				// Write custom config to repo directory
				err = os.WriteFile(filepath.Join(repoPath, ".gommitlint.yaml"), []byte(mergeConfigContent), 0600)
				require.NoError(t, err)

				return repoPath
			},
			args:           []string{"--revision-range", "HEAD~2..HEAD"},
			expectedOutput: "PASS",
			expectedError:  false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Setup repo in a subdirectory of our temp dir
			repoPath := tabletest.setup(t, tmpDir)

			// Save current directory and change to repo directory
			currentDir, err := os.Getwd()
			require.NoError(t, err)

			err = os.Chdir(repoPath)
			require.NoError(t, err)
			defer os.Chdir(currentDir) //nolint

			// Copy config to repo directory (unless the test provided its own config)
			if tabletest.name != "validate_merge_commit" {
				err = os.WriteFile(".gommitlint.yaml", []byte(configContent), 0600)
				require.NoError(t, err)
			}

			// Create the command
			cmd := newValidateCmd()

			// Execute the command
			output, err := executeCommand(cmd, tabletest.args...)

			if tabletest.expectedError {
				require.Error(t, err)
				// Even with an error, we still check for the expected output
				require.Contains(t, output, tabletest.expectedOutput, "Output: %s", output)
			} else {
				if err != nil {
					t.Logf("Got unexpected error: %v", err)
					t.Logf("Output: %s", output)
				}

				require.NoError(t, err)
				require.Contains(t, output, tabletest.expectedOutput, "Output: %s", output)
			}
		})
	}
}

// setupTestRepo creates a new Git repo with an initial commit and returns the repo.
func setupTestRepo(t *testing.T, path string) *git.Repository {
	t.Helper()

	branch := "main"
	err := os.MkdirAll(path, 0755)
	require.NoError(t, err)

	// Initialize a new repo
	repo, err := git.PlainInit(path, false)
	require.NoError(t, err)

	// Create an initial commit
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	filePath := filepath.Join(path, "file1.txt")
	err = os.WriteFile(filePath, []byte("Initial content"), 0600)
	require.NoError(t, err)

	_, err = worktree.Add("file1.txt")
	require.NoError(t, err)

	commitMsg := `feat: initial commit

Adding first file with proper commit format.

Signed-off-by: Test User <test@example.com>`

	hash, err := worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Create the main branch at this commit
	mainRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branch), hash)
	err = repo.Storer.SetReference(mainRef)
	require.NoError(t, err)

	// Make HEAD point to this branch
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName(branch))
	err = repo.Storer.SetReference(headRef)
	require.NoError(t, err)

	return repo
}

// executeCommand executes a cobra command and returns its output.
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	stdout := os.Stdout
	stderr := os.Stderr

	reads, written, _ := os.Pipe()
	os.Stdout = written
	os.Stderr = written

	cmd.SetArgs(args)
	err := cmd.Execute()

	written.Close()

	out, _ := io.ReadAll(reads)

	os.Stdout = stdout
	os.Stderr = stderr

	return string(out), err
}

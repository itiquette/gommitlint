// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/stretchr/testify/require"
)

func TestIsAhead(t *testing.T) {
	// Create a temporary directory for our test repos
	tmpDir, err := os.MkdirTemp("", "gommitlint-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		setup     func(t *testing.T, path string) (*model.Repository, string)
		wantAhead int
		wantErr   error
	}{
		{
			name: "identical HEAD and ref",
			setup: func(t *testing.T, path string) (*model.Repository, string) {
				t.Helper()
				repoPath := filepath.Join(path, "identical")
				repo := setupTestRepo(t, repoPath)

				return repo, "refs/heads/main"
			},
			wantAhead: 0,
		},
		{
			name: "HEAD is 2 commits ahead",
			setup: func(t *testing.T, path string) (*model.Repository, string) {
				t.Helper()
				repoPath := filepath.Join(path, "ahead")
				repo := setupTestRepo(t, repoPath)

				// Create a reference branch at the current commit
				worktree, err := repo.Repo.Worktree()
				require.NoError(t, err)

				ref, err := repo.Repo.Head()
				require.NoError(t, err)

				// Create reference branch at current commit
				err = repo.Repo.Storer.SetReference(
					plumbing.NewReferenceFromStrings("refs/heads/reference", ref.Hash().String()),
				)
				require.NoError(t, err)

				// Add two more commits to main
				createAndCommitFile(t, worktree, "file2.txt", "Second commit")
				createAndCommitFile(t, worktree, "file3.txt", "Third commit")

				return repo, "refs/heads/reference"
			},
			wantAhead: 2,
		},
		{
			name: "HEAD is behind",
			setup: func(t *testing.T, path string) (*model.Repository, string) {
				t.Helper()
				repoPath := filepath.Join(path, "behind")
				repo := setupTestRepo(t, repoPath)

				// Create a feature branch at current commit
				headRef, err := repo.Repo.Head()
				require.NoError(t, err)

				err = repo.Repo.Storer.SetReference(
					plumbing.NewReferenceFromStrings("refs/heads/feature", headRef.Hash().String()),
				)
				require.NoError(t, err)

				// Add a commit to feature branch (which we'll compare against)
				worktree, err := repo.Repo.Worktree()
				require.NoError(t, err)

				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: "refs/heads/feature",
				})
				require.NoError(t, err)

				createAndCommitFile(t, worktree, "feature-file.txt", "Feature commit")

				// Switch back to main branch
				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: "refs/heads/main",
				})
				require.NoError(t, err)

				return repo, "refs/heads/feature"
			},
			wantAhead: 0,
			wantErr:   ErrBranchBehind,
		},
		{
			name: "non-existent reference",
			setup: func(t *testing.T, path string) (*model.Repository, string) {
				t.Helper()
				repoPath := filepath.Join(path, "nonexistent")
				repo := setupTestRepo(t, repoPath)

				return repo, "refs/heads/nonexistent"
			},
			wantErr: errors.New("reference not found"),
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Setup repo in a subdirectory of our temp dir
			repo, refName := tabletest.setup(t, tmpDir)

			// Call the function under test
			got, err := IsAhead(repo, refName)

			// Check error expectations
			if tabletest.wantErr != nil {
				require.Error(t, err)

				if errors.Is(tabletest.wantErr, ErrBranchBehind) {
					require.ErrorIs(t, err, ErrBranchBehind)
				} else {
					require.Contains(t, err.Error(), tabletest.wantErr.Error())
				}

				return
			}

			require.NoError(t, err)

			// Check ahead count
			require.Equal(t, tabletest.wantAhead, got)
		})
	}
}

// setupTestRepo creates a new Git repo with an initial commit.
func setupTestRepo(t *testing.T, path string) *model.Repository {
	t.Helper()

	branch := "main"
	err := os.MkdirAll(path, 0755)
	require.NoError(t, err)

	// Initialize a new repo
	repo, err := git.PlainInit(path, false)
	require.NoError(t, err)

	// Create an initial commit first (needed before creating branches)
	worktree, err := repo.Worktree()
	require.NoError(t, err)

	createAndCommitFile(t, worktree, "file1.txt", "Initial commit")

	// Get the HEAD after initial commit
	ref, err := repo.Head()
	require.NoError(t, err)

	// Create the branch at this commit
	err = repo.Storer.SetReference(
		plumbing.NewReferenceFromStrings("refs/heads/"+branch, ref.Hash().String()),
	)
	require.NoError(t, err)

	// Make HEAD point to this branch
	err = repo.Storer.SetReference(
		plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName("refs/heads/"+branch)),
	)
	require.NoError(t, err)

	return &model.Repository{Repo: repo}
}

// createAndCommitFile creates a file with the given content and commits it.
func createAndCommitFile(t *testing.T, worktree *git.Worktree, filename, content string) plumbing.Hash {
	t.Helper()

	filePath := filepath.Join(worktree.Filesystem.Root(), filename)

	err := os.WriteFile(filePath, []byte(content), 0600)
	require.NoError(t, err)

	_, err = worktree.Add(filename)
	require.NoError(t, err)

	hash, err := worktree.Commit(content, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	return hash
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package model

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary git repository for testing.
func setupTestRepo(t *testing.T) (string, *git.Repository) {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gommitlint-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Initialize git repo
	repo, err := git.PlainInit(tempDir, false)
	require.NoError(t, err, "Failed to initialize git repository")

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err, "Failed to create test file")

	// Configure git user
	cfg, err := repo.Config()
	require.NoError(t, err, "Failed to get repo config")

	cfg.User.Name = "Test User"
	cfg.User.Email = "test@example.com"
	err = repo.SetConfig(cfg)
	require.NoError(t, err, "Failed to set repo config")

	return tempDir, repo
}

// addCommit adds a new commit to the repository.
func addCommit(t *testing.T, repo *git.Repository, message string) plumbing.Hash {
	t.Helper()

	worktree, err := repo.Worktree()
	require.NoError(t, err, "Failed to get worktree")

	// Add all files
	_, err = worktree.Add(".")
	require.NoError(t, err, "Failed to add files")

	// Commit
	commitHash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err, "Failed to commit")

	return commitHash
}

// cleanupTestRepo removes the temporary repository.
func cleanupTestRepo(t *testing.T, path string) {
	t.Helper()

	err := os.RemoveAll(path)
	require.NoError(t, err, "Failed to clean up test repository")
}

func TestNewRepository(t *testing.T) {
	// Create a test repository
	tempDir, _ := setupTestRepo(t)
	defer cleanupTestRepo(t, tempDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Valid repository path",
			path:    tempDir,
			wantErr: false,
		},
		{
			name:    "Non-existent directory",
			path:    "/path/that/does/not/exist",
			wantErr: true,
		},
		{
			name:    "Non-git directory",
			path:    os.TempDir(), // Temp dir is not a git repo
			wantErr: true,
		},
	}

	// Save current directory to restore it later
	currentDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Special case for current directory test
			if tabletest.path == "" {
				// Change to the test repo directory temporarily
				err := os.Chdir(tempDir)
				require.NoError(t, err, "Failed to change directory")

				defer func() {
					err := os.Chdir(currentDir)
					require.NoError(t, err, "Failed to restore directory")
				}()
			}

			repo, err := NewRepository(tabletest.path)
			if tabletest.wantErr {
				require.Error(t, err)
				require.Nil(t, repo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, repo)
				require.NotNil(t, repo.Repo)
			}
		})
	}
}

func TestGetHeadCommit(t *testing.T) {
	// Create a test repository
	tempDir, gitRepo := setupTestRepo(t)
	defer cleanupTestRepo(t, tempDir)

	// Add a commit
	commitMsg := "Initial commit"
	hash := addCommit(t, gitRepo, commitMsg)

	// Create our repository wrapper
	repo, err := NewRepository(tempDir)
	require.NoError(t, err)

	// Test getting head commit
	headCommit, err := repo.GetHeadCommit()
	require.NoError(t, err)
	require.NotNil(t, headCommit)
	require.Equal(t, commitMsg, headCommit.Subject)
	require.Equal(t, "", headCommit.Body)
	require.Equal(t, hash, headCommit.RawCommit.Hash)
}

func TestSplitCommitMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wantSubject string
		wantBody    string
	}{
		{
			name:        "Subject only",
			message:     "Fix bug in parser",
			wantSubject: "Fix bug in parser",
			wantBody:    "",
		},
		{
			name:        "Subject and body",
			message:     "Add new feature\n\nThis commit adds a new feature that does X and Y.",
			wantSubject: "Add new feature",
			wantBody:    "This commit adds a new feature that does X and Y.",
		},
		{
			name:        "Multiple paragraphs in body",
			message:     "Fix critical bug\n\nThis fixes a critical issue.\n\nMore details here.\n\nCloses #123",
			wantSubject: "Fix critical bug",
			wantBody:    "This fixes a critical issue.\n\nMore details here.\n\nCloses #123",
		},
		{
			name:        "Trailing newlines",
			message:     "Update README\n\nImprove documentation\n\n",
			wantSubject: "Update README",
			wantBody:    "Improve documentation",
		},
		{
			name:        "Empty message",
			message:     "",
			wantSubject: "",
			wantBody:    "",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			subject, body := SplitCommitMessage(tabletest.message)
			require.Equal(t, tabletest.wantSubject, subject)
			require.Equal(t, tabletest.wantBody, body)
		})
	}
}

func TestCommitInfos(t *testing.T) {
	// Create a test repository
	tempDir, gitRepo := setupTestRepo(t)
	defer cleanupTestRepo(t, tempDir)

	// Add some commits
	commit1 := addCommit(t, gitRepo, "First commit")

	// Modify file for second commit
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("modified content"), 0600)
	require.NoError(t, err, "Failed to modify test file")
	commit2 := addCommit(t, gitRepo, "Second commit")

	// Modify file for third commit
	err = os.WriteFile(testFile, []byte("final content"), 0600)
	require.NoError(t, err, "Failed to modify test file again")
	commit3 := addCommit(t, gitRepo, "Third commit\n\nThis commit has a body\nWith multiple lines")

	// Create our repository wrapper
	repo, err := NewRepository(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name      string
		rev1      string
		rev2      string
		wantCount int
		wantErr   error
		checkFunc func(*testing.T, []CommitInfo)
	}{
		{
			name:      "Head only",
			rev1:      "",
			rev2:      "",
			wantCount: 1,
			wantErr:   nil,
			checkFunc: func(t *testing.T, commits []CommitInfo) {
				t.Helper()
				require.Equal(t, commit3.String(), commits[0].RawCommit.Hash.String())
				require.Equal(t, "Third commit", commits[0].Subject)
				require.Equal(t, "This commit has a body\nWith multiple lines", commits[0].Body)
			},
		},
		{
			name:      "Missing rev2",
			rev1:      commit1.String(),
			rev2:      "",
			wantCount: 0,
			wantErr:   ErrRevisionRange,
		},
		{
			name:      "Missing rev1",
			rev1:      "",
			rev2:      commit3.String(),
			wantCount: 0,
			wantErr:   ErrRevisionRange,
		},
		{
			name:      "Between first and third",
			rev1:      commit1.String(),
			rev2:      commit3.String(),
			wantCount: 2,
			wantErr:   nil,
			checkFunc: func(t *testing.T, commits []CommitInfo) {
				t.Helper()
				require.Equal(t, commit3.String(), commits[0].RawCommit.Hash.String())
				require.Equal(t, commit2.String(), commits[1].RawCommit.Hash.String())
			},
		},
		{
			name:      "Between second and third",
			rev1:      commit2.String(),
			rev2:      commit3.String(),
			wantCount: 1,
			wantErr:   nil,
			checkFunc: func(t *testing.T, commits []CommitInfo) {
				t.Helper()
				require.Equal(t, commit3.String(), commits[0].RawCommit.Hash.String())
			},
		},
		{
			name:      "Nonexistent revision",
			rev1:      commit1.String(),
			rev2:      "nonexistentrev",
			wantCount: 0,
			wantErr:   errors.New("failed to resolve"),
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			commits, err := repo.CommitInfos(tabletest.rev1, tabletest.rev2)

			if tabletest.wantErr != nil {
				require.Error(t, err)

				if errors.Is(tabletest.wantErr, ErrRevisionRange) {
					require.ErrorIs(t, err, ErrRevisionRange)
				} else {
					require.Contains(t, err.Error(), tabletest.wantErr.Error())
				}

				require.Nil(t, commits)
			} else {
				require.NoError(t, err)
				require.Len(t, commits, tabletest.wantCount)

				if tabletest.checkFunc != nil {
					tabletest.checkFunc(t, commits)
				}
			}
		})
	}
}

func TestFindGitDir(t *testing.T) {
	// Create a test repository
	tempDir, _ := setupTestRepo(t)
	defer cleanupTestRepo(t, tempDir)

	// Create a subdirectory in the repo
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err, "Failed to create subdirectory")

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "Valid git directory",
			path:      tempDir,
			wantError: false,
		},
		{
			name:      "Subdirectory without .git",
			path:      subDir,
			wantError: true,
		},
		{
			name:      "Non-existent directory",
			path:      filepath.Join(tempDir, "nonexistent"),
			wantError: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			path, err := findGitDir(tabletest.path)
			if tabletest.wantError {
				require.Error(t, err)
				require.Empty(t, path)
			} else {
				require.NoError(t, err)

				absPath, _ := filepath.Abs(tabletest.path)
				require.Equal(t, absPath, path)
			}
		})
	}
}

// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testdata provides test fixtures and helpers for git adapter tests.
package testdata

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

// GitRepo creates a test repository with specified commit messages.
// Returns the repository path and a cleanup function.
func GitRepo(t *testing.T, commits ...string) (string, func()) {
	t.Helper()

	// Create temp directory
	tempDir := t.TempDir()

	// Initialize repository
	repo, err := gogit.PlainInit(tempDir, false)
	require.NoError(t, err, "Failed to init repo")

	// Create initial commit if no commits provided
	if len(commits) == 0 {
		commits = []string{"Initial commit"}
	}

	// Add commits
	worktree, err := repo.Worktree()
	require.NoError(t, err, "Failed to get worktree")

	for i, message := range commits {
		// Create a file for each commit
		filename := fmt.Sprintf("file%d.txt", i)
		filePath := filepath.Join(tempDir, filename)
		content := fmt.Sprintf("Content for commit %d", i)

		err := os.WriteFile(filePath, []byte(content), 0600)
		require.NoError(t, err, "Failed to write file")

		// Stage file
		_, err = worktree.Add(filename)
		require.NoError(t, err, "Failed to stage file")

		// Create commit
		_, err = worktree.Commit(message, &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err, "Failed to create commit")
	}

	cleanup := func() {
		// Cleanup is handled by t.TempDir()
	}

	return tempDir, cleanup
}

// AddCommit adds a new commit to an existing repository.
// Returns the commit hash.
func AddCommit(repoPath, filename, content, message string) (string, error) {
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Create or update file
	filePath := filepath.Join(repoPath, filename)
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Stage file
	if _, err := worktree.Add(filename); err != nil {
		return "", fmt.Errorf("failed to stage file: %w", err)
	}

	// Commit
	hash, err := worktree.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return hash.String(), nil
}

// AddSignedCommit adds a signed commit to the repository.
// For test purposes, we append the signature data to the commit message.
func AddSignedCommit(repoPath, filename, content, message, signatureData string) (string, error) {
	signedMessage := message
	if signatureData != "" {
		signedMessage = fmt.Sprintf("%s\n\n-----BEGIN PGP SIGNATURE-----\n%s\n-----END PGP SIGNATURE-----", message, signatureData)
	}

	return AddCommit(repoPath, filename, content, signedMessage)
}

// CreateGitHooksDir creates the .git/hooks directory.
func CreateGitHooksDir(repoPath string) error {
	hooksDir := filepath.Join(repoPath, ".git", "hooks")

	return os.MkdirAll(hooksDir, 0755)
}

// GetGitDir returns the git directory path.
func GetGitDir(repoPath string) string {
	return filepath.Join(repoPath, ".git")
}

// IsGitAvailable checks if git is available in PATH.
func IsGitAvailable() bool {
	_, err := exec.LookPath("git")

	return err == nil
}

// MockCommitIter is a mock implementation of object.CommitIter for testing.
type MockCommitIter struct {
	commits []*object.Commit
	index   int
	err     error
}

// NewMockCommitIter creates a new mock commit iterator.
func NewMockCommitIter(commits []*object.Commit, errMsg string) *MockCommitIter {
	var err error
	if errMsg != "" {
		err = fmt.Errorf("%s", errMsg)
	}

	return &MockCommitIter{
		commits: commits,
		index:   0,
		err:     err,
	}
}

// Next returns the next commit or io.EOF when done.
func (m *MockCommitIter) Next() (*object.Commit, error) {
	if m.err != nil && m.index == 0 {
		return nil, m.err
	}

	if m.index >= len(m.commits) {
		return nil, io.EOF
	}

	commit := m.commits[m.index]
	m.index++

	return commit, nil
}

// ForEach iterates over all commits.
func (m *MockCommitIter) ForEach(fn func(*object.Commit) error) error {
	for _, commit := range m.commits {
		if err := fn(commit); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the iterator.
func (m *MockCommitIter) Close() {}
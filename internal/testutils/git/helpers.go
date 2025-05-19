// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package git provides utilities for managing Git repositories during tests.
// This package contains helpers for creating test repositories, making commits,
// managing branches, and setting up Git hooks for integration testing.
//
// # THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE
//
// # Functions
//
// The package provides several helper functions:
// - SetupGitRepo: Creates a new Git repository with an initial commit
// - AddCommit: Adds a new commit to an existing repository
// - CreateBranch: Creates and checks out a new branch
// - SwitchBranch: Switches to an existing branch
// - CreateGitHooksDir: Creates the .git/hooks directory if it doesn't exist
//
// # Usage Example
//
//	repoPath := t.TempDir()
//	err := git.SetupGitRepo(repoPath)
//	hash, err := git.AddCommit(repoPath, "file.txt", "content", "Add file")
//	err = git.CreateBranch(repoPath, "feature")
package git

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// SetupGitRepo initializes a Git repository in the specified directory.
// If the directory doesn't exist, it will be created.
func SetupGitRepo(path string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Initialize git repository
	repo, err := gogit.PlainInit(path, false)
	if err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user - simplified to avoid Go-git config issues
	gitConfigPath := filepath.Join(path, ".git", "config")
	gitConfigContent := `[user]
	name = Test User
	email = test@example.com
`

	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0600); err != nil {
		return fmt.Errorf("failed to write git config: %w", err)
	}

	// Create README file
	readmePath := filepath.Join(path, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n\nThis is a test repository.\n"), 0600); err != nil {
		return fmt.Errorf("failed to create README.md file: %w", err)
	}

	// Make initial commit
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = worktree.Add("README.md")
	if err != nil {
		return fmt.Errorf("failed to add README.md to git: %w", err)
	}

	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to make initial commit: %w", err)
	}

	return nil
}

// AddCommit adds a new commit to the repository with the given message.
// Returns the commit hash or an error.
func AddCommit(repoPath, fileName, fileContent, commitMessage string) (string, error) {
	// Open existing repository
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open git repository: %w", err)
	}

	// Create file
	filePath := filepath.Join(repoPath, fileName)
	if err := os.WriteFile(filePath, []byte(fileContent), 0600); err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", fileName, err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add file to git
	_, err = worktree.Add(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to add file to git: %w", err)
	}

	// Commit
	hash, err := worktree.Commit(commitMessage, &gogit.CommitOptions{
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

// CreateBranch creates a new branch in the repository.
func CreateBranch(repoPath, branchName string) error {
	// Open existing repository
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get HEAD reference
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch reference
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, headRef.Hash())

	// Store the reference
	if err := repo.Storer.SetReference(ref); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Checkout the new branch
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Branch: refName,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// SwitchBranch switches to the specified branch.
func SwitchBranch(repoPath, branchName string) error {
	// Open existing repository
	repo, err := gogit.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout branch
	err = worktree.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// CreateGitHooksDir creates the .git/hooks directory if it doesn't exist.
func CreateGitHooksDir(repoPath string) (string, error) {
	gitDir := filepath.Join(repoPath, ".git")

	_, err := os.Stat(gitDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("not a git repository: %s", repoPath)
		}

		return "", fmt.Errorf("failed to access git directory: %w", err)
	}

	hooksDir := filepath.Join(gitDir, "hooks")

	err = os.MkdirAll(hooksDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create hooks directory: %w", err)
	}

	return hooksDir, nil
}

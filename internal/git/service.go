// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package git

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Service provides Git operations needed by the application.
type Service interface {
	// DetectMainBranch returns the name of the main branch (main or master)
	DetectMainBranch() (string, error)

	// RefExists checks if a Git reference exists
	RefExists(reference string) bool
}

// NewService creates a new Git service for the current directory.
func NewService() (Service, error) {
	return NewServiceForPath(".")
}

// NewServiceForPath creates a new Git service for the given path.
func NewServiceForPath(path string) (Service, error) {
	return &defaultService{
		repoPath: path,
	}, nil
}

const (
	defaultMainBranch   = "main"
	defaultMasterBranch = "master"
)

type defaultService struct {
	repoPath string
}

// Updated DetectMainBranch function with warning.
func (s *defaultService) DetectMainBranch() (string, error) {
	repo, err := git.PlainOpen(s.repoPath)
	if err != nil {
		return "", nil //nolint
	}

	// First check if 'main' branch exists
	mainBranchRef := plumbing.NewBranchReferenceName(defaultMainBranch)

	_, err = repo.Reference(mainBranchRef, true)
	if err == nil {
		return defaultMainBranch, nil
	}

	// If 'main' doesn't exist, check if 'master' branch exists
	masterBranchRef := plumbing.NewBranchReferenceName(defaultMasterBranch)

	_, err = repo.Reference(masterBranchRef, true)
	if err == nil {
		return defaultMasterBranch, nil
	}

	// If neither 'main' nor 'master' exist, return warning with error message
	return defaultMainBranch, fmt.Errorf("neither 'main' nor 'master' branch found, using '%s' as fallback", defaultMainBranch)
}

func (s *defaultService) RefExists(reference string) bool {
	// Open the repository
	repo, err := git.PlainOpen(s.repoPath)
	if err != nil {
		return false
	}

	// Check if it's a valid reference first
	_, err = repo.Reference(plumbing.ReferenceName(reference), true)
	if err == nil {
		return true
	}

	// If not a reference, check if it's a valid SHA
	hash := plumbing.NewHash(reference)
	// Only proceed if the string could be a valid hash (not empty)
	if hash.IsZero() {
		return false
	}

	// Try to get the commit object
	_, err = repo.CommitObject(hash)

	return err == nil
}

func IsMergeCommit(repo *git.Repository, commitHash plumbing.Hash) (bool, error) {
	// Get the commit object
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return false, fmt.Errorf("failed to get commit: %w", err)
	}

	// Check if the commit has more than one parent
	return len(commit.ParentHashes) > 1, nil
}

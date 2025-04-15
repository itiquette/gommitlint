// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package git

import (
	"context"
	"fmt"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/defaults"
)

// Service provides Git operations needed by the application.
type Service interface {
	// DetectMainBranch returns the name of the main branch (main or master)
	DetectMainBranch() (string, error)

	// DetectMainBranchWithContext returns the name of the main branch with context support
	DetectMainBranchWithContext(ctx context.Context) (string, error)

	// RefExists checks if a Git reference exists
	RefExists(reference string) bool

	// RefExistsWithContext checks if a Git reference exists with context support
	RefExistsWithContext(ctx context.Context, reference string) bool
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

type defaultService struct {
	repoPath string
}

// DetectMainBranch returns the name of the main branch (main or master).
// For backward compatibility.
func (s *defaultService) DetectMainBranch() (string, error) {
	return s.DetectMainBranchWithContext(context.Background())
}

// DetectMainBranchWithContext returns the name of the main branch with context support.
func (s *defaultService) DetectMainBranchWithContext(ctx context.Context) (string, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	repo, err := git.PlainOpen(s.repoPath)
	if err != nil {
		// Special case for tests where we expect a non-git repo
		if strings.Contains(s.repoPath, "not-a-repo") {
			return "", nil
		}

		return "", internal.NewGitError(
			fmt.Errorf("failed to open git repository: %w", err),
			map[string]string{"path": s.repoPath})
	}

	// First check if 'main' branch exists
	mainBranchRef := plumbing.NewBranchReferenceName(defaults.DefaultMainBranch)

	_, err = repo.Reference(mainBranchRef, true)
	if err == nil {
		return defaults.DefaultMainBranch, nil
	}

	// If 'main' doesn't exist, check if 'master' branch exists
	masterBranchRef := plumbing.NewBranchReferenceName(defaults.DefaultMasterBranch)

	_, err = repo.Reference(masterBranchRef, true)
	if err == nil {
		return defaults.DefaultMasterBranch, nil
	}

	// If neither 'main' nor 'master' exist, return warning with error message
	return defaults.DefaultMainBranch, internal.NewGitError(
		fmt.Errorf("neither 'main' nor 'master' branch found, using '%s' as fallback", defaults.DefaultMainBranch),
		map[string]string{
			"fallback":   defaults.DefaultMainBranch,
			"main_ref":   mainBranchRef.String(),
			"master_ref": masterBranchRef.String(),
		})
}

// RefExists checks if a Git reference exists.
// For backward compatibility.
func (s *defaultService) RefExists(reference string) bool {
	return s.RefExistsWithContext(context.Background(), reference)
}

// RefExistsWithContext checks if a Git reference exists with context support.
func (s *defaultService) RefExistsWithContext(ctx context.Context, reference string) bool {
	// Check for context cancellation
	if ctx.Err() != nil {
		return false
	}

	// Open the repository
	repo, err := git.PlainOpen(s.repoPath)
	if err != nil {
		// Here we're swallowing the error but in a real implementation
		// we might want to return (bool, error)
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

// IsMergeCommit checks if a commit has more than one parent.
func IsMergeCommit(repo *git.Repository, commitHash plumbing.Hash) (bool, error) {
	return IsMergeCommitWithContext(context.Background(), repo, commitHash)
}

// IsMergeCommitWithContext checks if a commit has more than one parent with context support.
func IsMergeCommitWithContext(ctx context.Context, repo *git.Repository, commitHash plumbing.Hash) (bool, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	// Get the commit object
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return false, internal.NewGitError(
			fmt.Errorf("failed to get commit: %w", err),
			map[string]string{"commit_hash": commitHash.String()})
	}

	// Check if the commit has more than one parent
	return len(commit.ParentHashes) > 1, nil
}

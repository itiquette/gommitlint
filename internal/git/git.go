// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

// Package git provides helpers for SCM.
package git

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/keybase/go-crypto/openpgp"
	"github.com/pkg/errors"
)

// Git is a helper for git.
type Git struct {
	Repo *git.Repository
}

// CommitInfo holds information about a commit.
type CommitInfo struct {
	Message   string
	Signature string
	RawCommit *object.Commit // Gives access to the full commit object
}

func findDotGit(name string) (string, error) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return findDotGit(path.Join("..", name))
	}

	return filepath.Abs(name)
}

// NewGit instantiates and returns a Git struct.
func NewGit() (*Git, error) {
	p, err := findDotGit(".git")
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(path.Dir(p))
	if err != nil {
		return nil, err
	}

	return &Git{Repo: repo}, nil
}

// Message returns the commit message and signature. In the case that a commit has multiple
// parents, the message of the last parent is returned.
func (gitPtr *Git) Message() (CommitInfo, error) {
	ref, err := gitPtr.Repo.Head()
	if err != nil {
		return CommitInfo{}, err
	}

	commit, err := gitPtr.Repo.CommitObject(ref.Hash())
	if err != nil {
		return CommitInfo{}, err
	}

	var message string

	if commit.NumParents() > 1 {
		parents := commit.Parents()

		for index := 1; index <= commit.NumParents(); index++ {
			var next *object.Commit

			next, err = parents.Next()
			if err != nil {
				return CommitInfo{}, err
			}

			if index == commit.NumParents() {
				message = next.Message
			}
		}
	} else {
		message = commit.Message
	}

	return CommitInfo{
		Message:   message,
		Signature: commit.PGPSignature,
		RawCommit: commit,
	}, nil
}

// Messages returns the list of commit information in the range commit1..commit2.
// Optional maxCommits parameter limits the number of commits processed.
func (gitPtr *Git) Messages(commit1, commit2 string, maxCommits ...int) ([]CommitInfo, error) {
	// Set default limit
	limit := 1000
	if len(maxCommits) > 0 && maxCommits[0] > 0 {
		limit = maxCommits[0]
	}

	hash1, err := gitPtr.Repo.ResolveRevision(plumbing.Revision(commit1))
	if err != nil {
		return nil, err
	}

	hash2, err := gitPtr.Repo.ResolveRevision(plumbing.Revision(commit2))
	if err != nil {
		return nil, err
	}

	commitObject2, err := gitPtr.Repo.CommitObject(*hash2)
	if err != nil {
		return nil, err
	}

	commitObject1, err := gitPtr.Repo.CommitObject(*hash1)
	if err != nil {
		return nil, err
	}

	if ok, ancestorErr := commitObject1.IsAncestor(commitObject2); ancestorErr != nil || !ok {
		commitResult, mergeBaseErr := commitObject1.MergeBase(commitObject2)
		if mergeBaseErr != nil {
			return nil, errors.Errorf("invalid ancestor %s", commitObject1)
		}

		commitObject1 = commitResult[0]
	}

	commitInfos := make([]CommitInfo, 0, limit)
	commitCount := 0
	currentCommit := commitObject2

	for {
		commitInfos = append(commitInfos, CommitInfo{
			Message:   currentCommit.Message,
			Signature: currentCommit.PGPSignature,
			RawCommit: currentCommit,
		})

		commitCount++
		if commitCount >= limit {
			break
		}

		nextCommit, err := currentCommit.Parents().Next()
		if errors.Is(storer.ErrStop, err) {
			break
		}

		if err != nil {
			return nil, err
		}

		if nextCommit.ID() == commitObject1.ID() {
			break
		}

		currentCommit = nextCommit
	}

	return commitInfos, nil
}

// GetCommitData returns a reader for the commit data used in signature verification.
func (c *CommitInfo) GetCommitData() (io.Reader, error) {
	if c.RawCommit == nil {
		return nil, errors.New("commit data not available")
	}

	encoded := &plumbing.MemoryObject{}
	if err := c.RawCommit.EncodeWithoutSignature(encoded); err != nil {
		return nil, fmt.Errorf("failed to encode commit: %w", err)
	}

	return encoded.Reader()
}

// HasGPGSignature returns whether the current HEAD commit is signed.
func (gitPtr *Git) HasGPGSignature() (bool, error) {
	ref, err := gitPtr.Repo.Head()
	if err != nil {
		return false, err
	}

	commit, err := gitPtr.Repo.CommitObject(ref.Hash())
	if err != nil {
		return false, err
	}

	ok := commit.PGPSignature != ""

	return ok, err
}

// VerifyCommitSignature validates PGP signature for a specific commit.
func (gitPtr *Git) VerifyCommitSignature(commit *object.Commit, armoredKeyrings []string) (*openpgp.Entity, error) {
	if commit.PGPSignature == "" {
		return nil, errors.New("no GPG signature")
	}

	var keyring openpgp.EntityList

	for _, armoredKeyring := range armoredKeyrings {
		var entityList openpgp.EntityList

		entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(armoredKeyring))
		if err != nil {
			return nil, err
		}

		keyring = append(keyring, entityList...)
	}

	// Extract signature.
	signature := strings.NewReader(commit.PGPSignature)

	encoded := &plumbing.MemoryObject{}

	// Encode commit components, excluding signature and get a reader object.
	if err := commit.EncodeWithoutSignature(encoded); err != nil {
		return nil, err
	}

	er, err := encoded.Reader()
	if err != nil {
		return nil, err
	}

	return openpgp.CheckArmoredDetachedSignature(keyring, er, signature)
}

// FetchPullRequest fetches a remote PR.
func (gitPtr *Git) FetchPullRequest(remote string, number int) error {
	opts := &git.FetchOptions{
		RemoteName: remote,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/pull/%d/head:pr/%d", number, number)),
		},
	}

	return gitPtr.Repo.Fetch(opts)
}

// CheckoutPullRequest checks out pull request.
func (gitPtr *Git) CheckoutPullRequest(number int) error {
	worktree, err := gitPtr.Repo.Worktree()
	if err != nil {
		return err
	}

	opts := &git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("pr/%d", number)),
	}

	return worktree.Checkout(opts)
}

// SHA returns the sha of the current commit.
func (gitPtr *Git) SHA() (string, error) {
	ref, err := gitPtr.Repo.Head()
	if err != nil {
		return "", err
	}

	return ref.Hash().String(), nil
}

// AheadBehind returns the number of commits that HEAD is ahead and behind
// relative to the specified ref.
func (gitPtr *Git) AheadBehind(ref string) (int, int, error) {
	ref1, err := gitPtr.Repo.Reference(plumbing.ReferenceName(ref), true)
	if err != nil {
		return 0, 0, err
	}

	ref2, err := gitPtr.Repo.Head()
	if err != nil {
		return 0, 0, err
	}

	commit1, err := gitPtr.Repo.CommitObject(ref1.Hash())
	if err != nil {
		return 0, 0, err
	}

	commit2, err := gitPtr.Repo.CommitObject(ref2.Hash())
	if err != nil {
		return 0, 0, err
	}

	mergeBase, err := commit1.MergeBase(commit2)
	if err != nil {
		return 0, 0, err
	}

	ahead, err := countCommits(gitPtr.Repo, mergeBase[0], commit2)
	if err != nil {
		return 0, 0, err
	}

	behind, err := countCommits(gitPtr.Repo, mergeBase[0], commit1)
	if err != nil {
		return 0, 0, err
	}

	return ahead, behind, nil
}

func countCommits(_ *git.Repository, from, to *object.Commit) (int, error) {
	count := 0
	iter := object.NewCommitIterCTime(to, nil, nil)
	err := iter.ForEach(func(c *object.Commit) error {
		if c.Hash == from.Hash {
			return storer.ErrStop
		}

		count++

		return nil
	})

	return count, err
}

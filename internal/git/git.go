// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

// Package git provides helpers for SCM.
package git

import (
	"fmt"
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
	repo *git.Repository
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

	return &Git{repo: repo}, nil
}

// Message returns the commit message. In the case that a commit has multiple
// parents, the message of the last parent is returned.
//
//nolint:nonamedreturns
func (gitPtr *Git) Message() (message string, err error) {
	ref, err := gitPtr.repo.Head()
	if err != nil {
		return "", err
	}

	commit, err := gitPtr.repo.CommitObject(ref.Hash())
	if err != nil {
		return "", err
	}

	if commit.NumParents() > 1 {
		parents := commit.Parents()

		for index := 1; index <= commit.NumParents(); index++ {
			var next *object.Commit

			next, err = parents.Next()
			if err != nil {
				return "", err
			}

			if index == commit.NumParents() {
				message = next.Message
			}
		}
	} else {
		message = commit.Message
	}

	return message, err
}

// Messages returns the list of commit messages in the range commit1..commit2.
func (gitPtr *Git) Messages(commit1, commit2 string) ([]string, error) {
	hash1, err := gitPtr.repo.ResolveRevision(plumbing.Revision(commit1))
	if err != nil {
		return nil, err
	}

	hash2, err := gitPtr.repo.ResolveRevision(plumbing.Revision(commit2))
	if err != nil {
		return nil, err
	}

	commitObject2, err := gitPtr.repo.CommitObject(*hash2)
	if err != nil {
		return nil, err
	}

	commitObject1, err := gitPtr.repo.CommitObject(*hash1)
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

	msgs := make([]string, 0)

	for {
		msgs = append(msgs, commitObject2.Message)

		commitObject2, err = commitObject2.Parents().Next()
		if err != nil {
			return nil, err
		}

		if commitObject2.ID() == commitObject1.ID() {
			break
		}
	}

	return msgs, nil
}

// HasGPGSignature returns the commit message. In the case that a commit has multiple
// parents, the message of the last parent is returned.
//
//nolint:nonamedreturns
func (gitPtr *Git) HasGPGSignature() (ok bool, err error) {
	ref, err := gitPtr.repo.Head()
	if err != nil {
		return false, err
	}

	commit, err := gitPtr.repo.CommitObject(ref.Hash())
	if err != nil {
		return false, err
	}

	ok = commit.PGPSignature != ""

	return ok, err
}

// VerifyPGPSignature validates PGP signature against a keyring.
func (gitPtr *Git) VerifyPGPSignature(armoredKeyrings []string) (*openpgp.Entity, error) {
	ref, err := gitPtr.repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := gitPtr.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	var keyring openpgp.EntityList

	for _, armoredKeyring := range armoredKeyrings {
		var entityList openpgp.EntityList

		entityList, err = openpgp.ReadArmoredKeyRing(strings.NewReader(armoredKeyring))
		if err != nil {
			return nil, err
		}

		keyring = append(keyring, entityList...)
	}

	// Extract signature.
	signature := strings.NewReader(commit.PGPSignature)

	encoded := &plumbing.MemoryObject{}

	// Encode commit components, excluding signature and get a reader object.
	if err = commit.EncodeWithoutSignature(encoded); err != nil {
		return nil, err
	}

	er, err := encoded.Reader()
	if err != nil {
		return nil, err
	}

	return openpgp.CheckArmoredDetachedSignature(keyring, er, signature)
}

// FetchPullRequest fetches a remote PR.
//
//nolint:nonamedreturns
func (gitPtr *Git) FetchPullRequest(remote string, number int) (err error) {
	opts := &git.FetchOptions{
		RemoteName: remote,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/pull/%d/head:pr/%d", number, number)),
		},
	}

	return gitPtr.repo.Fetch(opts)
}

// CheckoutPullRequest checks out pull request.
//
//nolint:nonamedreturns
func (gitPtr *Git) CheckoutPullRequest(number int) (err error) {
	worktree, err := gitPtr.repo.Worktree()
	if err != nil {
		return err
	}

	opts := &git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("pr/%d", number)),
	}

	return worktree.Checkout(opts)
}

// SHA returns the sha of the current commit.
//
//nolint:nonamedreturns
func (gitPtr *Git) SHA() (sha string, err error) {
	ref, err := gitPtr.repo.Head()
	if err != nil {
		return sha, err
	}

	sha = ref.Hash().String()

	return sha, nil
}

// AheadBehind returns the number of commits that HEAD is ahead and behind
// relative to the specified ref.
//
//nolint:nonamedreturns
func (gitPtr *Git) AheadBehind(ref string) (ahead, behind int, err error) {
	ref1, err := gitPtr.repo.Reference(plumbing.ReferenceName(ref), false)
	if err != nil {
		return 0, 0, err
	}

	ref2, err := gitPtr.repo.Head()
	if err != nil {
		return 0, 0, err
	}

	commit2, err := object.GetCommit(gitPtr.repo.Storer, ref2.Hash())
	if err != nil {
		return 0, 0, nil //nolint:nilerr
	}

	var count int

	iter := object.NewCommitPreorderIter(commit2, nil, nil)

	err = iter.ForEach(func(comm *object.Commit) error {
		if comm.Hash != ref1.Hash() {
			count++

			return nil
		}

		return storer.ErrStop
	})
	if err != nil {
		return 0, 0, nil //nolint:nilerr
	}

	return count, 0, nil
}

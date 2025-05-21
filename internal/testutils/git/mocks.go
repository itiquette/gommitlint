// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing/object"
)

// MockCommitIter is a mock implementation of object.CommitIter for testing.
// It allows you to control the behavior of a commit iterator in tests.
type MockCommitIter struct {
	Commits    []*object.Commit
	Index      int
	StopAtHash string // If set, ForEach will stop when it encounters this hash
}

// Next returns the next commit in the iterator.
func (m *MockCommitIter) Next() (*object.Commit, error) {
	if m.Index >= len(m.Commits) {
		return nil, errors.New("end of iterator")
	}

	commit := m.Commits[m.Index]
	m.Index++

	return commit, nil
}

// ForEach calls the callback function for each commit.
func (m *MockCommitIter) ForEach(callback func(*object.Commit) error) error {
	for _, commit := range m.Commits {
		// If stopAtHash is set and this is the commit with that hash, don't process it
		// but return a "stop" error to simulate the real behavior
		if m.StopAtHash != "" && commit != nil && commit.Hash.String() == m.StopAtHash {
			return errors.New("stop")
		}

		err := callback(commit)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close is a no-op required by the CommitIter interface.
func (m *MockCommitIter) Close() {}

// NewMockCommitIter creates a new MockCommitIter with the given commits.
func NewMockCommitIter(commits []*object.Commit, stopAtHash string) *MockCommitIter {
	return &MockCommitIter{
		Commits:    commits,
		Index:      0,
		StopAtHash: stopAtHash,
	}
}

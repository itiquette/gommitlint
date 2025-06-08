// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"testing"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
	"github.com/stretchr/testify/require"
)

func TestFindHookPath(t *testing.T) {
	tests := []struct {
		name          string
		repoPath      string
		hookType      string
		validator     cliAdapter.PathValidator
		expectedPath  string
		expectedError bool
	}{
		{
			name:     "valid path with mock validator",
			repoPath: "/test/repo",
			hookType: "commit-msg",
			validator: &mockHookPathValidator{
				returnPath: "/test/repo/.git/hooks/commit-msg",
			},
			expectedPath:  "/test/repo/.git/hooks/commit-msg",
			expectedError: false,
		},
		{
			name:     "error from validator",
			repoPath: "/test/repo",
			hookType: "commit-msg",
			validator: &mockHookPathValidator{
				shouldError: true,
			},
			expectedPath:  "",
			expectedError: true,
		},
		{
			name:          "nil validator uses default",
			repoPath:      "/test/repo",
			hookType:      "commit-msg",
			validator:     nil,
			expectedPath:  "",
			expectedError: true, // Default validator will fail without real git repo
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			path, err := FindHookPath(testCase.repoPath, testCase.hookType, testCase.validator)

			if testCase.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedPath, path)
			}
		})
	}
}

func TestEnsureHooksDirectory(t *testing.T) {
	tests := []struct {
		name          string
		repoPath      string
		validator     cliAdapter.PathValidator
		expectedError bool
	}{
		{
			name:     "successful directory creation",
			repoPath: "/test/repo",
			validator: &mockHookPathValidator{
				ensureSuccess: true,
			},
			expectedError: false,
		},
		{
			name:     "error from validator",
			repoPath: "/test/repo",
			validator: &mockHookPathValidator{
				shouldError: true,
			},
			expectedError: true,
		},
		{
			name:          "nil validator uses default",
			repoPath:      "/test/repo",
			validator:     nil,
			expectedError: true, // Default validator will fail without real git repo
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := EnsureHooksDirectory(testCase.repoPath, testCase.validator)

			if testCase.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// mockHookPathValidator is a test implementation of PathValidator for hook path tests.
type mockHookPathValidator struct {
	returnPath    string
	shouldError   bool
	ensureSuccess bool
}

func (m *mockHookPathValidator) ValidateHookPath(_, _ string) (string, error) {
	if m.shouldError {
		return "", os.ErrNotExist
	}

	return m.returnPath, nil
}

func (m *mockHookPathValidator) EnsureHooksDirectory(_ string) error {
	if m.shouldError {
		return os.ErrNotExist
	}

	if !m.ensureSuccess {
		return os.ErrPermission
	}

	return nil
}

// Ensure mock implements the interface.
var _ cliAdapter.PathValidator = (*mockHookPathValidator)(nil)

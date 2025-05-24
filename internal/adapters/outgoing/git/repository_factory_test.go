// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"os"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	testgit "github.com/itiquette/gommitlint/internal/testutils/git"
	"github.com/stretchr/testify/require"
)

func TestRepositoryFactory(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gommitlint-factory-test")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Initialize a git repository
	_, tempGitPath := testgit.SetupTestRepo(t)

	// We'll use the actual Git repo path for our tests
	tempDir = tempGitPath

	// Create context
	ctx := testcontext.CreateTestContext()

	// Create factory
	factory, err := NewRepositoryFactory(ctx, tempDir)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, factory, "Factory should not be nil")

	// Test creating different interface implementations
	t.Run("Create CommitRepository", func(t *testing.T) {
		service := factory.CreateCommitRepository()
		require.NotNil(t, service, "CommitRepository should not be nil")
		require.Implements(t, (*domain.CommitRepository)(nil), service)
	})

	t.Run("Create RepositoryInfoProvider", func(t *testing.T) {
		provider := factory.CreateRepositoryInfoProvider()
		require.NotNil(t, provider, "RepositoryInfoProvider should not be nil")
		require.Implements(t, (*domain.RepositoryInfoProvider)(nil), provider)
	})

	t.Run("Create CommitAnalyzer", func(t *testing.T) {
		analyzer := factory.CreateCommitAnalyzer()
		require.NotNil(t, analyzer, "CommitAnalyzer should not be nil")
		require.Implements(t, (*domain.CommitAnalyzer)(nil), analyzer)
	})
}

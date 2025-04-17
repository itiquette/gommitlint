// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"os"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryFactory(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gommitlint-factory-test")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Initialize a git repository
	_, tempGitPath := setupTestRepo(t)

	// We'll use the actual Git repo path for our tests
	tempDir = tempGitPath

	// Create factory
	factory, err := NewRepositoryFactory(tempDir)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, factory, "Factory should not be nil")

	// Test creating different interface implementations
	t.Run("Create CommitReader", func(t *testing.T) {
		reader := factory.CreateCommitReader()
		assert.NotNil(t, reader, "CommitReader should not be nil")
		assert.Implements(t, (*domain.CommitReader)(nil), reader)
	})

	t.Run("Create CommitHistoryReader", func(t *testing.T) {
		reader := factory.CreateHistoryReader()
		assert.NotNil(t, reader, "CommitHistoryReader should not be nil")
		assert.Implements(t, (*domain.CommitHistoryReader)(nil), reader)
	})

	t.Run("Create RepositoryInfoProvider", func(t *testing.T) {
		provider := factory.CreateInfoProvider()
		assert.NotNil(t, provider, "RepositoryInfoProvider should not be nil")
		assert.Implements(t, (*domain.RepositoryInfoProvider)(nil), provider)
	})

	t.Run("Create CommitAnalyzer", func(t *testing.T) {
		analyzer := factory.CreateCommitAnalyzer()
		assert.NotNil(t, analyzer, "CommitAnalyzer should not be nil")
		assert.Implements(t, (*domain.CommitAnalyzer)(nil), analyzer)
	})

	t.Run("Create Full Service", func(t *testing.T) {
		service := factory.CreateFullService()
		assert.NotNil(t, service, "GitRepositoryService should not be nil")
		assert.Implements(t, (*domain.GitRepositoryService)(nil), service)

		// Verify it implements all the specialized interfaces
		assert.Implements(t, (*domain.CommitReader)(nil), service)
		assert.Implements(t, (*domain.CommitHistoryReader)(nil), service)
		assert.Implements(t, (*domain.RepositoryInfoProvider)(nil), service)
		assert.Implements(t, (*domain.CommitAnalyzer)(nil), service)
	})
}

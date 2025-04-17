# Migration Guide

This document provides guidance for migrating from the original monolithic design to the new hexagonal architecture in Gommitlint.

## Overview

Gommitlint has undergone architectural improvements to enhance maintainability, testability, and extensibility. The key changes include:

1. Introduction of domain collections and specialized interfaces
2. Refactoring of repository implementation
3. Addition of a repository factory
4. Improved error handling and testing

## Migration Steps

### Step 1: Understanding the New Architecture

Before migrating, familiarize yourself with the hexagonal architecture concept and the new structure as described in `ARCHITECTURE.md`.

### Step 2: Migrating from Original GitRepositoryService

**Original Approach:**
```go
// Service depends on the entire GitRepositoryService
type MyService struct {
    repo *git.RepositoryAdapter
}

func (s *MyService) DoSomething() {
    commit := s.repo.GetCommit("abc123")
    // Do something with commit
}
```

**New Approach:**
```go
// Service depends only on the interfaces it needs
type MyService struct {
    commitReader domain.CommitReader
}

func NewMyService(factory *git.RepositoryFactory) *MyService {
    return &MyService{
        commitReader: factory.CreateCommitReader(),
    }
}

func (s *MyService) DoSomething() {
    commit := s.commitReader.GetCommit("abc123")
    // Do something with commit
}
```

### Step 3: Using Domain Collections

**Original Approach:**
```go
commits, err := repo.GetHeadCommits(10)
// Filter merge commits
filteredCommits := make([]*domain.CommitInfo, 0, len(commits))
for _, commit := range commits {
    if !commit.IsMergeCommit {
        filteredCommits = append(filteredCommits, commit)
    }
}
```

**New Approach:**
```go
commits, err := repo.GetHeadCommits(10)
collection := domain.NewCommitCollection(commits)
filteredCommits := collection.FilterMergeCommits().GetCommits()
```

### Step 4: Working with Repository Helpers

Repository helper methods have been extracted to simplify common operations:

**Original Approach:**
```go
// Complex logic in each method
func (g *RepositoryAdapter) GetHeadCommits(count int) ([]*domain.CommitInfo, error) {
    // Get HEAD reference
    ref, err := g.repo.Head()
    if err != nil {
        return nil, fmt.Errorf("failed to get HEAD: %w", err)
    }
    
    // Create commit iterator
    iter, err := g.repo.Log(&git.LogOptions{From: ref.Hash()})
    if err != nil {
        return nil, fmt.Errorf("failed to get log: %w", err)
    }
    
    // Collect commits
    var commits []*domain.CommitInfo
    err = iter.ForEach(func(c *object.Commit) error {
        // Convert each commit
        // Limited by count
        // Additional logic
        return nil
    })
    
    return commits, nil
}
```

**New Approach:**
```go
// Using helper methods
func (g *RepositoryAdapter) GetHeadCommits(count int) ([]*domain.CommitInfo, error) {
    // Resolve HEAD hash
    headHash, err := g.resolveRevision("")
    if err != nil {
        return nil, err
    }
    
    // Create iterator
    iter, err := g.createCommitIterator(headHash)
    if err != nil {
        return nil, err
    }
    
    // Collect and convert commits with limit
    return g.collectAndConvertCommits(iter, count, nil)
}
```

### Step 5: Updating Rule Implementations

Rules that previously depended on the GitRepository struct now use specific interfaces:

**Original Approach:**
```go
// CommitsAheadRule using GitRepository struct
type CommitsAheadRule struct {
    // Other fields
    repositoryGetter func() domain.GitRepository
}

func (r *CommitsAheadRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
    repo := r.repositoryGetter()
    
    if repo.Repo == nil {
        // Error handling
    }
    
    // Implementation
}
```

**New Approach:**
```go
// CommitsAheadRule using CommitAnalyzer interface
type CommitsAheadRule struct {
    // Other fields
    repositoryGetter func() domain.CommitAnalyzer
}

func (r *CommitsAheadRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
    analyzer := r.repositoryGetter()
    
    commitsAhead, err := analyzer.GetCommitsAhead(r.ref)
    if err != nil {
        // Error handling
    }
    
    // Implementation using interface methods
}
```

## Benefits of Migration

1. **Reduced Coupling**: Services depend only on the interfaces they need
2. **Improved Testability**: Easier to mock specific interfaces
3. **Better Code Organization**: Clear separation of concerns
4. **Enhanced Maintainability**: Smaller, focused components
5. **Fluent API**: Domain collections provide a more intuitive API
6. **Consistent Error Handling**: Structured approach to error handling
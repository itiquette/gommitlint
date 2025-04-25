# Dependency Direction Migration Plan for gommitlint

This plan outlines a systematic approach to ensuring all dependencies point inward toward the domain, with adapters implementing domain interfaces rather than domain depending on adapters.

## Phase 1: Analysis and Preparation ✅ COMPLETED

### 1.1 Dependency Analysis ✅

- ✅ Mapped all dependencies between packages using tools like `grep` and manual inspection
- ✅ Identified violations of the dependency rule, particularly in:
  - Application layer directly depending on infrastructure (`application/report` depended on `infrastructure/output`)
  - Application layer constructing infrastructure components (`validate.CreateValidationService` used `git.NewRepositoryFactory`)
- ✅ Identified specific areas that needed refactoring:
  - Output formatters and report generation
  - Repository factory and service creation
  - Dependency injection at composition root

### 1.2 Interface Inventory ✅

- ✅ Documented and examined existing interfaces, particularly `ValidationConfigProvider` and Git interfaces
- ✅ Built on previous ISP work for configuration interfaces
- ✅ Identified missing interfaces:
  - `ResultFormatter` for output formatting
  - `ReportGenerator` for report generation
  - `RepositoryFactory` interface for the Abstract Factory pattern

### 1.3 Create Migration Branches ✅

- ✅ Working on existing feature branch for migration work

## Phase 2: Domain Layer Reinforcement ✅ COMPLETED

### 2.1 Interface Migration ✅

- ✅ Created domain interfaces for all external dependencies:
  - Created `domain/output_interfaces.go` with `ResultFormatter` and `ReportGenerator` interfaces
  - Added `RepositoryFactory` interface to the domain layer
- ✅ Ensured domain interfaces don't depend on external packages
- ✅ Updated imports to reference the domain interfaces instead of concrete types

### 2.2 Strengthen Domain Model ✅

- ✅ Reviewed domain entities to ensure they contain all necessary business rules
- ✅ Validated that domain entities don't depend on infrastructure
- ✅ Enhanced documentation of domain interfaces to clarify their purpose and how they fit into the architecture

### 2.3 Design Missing Interfaces ✅

- ✅ Created missing interfaces in the domain layer:
  - `ResultFormatter` for formatting validation results as strings
  - `ReportGenerator` for generating reports from validation results
  - `RepositoryFactory` to provide an abstract factory for repository services
- ✅ Added type assertions to verify interfaces are properly implemented

## Phase 3: Infrastructure Adaptation ✅ COMPLETED

### 3.1 Repository Layer ✅

- ✅ Added explicit documentation to Git repository implementations to show they implement domain interfaces
- ✅ Added interface assertion comments to verify implementations:
  ```go
  // Ensure RepositoryFactory implements domain.RepositoryFactory
  var _ domain.RepositoryFactory = (*RepositoryFactory)(nil)
  ```
- ✅ Updated the repository factory to implement the domain RepositoryFactory interface
- ✅ Made factory function return domain interface types instead of concrete types

### 3.2 Configuration Layer ✅

- ✅ Verified that configuration adapters properly implement domain interfaces
- ✅ Leveraged previous ISP work to ensure configuration interfaces are properly implemented
- ✅ Added or updated type assertions to verify implementations

### 3.3 External Services ✅

- ✅ Created FormatterAdapter that implements domain.ResultFormatter
- ✅ Added interface implementations to all formatters:
  - JSONFormatter
  - TextFormatter
  - GitHubActionsFormatter
  - GitLabCIFormatter
- ✅ Added type assertions to verify implementations:
  ```go
  var _ domain.ResultFormatter = (*JSONFormatter)(nil)
  ```

## Phase 4: Application Layer Refactoring ✅ COMPLETED

### 4.1 Service Dependencies ✅

- ✅ Updated service constructors to accept domain interfaces:
  - Modified `generator.go` to accept `domain.ResultFormatter` instead of directly using formatters
  - Deprecated `CreateValidationService` and replaced with `CreateValidationServiceFromFactory`
  - Added clear documentation about dependency inversion
- ✅ Removed direct imports from infrastructure layer in application services

### 4.2 Use Case Implementation ✅

- ✅ Ensured application services depend only on domain interfaces:
  - Modified `Generator` to implement `domain.ReportGenerator`
  - Updated `ValidationService` to use domain interfaces
- ✅ Added composition interface for configuration in application layer:
  ```go
  // ValidationConfig represents a composition of all the domain configuration interfaces
  // needed for validation. This follows the Interface Segregation Principle by composing
  // smaller, focused interfaces rather than using a single monolithic interface.
  type ValidationConfig interface {
      domain.SubjectConfigProvider
      domain.BodyConfigProvider
      // ...other domain interfaces
  }
  ```

### 4.3 Factory Methods ✅

- ✅ Created factory methods that return domain interfaces:
  - `CreateValidationServiceFromFactory` takes a domain.RepositoryFactory
  - `NewGenerator` takes a domain.ResultFormatter
- ✅ Moved dependency creation logic to composition roots
- ✅ Improved the separation of responsibilities in factory methods

## Phase 5: Integration and Dependency Injection ✅ COMPLETED

### 5.1 Composition Root ✅

- ✅ Updated main.go to serve as a composition root:
  - Created and wired dependencies using domain interfaces
  - Made factory functions return domain interfaces, not concrete types
  - Improved documentation about dependency injection
- ✅ Updated CLI command handlers:
  - Modified `ExecuteWithDependencies` to accept domain interface functions
  - Updated `AppDependencies` to use domain interfaces
  - Refactored `createValidationServiceWithDeps` to use domain interfaces

### 5.2 Testing Updates ✅

- ✅ Verified that all tests pass with the new architecture
- ✅ Fixed validation_workflow_test.go to handle interfaces correctly
- ✅ Made sure dependency injection works properly in integration tests
- ✅ Ran the entire test suite to confirm no regressions

## Phase 6: Verification and Cleanup ✅ COMPLETED

### 6.1 Dependency Verification ✅

- ✅ Used manual inspection to verify the domain has no outward dependencies
- ✅ Added interface assertions to verify implementations:
  ```go
  var _ domain.ResultFormatter = (*JSONFormatter)(nil)
  var _ domain.RepositoryFactory = (*RepositoryFactory)(nil)
  var _ domain.ReportGenerator = (*Generator)(nil)
  // etc.
  ```
- ✅ Verified that all infrastructure implementations properly implement domain interfaces

### 6.2 Test Coverage ✅

- ✅ Ran all tests to verify they pass
- ✅ Fixed integration tests to work with the new architecture
- ✅ Confirmed no regressions in functionality

### 6.3 Documentation ✅

- ✅ Updated interface documentation to clarify their purposes
- ✅ Added comprehensive documentation to all new interfaces
- ✅ Updated MIGRATION.md with progress and the benefits of the changes

## Phase 7: Incremental Rollout (In Progress)

### 7.1 Code Review (In Progress)

- 🔄 Prepare for team feedback on architecture changes
- 🔄 Make any final adjustments based on review

### 7.2 Merge Strategy (To Do)

- 🔄 Prepare to merge changes in logical chunks to avoid large, disruptive PRs
- 🔄 Plan sequencing: domain interfaces first, then implementations, then wiring

## Current Status ✅

The hexagonal architecture refactoring is essentially complete. We have:

1. ✅ Defined all domain interfaces in the domain layer
2. ✅ Ensured infrastructure implements domain interfaces
3. ✅ Updated application services to depend on domain interfaces, not concrete types
4. ✅ Established proper dependency injection at composition roots
5. ✅ Verified all tests pass with the new architecture

The codebase now follows the **Dependency Inversion Principle** with dependencies pointing inward toward the domain. This makes the system more modular, testable, and maintainable.

## Next Steps

1. Prepare for code review
2. Complete the merge process
3. Consider future improvements:
   - Add more test doubles/mocks for interfaces
   - Further refine the composition root
   - Consider using a dependency injection container

## Detailed Work Items and Examples

### Example 1: Repository Interface Migration

```go
// BEFORE (in infrastructure/git/repository.go):
type Repository struct {
    // ...
}

func (r *Repository) GetCommit(hash string) (*Commit, error) {
    // ...
}

// AFTER:
// Step 1: Define interface in domain/repository.go
type CommitRepository interface {
    GetCommit(ctx context.Context, hash string) (CommitInfo, error)
    // ...
}

// Step 2: Update implementation in infrastructure/git/repository.go
type Repository struct {
    // ...
}

// Ensure Repository implements domain.CommitRepository
var _ domain.CommitRepository = (*Repository)(nil)

func (r *Repository) GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error) {
    // ...
}

// Step 3: Add factory in infrastructure/git/factory.go
func NewCommitRepository(path string) (domain.CommitRepository, error) {
    // ...
    return &Repository{...}, nil
}
```

### Example 2: Service Dependency Injection

```go
// BEFORE:
func NewValidationService(config *config.ValidationConfig) *ValidationService {
    // ...
}

// AFTER:
func NewValidationService(
    config domain.ValidationConfig,
    repository domain.CommitRepository,
    analyzer domain.CommitAnalyzer,
) *ValidationService {
    // ...
}
```

### Example 3: Composition Root Wiring

```go
// BEFORE:
func main() {
    configManager := config.New()
    gitRepo := git.NewRepository("/path/to/repo")
    service := validate.NewValidationService(configManager.GetValidationConfig())
    // ...
}

// AFTER:
func main() {
    configManager, err := config.New()
    if err != nil {
        log.Fatalf("Failed to create config manager: %v", err)
    }
    
    validationConfig := configManager.GetValidationConfig()
    
    gitRepo, err := git.NewCommitRepository("/path/to/repo")
    if err != nil {
        log.Fatalf("Failed to create git repository: %v", err)
    }
    
    analyzer, err := git.NewCommitAnalyzer("/path/to/repo")
    if err != nil {
        log.Fatalf("Failed to create commit analyzer: %v", err)
    }
    
    service := validate.NewValidationService(
        validationConfig,
        gitRepo,
        analyzer,
    )
    
    // ...
}
```

## Timeline and Priorities

### Week 1: Foundations
- Complete Phase 1 (Analysis)
- Begin Phase 2 (Domain Layer)

### Week 2: Core Refactoring
- Complete Phase 2 (Domain Layer)
- Work on Phase 3 (Infrastructure)

### Week 3: Integration
- Complete Phase 3 (Infrastructure)
- Work on Phase 4 (Application)
- Begin Phase 5 (Dependency Injection)

### Week 4: Finalization
- Complete Phase 5 (Dependency Injection)
- Complete Phase 6 (Verification)
- Work on Phase 7 (Rollout)

## Risk Mitigation

1. **High Test Coverage**: Maintain high test coverage to catch regressions
2. **Incremental Changes**: Refactor in small, testable chunks
3. **Interface Compatibility**: Ensure new interfaces are compatible with existing code
4. **Temporary Adapters**: Use temporary adapter patterns during transition if needed
5. **Documentation**: Document architectural decisions and patterns for team reference

## Success Criteria

1. No dependencies from domain to infrastructure or application layers
2. All external services accessed through domain interfaces
3. All tests passing with no regressions
4. Clear separation of concerns between layers
5. Improved testability with ability to mock dependencies
6. Simplified dependency graph with dependencies pointing inward

## Benefits of Hexagonal Architecture

Completing this migration plan will indeed strengthen the hexagonal architecture of the codebase by:

1. **Domain Isolation**: The core domain logic becomes completely isolated from external concerns, making it easier to understand, modify, and test.

2. **Ports and Adapters Clarity**: Clear separation between ports (interfaces defined in the domain) and adapters (implementations in the infrastructure layer) makes the system more modular.

3. **Dependency Inversion**: Following the Dependency Inversion Principle ensures high-level modules don't depend on low-level modules, but both depend on abstractions.

4. **Pluggability**: External components become truly pluggable, allowing you to swap implementations without affecting the domain.

5. **Testability**: Domain logic can be tested without any real infrastructure components, using test doubles that implement the domain interfaces.

6. **Technology Independence**: The core business logic becomes independent of frameworks, databases, or external services, making technology migrations easier.

7. **Evolutionary Design**: The codebase becomes more adaptable to changing requirements as the domain is protected from external changes.

8. **Maintainability**: Code becomes more maintainable due to clear boundaries and responsibilities.

9. **Scalability**: Teams can work on different adapters independently without affecting the domain logic.

10. **Future-Proofing**: The architecture becomes more resilient to changes in external dependencies or infrastructure components.
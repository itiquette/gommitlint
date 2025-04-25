# Hexagonal Architecture Migration Plan

This document outlines a structured plan to improve the hexagonal architecture of the codebase while maintaining simplicity, coherence, and idiomatic Go patterns.

## Phase 1: Foundation Restructuring

**Goal:** Create a solid architectural foundation by addressing core structural issues

### 1A: Interface Segregation

- [x] Break down `RuleValidationConfig` (87 methods) into domain-specific interfaces
- [x] Create focused interfaces like `SubjectConfigProvider`, `BodyConfigProvider`, `SecurityConfigProvider`
- [x] Update rule implementations to depend on specific interfaces they need
- [x] Create composite interfaces where needed (e.g., `ValidationConfigProvider` that embeds others)

### 1B: Dependency Direction Correction

- [x] Refactor configuration package to implement domain interfaces, not vice versa
- [x] Move interface definitions to domain layer if they're in the wrong place
- [x] Ensure all dependencies point inward toward domain core

**Success Criteria:**
- Smaller, focused interfaces with clear purposes ✅
- No circular dependencies between packages ✅
- All dependencies point toward the domain ✅

**Progress on Rule Implementation:**
- [x] SubjectLengthRule updated to use SubjectConfigProvider
- [x] CommitBodyRule updated to use BodyConfigProvider
- [x] JiraReferenceRule updated to use JiraConfigProvider and ConventionalConfigProvider
- [x] ConventionalCommitRule updated to use ConventionalConfigProvider
- [x] ImperativeVerbRule updated to use SubjectConfigProvider and ConventionalConfigProvider
- [x] SubjectCaseRule updated to use SubjectConfigProvider and ConventionalConfigProvider
- [x] Remaining rules to be updated:
  - [x] SubjectSuffix
  - [x] Signature
  - [x] SignOff
  - [x] Spell
  - [x] CommitsAhead

## Phase 2: Dependency Management

**Goal:** Improve testability and make dependencies explicit and clear

### 2A: Explicit Constructor Injection

- [x] Refactor services to accept dependencies rather than creating them internally
- [x] Convert `CreateValidationService()` to accept all dependencies explicitly
- [x] Create a simple composition root in main package for wiring dependencies
- [x] Remove internal dependency creation from service constructors

### 2B: Simplified Factory Pattern

- [x] Reduce unnecessary factory patterns while maintaining clean interfaces
- [x] Simplify repository factory to a single constructor function (`NewRepositoryServices`)
- [x] Maintain interface boundaries but reduce indirection layers
- [x] Consolidate redundant factory patterns throughout the codebase

**Success Criteria:**
- ✅ All dependencies explicitly injected in constructors
- ✅ Clear composition at the application entry point
- ✅ Reduced complexity in object creation patterns

## Phase 3: Domain Integrity

**Goal:** Ensure domain logic stays in the domain layer, strengthening the core

### 3A: Domain Logic Repatriation

- [x] Move domain logic from application layer back to domain layer
- [x] Ensure domain entities are complete and self-contained (CommitInfo with author info)
- [x] Implement proper value semantics throughout domain entities
- [x] Remove infrastructure dependencies from domain logic (no direct go-git access)

### 3B: Consistent Context Handling

- [x] Enhance adapter implementation to extract and map all necessary data
- [x] Fix documentation to conform to Go standards
- [x] Ensure domain entities don't rely on infrastructure details
- [x] Create clear boundaries between infrastructure and domain concerns

**Success Criteria:**
- ✅ All domain logic resides in domain layer
- ✅ Clean separation between domain and application concerns
- ✅ Consistent context handling throughout the codebase

## Phase 4: Simplification & Cleanup

**Goal:** Remove unnecessary complexity while maintaining architectural integrity

### 4A: Configuration System Streamlining

- [ ] Consolidate config loader, provider, manager into a single config service
- [ ] Ensure it implements the domain interfaces created in Phase 1
- [ ] Remove duplicated validation logic

### 4B: Rule Registration Refactoring

- [ ] Replace complex registration with simpler pattern (e.g., functional options)
- [ ] Maintain flexibility for enabling/disabling rules
- [ ] Improve discovery of available rules

**Success Criteria:**
- Simpler configuration system with fewer abstractions
- Easier rule registration and discovery
- Maintained architectural boundaries with less code

## Phase 5: Integration & Refinement

**Goal:** Ensure all components work together coherently after changes

### 5A: Integration Testing

- [ ] Create end-to-end tests for main workflows
- [ ] Verify rule validation still works correctly
- [ ] Check configuration loading and application

### 5B: Documentation Update

- [ ] Update architecture documentation to reflect new structure
- [ ] Document the simplified interfaces and their responsibilities
- [ ] Provide examples of proper usage of the architecture

**Success Criteria:**
- All tests pass with new architecture
- Updated documentation that reflects the improved design
- Simplified yet still powerful architecture

## Implementation Principles

1. **Progressive Completion** - Finish each phase before moving to the next
2. **Test Continuity** - Ensure tests pass after each change
3. **Simplicity Focus** - Add abstractions only when absolutely necessary
4. **Go Idioms** - Prefer simple interfaces, composition, and explicit code
5. **Clean Design** - Breaking changes are acceptable if they simplify the code
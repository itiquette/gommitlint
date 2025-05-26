# Functional Refactoring Plan

## Overview
Fix adapter dependencies and imperative code to achieve clean functional architecture.

## Phase 1: Fix Adapter Dependencies ✅ COMPLETE

### Goal
Remove direct log adapter imports from all adapters. Inject logger through constructors.

### Tasks

#### 1.1 Git Adapter

- [x] Remove log import from `internal/adapters/outgoing/git/repository.go`
- [x] Add logger parameter to `NewRepository`
- [x] Remove log import from `internal/adapters/outgoing/git/repository_factory.go`
- [x] Add logger parameter to `NewRepositoryFactory`
- [x] Remove log import from `internal/adapters/outgoing/git/repository_helpers.go`
- [x] Update all git adapter methods to use injected logger

#### 1.2 Output Adapters

- [x] Remove log import from `internal/adapters/outgoing/output/json.go`
- [x] Remove log import from `internal/adapters/outgoing/output/github.go`
- [x] Remove log import from `internal/adapters/outgoing/output/gitlab.go`
- [x] Remove log import from `internal/adapters/outgoing/output/text.go`
- [x] Add logger to output adapter constructors if needed (decided logging not needed for stateless formatters)

#### 1.3 CLI Adapters

- [x] Remove log import from `internal/adapters/incoming/cli/validatecmd.go` (using contextx.GetLogger)
- [x] Keep log import in `internal/adapters/incoming/cli/rootcmd.go` (needed for logger initialization)
- [x] Get logger from context instead of importing (in validatecmd)

#### 1.4 Update Factories and Composition

- [x] Update git repository factory to accept and pass logger
- [x] Update all factory methods to accept and pass logger
- [x] Fix compilation errors and unused parameters/functions
- [x] Fix integration tests that expect logger in context

## Phase 2: Fix Imperative Code in Rules ✅ COMPLETE

### Goal
Transform mutable, imperative validation code into pure functional code.

### Tasks

#### 2.1 Create Functional Utilities

- [x] Created `internal/common/functional/validation.go`
- [x] Added error collection helpers (CombineErrors, AllErrors)
- [x] Added conditional validation helpers

#### 2.2 Refactor Simple Rules

- [x] SubjectLengthRule - already using value receivers and functional patterns
- [x] SubjectCaseRule - already using value receivers and functional patterns
- [x] SubjectSuffixRule - already using value receivers and functional patterns

#### 2.3 Refactor Complex Rules

- [x] CommitBodyRule - refactored to use functional.AllErrors pattern
- [x] ConventionalCommitRule - already using value receivers and functional patterns
- [x] ImperativeRule - already using value receivers and functional patterns
- [x] JiraRule - already using value receivers and functional patterns
- [x] IdentityRule - already using value receivers and functional patterns

#### 2.4 Fix Option Functions

- [x] All option functions already return new values (not mutate)
- [x] All rules already use value receivers
- [x] No pointer receivers in rule methods

#### 2.5 Additional Rules Verified

- [x] SignatureRule - already using value receivers and functional patterns
- [x] SignOffRule - already using value receivers and functional patterns
- [x] SpellRule - already using value receivers and functional patterns
- [x] BranchAheadRule - already using value receivers and functional patterns

## Phase 3: Fix Remaining Pointer Receivers

### Goal
Complete the functional transformation by removing remaining pointer receivers in non-rule code.

### Tasks

#### 3.1 Config Interfaces

- [x] Fix emptyConfig methods in `internal/common/config/interface.go` - changed to value receivers
- [x] Updated GetConfig to return value instead of pointer

#### 3.2 Composition Container

- [x] Reviewed Container methods in `internal/composition/container.go`
- [x] Kept pointer receivers due to mutex and thread-safety requirements

#### 3.3 Config Loader

- [x] Fixed Loader methods in `internal/config/loader.go`
- [x] Changed Load() and LoadFromPath() to return new Loader instances

#### 3.4 Crypto Verifiers

- [x] Fixed SSH verifier to use value receivers
- [x] Fixed GPG verifier to use value receivers
- [x] Ensured immutable verification patterns

#### 3.5 Factories

- [x] SimpleRuleFactory already uses value receivers
- [x] OutgoingAdapterFactory already uses value receivers

#### 3.6 Test Utilities (Lower Priority)

- [ ] PrintLogger in testutils can remain with pointer receivers (test code)

## Success Criteria

- [x] All tests pass ✅
- [x] `make quality` passes ✅
- [x] No adapter imports other adapters ✅
- [x] All rules use immutable patterns ✅
- [x] No unnecessary pointer receivers in production code ✅
- [x] Functional patterns throughout where appropriate ✅

## Current Status

✅ **Phase 1 COMPLETE**:

- Successfully removed all cross-adapter dependencies
- Logger injection implemented throughout
- Code compiles cleanly and passes all quality checks
- All tests pass including integration tests

✅ **Phase 2 COMPLETE**:

- All rules verified to use value receivers
- CommitBodyRule refactored to use functional error collection
- Created functional validation utilities
- All rule option functions return new values
- No pointer receivers in any rule code

✅ **Phase 3 COMPLETE**:

- Fixed emptyConfig to use value receivers
- Config Loader now returns new instances (functional pattern)
- Crypto verifiers (SSH & GPG) now use value receivers
- Container keeps pointer receivers for valid thread-safety reasons
- All production code now uses appropriate receiver types
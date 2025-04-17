# Gommitlint Architecture

This document outlines the architecture of Gommitlint, a commit message validation tool built using hexagonal (ports and adapters) architecture.

## Hexagonal Architecture Overview

Gommitlint follows the hexagonal architecture pattern (also known as ports and adapters), where:

- The **domain** layer contains the core business logic and entities
- The **ports** define interfaces to interact with the domain
- The **adapters** implement these interfaces to connect the domain with external systems

This architectural approach provides several benefits:
- Clear separation of concerns
- Domain logic isolated from infrastructure details
- Testability through interface-based design
- Flexibility to change implementation details without affecting core business logic

## Project Structure

### Domain Layer

The domain layer is the heart of the application, containing the core business logic and entities.

- `domain/commit.go` - Core domain entities like `CommitInfo` and interfaces
- `domain/result.go` - Validation result models
- `domain/rule.go` - Rule interfaces for validation rules
- `domain/git_interfaces.go` - Specialized interfaces for Git operations
- `domain/commit_collection.go` - Domain collection for operations on groups of commits

### Core Layer

The core layer contains the business rules implementation.

- `core/rules/` - Implementation of the validation rules
- `core/validation/` - Core validation engine and rule provider

### Ports

Ports define the interfaces through which the application interacts with the outside world.

- `ports/cli/` - Command-line interface for the application
- `ports/fileio/` - File I/O interfaces

### Adapters (Infrastructure)

Adapters implement the interfaces defined by the ports to connect with external systems.

- `infrastructure/git/` - Git repository adapter
- `infrastructure/output/` - Output formatters (JSON, text)
- `infrastructure/config/` - Configuration provider

### Application Layer

The application layer orchestrates the domain logic to solve specific use cases.

- `application/validate/` - Validation service for commit validation
- `application/report/` - Report generation service

## Interface Segregation

Following the Interface Segregation Principle, interfaces are defined based on client requirements:

1. `CommitReader` - For reading individual commits
2. `CommitHistoryReader` - For accessing commit history
3. `RepositoryInfoProvider` - For general repository information
4. `CommitAnalyzer` - For advanced commit analysis

These specialized interfaces allow clients to depend only on what they need, reducing coupling.

## Repository Pattern

For Git operations, Gommitlint uses:

1. Repository interfaces in the domain layer
2. A repository adapter in the infrastructure layer
3. A factory to create specific repository interfaces
4. Helper methods to reduce complexity

## Dependency Injection

Dependencies are injected through:

1. Constructor parameters
2. Factory pattern for creating repositories
3. Functional options pattern for configuration

## Error Handling

The application uses a structured approach to error handling:

1. Domain-specific errors with context information
2. Error wrapping for maintaining context
3. Standardized validation error format

## Validation Rules

Validation rules follow a plugin-based approach:

1. All rules implement the `Rule` interface
2. Rules are registered through a provider
3. Rule configuration is centralized
4. Each rule focuses on a specific aspect of commit message validation

## Testing

The application is designed for testability:

1. Interface-based design allows for easy mocking
2. Table-driven tests for comprehensive coverage
3. Clear separation of concerns to facilitate unit testing
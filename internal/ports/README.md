# Hexagonal Architecture Ports

This package documents all port interfaces in the gommitlint system following hexagonal architecture principles.

## Overview

In hexagonal architecture (also known as ports and adapters), the core domain is isolated from external concerns through interfaces (ports). These ports define contracts that adapters must implement.

Due to Go's type system and to avoid circular dependencies, the actual interface definitions remain in their respective packages. This package serves as a central documentation point for all ports in the system.

## Port Locations

### Domain Ports

These interfaces are defined in the `internal/domain/` package and represent core business contracts:

#### Repository Interfaces (`internal/domain/commitinterfaces.go`)
- **`CommitRepository`** - Provides access to Git commit information
- **`RepositoryInfoProvider`** - Provides repository metadata
- **`CommitAnalyzer`** - Provides commit analysis operations

#### Cryptographic Interfaces (`internal/domain/crypto_interfaces.go`)
- **`CryptoVerifier`** - Verifies commit signatures
- **`CryptoKeyRepository`** - Manages cryptographic keys

#### Formatting Interface (`internal/domain/interfaces.go`)
- **`Formatter`** - Formats validation results for output

### Adapter Ports

These interfaces are defined by adapters following the dependency inversion principle:

#### Logger Interface (`internal/adapters/*/ports.go`)
- **`Logger`** - Provides logging capabilities
- Each adapter defines its own Logger interface to avoid coupling

### Application Ports

#### Validation Service (`internal/application/validation/`)
- **`ValidationService`** - The main validation service interface (typically not explicitly defined as the Service struct itself acts as the port)

## Architectural Principles

1. **Dependency Inversion**: High-level modules (domain) don't depend on low-level modules (adapters). Both depend on abstractions (interfaces).

2. **Interface Segregation**: Interfaces are kept small and focused on specific responsibilities.

3. **Explicit Dependencies**: All dependencies are passed explicitly through constructors or function parameters.

4. **No Circular Dependencies**: Interfaces are placed in packages to avoid import cycles.

## Implementation Guidelines

When implementing an adapter:

1. Check which domain interface(s) your adapter should implement
2. Ensure your adapter satisfies the interface with compile-time checks:
   ```go
   var _ domain.CommitRepository = (*MyAdapter)(nil)
   ```
3. Define adapter-specific interfaces (like Logger) in your adapter package
4. Pass dependencies explicitly through constructors

## Example

```go
// Adapter implementing domain interface
package git

import "github.com/itiquette/gommitlint/internal/domain"

// Repository adapts git operations to the domain model
type Repository struct {
    // fields...
}

// Compile-time check
var _ domain.CommitRepository = (*Repository)(nil)

// Implementation of interface methods...
```
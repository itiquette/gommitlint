# Crypto Package Architecture

This document describes the architecture of the crypto package, which provides signature and identity verification for commit validation.

## Overview

The crypto architecture provides a clean, maintainable implementation for signature verification by:

1. Using domain models for cryptographic signatures and identities
2. Separating verification logic from validation rules
3. Following functional programming principles and value semantics
4. Aligning with the hexagonal architecture pattern

## Architecture

### Domain Models

The domain models define the core concepts:

- `crypto.Signature`: Represents a cryptographic signature with its type and data
- `crypto.Identity`: Represents a signer or committer identity with name and email
- `crypto.VerificationResult`: Represents the outcome of signature verification

These models are pure value types with immutable semantics, located in the `internal/domain/crypto` package.

### Core Implementation

The core implementations provide the verification logic:

- `crypto.VerifierFactory`: Creates the appropriate verifier for a signature
- `gpg.Verifier`: Implements GPG signature verification
- `ssh.Verifier`: Implements SSH signature verification

These implementations are located in the `internal/core/crypto` package with specialized subpackages for each signature type.

### Adapters

The adapters handle external dependencies:

- `crypto.KeyRepository`: Interface for accessing cryptographic keys
- `crypto.FileSystemKeyRepository`: Implementation that uses the local filesystem
- `crypto.VerificationAdapter`: High-level adapter for verification operations

These adapters are located in the `internal/adapters/outgoing/crypto` package.

### Utility Packages

The crypto functionality relies on several utility packages:

- `internal/common/fsutils`: Filesystem utilities for path sanitization and file operations
- `internal/common/encutils`: Encoding utilities for base64 operations

These utility packages follow functional programming principles with pure functions and value semantics.

### Rules

The rules use the crypto domain for validation:

- `SignatureRule`: Validates that commits have proper signatures
- `IdentityRule`: Validates that signatures match committer identity

## Using the New Architecture

### For Rules

Rules should use the crypto domain models directly:

```go
// Create a signature from a string
signature := crypto.NewSignature(commit.Signature)

// Check properties
if signature.Type() == crypto.SignatureTypeGPG {
    // Handle GPG signature
}

// Create identities
authorIdentity := crypto.NewIdentity(commit.AuthorName, commit.AuthorEmail)
signerIdentity := crypto.NewIdentityFromString("Name <email@example.com>")

// Compare identities
if authorIdentity.Matches(signerIdentity) {
    // Identities match
}
```

### For Verification

Use the verification adapter:

```go
// Create repository and adapter
repository := crypto.NewFileSystemKeyRepository(keyDir)
verifier := crypto.NewVerificationAdapter(repository)

// Verify a commit
result, err := verifier.VerifyCommit(ctx, commit)

// Check result
if result.IsVerified() {
    // Signature verified successfully
    identity := result.Identity()
    fmt.Println("Verified by:", identity.String())
} else {
    fmt.Println("Verification failed:", result.ErrorMessage())
}
```

## Benefits of the New Architecture

- **Clean domain model**: Pure value types for signatures and identities
- **Better testability**: Interfaces for verification allow easy mocking
- **Immutability**: No state modification in core functions
- **Simpler rules**: Focused validation logic
- **Extensibility**: Easy to add new signature types or verifiers

## Update Tests

The test fixtures and mocks have been updated to work with the new architecture. Use the new test helpers in your tests.

## Verifying the Migration

### Running Quality Checks

To ensure the refactored code meets quality standards:

```bash
# Run all linters
make quality

# Run specific linters
make quality/golangcilint
make quality/govet
make quality/revive
```

These checks will verify:
- Proper documentation (including deprecation notices)
- No unused code
- Consistent formatting
- No import cycles
- Correct error handling patterns

### Running Tests

To verify the new implementation works correctly:

```bash
# Run all tests
make test

# Run only crypto-related tests
go test -v ./internal/domain/crypto/...
go test -v ./internal/core/crypto/...
go test -v ./internal/adapters/outgoing/crypto/...

# Run tests for utility packages
go test -v ./internal/common/fsutils/...
go test -v ./internal/common/encutils/...
```

### Package Structure

The crypto functionality is organized across the following packages to maintain hexagonal architecture:

```
internal/
├── domain/
│   └── crypto/             # Domain models and interfaces
│       ├── identity.go     # Identity model
│       ├── signature.go    # Signature model
│       └── verification.go # Verification interfaces and results
│
├── core/
│   └── crypto/             # Core implementations
│       ├── factory.go      # Verifier factory
│       ├── gpg/            # GPG verification 
│       └── ssh/            # SSH verification
│
├── adapters/
│   └── outgoing/
│       └── crypto/         # External adapters
│           ├── repository.go     # Key repository interface
│           └── verification.go   # Verification adapter
│
├── common/
│   ├── fsutils/            # Filesystem utilities
│   └── encutils/           # Encoding utilities
└── ...
```

## Questions and Support

If you have questions about the migration, please open an issue on the project repository.
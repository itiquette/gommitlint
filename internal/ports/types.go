// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package ports

// This file provides documentation for all port interfaces in the system.
// The actual interfaces are defined in their respective packages to avoid circular dependencies.

// Domain Ports (defined in internal/domain/)
// ==========================================

// CommitRepository (domain.CommitRepository) provides access to Git commit information.
// Methods:
//   - GetCommit(ctx, hash) (CommitInfo, error)
//   - GetCommits(ctx, limit) ([]CommitInfo, error)
//   - GetCommitRange(ctx, fromHash, toHash) ([]CommitInfo, error)
//   - GetHeadCommits(ctx, count) ([]CommitInfo, error)

// RepositoryInfoProvider (domain.RepositoryInfoProvider) provides repository metadata.
// Methods:
//   - GetCurrentBranch(ctx) (string, error)
//   - GetRepositoryName(ctx) string
//   - IsValid(ctx) (bool, error)

// CommitAnalyzer (domain.CommitAnalyzer) provides commit analysis operations.
// Methods:
//   - GetCommitsAhead(ctx, ref) (int, error)

// CryptoVerifier (domain.CryptoVerifier) verifies commit signatures.
// Methods:
//   - VerifyCommit(ctx, commit) (VerificationResult, error)

// CryptoKeyRepository (domain.CryptoKeyRepository) manages cryptographic keys.
// Methods:
//   - GetKeyDirectory() string
//   - FindKeyFiles(extensions) ([]string, error)
//   - ReadKeyFile(path) ([]byte, error)

// Formatter (format.Formatter) formats validation results.
// Methods:
//   - Format(ctx, results) string
//   - ContentType() string

// Adapter Ports (defined in internal/adapters/*/ports.go)
// =======================================================

// Logger provides logging capabilities.
// This interface is defined in multiple packages following dependency inversion.
// Methods:
//   - Debug(msg, keysAndValues...)
//   - Info(msg, keysAndValues...)
//   - Warn(msg, keysAndValues...)
//   - Error(msg, keysAndValues...)

// Application Ports (defined in internal/application/*/ports.go)
// ==============================================================

// ValidationService provides commit validation operations.
// This is typically implemented by the validation.Service type.
// Methods:
//   - ValidateCommit(ctx, hash) (ValidationResult, error)
//   - ValidateCommits(ctx, hashes) ([]ValidationResult, error)
//   - ValidateBranch(ctx, branch, baseBranch) ([]ValidationResult, error)
//   - ValidateRange(ctx, fromHash, toHash) ([]ValidationResult, error)
//   - ValidateMessage(ctx, message) (ValidationResult, error)
//   - ValidateMessagesFromFile(ctx, filePath) ([]ValidationResult, error)

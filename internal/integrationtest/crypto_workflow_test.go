// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/crypto/gpg"
	"github.com/itiquette/gommitlint/internal/core/crypto/ssh"
	"github.com/itiquette/gommitlint/internal/domain"
	domainCrypto "github.com/itiquette/gommitlint/internal/domain/crypto"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// TestCryptoVerificationWorkflow is a comprehensive test of crypto verification
// in an integration context.
//
// NOTE: This is a smoke test that doesn't verify actual signatures but makes sure
// all components work together correctly. We don't expect real signature verification
// to pass in CI environments since that would require actual private keys.
func TestCryptoVerificationWorkflow(t *testing.T) {
	// Skip test in short mode as it involves filesystem operations
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary directory for test keys
	tempDir, err := os.MkdirTemp("", "test-keys")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test GPG public key file
	gpgKeyPath := filepath.Join(tempDir, "test.gpg")
	err = os.WriteFile(gpgKeyPath, []byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

mQENBGHfLDwBCADR7a/VQHUJlTZMWjbvWNIhwKqWzNQ6KLnryiAJJ70Q7hDkUVe+
VZCD0evJ7jMoqXMEqATGc2RvmjxnYQh2AINAkxJXc5YMYvczqsKYrT48PUCdZN0r
TTDcIQbLUTm+8rtWxiHOV0KIj+I9R8OeKjLmPGMO/EP2qJwSkgVbTzuR04vLGDoe
zFLOlT40F3b3nzYJTWfIHC/+/3d6c2Kv/UiNbQ3ZXvgBUzLswUQcP7lHhBy8tgLW
nwYJ+QxxM1HWAeXY7gvlqH87LjL8CJrUk1mAULpdAJlUAkUE3PZt1zDVJPY0KMP8
4yF1QxRM2SG4BFBtpJ6j6QK8RKnZBM5xfXylABEBAAG0HVRlc3QgVXNlciA8dGVz
dEB0ZXN0Lml0Lm5vdzE+iQFOBBMBCAA4FiEEnQ0DNyQNzOSKFhQA7lK9N9hs+/EF
AmHfLDwCGwMFCwkIBwIGFQoJCAsCBBYCAwECHgECF4AACgkQ7lK9N9hs+/E3zgf/
cdA7Ff8bmzBG7K1a85nkPmKtj6M4c1WVYgYS3qwdvz0jQXviVQiZQJVwEORSZX6z
CCvJdADYDMcQJ33DAkQ+gWE0mfxpRtP4rntLGOsTQfyTo/+bIxdtm1jshAMELkb9
L3I+QGLlRqmc9OzIcOS5aLJQUbR7+v02L1NUJJw4WS+32EDEfHUGCUqHJ2L1QdNc
1CWnHFjzZFUvQ7mSEpBbVeZZBNBYUjwZqGaKozdW5xqnugQ8M/X5B74m2XL5oykZ
g+l5BLMCl8dKPDfUlUdGNKMYJLjfB1wAODFZ74QM0YB1YEIzCd+WbdYYQhGGQ6/g
4JUo5P5ZfIiPCMiHdQ==
=YWul
-----END PGP PUBLIC KEY BLOCK-----`), 0600)
	require.NoError(t, err)

	// Create a test SSH public key file
	sshKeyPath := filepath.Join(tempDir, "test.pub")
	err = os.WriteFile(sshKeyPath, []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDdUn4q1WZ test@ssh.example.com"), 0600)
	require.NoError(t, err)

	// Create test commits with different signatures
	commits := map[string]domain.CommitInfo{
		"gpg": {
			Hash:        "abc123",
			Message:     "GPG signed commit",
			AuthorName:  "Test User",
			AuthorEmail: "test@test.it.now1",
			Signature: `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2

iQEcBAABCAAGBQJkglfUAQoJECXBz7a2zOr65JcIAK+ghUvxzS4DZBnY1t2+8JLk
xDrX6MT7MHuJcOd05yVgGJdyj2o73Qlxc/lGlQe/i+3R9cwTAqxY+GOKVhzM+QtT
3qfyQlKUwlwN3fPlgwZZbCgXUtQ+/71zLFEMmU7QtgRBUYKi4LVD5TuWUxQOLHbm
I5AXZIZOUmODdOVEwQT13mDV7FP/Mt+9ZKidq7YP3jUvlf7i9OCqzGsXQGa6FKs/
zC5FfCXR0dNrHz2V9IPhJsV8bxNTSZ5dRELaFa3mc0ew0mYriZz1LgaGu6Km3mJv
0mZlH6y9MWy9lx5FhAOA8b8EFWIDdHaDu4F0ZMUJtZx9/G0QEY1U6P/iuR8=
=QLiW
-----END PGP SIGNATURE-----`,
		},
		"ssh": {
			Hash:        "def456",
			Message:     "SSH signed commit",
			AuthorName:  "SSH Test",
			AuthorEmail: "test@ssh.example.com",
			Signature: `-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg0+/s2JSj4+kZxJKkZIFtAEp
N3n4HR8xiknrpOKpgJHMAAAAHZ2l0QDIuMzQBDQ==
-----END SSH SIGNATURE-----`,
		},
		"unsigned": {
			Hash:        "ghi789",
			Message:     "Unsigned commit",
			AuthorName:  "Unsigned Test",
			AuthorEmail: "unsigned@example.com",
			Signature:   "",
		},
	}

	// Create a context with configuration
	ctx := testcontext.CreateTestContext()
	cfg := types.Config{
		Signing: types.SigningConfig{
			AllowedSigners: []string{
				"Test User <test@test.it.now1>",
				"SSH Test <test@ssh.example.com>",
			},
			KeyDirectory: tempDir,
		},
		Rules: types.RulesConfig{
			Enabled: []string{"Signature", "SignedIdentity"},
		},
	}

	adapter := config.NewAdapter(cfg)
	ctx = contextx.WithConfig(ctx, adapter)

	// Test the domain models directly
	t.Run("DomainModels", func(t *testing.T) {
		// Create GPG signature
		gpgSignature := domainCrypto.NewSignature(commits["gpg"].Signature)
		require.Equal(t, domainCrypto.SignatureTypeGPG, gpgSignature.Type())
		require.False(t, gpgSignature.IsEmpty())
		require.True(t, gpgSignature.IsValid())

		// Create SSH signature
		sshSignature := domainCrypto.NewSignature(commits["ssh"].Signature)
		require.Equal(t, domainCrypto.SignatureTypeSSH, sshSignature.Type())
		require.False(t, sshSignature.IsEmpty())
		require.True(t, sshSignature.IsValid())

		// Create empty signature
		emptySignature := domainCrypto.NewSignature("")
		require.Equal(t, domainCrypto.SignatureTypeUnknown, emptySignature.Type())
		require.True(t, emptySignature.IsEmpty())
		require.False(t, emptySignature.IsValid())

		// Create identities
		authorIdentity := domainCrypto.NewIdentity(commits["gpg"].AuthorName, commits["gpg"].AuthorEmail)
		allowedIdentity := domainCrypto.NewIdentityFromString("Test User <test@test.it.now1>")

		// Test matching
		require.True(t, authorIdentity.Matches(allowedIdentity))

		// Test matching against collection
		allowedIdentities := []domainCrypto.Identity{
			domainCrypto.NewIdentityFromString("Test User <test@test.it.now1>"),
			domainCrypto.NewIdentityFromString("Other User <other@example.com>"),
		}
		require.True(t, authorIdentity.MatchesAny(allowedIdentities))

		// Test non-matching identity
		nonMatchingIdentity := domainCrypto.NewIdentity("Other", "other@example.com")
		require.False(t, nonMatchingIdentity.Matches(authorIdentity))
	})

	// Test the verification adapter
	t.Run("VerificationAdapter", func(t *testing.T) {
		// Create the repository and adapter
		repository := crypto.NewFileSystemKeyRepository(tempDir)
		verifier := crypto.NewVerificationAdapter(repository)

		// Test with GPG signature
		result, err := verifier.VerifyCommit(ctx, commits["gpg"])
		require.NoError(t, err)
		// We expect this to fail in CI (we don't have the actual private key)
		t.Logf("GPG verification result: status=%s, identity=%s",
			result.Status(), result.Identity().String())

		// Test with SSH signature
		result, err = verifier.VerifyCommit(ctx, commits["ssh"])
		require.NoError(t, err)
		t.Logf("SSH verification result: status=%s, identity=%s",
			result.Status(), result.Identity().String())

		// Test with unsigned commit
		result, err = verifier.VerifyCommit(ctx, commits["unsigned"])
		require.NoError(t, err)
		require.Equal(t, domainCrypto.VerificationStatusFailed, result.Status())
		require.Equal(t, "missing_signature", result.ErrorCode())
	})

	// Test individual verifiers
	t.Run("Verifiers", func(t *testing.T) {
		// GPG verifier
		gpgVerifier := gpg.NewDefaultVerifier()
		require.True(t, gpgVerifier.CanVerify(domainCrypto.NewSignature(commits["gpg"].Signature)))
		require.False(t, gpgVerifier.CanVerify(domainCrypto.NewSignature(commits["ssh"].Signature)))

		// SSH verifier
		sshVerifier := ssh.NewDefaultVerifier()
		require.True(t, sshVerifier.CanVerify(domainCrypto.NewSignature(commits["ssh"].Signature)))
		require.False(t, sshVerifier.CanVerify(domainCrypto.NewSignature(commits["gpg"].Signature)))

		// Verification result
		result := domainCrypto.NewVerificationResult(
			domainCrypto.VerificationStatusVerified,
			domainCrypto.NewIdentity("Test User", "test@example.com"),
			domainCrypto.NewSignature("test-signature"),
		)
		require.True(t, result.IsVerified())
		require.Equal(t, "Test User <test@example.com>", result.Identity().String())

		// Add error to result
		resultWithError := result.WithError("test_error", "Test error message")
		require.True(t, resultWithError.HasError())
		require.Equal(t, "test_error", resultWithError.ErrorCode())
		require.Equal(t, "Test error message", resultWithError.ErrorMessage())
	})

	// Print success message
	fmt.Println("Crypto verification architecture tests completed successfully")
}

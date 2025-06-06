// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  domain.SignatureType
		wantEmpty bool
		wantValid bool
	}{
		{
			name:      "Empty signature",
			input:     "",
			wantType:  domain.SignatureTypeUnknown,
			wantEmpty: true,
			wantValid: false,
		},
		{
			name: "GPG signature",
			input: `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2

iQEcBAABCAAGBQJkglfUAQoJECXBz7a2zOr65JcIAK+ghUvxzS4DZBnY1t2+8JLk
xDrX6MT7MHuJcOd05yVgGJdyj2o73Qlxc/lGlQe/i+3R9cwTAqxY+GOKVhzM+QtT
3qfyQlKUwlwN3fPlgwZZbCgXUtQ+/71zLFEMmU7QtgRBUYKi4LVD5TuWUxQOLHbm
I5AXZIZOUmODdOVEwQT13mDV7FP/Mt+9ZKidq7YP3jUvlf7i9OCqzGsXQGa6FKs/
zC5FfCXR0dNrHz2V9IPhJsV8bxNTSZ5dRELaFa3mc0ew0mYriZz1LgaGu6Km3mJv
0mZlH6y9MWy9lx5FhAOA8b8EFWIDdHaDu4F0ZMUJtZx9/G0QEY1U6P/iuR8=
=QLiW
-----END PGP SIGNATURE-----`,
			wantType:  domain.SignatureTypeGPG,
			wantEmpty: false,
			wantValid: true,
		},
		{
			name: "SSH signature",
			input: `-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAgm5CqhCBSQDCbXJc94GFxy+9EYhYt
0zn/S58cPKoFNRMAAAAHZ2l0LXNzaAAAAACzAAAAMAAAAAtzdWJqZWN0CUFkZCBmZWF0dXJl
LwAAACRrZXktMS11c2VyQHVuaXQuZXhhbXBsZQAAAAAAAAAAAAAAAA==
-----END SSH SIGNATURE-----`,
			wantType:  domain.SignatureTypeSSH,
			wantEmpty: false,
			wantValid: true,
		},
		{
			name:      "SSH signature with format:blob",
			input:     "ssh-rsa:AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
			wantType:  domain.SignatureTypeSSH,
			wantEmpty: false,
			wantValid: true,
		},
		{
			name:      "Unknown signature",
			input:     "some-random-data",
			wantType:  domain.SignatureTypeUnknown,
			wantEmpty: false,
			wantValid: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			sig := domain.NewSignature(testCase.input)

			require.Equal(t, testCase.wantType, sig.Type())
			require.Equal(t, testCase.input, sig.Data())
			require.Equal(t, testCase.wantEmpty, sig.IsEmpty())
			require.Equal(t, testCase.wantValid, sig.IsValid())
		})
	}
}

func TestSignatureWithType(t *testing.T) {
	sig := domain.NewSignatureWithType("test-data", domain.SignatureTypeGPG)
	require.Equal(t, domain.SignatureTypeGPG, sig.Type())
	require.Equal(t, "test-data", sig.Data())
	require.False(t, sig.IsEmpty())
}

func TestSignatureString(t *testing.T) {
	sig := domain.NewSignature("short-signature")
	require.Contains(t, sig.String(), "short-signature")

	longSig := domain.NewSignature("this-is-a-very-long-signature-that-should-be-truncated-in-the-string-output")
	strOutput := longSig.String()
	require.Contains(t, strOutput, "...")
	require.Less(t, len(strOutput), len(longSig.Data()))
}

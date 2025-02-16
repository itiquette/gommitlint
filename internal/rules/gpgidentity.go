// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

type GPGIdentityCheck struct {
	errors   []error
	identity string
}

func (g GPGIdentityCheck) Status() string {
	return "GPG Identity"
}

func (g GPGIdentityCheck) Message() string {
	if len(g.errors) != 0 {
		return g.errors[0].Error()
	}

	return fmt.Sprintf("Signed by %q", g.identity)
}

func (g GPGIdentityCheck) Errors() []error {
	return g.errors
}

func ValidateGPGIdentity(signature string, commitData *object.Commit, pubKeyDir string) GPGIdentityCheck {
	check := GPGIdentityCheck{}

	if pubKeyDir == "" {
		check.errors = append(check.errors, errors.New("no public key directory provided"))

		return check
	}

	if signature == "" {
		check.errors = append(check.errors, errors.New("commit is not signed"))

		return check
	}

	entries, err := os.ReadDir(pubKeyDir)
	if err != nil {
		check.errors = append(check.errors, fmt.Errorf("failed to load public keys: %w", err))

		return check
	}

	var entityList []*openpgp.Entity

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .pub, .gpg, or .asc files
		name := entry.Name()
		if !strings.HasSuffix(name, ".pub") && !strings.HasSuffix(name, ".gpg") && !strings.HasSuffix(name, ".asc") {
			continue
		}

		keyData, err := os.ReadFile(filepath.Join(pubKeyDir, name))
		if err != nil {
			continue // Skip files we can't read
		}

		entities, err := openpgp.ReadArmoredKeyRing(strings.NewReader(string(keyData)))
		if err != nil {
			continue // Skip invalid key files
		}

		entityList = append(entityList, entities...)
	}

	if len(entityList) == 0 {
		check.errors = append(check.errors, errors.New("no public keys found in directory"))

		return check
	}
	// Get commit data for signature verification
	encoded := &plumbing.MemoryObject{}
	_ = commitData.EncodeWithoutSignature(encoded)

	// Get reader with commit data
	commitDataReader, _ := encoded.Reader()
	// Try to verify the signature with any of the public keys
	signatureReader := strings.NewReader(signature)
	for _, entity := range entityList {
		// Try to verify signature
		verifiedEntity, err := openpgp.CheckArmoredDetachedSignature(
			openpgp.EntityList{entity},
			commitDataReader, // Empty content
			signatureReader,
			nil,
		)
		if err == nil && verifiedEntity != nil {
			// Found a matching key, get the identity
			for _, identity := range verifiedEntity.Identities {
				check.identity = identity.Name

				return check
			}
		}

		signatureReader.Reset(signature)
	}

	check.errors = append(check.errors, errors.New("no valid signature found with trusted keys"))

	return check
}

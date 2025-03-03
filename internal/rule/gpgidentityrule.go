// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

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

type GPGIdentityRule struct {
	RuleErrors []error
	Identity   string
}

func (g GPGIdentityRule) Name() string {
	return "GPGIdentityRule"
}

func (g GPGIdentityRule) Result() string {
	if len(g.RuleErrors) != 0 {
		return g.RuleErrors[0].Error()
	}

	return fmt.Sprintf("Signed by %q", g.Identity)
}

func (g GPGIdentityRule) Errors() []error {
	return g.RuleErrors
}

func ValidateGPGIdentity(signature string, commitData *object.Commit, pubKeyDir string) GPGIdentityRule {
	rule := GPGIdentityRule{}

	if pubKeyDir == "" {
		rule.RuleErrors = append(rule.RuleErrors, errors.New("no public key directory provided"))

		return rule
	}

	if signature == "" {
		rule.RuleErrors = append(rule.RuleErrors, errors.New("commit is not signed"))

		return rule
	}

	entries, err := os.ReadDir(pubKeyDir)
	if err != nil {
		rule.RuleErrors = append(rule.RuleErrors, fmt.Errorf("failed to load public keys: %w", err))

		return rule
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
		rule.RuleErrors = append(rule.RuleErrors, errors.New("no public keys found in directory"))

		return rule
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
				rule.Identity = identity.Name

				return rule
			}
		}

		signatureReader.Reset(signature)
	}

	rule.RuleErrors = append(rule.RuleErrors, errors.New("no valid signature found with trusted keys"))

	return rule
}

// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/google/go-github/v60/github"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/janderssonse/gommitlint/internal/git"
	"github.com/janderssonse/gommitlint/internal/interfaces"
)

// GPGIdentityCheck ensures that the commit is cryptographically signed using known identity.
type GPGIdentityCheck struct {
	errors   []error
	identity string
}

// Status returns the name of the check.
func (gpgIdentityCheck GPGIdentityCheck) Status() string {
	return "GPG Identity"
}

// Message returns to check message.
func (gpgIdentityCheck GPGIdentityCheck) Message() string {
	if len(gpgIdentityCheck.errors) != 0 {
		return gpgIdentityCheck.errors[0].Error()
	}

	return fmt.Sprintf("Signed by %q", gpgIdentityCheck.identity)
}

// Errors returns any violations of the check.
func (gpgIdentityCheck GPGIdentityCheck) Errors() []error {
	return gpgIdentityCheck.errors
}

// ValidateGPGIdentity checks the commit GPG signature for a known identity.
func ValidateGPGIdentity(gitPtr *git.Git, githubOrg string) interfaces.Check { //nolint:ireturn
	check := &GPGIdentityCheck{}

	switch {
	case githubOrg != "":
		githubClient := github.NewClient(nil)

		list, _, err := githubClient.Organizations.ListMembers(context.Background(), githubOrg, &github.ListMembersOptions{})
		if err != nil {
			check.errors = append(check.errors, err)

			return check
		}

		members := make([]string, len(list))

		for index := range list {
			members[index] = list[index].GetLogin()
		}

		keyrings, err := getKeyring(context.Background(), members)
		if err != nil {
			check.errors = append(check.errors, err)

			return check
		}

		entity, err := gitPtr.VerifyPGPSignature(keyrings)
		if err != nil {
			check.errors = append(check.errors, err)

			return check
		}

		for identity := range entity.Identities {
			check.identity = identity

			break
		}
	default:
		check.errors = append(check.errors, errors.New("no signature identity configuration found"))
	}

	return check
}

func getKeyring(ctx context.Context, members []string) ([]string, error) {
	var (
		result []string
		mutex  sync.Mutex
	)

	errorgroup, ctx := errgroup.WithContext(ctx)

	for _, member := range members {
		errorgroup.Go(func() error {
			key, err := getKey(ctx, member)

			mutex.Lock()
			result = append(result, key)
			mutex.Unlock()

			return err
		})
	}

	err := errorgroup.Wait()

	return result, err
}

func getKey(ctx context.Context, login string) (string, error) {
	// GitHub client doesn't have a method to fetch a key unauthenticated
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://github.com/%s.gpg", login), nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)

	return string(buf), err
}

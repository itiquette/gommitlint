// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

// Package validator defines commit policy validation.
package validator

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	"github.com/janderssonse/gommitlint/internal/policy"
	"github.com/janderssonse/gommitlint/internal/policy/commit"
)

// Gommitlint is a struct that gommitlint.yaml gets decoded into.
type Gommitlint struct {
	CommitDeclaration *commit.Commit `yaml:"commit"`
}

// New loads the gommitlint.yaml file and unmarshals it into a Gommitlint struct.
func New() (*Gommitlint, error) {
	gommitlint := &Gommitlint{}

	configBytes, err := os.ReadFile(".gommitlint.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configBytes, gommitlint) //nolint
	if err != nil {
		return nil, err
	}

	return gommitlint, nil
}

// Validate enforces all policies defined in the gommitlint.yaml file.
func (c *Gommitlint) Validate(setters ...policy.Option) error {
	opts := policy.NewDefaultOptions(setters...)

	const padding = 8
	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	fmt.Fprintln(tabWriter, "POLICY\tCHECK\tSTATUS\tMESSAGE\t")

	pass := true

	commitConf, err := c.convertDeclarations()
	if err != nil {
		return fmt.Errorf("failed to convert declarations: %w", err)
	}

	report, err := commitConf.Compliance(opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, check := range report.Checks() {
		if len(check.Errors()) != 0 {
			for _, err := range check.Errors() {
				fmt.Fprintf(tabWriter, "%s\t%s\t%s\t%v\t\n", "commit", check.Name(), "FAILED", err)
			}

			pass = false
		} else {
			fmt.Fprintf(tabWriter, "%s\t%s\t%s\t%s\t\n", "commit", check.Name(), "PASS", check.Message())
		}
	}

	tabWriter.Flush()

	if !pass {
		return errors.New("1 or more policy failed")
	}

	return nil
}

func (c *Gommitlint) convertDeclarations() (*commit.Commit, error) {
	// spec["gpg"] = map[string]interface{}{
	// 	"required": val,
	// }
	var cmt commit.Commit

	if err := mapstructure.Decode(c.CommitDeclaration, &cmt); err != nil {
		return nil, fmt.Errorf("failed to convert commit policy: %w", err)
	}

	return &cmt, nil
}

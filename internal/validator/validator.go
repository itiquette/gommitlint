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
	Policies []*PolicyDeclaration `yaml:"policies"`
}

// PolicyDeclaration allows a user to declare an arbitrary type along with a
// spec that will be decoded into the appropriate concrete type.
type PolicyDeclaration struct {
	Type string      `yaml:"type"`
	Spec interface{} `yaml:"spec"`
}

// New loads the gommitlint.yaml file and unmarshals it into a Gommitlint struct.
func New() (*Gommitlint, error) {
	gommitlint := &Gommitlint{}

	configBytes, err := os.ReadFile(".gommitlint.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configBytes, gommitlint)
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

	policiesWithTypes, err := c.convertDeclarations()
	if err != nil {
		return fmt.Errorf("failed to convert declarations: %w", err)
	}

	for _, policyWithType := range policiesWithTypes {
		report, err := policyWithType.policy.Compliance(opts)
		if err != nil {
			log.Fatal(err)
		}

		for _, check := range report.Checks() {
			if len(check.Errors()) != 0 {
				for _, err := range check.Errors() {
					fmt.Fprintf(tabWriter, "%s\t%s\t%s\t%v\t\n", policyWithType.Type, check.Name(), "FAILED", err)
				}

				pass = false
			} else {
				fmt.Fprintf(tabWriter, "%s\t%s\t%s\t%s\t\n", policyWithType.Type, check.Name(), "PASS", check.Message())
			}
		}
	}

	tabWriter.Flush()

	if !pass {
		return errors.New("1 or more policy failed")
	}

	return nil
}

type policyWithType struct {
	policy policy.Policy
	Type   string
}

func (c *Gommitlint) convertDeclarations() ([]policyWithType, error) {
	var (
		policies = make([]policyWithType, 0, len(c.Policies))
	)

	for _, policy := range c.Policies {
		switch policy.Type {
		case "commit":
			// backwards compatibility, convert `gpg: bool` into `gpg: required: bool`
			if spec, ok := policy.Spec.(map[interface{}]interface{}); ok {
				if gpg, ok := spec["gpg"]; ok {
					if val, ok := gpg.(bool); ok {
						spec["gpg"] = map[string]interface{}{
							"required": val,
						}
					}
				}
			}

			var cmt commit.Commit

			if err := mapstructure.Decode(policy.Spec, &cmt); err != nil {
				return nil, fmt.Errorf("failed to convert commit policy: %w", err)
			}

			policies = append(policies, policyWithType{
				Type:   policy.Type,
				policy: &cmt,
			})
		default:
			return nil, fmt.Errorf("invalid policy type: %s", policy.Type)
		}
	}

	return policies, nil
}

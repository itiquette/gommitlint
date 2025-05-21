// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package clitests provides test helpers for CLI-related functionality.
package clitests

import (
	"io"
	"reflect"
)

// HookInstallationParams is a version of HookInstallationParameters
// without directly importing it to avoid import cycles.
type HookInstallationParams struct {
	Force    bool
	RepoPath string
	HookType string
}

// HookRemovalParams is a version of HookRemovalParameters
// without directly importing it to avoid import cycles.
type HookRemovalParams struct {
	RepoPath    string
	SkipConfirm bool
	HookType    string
	Output      io.Writer
	Input       io.Reader
}

// WithHookType returns a new HookInstallationParameters with the hook type updated.
// This is a test-only function for use in test code.
func WithHookType(params interface{}, hookType string) interface{} {
	// Use reflection to set the field without depending on the exact type
	val := reflect.ValueOf(params)
	result := reflect.New(val.Type()).Elem()
	result.Set(val)

	// Set the HookType field
	result.FieldByName("HookType").SetString(hookType)

	return result.Interface()
}

// WithForce returns a new HookInstallationParameters with the force flag updated.
// This is a test-only function for use in test code.
func WithForce(params interface{}, force bool) interface{} {
	// Use reflection to set the field without depending on the exact type
	val := reflect.ValueOf(params)
	result := reflect.New(val.Type()).Elem()
	result.Set(val)

	// Set the Force field
	result.FieldByName("Force").SetBool(force)

	return result.Interface()
}

// WithRepoPath returns a new HookInstallationParameters with the repo path updated.
// This is a test-only function for use in test code.
func WithRepoPath(params interface{}, repoPath string) interface{} {
	// Use reflection to set the field without depending on the exact type
	val := reflect.ValueOf(params)
	result := reflect.New(val.Type()).Elem()
	result.Set(val)

	// Set the RepoPath field
	result.FieldByName("RepoPath").SetString(repoPath)

	return result.Interface()
}

// WithHookTypeRemoval returns a new HookRemovalParameters with the hook type updated.
// This is a test-only function for use in test code.
func WithHookTypeRemoval(params interface{}, hookType string) interface{} {
	// Use reflection to set the field without depending on the exact type
	val := reflect.ValueOf(params)
	result := reflect.New(val.Type()).Elem()
	result.Set(val)

	// Set the HookType field
	result.FieldByName("HookType").SetString(hookType)

	return result.Interface()
}

// WithSkipConfirm returns a new HookRemovalParameters with the skip confirm flag updated.
// This is a test-only function for use in test code.
func WithSkipConfirm(params interface{}, skipConfirm bool) interface{} {
	// Use reflection to set the field without depending on the exact type
	val := reflect.ValueOf(params)
	result := reflect.New(val.Type()).Elem()
	result.Set(val)

	// Set the SkipConfirm field
	result.FieldByName("SkipConfirm").SetBool(skipConfirm)

	return result.Interface()
}

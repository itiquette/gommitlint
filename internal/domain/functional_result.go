// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package functional provides functional programming utilities for error handling and data transformations.
// This package follows functional programming principles with value semantics and immutable data structures.
//
//nolint:varnamelen // Short parameter names are idiomatic in functional programming
package domain

import (
	"errors"
	"fmt"
)

// Result represents a value that may contain either a successful result or an error.
// This is a functional approach to error handling that allows for composable operations.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a Result with a successful value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, err: nil}
}

// Err creates a Result with an error.
func Err[T any](err error) Result[T] {
	var zero T

	return Result[T]{value: zero, err: err}
}

// IsOk returns true if the Result contains a successful value.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the Result contains an error.
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap returns the value and error.
// This allows integration with traditional Go error handling.
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// UnwrapOr returns the value if Ok, otherwise returns the provided default.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.IsOk() {
		return r.value
	}

	return defaultValue
}

// Map applies a function to the value if Ok, otherwise returns the error.
// This allows for functional transformation chains.
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.IsErr() {
		return Err[U](r.err)
	}

	return Ok(fn(r.value))
}

// FlatMap applies a function that returns a Result to the value if Ok.
// This allows for chaining operations that may fail.
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.IsErr() {
		return Err[U](r.err)
	}

	return fn(r.value)
}

// MapErr applies a function to the error if Err, otherwise returns the value.
// This allows for error transformation and recovery.
func MapErr[T any](r Result[T], fn func(error) error) Result[T] {
	if r.IsOk() {
		return r
	}

	return Err[T](fn(r.err))
}

// Filter applies a predicate to the value if Ok.
// If the predicate returns false, returns an error with the provided message.
func Filter[T any](r Result[T], predicate func(T) bool, errorMsg string) Result[T] {
	if r.IsErr() {
		return r
	}

	if predicate(r.value) {
		return r
	}

	return Err[T](errors.New(errorMsg))
}

// Recover attempts to recover from an error using the provided function.
// If the Result is Ok, it returns unchanged. If Err, applies the recovery function.
func Recover[T any](r Result[T], recovery func(error) Result[T]) Result[T] {
	if r.IsOk() {
		return r
	}

	return recovery(r.err)
}

// Combine combines two Results using the provided function.
// If both are Ok, applies the function. If either is Err, returns the first error.
func Combine[T, U, V any](r1 Result[T], r2 Result[U], fn func(T, U) V) Result[V] {
	if r1.IsErr() {
		return Err[V](r1.err)
	}

	if r2.IsErr() {
		return Err[V](r2.err)
	}

	return Ok(fn(r1.value, r2.value))
}

// FromGoResult creates a Result from a traditional Go (value, error) pair.
func FromGoResult[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}

	return Ok(value)
}

// Try executes a function that may panic and converts panics to errors.
// This provides safe execution of potentially unsafe operations.
func Try[T any](fn func() T) Result[T] {
	var result Result[T]

	defer func() {
		if r := recover(); r != nil {
			result = Err[T](fmt.Errorf("panic recovered: %v", r))
		}
	}()

	result = Ok(fn())

	return result
}

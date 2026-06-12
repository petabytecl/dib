package cli

import (
	"errors"
	"fmt"
)

// ErrInvalidInvocation identifies malformed invocation boundary input.
var ErrInvalidInvocation = errors.New("invalid cli invocation")

// InvocationError reports an inspectable invocation construction failure.
type InvocationError struct {
	category error
	problem  string
}

func newInvocationError(category error, problem string) *InvocationError {
	return &InvocationError{category: category, problem: problem}
}

func (e *InvocationError) Error() string {
	if e == nil {
		return "cli: invocation error"
	}
	if e.problem == "" {
		return fmt.Sprintf("cli: %v", e.category)
	}
	return fmt.Sprintf("cli: %v: %s", e.category, e.problem)
}

// Unwrap exposes the sentinel category for errors.Is.
func (e *InvocationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Category returns the invocation diagnostic category.
func (e *InvocationError) Category() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Problem returns a value-free description of the invalid invocation shape.
func (e *InvocationError) Problem() string {
	if e == nil {
		return ""
	}
	return e.problem
}

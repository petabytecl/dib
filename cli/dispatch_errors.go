package cli

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNoHandler identifies a successfully routed command with no registered handler.
var ErrNoHandler = errors.New("cli command has no handler")

// DispatchError reports an inspectable high-level dispatch failure.
type DispatchError struct {
	path     []string
	category error
}

func newDispatchError(path []string, category error) *DispatchError {
	return &DispatchError{
		path:     append([]string(nil), path...),
		category: category,
	}
}

func (e *DispatchError) Error() string {
	if e == nil {
		return "cli: dispatch error"
	}
	if len(e.path) == 0 {
		return fmt.Sprintf("cli: %v", e.category)
	}
	return fmt.Sprintf("cli: %v for route %q", e.category, strings.Join(e.path, " "))
}

// Unwrap exposes the sentinel category for errors.Is.
func (e *DispatchError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Category returns the dispatch diagnostic category.
func (e *DispatchError) Category() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Path returns the canonical command route path associated with the failure.
func (e *DispatchError) Path() []string {
	if e == nil {
		return nil
	}
	return append([]string(nil), e.path...)
}

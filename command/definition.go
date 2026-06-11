package command

import (
	"strings"
)

// NameError reports command name validation failure.
type NameError struct{}

// Error returns a stable diagnostic without echoing rejected input.
func (*NameError) Error() string {
	return "command name must not be empty or whitespace"
}

// Definition is an immutable command definition.
//
// Use NewDefinition to create validated command definitions. The zero value is
// not a validated command definition.
type Definition struct {
	name string
}

// NewDefinition returns a command definition with a stable name.
//
// Invalid names return *NameError so callers can inspect failures with
// errors.As.
func NewDefinition(name string) (Definition, error) {
	if strings.TrimSpace(name) == "" {
		return Definition{}, &NameError{}
	}

	return Definition{name: name}, nil
}

// Name returns the command definition name supplied at construction.
func (d Definition) Name() string {
	return d.name
}

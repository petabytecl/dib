package command

import (
	"errors"
	"fmt"
)

// ErrUnknownCommand identifies a routing failure caused by an unmatched command token.
var ErrUnknownCommand = errors.New("unknown command")

// UnknownCommandError reports an inspectable command routing failure.
type UnknownCommandError struct {
	token      string
	parentPath []string
}

func newUnknownCommandError(token string, parentPath []Definition) *UnknownCommandError {
	return &UnknownCommandError{
		token:      token,
		parentPath: pathNames(parentPath),
	}
}

func (e *UnknownCommandError) Error() string {
	if e == nil {
		return "command: unknown command"
	}
	return fmt.Sprintf("command: %v %q", ErrUnknownCommand, e.token)
}

// Unwrap exposes ErrUnknownCommand for errors.Is.
func (e *UnknownCommandError) Unwrap() error {
	if e == nil {
		return nil
	}
	return ErrUnknownCommand
}

// Token returns the unmatched source token.
func (e *UnknownCommandError) Token() string {
	if e == nil {
		return ""
	}
	return e.token
}

// ParentPath returns the canonical command path matched before the failure.
func (e *UnknownCommandError) ParentPath() []string {
	if e == nil {
		return nil
	}
	return append([]string(nil), e.parentPath...)
}

package command

import (
	"errors"
	"fmt"
)

// ErrUnknownCommand identifies a routing failure caused by an unmatched command token.
var ErrUnknownCommand = errors.New("unknown command")

// ErrInvalidCommandAlias identifies alias metadata that cannot be used for lookup.
var ErrInvalidCommandAlias = errors.New("invalid command alias")

// ErrDuplicateCommandToken identifies an ambiguous command lookup token.
var ErrDuplicateCommandToken = errors.New("duplicate command token")

// ErrFlagComposition identifies a command route flag composition setup failure.
var ErrFlagComposition = errors.New("command flag composition failed")

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

// AliasError reports invalid alias metadata for one command definition.
type AliasError struct {
	parentPath []string
	command    string
	alias      string
}

func newAliasError(parentPath []string, command, alias string) *AliasError {
	return &AliasError{
		parentPath: append([]string(nil), parentPath...),
		command:    command,
		alias:      alias,
	}
}

func (e *AliasError) Error() string {
	if e == nil {
		return "command: invalid command alias"
	}
	return fmt.Sprintf("command: %v %q for %q", ErrInvalidCommandAlias, e.alias, e.command)
}

// Unwrap exposes ErrInvalidCommandAlias for errors.Is.
func (e *AliasError) Unwrap() error {
	if e == nil {
		return nil
	}
	return ErrInvalidCommandAlias
}

// ParentPath returns the canonical parent path containing the command.
func (e *AliasError) ParentPath() []string {
	if e == nil {
		return nil
	}
	return append([]string(nil), e.parentPath...)
}

// Command returns the canonical command name whose alias is invalid.
func (e *AliasError) Command() string {
	if e == nil {
		return ""
	}
	return e.command
}

// Alias returns the invalid alias token.
func (e *AliasError) Alias() string {
	if e == nil {
		return ""
	}
	return e.alias
}

// TokenConflictError reports an ambiguous command lookup token in one parent scope.
type TokenConflictError struct {
	parentPath       []string
	token            string
	firstCommand     string
	collidingCommand string
}

func newTokenConflictError(parentPath []string, token, firstCommand, collidingCommand string) *TokenConflictError {
	return &TokenConflictError{
		parentPath:       append([]string(nil), parentPath...),
		token:            token,
		firstCommand:     firstCommand,
		collidingCommand: collidingCommand,
	}
}

func (e *TokenConflictError) Error() string {
	if e == nil {
		return "command: duplicate command token"
	}
	return fmt.Sprintf("command: %v %q under %v", ErrDuplicateCommandToken, e.token, e.parentPath)
}

// Unwrap exposes ErrDuplicateCommandToken for errors.Is.
func (e *TokenConflictError) Unwrap() error {
	if e == nil {
		return nil
	}
	return ErrDuplicateCommandToken
}

// ParentPath returns the canonical parent path where the token conflict occurred.
func (e *TokenConflictError) ParentPath() []string {
	if e == nil {
		return nil
	}
	return append([]string(nil), e.parentPath...)
}

// Token returns the ambiguous lookup token.
func (e *TokenConflictError) Token() string {
	if e == nil {
		return ""
	}
	return e.token
}

// FirstCommand returns the first canonical child command registered for the token.
func (e *TokenConflictError) FirstCommand() string {
	if e == nil {
		return ""
	}
	return e.firstCommand
}

// CollidingCommand returns the canonical child command that collided with the token.
func (e *TokenConflictError) CollidingCommand() string {
	if e == nil {
		return ""
	}
	return e.collidingCommand
}

// FlagCompositionError reports an inspectable flag composition failure for one
// canonical command path.
type FlagCompositionError struct {
	path  []string
	scope string
	cause error
}

func newFlagCompositionError(path []Definition, scope string, cause error) *FlagCompositionError {
	return &FlagCompositionError{
		path:  pathNames(path),
		scope: scope,
		cause: cause,
	}
}

func (e *FlagCompositionError) Error() string {
	if e == nil {
		return "command: flag composition failed"
	}
	return fmt.Sprintf("command: %v at %v during %s flag composition: %v", ErrFlagComposition, e.path, e.scope, e.cause)
}

// Unwrap exposes both the command composition category and the underlying
// flags package diagnostic for errors.Is and errors.As.
func (e *FlagCompositionError) Unwrap() []error {
	if e == nil {
		return nil
	}
	if e.cause == nil {
		return []error{ErrFlagComposition}
	}
	return []error{ErrFlagComposition, e.cause}
}

// Path returns the canonical command path whose flags failed to compose.
func (e *FlagCompositionError) Path() []string {
	if e == nil {
		return nil
	}
	return append([]string(nil), e.path...)
}

// Scope returns the composition scope associated with the failure.
func (e *FlagCompositionError) Scope() string {
	if e == nil {
		return ""
	}
	return e.scope
}

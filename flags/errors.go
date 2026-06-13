package flags

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by the flags package; match them with errors.Is.
var (
	ErrInvalidDefinition       = errors.New("invalid flag definition")
	ErrDuplicateName           = errors.New("duplicate flag name")
	ErrDuplicateNormalizedName = errors.New("duplicate normalized flag name")
	ErrDuplicateShorthand      = errors.New("duplicate flag shorthand")
	ErrInvalidShorthand        = errors.New("invalid flag shorthand")
	ErrInvalidNoOptionDefault  = errors.New("invalid flag no-option default")
	ErrUnknownFlag             = errors.New("unknown flag")
	ErrMissingValue            = errors.New("missing flag value")
	ErrDuplicateValue          = errors.New("duplicate flag value")
	ErrConversion              = errors.New("flag value conversion failed")
	ErrInvalidGroup            = errors.New("invalid shorthand group")
	ErrHelpRequest             = errors.New("flag help request")
)

// DefinitionError reports an inspectable setup-time definition failure.
type DefinitionError struct {
	name           string
	shorthand      string
	collidingName  string
	normalizedName string
	category       error
}

func newDefinitionError(name, shorthand string, category error) *DefinitionError {
	return &DefinitionError{name: name, shorthand: shorthand, category: category}
}

func newInvalidNormalizedDefinitionError(name, normalizedName string) *DefinitionError {
	return &DefinitionError{
		name:           name,
		normalizedName: normalizedName,
		category:       ErrInvalidDefinition,
	}
}

func newNormalizedDefinitionError(name, collidingName, normalizedName string) *DefinitionError {
	return &DefinitionError{
		name:           name,
		collidingName:  collidingName,
		normalizedName: normalizedName,
		category:       ErrDuplicateNormalizedName,
	}
}

func (e *DefinitionError) Error() string {
	if e == nil {
		return "flags: definition error"
	}

	message := fmt.Sprintf("flags: %v", e.category)
	if e.name != "" {
		message += fmt.Sprintf(" for %q", e.name)
	}
	if e.shorthand != "" {
		message += fmt.Sprintf(" shorthand %q", e.shorthand)
	}
	if e.collidingName != "" {
		message += fmt.Sprintf(" collides with %q", e.collidingName)
	}
	if e.normalizedName != "" {
		message += fmt.Sprintf(" normalized as %q", e.normalizedName)
	}
	return message
}

func (e *DefinitionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Name returns the long flag name associated with the setup failure.
func (e *DefinitionError) Name() string {
	if e == nil {
		return ""
	}
	return e.name
}

// Shorthand returns the shorthand associated with the setup failure.
func (e *DefinitionError) Shorthand() string {
	if e == nil {
		return ""
	}
	return e.shorthand
}

// CollidingName returns the other long flag name involved in a setup collision.
func (e *DefinitionError) CollidingName() string {
	if e == nil {
		return ""
	}
	return e.collidingName
}

// NormalizedName returns the normalized long-name key associated with the setup failure.
func (e *DefinitionError) NormalizedName() string {
	if e == nil {
		return ""
	}
	return e.normalizedName
}

// ValueError reports an inspectable flag value conversion failure.
type ValueError struct {
	name  string
	kind  Kind
	cause error
}

func newValueError(name string, kind Kind, cause error) *ValueError {
	return &ValueError{name: name, kind: kind, cause: cause}
}

func (e *ValueError) Error() string {
	if e == nil {
		return "flags: value error"
	}
	return fmt.Sprintf("flags: %v for %q as %s", ErrConversion, e.name, e.kind)
}

func (e *ValueError) Unwrap() []error {
	if e == nil {
		return nil
	}
	if e.cause == nil {
		return []error{ErrConversion}
	}
	return []error{ErrConversion, e.cause}
}

// Name returns the flag name associated with the conversion failure.
func (e *ValueError) Name() string {
	if e == nil {
		return ""
	}
	return e.name
}

// Kind returns the value kind associated with the conversion failure.
func (e *ValueError) Kind() Kind {
	if e == nil {
		return KindString
	}
	return e.kind
}

// ParseError reports an inspectable command-line parse failure.
type ParseError struct {
	category       error
	token          string
	name           string
	normalizedName string
	definition     Definition
	hasDefinition  bool
	cause          error
}

func newParseError(category error, token, name, normalizedName string, def Definition, hasDefinition bool, cause error) *ParseError {
	return &ParseError{
		category:       category,
		token:          token,
		name:           name,
		normalizedName: normalizedName,
		definition:     def,
		hasDefinition:  hasDefinition,
		cause:          cause,
	}
}

func (e *ParseError) Error() string {
	if e == nil {
		return "flags: parse error"
	}

	message := fmt.Sprintf("flags: %v", e.category)
	if e.token != "" {
		message += fmt.Sprintf(" at %q", e.token)
	}
	if e.name != "" {
		message += fmt.Sprintf(" for %q", e.name)
	}
	return message
}

func (e *ParseError) Unwrap() []error {
	if e == nil {
		return nil
	}
	if e.cause == nil {
		return []error{e.category}
	}
	return []error{e.category, e.cause}
}

// Category returns the sentinel parse failure category.
func (e *ParseError) Category() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Token returns the source flag token without any attached value.
func (e *ParseError) Token() string {
	if e == nil {
		return ""
	}
	return e.token
}

// Name returns the raw source flag identifier: the long flag name or shorthand.
func (e *ParseError) Name() string {
	if e == nil {
		return ""
	}
	return e.name
}

// NormalizedName returns the stable lookup key for the matched source spelling.
func (e *ParseError) NormalizedName() string {
	if e == nil {
		return ""
	}
	return e.normalizedName
}

// Definition returns the canonical definition associated with the parse failure.
func (e *ParseError) Definition() (Definition, bool) {
	if e == nil || !e.hasDefinition {
		return Definition{}, false
	}
	return e.definition, true
}

package config

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidDefinition      = errors.New("invalid config definition")
	ErrDuplicateKey           = errors.New("duplicate config key")
	ErrDuplicateNormalizedKey = errors.New("duplicate normalized config key")
	ErrInvalidDefault         = errors.New("invalid config default")
)

// DefinitionError reports an inspectable setup-time config definition failure.
type DefinitionError struct {
	key             string
	collidingKey    string
	normalizedKey   string
	kind            Kind
	provenance      string
	diagnosticValue string
	redacted        bool
	category        error
}

func newDefinitionError(key string, collidingKey string, normalizedKey string, kind Kind, diagnosticValue string, redacted bool, category error) *DefinitionError {
	provenance := ""
	if category == ErrInvalidDefault {
		provenance = SourceDefault
	}
	return &DefinitionError{
		key:             key,
		collidingKey:    collidingKey,
		normalizedKey:   normalizedKey,
		kind:            kind,
		provenance:      provenance,
		diagnosticValue: diagnosticValue,
		redacted:        redacted,
		category:        category,
	}
}

func formatInvalidDefault(def Definition) string {
	if def.sensitive {
		return ""
	}
	return fmt.Sprintf("%v", def.defaultValue)
}

func (e *DefinitionError) Error() string {
	if e == nil {
		return "config: definition error"
	}

	message := fmt.Sprintf("config: %v", e.category)
	if e.key != "" {
		message += fmt.Sprintf(" for %q", e.key)
	}
	if e.kind.String() != "unknown" {
		message += fmt.Sprintf(" as %s", e.kind)
	}
	if e.provenance != "" {
		message += fmt.Sprintf(" from %s", e.provenance)
	}
	if e.collidingKey != "" {
		message += fmt.Sprintf(" collides with %q", e.collidingKey)
	}
	if e.normalizedKey != "" {
		message += fmt.Sprintf(" normalized as %q", e.normalizedKey)
	}
	if e.redacted {
		message += " value redacted"
	} else if e.diagnosticValue != "" {
		message += fmt.Sprintf(" value %q", e.diagnosticValue)
	}
	return message
}

func (e *DefinitionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Key returns the config key associated with the setup failure.
func (e *DefinitionError) Key() string {
	if e == nil {
		return ""
	}
	return e.key
}

// CollidingKey returns the other config key involved in a setup collision.
func (e *DefinitionError) CollidingKey() string {
	if e == nil {
		return ""
	}
	return e.collidingKey
}

// NormalizedKey returns the normalized key associated with the setup failure.
func (e *DefinitionError) NormalizedKey() string {
	if e == nil {
		return ""
	}
	return e.normalizedKey
}

// Kind returns the expected value kind associated with the setup failure.
func (e *DefinitionError) Kind() Kind {
	if e == nil {
		return KindString
	}
	return e.kind
}

// Provenance returns the source label associated with the setup failure.
func (e *DefinitionError) Provenance() string {
	if e == nil {
		return ""
	}
	return e.provenance
}

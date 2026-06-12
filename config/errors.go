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
	ErrInvalidSource          = errors.New("invalid config source")
	ErrUnknownSourceKey       = errors.New("unknown config source key")
	ErrDuplicateBinding       = errors.New("duplicate config flag binding")
	ErrSourceRead             = errors.New("config source read failure")
	ErrJSONDecode             = errors.New("config JSON decode failure")
	ErrSourceConversion       = errors.New("config source value conversion failure")

	ErrKeyNotFound   = errors.New("config key not found")
	ErrKeyAbsent     = errors.New("config key has no value")
	ErrGetConversion = errors.New("config value type mismatch")
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

// Redacted reports whether the raw definition value was omitted from diagnostics.
func (e *DefinitionError) Redacted() bool {
	if e == nil {
		return false
	}
	return e.redacted
}

// Category returns the definition diagnostic category.
func (e *DefinitionError) Category() error {
	if e == nil {
		return nil
	}
	return e.category
}

// SourceError reports an inspectable config source ingestion failure.
type SourceError struct {
	key             string
	source          string
	envName         string
	jsonPath        string
	jsonReaderLabel string
	kind            Kind
	diagnosticValue string
	redacted        bool
	category        error
	cause           error
}

func newSourceError(key string, source string, envName string, jsonPath string, jsonReaderLabel string, kind Kind, diagnosticValue string, redacted bool, category error, cause error) *SourceError {
	if redacted {
		diagnosticValue = ""
	}
	return &SourceError{
		key:             key,
		source:          source,
		envName:         envName,
		jsonPath:        jsonPath,
		jsonReaderLabel: jsonReaderLabel,
		kind:            kind,
		diagnosticValue: diagnosticValue,
		redacted:        redacted,
		category:        category,
		cause:           cause,
	}
}

func (e *SourceError) Error() string {
	if e == nil {
		return "config: source error"
	}
	message := fmt.Sprintf("config: %v", e.category)
	if e.key != "" {
		message += fmt.Sprintf(" for %q", e.key)
	}
	if e.kind.String() != "unknown" {
		message += fmt.Sprintf(" as %s", e.kind)
	}
	if e.source != "" {
		message += fmt.Sprintf(" from %s", e.source)
	}
	if e.envName != "" {
		message += fmt.Sprintf(" env %q", e.envName)
	}
	if e.jsonPath != "" {
		message += fmt.Sprintf(" path %q", e.jsonPath)
	}
	if e.jsonReaderLabel != "" {
		message += fmt.Sprintf(" reader %q", e.jsonReaderLabel)
	}
	if e.redacted {
		message += " value redacted"
	} else if e.diagnosticValue != "" {
		message += fmt.Sprintf(" value %q", e.diagnosticValue)
	}
	if e.cause != nil && !e.redacted {
		message += fmt.Sprintf(": %v", e.cause)
	}
	return message
}

func (e *SourceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *SourceError) Is(target error) bool {
	if e == nil {
		return false
	}
	return target == e.category || errors.Is(e.cause, target)
}

// Category returns the source diagnostic category.
func (e *SourceError) Category() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Key returns the config key associated with the source failure.
func (e *SourceError) Key() string {
	if e == nil {
		return ""
	}
	return e.key
}

// Source returns the provenance source label associated with the failure.
func (e *SourceError) Source() string {
	if e == nil {
		return ""
	}
	return e.source
}

// EnvName returns the environment variable associated with the failure.
func (e *SourceError) EnvName() string {
	if e == nil {
		return ""
	}
	return e.envName
}

// JSONPath returns the JSON file path associated with the failure.
func (e *SourceError) JSONPath() string {
	if e == nil {
		return ""
	}
	return e.jsonPath
}

// JSONReaderLabel returns the JSON reader label associated with the failure.
func (e *SourceError) JSONReaderLabel() string {
	if e == nil {
		return ""
	}
	return e.jsonReaderLabel
}

// Kind returns the expected value kind associated with the failure.
func (e *SourceError) Kind() Kind {
	if e == nil {
		return KindString
	}
	return e.kind
}

// Redacted reports whether the raw source value was omitted from diagnostics.
func (e *SourceError) Redacted() bool {
	if e == nil {
		return false
	}
	return e.redacted
}

// Cause returns the underlying source cause when it is safe to expose.
func (e *SourceError) Cause() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// GetError reports an inspectable typed getter retrieval failure.
type GetError struct {
	key         string
	kind        Kind
	wantKind    Kind
	sourceLabel string
	redacted    bool
	category    error
}

func newGetError(key string, kind Kind, wantKind Kind, sourceLabel string, redacted bool, category error) *GetError {
	return &GetError{
		key:         key,
		kind:        kind,
		wantKind:    wantKind,
		sourceLabel: sourceLabel,
		redacted:    redacted,
		category:    category,
	}
}

func (e *GetError) Error() string {
	if e == nil {
		return "config: get error"
	}
	message := fmt.Sprintf("config: %v for %q", e.category, e.key)
	if e.category != ErrKeyNotFound && e.kind.String() != "unknown" {
		message += fmt.Sprintf(" as %s", e.kind)
	}
	if e.sourceLabel != "" {
		message += fmt.Sprintf(" from %s", e.sourceLabel)
	}
	if e.redacted {
		message += " value redacted"
	}
	if e.category == ErrGetConversion {
		message += fmt.Sprintf(": want %s", e.wantKind)
	}
	return message
}

func (e *GetError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.category
}

func (e *GetError) Is(target error) bool {
	if e == nil {
		return false
	}
	return target == e.category
}

// Key returns the config key that failed retrieval.
func (e *GetError) Key() string {
	if e == nil {
		return ""
	}
	return e.key
}

// Kind returns the actual registered kind of the config key.
func (e *GetError) Kind() Kind {
	if e == nil {
		return KindString
	}
	return e.kind
}

// WantKind returns the kind the caller's getter requested.
func (e *GetError) WantKind() Kind {
	if e == nil {
		return KindString
	}
	return e.wantKind
}

// SourceLabel returns the provenance label when the value was present but wrong kind.
func (e *GetError) SourceLabel() string {
	if e == nil {
		return ""
	}
	return e.sourceLabel
}

// Redacted reports whether the key's raw value is sensitive and omitted from diagnostics.
func (e *GetError) Redacted() bool {
	if e == nil {
		return false
	}
	return e.redacted
}

// Category returns the sentinel error category for this retrieval failure.
func (e *GetError) Category() error {
	if e == nil {
		return nil
	}
	return e.category
}

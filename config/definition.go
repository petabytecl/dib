package config

import (
	"strings"
	"time"
)

// Definition is an immutable config key definition value.
type Definition struct {
	key           string
	usage         string
	defaultValue  any
	hasDefault    bool
	kind          Kind
	sensitive     bool
	invalidOption bool
}

type optionState struct {
	defaultValue any
	hasDefault   bool
	sensitive    bool
}

// Option configures config definition metadata.
type Option func(optionState) optionState

// Default records the optional default value for a definition.
func Default(value any) Option {
	return func(state optionState) optionState {
		state.defaultValue = clonePublicValue(value)
		state.hasDefault = true
		return state
	}
}

// Sensitive marks a definition whose raw values must not appear in diagnostics.
func Sensitive() Option {
	return func(state optionState) optionState {
		state.sensitive = true
		return state
	}
}

// Define returns a config key definition with an optional default.
func Define(key string, kind Kind, usage string, opts ...Option) Definition {
	state := optionState{}
	invalidOption := false
	for _, opt := range opts {
		if opt == nil {
			invalidOption = true
			continue
		}
		state = opt(state)
	}

	return Definition{
		key:           key,
		usage:         usage,
		defaultValue:  clonePublicValue(state.defaultValue),
		hasDefault:    state.hasDefault,
		kind:          kind,
		sensitive:     state.sensitive,
		invalidOption: invalidOption,
	}
}

// String returns a string config key definition with a default value.
func String(key string, defaultValue string, usage string, opts ...Option) Definition {
	return Define(key, KindString, usage, withDefault(defaultValue, opts)...)
}

// Bool returns a bool config key definition with a default value.
func Bool(key string, defaultValue bool, usage string, opts ...Option) Definition {
	return Define(key, KindBool, usage, withDefault(defaultValue, opts)...)
}

// Int returns an int config key definition with a default value.
func Int(key string, defaultValue int, usage string, opts ...Option) Definition {
	return Define(key, KindInt, usage, withDefault(defaultValue, opts)...)
}

// Int64 returns an int64 config key definition with a default value.
func Int64(key string, defaultValue int64, usage string, opts ...Option) Definition {
	return Define(key, KindInt64, usage, withDefault(defaultValue, opts)...)
}

// Uint returns a uint config key definition with a default value.
func Uint(key string, defaultValue uint, usage string, opts ...Option) Definition {
	return Define(key, KindUint, usage, withDefault(defaultValue, opts)...)
}

// Uint64 returns a uint64 config key definition with a default value.
func Uint64(key string, defaultValue uint64, usage string, opts ...Option) Definition {
	return Define(key, KindUint64, usage, withDefault(defaultValue, opts)...)
}

// Float64 returns a float64 config key definition with a default value.
func Float64(key string, defaultValue float64, usage string, opts ...Option) Definition {
	return Define(key, KindFloat64, usage, withDefault(defaultValue, opts)...)
}

// Duration returns a time.Duration config key definition with a default value.
func Duration(key string, defaultValue time.Duration, usage string, opts ...Option) Definition {
	return Define(key, KindDuration, usage, withDefault(defaultValue, opts)...)
}

// StringList returns a string-list config key definition with a default value.
func StringList(key string, defaultValue []string, usage string, opts ...Option) Definition {
	return Define(key, KindStringList, usage, withDefault(cloneStringSlice(defaultValue), opts)...)
}

func withDefault(value any, opts []Option) []Option {
	next := make([]Option, 0, len(opts)+1)
	next = append(next, Default(value))
	next = append(next, opts...)
	return next
}

// Name returns the stable config key name.
func (d Definition) Name() string {
	return d.key
}

// Kind returns the expected config value kind.
func (d Definition) Kind() Kind {
	return d.kind
}

// Usage returns the documentation text for this config key.
func (d Definition) Usage() string {
	return d.usage
}

// Default returns the optional default value and whether one was configured.
func (d Definition) Default() (any, bool) {
	return clonePublicValue(d.defaultValue), d.hasDefault
}

// Sensitive reports whether raw values for this definition are sensitive.
func (d Definition) Sensitive() bool {
	return d.sensitive
}

func validateDefinition(def Definition) error {
	if def.invalidOption {
		return newDefinitionError(def.key, "", "", def.kind, "", false, ErrInvalidDefinition)
	}
	if invalidKey(def.key) {
		return newDefinitionError(def.key, "", "", def.kind, "", false, ErrInvalidDefinition)
	}
	if !validKind(def.kind) {
		return newDefinitionError(def.key, "", "", def.kind, "", false, ErrInvalidDefinition)
	}
	if def.hasDefault && !valueMatchesKind(def.kind, def.defaultValue) {
		return newDefinitionError(def.key, "", "", def.kind, formatInvalidDefault(def), def.sensitive, ErrInvalidDefault)
	}
	return nil
}

func invalidKey(key string) bool {
	if strings.TrimSpace(key) == "" {
		return true
	}
	if strings.HasPrefix(key, "-") {
		return true
	}
	return strings.ContainsAny(key, " \t\r\n")
}

func validKind(kind Kind) bool {
	switch kind {
	case KindString, KindBool, KindInt, KindInt64, KindUint, KindUint64, KindFloat64, KindDuration, KindStringList:
		return true
	default:
		return false
	}
}

func valueMatchesKind(kind Kind, value any) bool {
	switch kind {
	case KindString:
		_, ok := value.(string)
		return ok
	case KindBool:
		_, ok := value.(bool)
		return ok
	case KindInt:
		_, ok := value.(int)
		return ok
	case KindInt64:
		_, ok := value.(int64)
		return ok
	case KindUint:
		_, ok := value.(uint)
		return ok
	case KindUint64:
		_, ok := value.(uint64)
		return ok
	case KindFloat64:
		_, ok := value.(float64)
		return ok
	case KindDuration:
		_, ok := value.(time.Duration)
		return ok
	case KindStringList:
		_, ok := value.([]string)
		return ok
	default:
		return false
	}
}

func clonePublicValue(value any) any {
	switch typed := value.(type) {
	case []string:
		return cloneStringSlice(typed)
	case []any:
		return cloneAnySlice(typed)
	default:
		return value
	}
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	copied := make([]string, len(values))
	copy(copied, values)
	return copied
}

func cloneAnySlice(values []any) []any {
	if values == nil {
		return nil
	}

	copied := make([]any, len(values))
	for i, value := range values {
		copied[i] = clonePublicValue(value)
	}
	return copied
}

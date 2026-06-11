package flags

import (
	"reflect"
	"strings"
	"time"
)

// Definition is an immutable flag definition value.
type Definition struct {
	name                 string
	shorthand            string
	hasShorthand         bool
	usage                string
	defaultValue         any
	kind                 Kind
	parser               Parser
	repeatPolicy         RepeatPolicy
	arity                Arity
	noOptionDefault      any
	hasNoOptionDefault   bool
	hidden               bool
	deprecated           string
	sensitive            bool
	customParserRequired bool
	invalidOption        bool
}

type optionState struct {
	shorthand          string
	hasShorthand       bool
	hidden             bool
	deprecated         string
	sensitive          bool
	noOptionDefault    any
	hasNoOptionDefault bool
	repeatPolicy       RepeatPolicy
}

// Option configures flag definition metadata.
type Option func(optionState) optionState

// Shorthand records a one-rune shorthand name for a definition.
func Shorthand(shorthand string) Option {
	return func(state optionState) optionState {
		state.shorthand = shorthand
		state.hasShorthand = true
		return state
	}
}

// Hidden marks a definition as hidden from generated usage surfaces.
func Hidden() Option {
	return func(state optionState) optionState {
		state.hidden = true
		return state
	}
}

// Deprecated records a deprecation diagnostic for a definition.
func Deprecated(message string) Option {
	return func(state optionState) optionState {
		state.deprecated = message
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

// NoOptionDefault records the value used when an option appears without a value.
func NoOptionDefault(value any) Option {
	return func(state optionState) optionState {
		state.noOptionDefault = clonePublicValue(value)
		state.hasNoOptionDefault = true
		return state
	}
}

// Repeatable records that repeated occurrences accumulate values.
func Repeatable() Option {
	return func(state optionState) optionState {
		state.repeatPolicy = RepeatAccumulated
		return state
	}
}

// String returns a string flag definition.
func String(name string, defaultValue string, usage string, opts ...Option) Definition {
	return newDefinition(name, KindString, defaultValue, usage, ParserFunc(stringParser), ArityRequired, false, opts...)
}

// Bool returns a bool flag definition.
func Bool(name string, defaultValue bool, usage string, opts ...Option) Definition {
	return newDefinition(name, KindBool, defaultValue, usage, ParserFunc(boolParser), ArityOptional, false, opts...)
}

// Int returns an int flag definition.
func Int(name string, defaultValue int, usage string, opts ...Option) Definition {
	return newDefinition(name, KindInt, defaultValue, usage, ParserFunc(intParser), ArityRequired, false, opts...)
}

// Int64 returns an int64 flag definition.
func Int64(name string, defaultValue int64, usage string, opts ...Option) Definition {
	return newDefinition(name, KindInt64, defaultValue, usage, ParserFunc(int64Parser), ArityRequired, false, opts...)
}

// Uint returns a uint flag definition.
func Uint(name string, defaultValue uint, usage string, opts ...Option) Definition {
	return newDefinition(name, KindUint, defaultValue, usage, ParserFunc(uintParser), ArityRequired, false, opts...)
}

// Uint64 returns a uint64 flag definition.
func Uint64(name string, defaultValue uint64, usage string, opts ...Option) Definition {
	return newDefinition(name, KindUint64, defaultValue, usage, ParserFunc(uint64Parser), ArityRequired, false, opts...)
}

// Float64 returns a float64 flag definition.
func Float64(name string, defaultValue float64, usage string, opts ...Option) Definition {
	return newDefinition(name, KindFloat64, defaultValue, usage, ParserFunc(float64Parser), ArityRequired, false, opts...)
}

// Duration returns a time.Duration flag definition.
func Duration(name string, defaultValue time.Duration, usage string, opts ...Option) Definition {
	return newDefinition(name, KindDuration, defaultValue, usage, ParserFunc(durationParser), ArityRequired, false, opts...)
}

// StringList returns a repeated string-list flag definition.
func StringList(name string, defaultValue []string, usage string, opts ...Option) Definition {
	return newDefinition(name, KindStringList, cloneStringSlice(defaultValue), usage, ParserFunc(stringListParser), ArityRequired, false, opts...)
}

// Custom returns a flag definition backed by a caller-supplied parser.
func Custom(name string, kind Kind, defaultValue any, usage string, parser Parser, opts ...Option) Definition {
	return newDefinition(name, kind, defaultValue, usage, parser, ArityRequired, true, opts...)
}

func newDefinition(name string, kind Kind, defaultValue any, usage string, parser Parser, arity Arity, customParserRequired bool, opts ...Option) Definition {
	state := optionState{repeatPolicy: RepeatLast}
	invalidOption := false
	for _, opt := range opts {
		if opt == nil {
			invalidOption = true
			continue
		}
		state = opt(state)
	}

	return Definition{
		name:                 name,
		shorthand:            state.shorthand,
		hasShorthand:         state.hasShorthand,
		usage:                usage,
		defaultValue:         clonePublicValue(defaultValue),
		kind:                 kind,
		parser:               parser,
		repeatPolicy:         state.repeatPolicy,
		arity:                arity,
		noOptionDefault:      clonePublicValue(state.noOptionDefault),
		hasNoOptionDefault:   state.hasNoOptionDefault,
		hidden:               state.hidden,
		deprecated:           state.deprecated,
		sensitive:            state.sensitive,
		customParserRequired: customParserRequired,
		invalidOption:        invalidOption,
	}
}

// Name returns the long flag name.
func (d Definition) Name() string {
	return d.name
}

// Kind returns the value kind.
func (d Definition) Kind() Kind {
	return d.kind
}

// Default returns a defensive copy of the default value where the value is mutable.
func (d Definition) Default() any {
	return clonePublicValue(d.defaultValue)
}

// Usage returns the usage text.
func (d Definition) Usage() string {
	return d.usage
}

// Shorthand returns the optional one-rune shorthand.
func (d Definition) Shorthand() (string, bool) {
	return d.shorthand, d.hasShorthand
}

// Hidden reports whether this definition should be omitted from usage surfaces.
func (d Definition) Hidden() bool {
	return d.hidden
}

// Sensitive reports whether raw values for this definition are sensitive.
func (d Definition) Sensitive() bool {
	return d.sensitive
}

// Deprecated returns the optional deprecation diagnostic.
func (d Definition) Deprecated() string {
	return d.deprecated
}

// NoOptionDefault returns the optional value used when no option value appears.
func (d Definition) NoOptionDefault() (any, bool) {
	return clonePublicValue(d.noOptionDefault), d.hasNoOptionDefault
}

// RepeatPolicy returns how repeated values are represented.
func (d Definition) RepeatPolicy() RepeatPolicy {
	return d.repeatPolicy
}

// Arity returns how this definition consumes CLI values.
func (d Definition) Arity() Arity {
	return d.arity
}

// Parse converts one raw value using this definition's parser.
func (d Definition) Parse(raw string) (any, error) {
	if d.parser == nil {
		return nil, newValueError(d.name, d.kind, ErrInvalidDefinition)
	}

	value, err := d.parser.ParseFlagValue(raw)
	if err != nil {
		if d.sensitive {
			return nil, newValueError(d.name, d.kind, nil)
		}
		return nil, newValueError(d.name, d.kind, err)
	}
	if !valueMatchesKind(d.kind, value) {
		return nil, newValueError(d.name, d.kind, ErrInvalidDefinition)
	}
	return clonePublicValue(value), nil
}

func validateDefinition(def Definition) error {
	if def.invalidOption {
		return newDefinitionError(def.name, "", ErrInvalidDefinition)
	}
	if invalidName(def.name) {
		return newDefinitionError(def.name, "", ErrInvalidDefinition)
	}
	if !validKind(def.kind) {
		return newDefinitionError(def.name, "", ErrInvalidDefinition)
	}
	if def.customParserRequired && parserIsNil(def.parser) {
		return newDefinitionError(def.name, "", ErrInvalidDefinition)
	}
	if !defaultMatchesKind(def) {
		return newDefinitionError(def.name, "", ErrInvalidDefinition)
	}
	if def.hasShorthand && invalidShorthand(def.shorthand) {
		return newDefinitionError(def.name, def.shorthand, ErrInvalidShorthand)
	}
	if def.hasNoOptionDefault && !validNoOptionDefault(def) {
		return newDefinitionError(def.name, "", ErrInvalidNoOptionDefault)
	}
	return nil
}

func invalidName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return true
	}
	if strings.HasPrefix(name, "-") {
		return true
	}
	return strings.ContainsAny(name, " \t\r\n")
}

func invalidShorthand(shorthand string) bool {
	if strings.TrimSpace(shorthand) == "" {
		return true
	}
	return len([]rune(shorthand)) != 1 || shorthand == "-"
}

func validKind(kind Kind) bool {
	switch kind {
	case KindString, KindBool, KindInt, KindInt64, KindUint, KindUint64, KindFloat64, KindDuration, KindStringList:
		return true
	default:
		return false
	}
}

func parserIsNil(parser Parser) bool {
	if parser == nil {
		return true
	}

	value := reflect.ValueOf(parser)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func defaultMatchesKind(def Definition) bool {
	if def.defaultValue == nil {
		return def.kind == KindStringList
	}
	return valueMatchesKind(def.kind, def.defaultValue)
}

func validNoOptionDefault(def Definition) bool {
	if valueMatchesKind(def.kind, def.noOptionDefault) {
		return true
	}

	raw, ok := def.noOptionDefault.(string)
	if !ok {
		return false
	}
	_, err := def.Parse(raw)
	return err == nil
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
	return append([]string(nil), values...)
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

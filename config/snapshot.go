package config

const (
	// SourceDefault is the provenance label for values supplied by registered defaults.
	SourceDefault = "default"
	// SourceExplicit is the provenance label for caller-supplied explicit values.
	SourceExplicit = "explicit setter"
	// SourceFlagBinding is the provenance label for values supplied by explicit flag bindings.
	SourceFlagBinding = "flag binding"
	// SourceEnv is the provenance label for values supplied by injected env lookup.
	SourceEnv = "env"
	// SourceJSON is the provenance label for values supplied by JSON readers or paths.
	SourceJSON = "JSON"
)

// Snapshot is a self-contained view of config values for one resolution run.
type Snapshot struct {
	values     []Value
	byKey      map[string]int
	normalizer NameNormalizer
}

// Lookup returns the value state for a config key.
func (s Snapshot) Lookup(key string) (Value, bool) {
	if invalidKey(key) {
		return Value{}, false
	}
	normalizedKey := normalizeName(s.normalizer, key)
	index, ok := s.byKey[normalizedKey]
	if !ok {
		return Value{}, false
	}
	if index < 0 || index >= len(s.values) {
		return Value{}, false
	}
	return s.values[index].clone(), true
}

// Value records the resolved value state for one config key.
type Value struct {
	definition    Definition
	hasDefinition bool
	value         any
	hasValue      bool
	provenance    string
	source        Source
}

func newDefaultValue(def Definition) Value {
	value, hasDefault := def.Default()
	provenance := ""
	if hasDefault {
		provenance = SourceDefault
	}
	return Value{
		definition:    def,
		hasDefinition: true,
		value:         value,
		hasValue:      hasDefault,
		provenance:    provenance,
		source: Source{
			key:      def.key,
			label:    provenance,
			redacted: def.sensitive,
		},
	}
}

func newAbsentSourceValue(def Definition) Value {
	return Value{
		definition:    def,
		hasDefinition: true,
		source: Source{
			key:      def.key,
			redacted: def.sensitive,
		},
	}
}

func newSourceValue(def Definition, value any, source Source) Value {
	source.key = def.key
	source.redacted = def.sensitive
	return Value{
		definition:    def,
		hasDefinition: true,
		value:         clonePublicValue(value),
		hasValue:      true,
		provenance:    source.label,
		source:        source,
	}
}

func (v Value) clone() Value {
	return Value{
		definition:    v.definition,
		hasDefinition: v.hasDefinition,
		value:         clonePublicValue(v.value),
		hasValue:      v.hasValue,
		provenance:    v.provenance,
		source:        v.source,
	}
}

// Definition returns the registered definition associated with this value.
func (v Value) Definition() (Definition, bool) {
	return v.definition, v.hasDefinition
}

// Value returns the resolved value and whether a value is present.
func (v Value) Value() (any, bool) {
	return clonePublicValue(v.value), v.hasValue
}

// Provenance returns the source label for the resolved value.
func (v Value) Provenance() string {
	return v.provenance
}

// Source returns metadata describing where this value came from.
func (v Value) Source() Source {
	return v.source
}

// Source records safe provenance metadata for a value.
type Source struct {
	key             string
	label           string
	envName         string
	jsonPath        string
	jsonReaderLabel string
	redacted        bool
}

// Key returns the canonical registered config key for this source value.
func (s Source) Key() string {
	return s.key
}

// Label returns the provenance source label.
func (s Source) Label() string {
	return s.label
}

// EnvName returns the environment variable name used for env source values.
func (s Source) EnvName() string {
	return s.envName
}

// JSONPath returns the caller-supplied path used for JSON file source values.
func (s Source) JSONPath() string {
	return s.jsonPath
}

// JSONReaderLabel returns the caller-supplied label for JSON reader source values.
func (s Source) JSONReaderLabel() string {
	return s.jsonReaderLabel
}

// Redacted reports whether raw values for this source must be omitted from diagnostics.
func (s Source) Redacted() bool {
	return s.redacted
}

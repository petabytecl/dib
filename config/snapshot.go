package config

// SourceDefault is the provenance label for values supplied by registered defaults.
const SourceDefault = "default"

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
	}
}

func (v Value) clone() Value {
	return Value{
		definition:    v.definition,
		hasDefinition: v.hasDefinition,
		value:         clonePublicValue(v.value),
		hasValue:      v.hasValue,
		provenance:    v.provenance,
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

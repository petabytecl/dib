package config

// Set is an immutable collection of config key definitions and lookup indexes.
type Set struct {
	definitions []Definition
	byKey       map[string]int
	normalizer  NameNormalizer
}

// NewSet validates and returns an independent config definition set.
func NewSet(defs ...Definition) (Set, error) {
	return newSet(nil, defs...)
}

// NewNormalizedSet validates and returns a set that normalizes config keys for lookup.
func NewNormalizedSet(normalizer NameNormalizer, defs ...Definition) (Set, error) {
	return newSet(normalizer, defs...)
}

func newSet(normalizer NameNormalizer, defs ...Definition) (Set, error) {
	definitions := make([]Definition, 0, len(defs))
	byExactKey := make(map[string]int, len(defs))
	byKey := make(map[string]int, len(defs))

	for _, def := range defs {
		if err := validateDefinition(def); err != nil {
			return Set{}, err
		}
		if _, ok := byExactKey[def.key]; ok {
			return Set{}, newDefinitionError(def.key, "", "", def.kind, "", false, ErrDuplicateKey)
		}

		normalizedKey := normalizeName(normalizer, def.key)
		if invalidKey(normalizedKey) {
			return Set{}, newDefinitionError(def.key, "", normalizedKey, def.kind, "", false, ErrInvalidDefinition)
		}
		if existing, ok := byKey[normalizedKey]; ok {
			return Set{}, newDefinitionError(def.key, definitions[existing].key, normalizedKey, def.kind, "", false, ErrDuplicateNormalizedKey)
		}

		byExactKey[def.key] = len(definitions)
		byKey[normalizedKey] = len(definitions)
		definitions = append(definitions, def)
	}

	return Set{
		definitions: append([]Definition(nil), definitions...),
		byKey:       cloneIndex(byKey),
		normalizer:  normalizer,
	}, nil
}

// Len returns the number of definitions in the set.
func (s Set) Len() int {
	return len(s.definitions)
}

// Definitions returns definitions in deterministic registration order.
func (s Set) Definitions() []Definition {
	return append([]Definition(nil), s.definitions...)
}

// Lookup returns the definition for a config key.
func (s Set) Lookup(key string) (Definition, bool) {
	if invalidKey(key) {
		return Definition{}, false
	}

	index, ok := s.byKey[normalizeName(s.normalizer, key)]
	if !ok {
		return Definition{}, false
	}
	return s.definitions[index], true
}

// With returns a new Set containing existing definitions followed by defs.
func (s Set) With(defs ...Definition) (Set, error) {
	next := s.Definitions()
	next = append(next, defs...)
	return newSet(s.normalizer, next...)
}

// WithNormalizer returns a new Set with the same definitions and a new key normalizer.
func (s Set) WithNormalizer(normalizer NameNormalizer) (Set, error) {
	return newSet(normalizer, s.definitions...)
}

// DefaultSnapshot returns the default value state for this set.
func (s Set) DefaultSnapshot() Snapshot {
	values := make([]Value, 0, len(s.definitions))
	for _, def := range s.definitions {
		values = append(values, newDefaultValue(def))
	}
	return Snapshot{
		values:     values,
		byKey:      cloneIndex(s.byKey),
		normalizer: s.normalizer,
	}
}

func cloneIndex(index map[string]int) map[string]int {
	if index == nil {
		return nil
	}
	copied := make(map[string]int, len(index))
	for key, value := range index {
		copied[key] = value
	}
	return copied
}

package flags

// Set is an immutable collection of flag definitions and lookup indexes.
type Set struct {
	definitions []Definition
	byExactName map[string]int
	byName      map[string]int
	byShort     map[string]int
	normalizer  NameNormalizer
}

// NewSet validates and returns an independent flag definition set.
func NewSet(defs ...Definition) (Set, error) {
	return newSet(nil, defs...)
}

// NewNormalizedSet validates and returns a set that normalizes long flag names for lookup.
func NewNormalizedSet(normalizer NameNormalizer, defs ...Definition) (Set, error) {
	return newSet(normalizer, defs...)
}

func newSet(normalizer NameNormalizer, defs ...Definition) (Set, error) {
	definitions := make([]Definition, 0, len(defs))
	byExactName := make(map[string]int, len(defs))
	byName := make(map[string]int, len(defs))
	byShort := make(map[string]int)

	for _, def := range defs {
		if err := validateDefinition(def); err != nil {
			return Set{}, err
		}
		if _, ok := byExactName[def.name]; ok {
			return Set{}, newDefinitionError(def.name, "", ErrDuplicateName)
		}
		if def.hasShorthand {
			if _, ok := byShort[def.shorthand]; ok {
				return Set{}, newDefinitionError("", def.shorthand, ErrDuplicateShorthand)
			}
			byShort[def.shorthand] = len(definitions)
		}

		normalizedName := normalizeName(normalizer, def.name)
		if invalidName(normalizedName) {
			return Set{}, newInvalidNormalizedDefinitionError(def.name, normalizedName)
		}
		if existing, ok := byName[normalizedName]; ok {
			return Set{}, newNormalizedDefinitionError(def.name, definitions[existing].name, normalizedName)
		}

		byExactName[def.name] = len(definitions)
		byName[normalizedName] = len(definitions)
		definitions = append(definitions, def)
	}

	return Set{
		definitions: append([]Definition(nil), definitions...),
		byExactName: cloneIndex(byExactName),
		byName:      cloneIndex(byName),
		byShort:     cloneIndex(byShort),
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

// Lookup returns the definition for a long flag name.
func (s Set) Lookup(name string) (Definition, bool) {
	if invalidName(name) {
		return Definition{}, false
	}
	if _, shorthand := s.byShort[name]; shorthand {
		if _, exactName := s.byExactName[name]; !exactName {
			return Definition{}, false
		}
	}

	index, ok := s.byName[normalizeName(s.normalizer, name)]
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

// WithNormalizer returns a new Set with the same definitions and a new long-name normalizer.
func (s Set) WithNormalizer(normalizer NameNormalizer) (Set, error) {
	return newSet(normalizer, s.definitions...)
}

// DefaultSnapshot returns the default value state for this set.
func (s Set) DefaultSnapshot() Snapshot {
	values := make(map[string]ValueState, len(s.definitions))
	for _, def := range s.definitions {
		values[def.name] = newDefaultValueState(def)
	}
	return Snapshot{values: values}
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

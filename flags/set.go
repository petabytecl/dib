package flags

// Set is an immutable collection of flag definitions and lookup indexes.
type Set struct {
	definitions []Definition
	byName      map[string]int
	byShort     map[string]int
}

// NewSet validates and returns an independent flag definition set.
func NewSet(defs ...Definition) (Set, error) {
	definitions := make([]Definition, 0, len(defs))
	byName := make(map[string]int, len(defs))
	byShort := make(map[string]int)

	for _, def := range defs {
		if err := validateDefinition(def); err != nil {
			return Set{}, err
		}
		if _, ok := byName[def.name]; ok {
			return Set{}, newDefinitionError(def.name, "", ErrDuplicateName)
		}
		if def.hasShorthand {
			if _, ok := byShort[def.shorthand]; ok {
				return Set{}, newDefinitionError("", def.shorthand, ErrDuplicateShorthand)
			}
			byShort[def.shorthand] = len(definitions)
		}

		byName[def.name] = len(definitions)
		definitions = append(definitions, def)
	}

	return Set{
		definitions: append([]Definition(nil), definitions...),
		byName:      cloneIndex(byName),
		byShort:     cloneIndex(byShort),
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
	index, ok := s.byName[name]
	if !ok {
		return Definition{}, false
	}
	return s.definitions[index], true
}

// With returns a new Set containing existing definitions followed by defs.
func (s Set) With(defs ...Definition) (Set, error) {
	next := s.Definitions()
	next = append(next, defs...)
	return NewSet(next...)
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

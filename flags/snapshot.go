package flags

// Snapshot is a self-contained view of flag values for one parse run.
type Snapshot struct {
	values    map[string]ValueState
	remaining []string
}

// Lookup returns the value state for a long flag name.
func (s Snapshot) Lookup(name string) (ValueState, bool) {
	state, ok := s.values[name]
	if !ok {
		return ValueState{}, false
	}
	return state.clone(), true
}

// RemainingArgs returns the positional arguments left after parsing flags.
func (s Snapshot) RemainingArgs() []string {
	return append([]string(nil), s.remaining...)
}

// ValueState records the value state for one flag in a snapshot.
type ValueState struct {
	defaultValue any
	values       []any
	explicit     bool
	arity        Arity
	occurrences  []ValueOccurrence
}

func newDefaultValueState(def Definition) ValueState {
	return ValueState{
		defaultValue: clonePublicValue(def.defaultValue),
		values:       defaultValues(def),
		explicit:     false,
		arity:        def.arity,
		occurrences:  nil,
	}
}

func defaultValues(def Definition) []any {
	if def.kind == KindStringList {
		values, ok := def.defaultValue.([]string)
		if !ok {
			return nil
		}

		copied := make([]any, len(values))
		for i, value := range values {
			copied[i] = value
		}
		return copied
	}

	return []any{clonePublicValue(def.defaultValue)}
}

func (s ValueState) clone() ValueState {
	return ValueState{
		defaultValue: clonePublicValue(s.defaultValue),
		values:       cloneAnySlice(s.values),
		explicit:     s.explicit,
		arity:        s.arity,
		occurrences:  cloneOccurrences(s.occurrences),
	}
}

// Default returns the definition default captured in the snapshot.
func (s ValueState) Default() any {
	return clonePublicValue(s.defaultValue)
}

// Values returns the effective values captured in the snapshot.
func (s ValueState) Values() []any {
	return cloneAnySlice(s.values)
}

// Explicit reports whether this state came from explicit caller input.
func (s ValueState) Explicit() bool {
	return s.explicit
}

// Arity returns the value arity captured with the state.
func (s ValueState) Arity() Arity {
	return s.arity
}

// Occurrences returns the source occurrences that explicitly set this value.
func (s ValueState) Occurrences() []ValueOccurrence {
	return cloneOccurrences(s.occurrences)
}

// ValueOccurrence records one explicit CLI occurrence of a flag value.
type ValueOccurrence struct {
	spelling       string
	normalizedName string
	definition     Definition
}

func newValueOccurrence(spelling string, normalizedName string, definition Definition) ValueOccurrence {
	return ValueOccurrence{spelling: spelling, normalizedName: normalizedName, definition: definition}
}

// Spelling returns the raw source flag spelling without any attached value.
func (o ValueOccurrence) Spelling() string {
	return o.spelling
}

// NormalizedName returns the lookup key used for this occurrence.
func (o ValueOccurrence) NormalizedName() string {
	return o.normalizedName
}

// Definition returns the canonical definition matched by this occurrence.
func (o ValueOccurrence) Definition() Definition {
	return o.definition
}

func cloneOccurrences(occurrences []ValueOccurrence) []ValueOccurrence {
	if occurrences == nil {
		return nil
	}
	return append([]ValueOccurrence(nil), occurrences...)
}

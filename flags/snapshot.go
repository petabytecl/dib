package flags

// Snapshot is a self-contained view of flag values for one parse run.
type Snapshot struct {
	values map[string]ValueState
}

// Lookup returns the value state for a long flag name.
func (s Snapshot) Lookup(name string) (ValueState, bool) {
	state, ok := s.values[name]
	if !ok {
		return ValueState{}, false
	}
	return state.clone(), true
}

// ValueState records the value state for one flag in a snapshot.
type ValueState struct {
	defaultValue any
	values       []any
	explicit     bool
	arity        Arity
}

func newDefaultValueState(def Definition) ValueState {
	return ValueState{
		defaultValue: clonePublicValue(def.defaultValue),
		values:       defaultValues(def),
		explicit:     false,
		arity:        def.arity,
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

package flags

import "strings"

// Parse parses caller-supplied command-line arguments into an independent snapshot.
func (s Set) Parse(args []string) (Snapshot, error) {
	snapshot := s.DefaultSnapshot()

	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			snapshot.remaining = append(snapshot.remaining, args[i+1:]...)
			break
		}

		next, err := s.parseArg(args, i, &snapshot)
		if err != nil {
			return Snapshot{}, err
		}
		i = next
	}

	return snapshot, nil
}

// parseArg dispatches a single token to the long, short, or positional handler
// and returns the index of the last argument it consumed.
func (s Set) parseArg(args []string, index int, snapshot *Snapshot) (int, error) {
	arg := args[index]
	switch {
	case isLongFlagToken(arg):
		next, consumed, err := s.parseLong(args, index, snapshot)
		return resumeIndex(index, next, consumed, err)
	case isShortFlagToken(arg):
		next, consumed, err := s.parseShort(args, index, snapshot)
		return resumeIndex(index, next, consumed, err)
	default:
		snapshot.remaining = append(snapshot.remaining, arg)
		return index, nil
	}
}

// resumeIndex collapses a (next, consumed, err) handler result into the index
// the caller should resume from.
func resumeIndex(index, next int, consumed bool, err error) (int, error) {
	if err != nil {
		return index, err
	}
	if consumed {
		return next, nil
	}
	return index, nil
}

func (s Set) parseShort(args []string, index int, snapshot *Snapshot) (int, bool, error) {
	arg := args[index]
	token, shorthand, rawValue, hasAttachedValue := splitShortFlag(arg)
	if !hasAttachedValue && len([]rune(shorthand)) > 1 {
		return s.parseShortGroup(args, index, snapshot, shorthand)
	}

	def, ok := s.lookupShorthand(shorthand)
	if !ok {
		if shorthand == "h" {
			return index, false, newParseError(ErrHelpRequest, "--help", "help", "", Definition{}, false, nil)
		}
		return index, false, newParseError(ErrUnknownFlag, token, shorthand, "", Definition{}, false, nil)
	}
	return parseResolvedFlag(args, index, snapshot, def, token, shorthand, def.name, rawValue, hasAttachedValue, false)
}

func (s Set) parseLong(args []string, index int, snapshot *Snapshot) (int, bool, error) {
	arg := args[index]
	token, name, rawValue, hasAttachedValue := splitLongFlag(arg)
	normalizedName := normalizeName(s.normalizer, name)

	def, ok := s.Lookup(name)
	if !ok {
		if name == "help" {
			return index, false, newParseError(ErrHelpRequest, token, name, normalizedName, Definition{}, false, nil)
		}
		return index, false, newParseError(ErrUnknownFlag, token, name, normalizedName, Definition{}, false, nil)
	}
	return parseResolvedFlag(args, index, snapshot, def, token, name, normalizedName, rawValue, hasAttachedValue, true)
}

func (s Set) parseShortGroup(args []string, index int, snapshot *Snapshot, group string) (int, bool, error) {
	members := []rune(group)
	for i := range members {
		shorthand := string(members[i])
		token := shortGroupToken(members, i)
		occurrence := "-" + shorthand

		def, ok := s.lookupShorthand(shorthand)
		if !ok {
			return index, false, newParseError(ErrUnknownFlag, token, shorthand, "", Definition{}, false, nil)
		}

		if i == len(members)-1 {
			return parseResolvedFlagWithOccurrence(args, index, snapshot, def, token, occurrence, shorthand, def.name, "", false, false)
		}

		nextKnown := false
		if i+1 < len(members) {
			_, nextKnown = s.lookupShorthand(string(members[i+1]))
		}

		if def.arity == ArityOptional || (hasNoOptionDefault(def) && nextKnown) {
			if err := applyNoOptionParsedValue(snapshot, def, token, occurrence, shorthand, def.name); err != nil {
				return index, false, err
			}
			continue
		}
		if nextKnown {
			return index, false, newParseError(ErrInvalidGroup, token, shorthand, def.name, def, true, nil)
		}

		rawValue := string(members[i+1:])
		return parseResolvedFlagWithOccurrence(args, index, snapshot, def, token, occurrence, shorthand, def.name, rawValue, true, false)
	}

	return index, false, nil
}

func parseResolvedFlag(
	args []string,
	index int,
	snapshot *Snapshot,
	def Definition,
	token string,
	name string,
	lookupKey string,
	rawValue string,
	hasAttachedValue bool,
	stopBeforeLong bool,
) (int, bool, error) {
	return parseResolvedFlagWithOccurrence(args, index, snapshot, def, token, token, name, lookupKey, rawValue, hasAttachedValue, stopBeforeLong)
}

// resolvedRawValue carries how an unattached flag value was resolved. When done
// is true the value has already been applied and the caller should stop.
type resolvedRawValue struct {
	raw      string
	consumed bool
	done     bool
}

func parseResolvedFlagWithOccurrence(
	args []string,
	index int,
	snapshot *Snapshot,
	def Definition,
	token string,
	occurrence string,
	name string,
	lookupKey string,
	rawValue string,
	hasAttachedValue bool,
	stopBeforeLong bool,
) (int, bool, error) {
	if state, ok := snapshot.values[def.name]; ok && state.explicit && def.repeatPolicy != RepeatAccumulated {
		return index, false, newParseError(ErrDuplicateValue, token, name, lookupKey, def, true, nil)
	}

	valueRaw := rawValue
	consumedNext := false
	if !hasAttachedValue {
		resolved, err := resolveUnattachedValue(args, index, snapshot, def, token, occurrence, name, lookupKey, stopBeforeLong)
		if err != nil {
			return index, false, err
		}
		if resolved.done {
			return index, false, nil
		}
		valueRaw = resolved.raw
		consumedNext = resolved.consumed
	}

	value, err := def.Parse(valueRaw)
	if err != nil {
		return index, false, newParseError(ErrConversion, token, name, lookupKey, def, true, err)
	}
	if err = applyParsedValue(snapshot, def, token, occurrence, name, lookupKey, value); err != nil {
		return index, false, err
	}
	if consumedNext {
		return index + 1, true, nil
	}
	return index, false, nil
}

// resolveUnattachedValue determines the raw value for a flag supplied without an
// attached "=value", honouring arity and no-option defaults.
func resolveUnattachedValue(
	args []string,
	index int,
	snapshot *Snapshot,
	def Definition,
	token, occurrence, name, lookupKey string,
	stopBeforeLong bool,
) (resolvedRawValue, error) {
	switch def.arity {
	case ArityOptional:
		value, err := noOptionValue(def)
		if err != nil {
			return resolvedRawValue{done: true}, newParseError(ErrConversion, token, name, lookupKey, def, true, err)
		}
		return resolvedRawValue{done: true}, applyParsedValue(snapshot, def, token, occurrence, name, lookupKey, value)
	case ArityRequired:
		if hasFollowingValue(args, index, stopBeforeLong) {
			return resolvedRawValue{raw: args[index+1], consumed: true}, nil
		}
		if hasNoOptionDefault(def) {
			return resolvedRawValue{done: true}, applyNoOptionParsedValue(snapshot, def, token, occurrence, name, lookupKey)
		}
		return resolvedRawValue{done: true}, newParseError(ErrMissingValue, token, name, lookupKey, def, true, nil)
	default:
		return resolvedRawValue{}, nil
	}
}

// hasFollowingValue reports whether args[index+1] can serve as a required flag's value.
func hasFollowingValue(args []string, index int, stopBeforeLong bool) bool {
	if index+1 >= len(args) {
		return false
	}
	next := args[index+1]
	return next != "--" && (!stopBeforeLong || !isLongFlagToken(next))
}

func splitLongFlag(arg string) (string, string, string, bool) {
	body := strings.TrimPrefix(arg, "--")
	namePart, value, found := strings.Cut(body, "=")
	return "--" + namePart, namePart, value, found
}

func splitShortFlag(arg string) (string, string, string, bool) {
	body := strings.TrimPrefix(arg, "-")
	namePart, value, found := strings.Cut(body, "=")
	return "-" + namePart, namePart, value, found
}

func isLongFlagToken(arg string) bool {
	return strings.HasPrefix(arg, "--") && arg != "--"
}

func isShortFlagToken(arg string) bool {
	return strings.HasPrefix(arg, "-") && arg != "-" && !strings.HasPrefix(arg, "--")
}

func (s Set) lookupShorthand(shorthand string) (Definition, bool) {
	index, ok := s.byShort[shorthand]
	if !ok {
		return Definition{}, false
	}
	return s.definitions[index], true
}

func shortGroupToken(members []rune, index int) string {
	return "-" + string(members[:index+1])
}

func hasNoOptionDefault(def Definition) bool {
	_, ok := def.NoOptionDefault()
	return ok
}

func noOptionValue(def Definition) (any, error) {
	if value, ok := def.NoOptionDefault(); ok {
		if valueMatchesKind(def.kind, value) {
			return clonePublicValue(value), nil
		}
		raw, _ := value.(string)
		return def.Parse(raw)
	}
	if def.kind == KindBool {
		return def.Parse("true")
	}
	return nil, newValueError(def.name, def.kind, ErrMissingValue)
}

func applyNoOptionParsedValue(snapshot *Snapshot, def Definition, token, occurrence, name, normalizedName string) error {
	value, err := noOptionValue(def)
	if err != nil {
		return newParseError(ErrConversion, token, name, normalizedName, def, true, err)
	}
	return applyParsedValue(snapshot, def, token, occurrence, name, normalizedName, value)
}

func applyParsedValue(snapshot *Snapshot, def Definition, token, occurrence, name, normalizedName string, value any) error {
	state := snapshot.values[def.name]
	if state.explicit && def.repeatPolicy != RepeatAccumulated {
		return newParseError(ErrDuplicateValue, token, name, normalizedName, def, true, nil)
	}

	values := parsedValues(value)
	if state.explicit && def.repeatPolicy == RepeatAccumulated {
		state.values = append(state.values, values...)
	} else {
		state.values = values
	}
	state.explicit = true
	state.occurrences = append(state.occurrences, newValueOccurrence(occurrence, normalizedName, def))
	snapshot.values[def.name] = state
	return nil
}

func parsedValues(value any) []any {
	if values, ok := value.([]string); ok {
		parsed := make([]any, len(values))
		for i, value := range values {
			parsed[i] = value
		}
		return parsed
	}
	return []any{clonePublicValue(value)}
}

package flags

import "strings"

// Parse parses caller-supplied command-line arguments into an independent snapshot.
func (s Set) Parse(args []string) (Snapshot, error) {
	snapshot := s.DefaultSnapshot()

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			snapshot.remaining = append(snapshot.remaining, args[i+1:]...)
			break
		}
		if !isLongFlagToken(arg) {
			if isShortFlagToken(arg) {
				next, consumed, err := s.parseShort(args, i, &snapshot)
				if err != nil {
					return Snapshot{}, err
				}
				if consumed {
					i = next
				}
				continue
			}
			snapshot.remaining = append(snapshot.remaining, arg)
			continue
		}

		next, consumed, err := s.parseLong(args, i, &snapshot)
		if err != nil {
			return Snapshot{}, err
		}
		if consumed {
			i = next
		}
	}

	return snapshot, nil
}

func (s Set) parseShort(args []string, index int, snapshot *Snapshot) (int, bool, error) {
	arg := args[index]
	token, shorthand, rawValue, hasAttachedValue := splitShortFlag(arg)

	def, ok := s.lookupShorthand(shorthand)
	if !ok {
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
		return index, false, newParseError(ErrUnknownFlag, token, name, normalizedName, Definition{}, false, nil)
	}
	return parseResolvedFlag(args, index, snapshot, def, token, name, normalizedName, rawValue, hasAttachedValue, true)
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
	if state, ok := snapshot.values[def.name]; ok && state.explicit && def.repeatPolicy != RepeatAccumulated {
		return index, false, newParseError(ErrDuplicateValue, token, name, lookupKey, def, true, nil)
	}

	valueRaw := rawValue
	consumedNext := false
	if !hasAttachedValue {
		switch def.arity {
		case ArityOptional:
			value, err := noOptionValue(def)
			if err != nil {
				return index, false, newParseError(ErrConversion, token, name, lookupKey, def, true, err)
			}
			if err := applyParsedValue(snapshot, def, token, name, lookupKey, value); err != nil {
				return index, false, err
			}
			return index, false, nil
		case ArityRequired:
			if index+1 >= len(args) || args[index+1] == "--" || (stopBeforeLong && isLongFlagToken(args[index+1])) {
				return index, false, newParseError(ErrMissingValue, token, name, lookupKey, def, true, nil)
			}
			valueRaw = args[index+1]
			consumedNext = true
		default:
			valueRaw = ""
		}
	}

	value, err := def.Parse(valueRaw)
	if err != nil {
		return index, false, newParseError(ErrConversion, token, name, lookupKey, def, true, err)
	}
	if err := applyParsedValue(snapshot, def, token, name, lookupKey, value); err != nil {
		return index, false, err
	}
	if consumedNext {
		return index + 1, true, nil
	}
	return index, false, nil
}

func splitLongFlag(arg string) (token string, name string, rawValue string, hasAttachedValue bool) {
	body := strings.TrimPrefix(arg, "--")
	namePart, value, found := strings.Cut(body, "=")
	return "--" + namePart, namePart, value, found
}

func splitShortFlag(arg string) (token string, shorthand string, rawValue string, hasAttachedValue bool) {
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

func applyParsedValue(snapshot *Snapshot, def Definition, token string, name string, normalizedName string, value any) error {
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
	state.occurrences = append(state.occurrences, newValueOccurrence(token, normalizedName, def))
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

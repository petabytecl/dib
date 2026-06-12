package config

import "time"

// IsSet reports whether key has a resolved value in this snapshot, including defaults.
// Returns false for unregistered keys and registered keys with no value from any source
// and no default. Returns true for explicit zero values, empty strings, and nil-typed defaults.
func (s Snapshot) IsSet(key string) bool {
	v, ok := s.Lookup(key)
	if !ok {
		return false
	}
	_, hasValue := v.Value()
	return hasValue
}

// getTyped is an unexported generic helper shared by all typed getters.
// It handles the lookup, absence check, and type assertion in one place.
// wantKind identifies the kind the caller's getter expects (used in error reporting).
// convert performs the type assertion from any to the expected Go type T.
func getTyped[T any](s Snapshot, key string, wantKind Kind, convert func(any) (T, bool)) (T, error) {
	var zero T
	v, ok := s.Lookup(key)
	if !ok {
		return zero, newGetError(key, KindString, wantKind, "", false, ErrKeyNotFound)
	}
	def, _ := v.Definition()
	raw, hasValue := v.Value()
	if !hasValue {
		return zero, newGetError(key, def.kind, wantKind, "", def.sensitive, ErrKeyAbsent)
	}
	if def.kind != wantKind {
		return zero, newGetError(key, def.kind, wantKind, v.Provenance(), def.sensitive, ErrGetConversion)
	}
	val, ok := convert(raw)
	if !ok {
		return zero, newGetError(key, def.kind, wantKind, v.Provenance(), def.sensitive, ErrGetConversion)
	}
	return val, nil
}

// GetString returns the resolved string value for key.
// Returns ("", *GetError) if key is unregistered (ErrKeyNotFound),
// has no value (ErrKeyAbsent), or is not KindString (ErrGetConversion).
func (s Snapshot) GetString(key string) (string, error) {
	return getTyped(s, key, KindString, func(v any) (string, bool) {
		typed, ok := v.(string)
		return typed, ok
	})
}

// GetBool returns the resolved bool value for key.
func (s Snapshot) GetBool(key string) (bool, error) {
	return getTyped(s, key, KindBool, func(v any) (bool, bool) {
		typed, ok := v.(bool)
		return typed, ok
	})
}

// GetInt returns the resolved int value for key.
func (s Snapshot) GetInt(key string) (int, error) {
	return getTyped(s, key, KindInt, func(v any) (int, bool) {
		typed, ok := v.(int)
		return typed, ok
	})
}

// GetInt64 returns the resolved int64 value for key.
func (s Snapshot) GetInt64(key string) (int64, error) {
	return getTyped(s, key, KindInt64, func(v any) (int64, bool) {
		typed, ok := v.(int64)
		return typed, ok
	})
}

// GetUint returns the resolved uint value for key.
func (s Snapshot) GetUint(key string) (uint, error) {
	return getTyped(s, key, KindUint, func(v any) (uint, bool) {
		typed, ok := v.(uint)
		return typed, ok
	})
}

// GetUint64 returns the resolved uint64 value for key.
func (s Snapshot) GetUint64(key string) (uint64, error) {
	return getTyped(s, key, KindUint64, func(v any) (uint64, bool) {
		typed, ok := v.(uint64)
		return typed, ok
	})
}

// GetFloat64 returns the resolved float64 value for key.
func (s Snapshot) GetFloat64(key string) (float64, error) {
	return getTyped(s, key, KindFloat64, func(v any) (float64, bool) {
		typed, ok := v.(float64)
		return typed, ok
	})
}

// GetDuration returns the resolved time.Duration value for key.
func (s Snapshot) GetDuration(key string) (time.Duration, error) {
	return getTyped(s, key, KindDuration, func(v any) (time.Duration, bool) {
		typed, ok := v.(time.Duration)
		return typed, ok
	})
}

// GetStringList returns a defensive copy of the resolved []string value for key.
func (s Snapshot) GetStringList(key string) ([]string, error) {
	val, err := getTyped(s, key, KindStringList, func(v any) ([]string, bool) {
		typed, ok := v.([]string)
		return typed, ok
	})
	if err != nil {
		return nil, err
	}
	return append([]string(nil), val...), nil
}

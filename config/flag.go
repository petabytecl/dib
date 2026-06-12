package config

import "fmt"

// FlagValue carries a translated flag binding for a registered config key.
// The caller constructs these from a parsed flag snapshot without importing flags/.
type FlagValue struct {
	// ConfigKey is the registered config key this binding targets.
	ConfigKey string
	// ExplicitlySet is true only when the flag was explicitly provided on the CLI (not its default).
	ExplicitlySet bool
	// Value is the parsed flag value; only consumed when ExplicitlySet is true.
	Value any
}

// NewFlagSnapshot ingests caller-supplied flag bindings for registered config keys.
//
// Each FlagValue.ConfigKey must be registered in the set; an unknown key returns a
// typed *SourceError with ErrUnknownSourceKey. Duplicate ConfigKey entries (including
// entries that normalize to the same key) return a typed *SourceError with ErrDuplicateBinding
// before any values are stored. When ExplicitlySet is false, no value is injected for
// that tier — the flag's default does not enter the flag binding source.
func NewFlagSnapshot(set Set, values []FlagValue) (Snapshot, error) {
	snapshot := newSourceSnapshot(set)
	seen := make(map[string]struct{}, len(values))
	for _, fv := range values {
		def, index, ok := lookupSourceDefinition(set, fv.ConfigKey)
		if !ok {
			return Snapshot{}, newSourceError(fv.ConfigKey, SourceFlagBinding, "", "", "", Kind(-1), "", false, ErrUnknownSourceKey, nil)
		}
		if _, dup := seen[def.key]; dup {
			return Snapshot{}, newSourceError(def.key, SourceFlagBinding, "", "", "", def.kind, "", def.sensitive, ErrDuplicateBinding, nil)
		}
		seen[def.key] = struct{}{}
		if !fv.ExplicitlySet {
			continue
		}
		if !valueMatchesKind(def.kind, fv.Value) {
			raw := fmt.Sprintf("%v", fv.Value)
			return Snapshot{}, newSourceError(def.key, SourceFlagBinding, "", "", "", def.kind, raw, def.sensitive, ErrSourceConversion, nil)
		}
		snapshot.values[index] = newSourceValue(def, clonePublicValue(fv.Value), Source{label: SourceFlagBinding})
	}
	return snapshot, nil
}

package config

// Resolve returns a snapshot with each registered key resolved by the V1 precedence:
// explicit setter > flag binding > env > JSON > default.
//
// Passing a zero-value Snapshot{} for any tier is safe and equivalent to that tier
// having no values for any key. The returned snapshot is self-contained and holds no
// references into the caller-supplied source snapshots.
func Resolve(set Set, explicit, flag, env, jsonSrc Snapshot) Snapshot {
	result := set.DefaultSnapshot()
	// Layer sources from lowest to highest precedence. A higher-tier source value
	// overwrites the result for that key when the source has a value.
	for _, src := range []Snapshot{jsonSrc, env, flag, explicit} {
		for i, v := range src.values {
			if v.hasValue && i < len(result.values) {
				result.values[i] = v.clone()
			}
		}
	}
	return result
}

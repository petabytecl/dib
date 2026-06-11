package flags

// NameNormalizer maps long flag names to lookup keys for a Set.
//
// A nil normalizer keeps exact-name behavior.
//
// Normalizers must be deterministic and side-effect free. Set construction uses
// the normalizer to validate registered definition names and detect collisions;
// Lookup validates the raw long-name spelling before normalization and never
// resolves registered shorthand spellings as long-name aliases.
type NameNormalizer func(string) string

func normalizeName(normalizer NameNormalizer, name string) string {
	if normalizer == nil {
		return name
	}
	return normalizer(name)
}

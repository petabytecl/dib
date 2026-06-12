package config

// NameNormalizer maps config key names to lookup keys for a Set.
//
// A nil normalizer keeps exact-key behavior.
//
// Normalizers must be deterministic and side-effect free. Set construction uses
// the normalizer to validate registered definition keys and detect collisions.
type NameNormalizer func(string) string

func normalizeName(normalizer NameNormalizer, name string) string {
	if normalizer == nil {
		return name
	}
	return normalizer(name)
}

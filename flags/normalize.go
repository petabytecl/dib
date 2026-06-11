package flags

// NameNormalizer maps long flag names to lookup keys for a Set.
//
// A nil normalizer keeps exact-name behavior.
type NameNormalizer func(string) string

func normalizeName(normalizer NameNormalizer, name string) string {
	if normalizer == nil {
		return name
	}
	return normalizer(name)
}

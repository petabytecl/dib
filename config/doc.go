// Package config defines explicit configuration keys and default-resolution
// contracts.
//
// Dib configuration values are intended to be built from explicit instances and
// caller-owned inputs. Definitions are immutable after construction and expose
// stable key names, type expectations, optional defaults, usage text, and
// sensitivity metadata through accessors.
//
// Sets exact-match config keys by default. Callers can opt into key
// normalization with a deterministic NameNormalizer; setup rejects duplicate
// exact keys and duplicate normalized keys before any resolution occurs.
//
// DefaultSnapshot returns a self-contained snapshot of registered defaults.
// Snapshot lookup returns (Value, false) for unregistered keys, and a registered
// key with no default returns a Value whose Value method reports no value. Values
// supplied by defaults use the SourceDefault provenance label.
//
// Config resolution in this package does not read process arguments,
// environment variables, stdout, stderr, hidden caches, files, or package-level
// default configuration.
//
// Programmatic failures should use typed or sentinel errors so callers can
// inspect behavior without matching diagnostic strings.
package config

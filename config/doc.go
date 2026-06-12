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
// NewExplicitSnapshot ingests ordered caller-supplied assignments with
// SourceExplicit provenance. NewEnvSnapshot ingests only registered bindings
// through a caller-supplied EnvLookup function and uses SourceEnv provenance.
// LoadJSON and LoadJSONFile ingest JSON objects from explicit readers or paths,
// default to strict registered-key mode, and use SourceJSON provenance.
//
// NewFlagSnapshot ingests caller-supplied FlagValue bindings with SourceFlagBinding
// provenance. Each FlagValue carries a config key, an explicit-set indicator, and
// the parsed value. Only explicitly-set flag values enter the flag binding tier; a
// flag whose default was used does not inject any value (ExplicitlySet false leaves
// the tier absent for that key). Duplicate ConfigKey entries are setup errors.
//
// Resolve applies the V1 precedence order across independent source snapshots:
// explicit setter > flag binding > env > JSON > default. Passing a zero-value
// Snapshot{} for any tier is safe and treated as that tier having no values.
// The returned snapshot is self-contained and holds no references into its sources.
//
// Snapshot.IsSet reports whether a key has a resolved value in the snapshot,
// including defaults. It returns false for unregistered keys and registered
// keys with no value from any source. Explicit zero values, empty strings, and
// nil-typed defaults all count as set.
//
// Typed getters (GetString, GetBool, GetInt, GetInt64, GetUint, GetUint64,
// GetFloat64, GetDuration, GetStringList) retrieve values with their native
// Go types from a Snapshot. Each getter returns the value and a nil error on
// success, or a zero value and a *GetError on failure. GetStringList returns a
// defensive copy of the underlying slice.
//
// Three new sentinel errors cover typed retrieval failures:
//   - ErrKeyNotFound: the requested key is not registered in the set.
//   - ErrKeyAbsent: the key is registered but has no value from any source and
//     no default.
//   - ErrGetConversion: the key resolved but the caller requested a different
//     kind than what was registered.
//
// *GetError carries structured context about a retrieval failure. Callers can
// inspect it with errors.Is (matches the category sentinel) and errors.As.
// Accessors expose Key, Kind (actual registered kind), WantKind (caller's
// requested kind), SourceLabel (provenance when kind mismatch occurred),
// Redacted (whether the key is sensitive), and Category.
//
// Config APIs in this package do not read process arguments, live environment
// variables, stdout, stderr, hidden caches, ambient config files, or
// package-level default configuration. File reads occur only through explicit
// caller-supplied JSON paths.
//
// Programmatic failures should use typed or sentinel errors so callers can
// inspect behavior without matching diagnostic strings.
package config

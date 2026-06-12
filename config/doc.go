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
// Config source snapshots do not resolve cross-source precedence. That remains
// a separate resolution step for callers or later package APIs.
//
// Config APIs in this package do not read process arguments, live environment
// variables, stdout, stderr, hidden caches, ambient config files, or
// package-level default configuration. File reads occur only through explicit
// caller-supplied JSON paths.
//
// Programmatic failures should use typed or sentinel errors so callers can
// inspect behavior without matching diagnostic strings.
package config

// Package config defines explicit configuration keys, sources, and resolution
// contracts.
//
// Dib configuration values are intended to be built from explicit instances and
// caller-owned inputs. Definitions should remain immutable after construction,
// and resolution should return per-run snapshots rather than reading process
// arguments, environment variables, stdout, stderr, hidden caches, or
// package-level default configuration.
//
// Programmatic failures should use typed or sentinel errors so callers can
// inspect behavior without matching diagnostic strings.
package config

// Package command defines explicit command definitions for reusable CLI trees.
//
// Dib command values are intended to be built from explicit instances and
// caller-owned inputs. Definitions remain immutable after construction, and
// routing returns per-run snapshots rather than reading process arguments,
// environment variables, stdout, stderr, hidden caches, or package-level
// default commands.
//
// Programmatic failures should use typed or sentinel errors so callers can
// inspect behavior without matching diagnostic strings.
package command

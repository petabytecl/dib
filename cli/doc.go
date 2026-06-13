// Package cli provides optional composition support for caller-owned CLI inputs.
//
// The package defines small boundary values used to carry an explicit program
// name and user arguments into higher-level command, flag, and config
// composition. It is not a root facade or process-owning framework: callers
// choose when to read process state and pass those inputs explicitly.
//
// cli.Resolve is the primary composition entry point: it routes an Invocation
// through a Plan, builds the flag-tier config snapshot, resolves config by
// precedence, and returns a Result without invoking callbacks or owning process
// lifecycle.
//
// CLI values in this package do not read os.Args, call os.Exit, write streams,
// read environment variables, load files, execute callbacks, or use package-level
// default invocations.
package cli

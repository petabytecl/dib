// Package cli provides optional composition support for caller-owned CLI inputs.
//
// The package defines small boundary values used to carry an explicit program
// name and user arguments into higher-level command, flag, and config
// composition. It is not a root facade or process-owning framework: callers
// choose when to read process state and pass those inputs explicitly.
//
// cli.New is the high-level authoring entry point. It returns a root Command
// builder so packages can register command subtrees in separate files, then Run
// resolves the invocation and dispatches the matched Handler with one
// CommandContext argument. Resolve remains available as the low-level
// inspectable composition entry point.
//
// CLI values in this package do not read os.Args, call os.Exit, write streams,
// read environment variables, load files, or use package-level default
// invocations. High-level Run invokes only caller-registered handlers and
// returns errors for caller-owned logging and exit policy.
package cli

# Dib

Dib is a standard-library-only Go library for CLI command routing, flag parsing, and config resolution.

## Status

Dib v0 is an experimental API. Future v0 tags may change exported interfaces before a stable v1 contract is established. Go 1.26 or newer is required. The root module records `go 1.26`; no external dependencies are added to the module graph.

## Packages

Four public package surfaces compose the library:

| Package | Import path | Role |
| --- | --- | --- |
| `flags` | `github.com/petabytecl/dib/flags` | Explicit flag sets, long/short flags, shorthand groups, repeated values, and typed parse diagnostics. |
| `command` | `github.com/petabytecl/dib/command` | Command routing, nested trees, aliases, local/inherited flags, deterministic help/usage, and typed routing errors. |
| `config` | `github.com/petabytecl/dib/config` | Registered keys, explicit setters, flag/env/JSON bindings, precedence, typed getters, provenance, and redaction. |
| `cli` | `github.com/petabytecl/dib/cli` | Optional composition: carries `os.Args` as an explicit `Invocation`, routes commands, resolves config, and returns a `Result` without owning process lifecycle. |

The first three packages work independently: `flags` works without `command` or `config`; `command` does not depend on `config`; callers compose these surfaces explicitly. `cli` is an optional fourth surface that composes all three in a single call.

## Install

```
go get github.com/petabytecl/dib
```

Import the surfaces you need:

```go
import (
    "github.com/petabytecl/dib/cli"
    "github.com/petabytecl/dib/command"
    "github.com/petabytecl/dib/config"
    "github.com/petabytecl/dib/flags"
)
```

## Quickstart

### Flag parsing

```go
set, err := flags.NewSet(
    flags.String("host", "localhost", "server hostname"),
    flags.Int("port", 8080, "server port"),
    flags.Bool("verbose", false, "enable verbose output"),
)
if err != nil {
    log.Fatal(err)
}
snapshot, err := set.Parse(os.Args[1:])
if err != nil {
    log.Fatal(err)
}
if state, ok := snapshot.Lookup("host"); ok {
    fmt.Println(state.Values())
}
```

### Command routing

```go
serve, _ := command.NewDefinition("serve",
    command.Description("start the server"),
)
root, err := command.NewDefinition("app",
    command.Description("my application"),
    command.Children(serve),
)
if err != nil {
    log.Fatal(err)
}
result, err := root.Route(os.Args[1:])
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.PathNames())
```

### Config resolution

```go
set, err := config.NewSet(
    config.String("host", "localhost", "server hostname"),
)
if err != nil {
    log.Fatal(err)
}
envSnapshot, err := config.NewEnvSnapshot(set, os.Getenv, []config.EnvBinding{
    config.BindEnv("host", "APP_HOST"),
})
if err != nil {
    log.Fatal(err)
}
resolved := config.Resolve(set, config.Snapshot{}, config.Snapshot{}, envSnapshot, config.Snapshot{})
host, err := resolved.GetString("host")
if err != nil {
    log.Fatal(err)
}
fmt.Println(host)
```

### Using command, flags, and config together

`cli.Resolve` is the optional composition entry point. It accepts an explicit
`Invocation` (built from `os.Args`, `os.Args[0]`, and `os.Args[1:]` or any
caller-supplied slice) and a `Plan`, then routes commands, parses flags, and
resolves config in one call without owning process lifecycle.

```go
// Invocation boundary: os.Args[0] is the program; os.Args[1:] are the user args.
inv, err := cli.FromOSArgs(os.Args)
if err != nil {
    log.Fatal(err)
}

// Define a command tree with a --host flag on the serve sub-command.
serve, _ := command.NewDefinition("serve",
    command.Description("start the server"),
    command.LocalFlags(flags.String("host", "localhost", "server hostname")),
)
root, err := command.NewDefinition("app",
    command.Description("my application"),
    command.Children(serve),
)
if err != nil {
    log.Fatal(err)
}

// Register config keys.
set, err := config.NewSet(
    config.String("host", "localhost", "server hostname"),
)
if err != nil {
    log.Fatal(err)
}

// Bind the --host flag to the host config key and resolve everything.
plan := cli.NewPlan(root, set).
    WithBindings([]cli.FlagBinding{cli.BindFlag("host", "host")})

result, err := cli.Resolve(inv, plan)
if err != nil {
    log.Fatal(err)
}

host, _ := result.Config().GetString("host")
fmt.Println(result.Route().PathNames()) // e.g. [app serve]
fmt.Println(host)                       // e.g. example.com (from --host flag)
```

## Compatibility

Dib is a clean-room native Go API. It is not a source-compatible clone, not a drop-in replacement, and not a framework compatibility layer for Go `flag`, pflag, Cobra, Viper, or comparable projects. Familiar CLI concepts from those libraries are documented as supported, narrowed, omitted, or intentionally different.

See `docs/compatibility.md` for the full compatibility boundary table.

## Documentation

| Document | Contents |
| --- | --- |
| `docs/config-precedence.md` | Canonical config precedence order (`explicit setter > flag binding > env > JSON > default`) |
| `docs/diagnostics-and-errors.md` | Error taxonomy and diagnostic vocabulary |
| `docs/compatibility.md` | Compatibility boundaries vs Go `flag`, pflag, Cobra, Viper |
| `docs/behavior-matrices.md` | Consolidated adoption evidence |
| `docs/testing.md` | Local verification, lint, coverage, release gates |
| `docs/release-checklist.md` | Release evidence |
| `examples/migration/` | Executable migration examples |
| `examples/multicommand/` | CLI composition example |
| `CONTRIBUTING.md` | Contribution guidelines and clean-room policy |

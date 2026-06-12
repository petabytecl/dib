# Dib

Dib is a standard-library-only Go library for CLI command routing, flag parsing, and config resolution.

## Status

Dib v0 is an experimental API. Future v0 tags may change exported interfaces before a stable v1 contract is established. Go 1.26 or newer is required. The root module records `go 1.26`; no external dependencies are added to the module graph.

## Packages

Three independent public package surfaces compose the library:

| Package | Import path | Role |
| --- | --- | --- |
| `flags` | `github.com/petabytecl/dib/flags` | Explicit flag sets, long/short flags, shorthand groups, repeated values, and typed parse diagnostics. |
| `command` | `github.com/petabytecl/dib/command` | Command routing, nested trees, aliases, local/inherited flags, deterministic help/usage, and typed routing errors. |
| `config` | `github.com/petabytecl/dib/config` | Registered keys, explicit setters, flag/env/JSON bindings, precedence, typed getters, provenance, and redaction. |

Each package works independently: `flags` works without `command` or `config`; `command` does not depend on `config`; callers compose the three surfaces explicitly.

## Install

```
go get github.com/petabytecl/dib
```

Import the surfaces you need:

```go
import (
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
| `CONTRIBUTING.md` | Contribution guidelines and clean-room policy |

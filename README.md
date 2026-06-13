# Dib

Dib is a standard-library-only Go library for CLI command routing, flag parsing, and config resolution.

## Status

Dib v0 is an experimental API. Future v0 tags may change exported interfaces before a stable v1 contract is established. Go 1.26 or newer is required. The root module records `go 1.26`; no external dependencies are added to the module graph.

## Packages

Four public package surfaces compose the library:

| Package   | Import path                         | Role                                                                                                                                                    |
|-----------|-------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| `flags`   | `github.com/petabytecl/dib/flags`   | Explicit flag sets, long/short flags, shorthand groups, repeated values, and typed parse diagnostics.                                                   |
| `command` | `github.com/petabytecl/dib/command` | Command routing, nested trees, aliases, local/inherited flags, deterministic help/usage, and typed routing errors.                                      |
| `config`  | `github.com/petabytecl/dib/config`  | Registered keys, explicit setters, flag/env/JSON bindings, precedence, typed getters, provenance, and redaction.                                        |
| `cli`     | `github.com/petabytecl/dib/cli`     | Optional composition: builds distributed command trees, resolves route/flag/config state, dispatches handlers, and still leaves process lifecycle to callers. |

The first three packages work independently: `flags` works without `command` or `config`; `command` does not depend on `config`; callers compose these surfaces explicitly. `cli` is an optional fourth surface that provides a high-level command builder plus the lower-level `Resolve` API.

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

### Building a distributed command tree

`cli.New` returns a root command builder. Feature packages can receive that root
or any subcommand and attach their own command subtree in file order. `Run`
builds the immutable command/config plan, resolves the invocation, and invokes
the matched handler with one `cli.CommandContext` argument.

```go
func main() {
    root := cli.New("svcctl",
        cli.Description("service control"),
        cli.Config(config.String("target", "scrapd", "service target")),
    )

    registerServiceCommands(root)

    if _, err := root.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}

func registerServiceCommands(root *cli.Command) {
    service := root.Command("service",
        cli.Description("service commands"),
    )

    service.Command("start",
        cli.Flags(flags.String("target", "scrapd", "service target")),
        cli.Bindings(cli.BindFlag("target", "target")),
        cli.Handle(startService),
    )

    service.Command("stop",
        cli.Flags(flags.String("target", "scrapd", "service target")),
        cli.Bindings(cli.BindFlag("target", "target")),
        cli.Handle(stopService),
    )
}

func startService(cmd cli.CommandContext) error {
    target, err := cmd.Config().GetString("target")
    if err != nil {
        return err
    }
    return start(cmd.Context(), target)
}

func stopService(cmd cli.CommandContext) error {
    target, err := cmd.Config().GetString("target")
    if err != nil {
        return err
    }
    return stop(cmd.Context(), target)
}
```

`Run` does not call `os.Exit`, write output, read environment variables, or load
files. Handlers own their application behavior; the caller still owns logging,
IO, context cancellation, and exit-code policy. For tests or adapters that
already have stripped user args, use `RunArgs`.

### Using Resolve directly

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

The builder API is the recommended path for application CLIs. `Resolve` remains
available when a caller wants to inspect routing/config state and dispatch
manually. See `examples/multicommand/` for executable examples covering both
`Example_dispatchStartStop` and `Example_lowLevelDispatch`.

## Compatibility

Dib is a clean-room native Go API. It is not a source-compatible clone, not a drop-in replacement, and not a framework compatibility layer for Go `flag`, pflag, Cobra, Viper, or comparable projects. Familiar CLI concepts from those libraries are documented as supported, narrowed, omitted, or intentionally different.

See `docs/compatibility.md` for the full compatibility boundary table.

## Documentation

| Document                         | Contents                                                                                    |
|----------------------------------|---------------------------------------------------------------------------------------------|
| `docs/config-precedence.md`      | Canonical config precedence order (`explicit setter > flag binding > env > JSON > default`) |
| `docs/diagnostics-and-errors.md` | Error taxonomy and diagnostic vocabulary                                                    |
| `docs/compatibility.md`          | Compatibility boundaries vs Go `flag`, pflag, Cobra, Viper                                  |
| `docs/behavior-matrices.md`      | Consolidated adoption evidence                                                              |
| `docs/testing.md`                | Local verification, lint, coverage, release gates                                           |
| `docs/release-checklist.md`      | Release evidence                                                                            |
| `examples/migration/`            | Executable migration examples                                                               |
| `examples/multicommand/`         | CLI composition example                                                                     |
| `CONTRIBUTING.md`                | Contribution guidelines and clean-room policy                                               |
| `LICENSE`                        | MIT License                                                                                 |

## License

Dib is licensed under the [MIT License](LICENSE).

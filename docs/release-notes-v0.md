# Release Notes V0

## V0 Experimental API Status

Dib v0 is an experimental Go module API. Future v0 tags may change exported APIs
before a stable v1 contract, but that status does not relax correctness,
redaction, clean-room, standard-library-only dependency, or release-gate
expectations.

## Supported Go Version

Dib v0 requires Go 1.26 or newer. The root module records `go 1.26`, and CI
loads the Go version from `go.mod`.

## Release Gates

Release review for a v0 module tag records these local gates in
`docs/release-checklist.md`:

- `go test ./...`
- `go run ./tools/lint`
- `go vet ./...`
- `go run ./tools/coverage`
- `go run ./tools/depgate`
- `go test -race ./...`

Parser fuzz targets are release-candidate hardening evidence when parser
behavior changes or when release-candidate fuzz evidence is requested.

Epic 6: Release Hardening and Public Usage Onboarding added three formal release gates: the isolated lint gate (`go run ./tools/lint`), package-aware coverage validation for public runtime packages (`go run ./tools/coverage`), and public usage documentation (`README.md`).

## Compatibility And Migration

Dib is a clean-room native Go API. It is not a source-compatible clone, not a drop-in replacement, and not a framework compatibility layer for Go `flag`, pflag, Cobra, Viper, or comparable projects.

Compatibility boundaries are documented in `docs/compatibility.md`. Migration examples live under `examples/migration/` and demonstrate Dib-native APIs with
explicit instances, injected inputs, typed errors, deterministic rendering, and
redaction-safe reports. `README.md` provides public onboarding with install guidance, package roles, and a minimal flag/command/config quickstart.

## Release Scope

Dib v0 releases are Go module tags. This guidance does not define standalone
artifact publishing, container images, cluster manifests, completion-script
generation, or manpage generation.

## License

Dib v0 is published under the MIT License.

## Epic 7: CLI Composition Ergonomics

Epic 7 added the `cli` package as an optional fourth public package surface,
completing the CLI composition story for Dib v0:

- Added `cli.Invocation` to carry an explicit caller-supplied program name and
  user arguments as an immutable snapshot. `cli.FromOSArgs` accepts a full argv
  slice (where `os.Args[0]` is the program and `os.Args[1:]` are the user args)
  and `cli.FromArgs` accepts separate program and args values.
- Added `cli.Plan` to carry the root command definition, config set, optional
  source snapshots, and flag bindings as an immutable composition plan.
- Added `cli.Resolve` to route an invocation through a plan: routes commands via
  `command.Definition.Route`, builds the flag-tier config snapshot from
  `cli.FlagBinding` values, resolves config by precedence, and returns a
  `cli.Result` — without invoking callbacks, calling `os.Exit`, writing streams,
  reading env implicitly, or loading files.
- Added `cli.Result` to expose the invocation, command route, flag snapshot, and
  fully resolved config snapshot as a single immutable value.
- The `cli` package may be used optionally; `command`, `flags`, and `config`
  remain independently usable without it.
- Coverage gate extended to include `cli` at the same 85% threshold as the
  existing three public runtime packages.
- `examples/multicommand/` added as executable composition evidence; the
  `Example_composedCLI` function in `examples/multicommand/example_test.go`
  demonstrates the full composition path with caller-supplied inputs, and
  `Example_dispatchStartStop` demonstrates application-owned `start`/`stop`
  handler dispatch after `cli.Resolve`; `Example_lowLevelDispatch` preserves
  the manual dispatch pattern as low-level evidence.
- Post-Epic 7 issue #52 adds the high-level `cli.Command` builder,
  distributed subcommand registration, `cli.CommandContext`, and `Run` handler
  dispatch while keeping `Resolve` as the low-level inspectable path.
- GitHub tracking: Epic 7 issue #46, Story 7.4 issue #50, handler dispatch
  follow-up issue #52.

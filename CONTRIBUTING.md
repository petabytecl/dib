# Contributing

Before changing Dib, read the [clean-room policy](docs/clean-room-policy.md)
and update the [provenance log](docs/provenance-log.md) when your work uses
copied or adapted material allowed by that policy, generated or
reference-derived material, or a source whose influence should be audited.

## Clean-Room Contribution Checklist

- Preserve Dib's zero-runtime-dependency contract unless the architecture is
  updated first.
- Write compatibility examples, tests, docs, fixtures, behavior matrices, and
  fuzz seeds independently. Provenance records external influence; it does not
  permit copied source, tests, comments, examples, fixtures, internal names,
  file layout, README structure, or source-derived generated content from
  inspiration projects.
- Keep compatibility claims behavior-scoped and test-backed where practical.
- Do not describe Dib as source-compatible with Go `flag`, pflag, Cobra, Viper,
  or comparable projects.
- Do not add external runtime, test, or documentation-example dependencies
  unless the architecture is updated first.
- Do not copy third-party source, tests, comments, examples, fixtures, file
  layout, internal names, README structure, or source-derived generated content.
- Treat missing provenance as an acceptance blocker for copied, generated,
  adapted, or reference-derived artifacts.

## Local Verification

Run these baseline checks before handing work off for review:

```sh
golangci-lint run
go test ./...
go vet ./...
go run ./tools/depgate
```

The lint gate uses `golangci-lint` (pinned to `v2.10.1` in CI) with the strict
ruleset in `.golangci.yml`. See [`docs/testing.md`](docs/testing.md) for the
local command and version pin.

The dependency gate is the repository authority for Dib's
standard-library-only contract. Do not treat ad hoc `go list` commands as
release-candidate dependency evidence once `tools/depgate/` exists.

## Documentation Expectations

Documentation should be short, auditable, and independently written. If public
documentation or observed behavior informs an explanation, keep the wording in
Dib's project language and record provenance when the result is more than
inspiration-only.

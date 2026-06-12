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
- `go vet ./...`
- `go run ./tools/depgate`
- `go test -race ./...`

Parser fuzz targets are release-candidate hardening evidence when parser
behavior changes or when release-candidate fuzz evidence is requested.

## Compatibility And Migration

Dib is a clean-room native Go API. It is not a source-compatible clone, not a drop-in replacement, and not a framework compatibility layer for Go `flag`, pflag, Cobra, Viper, or comparable projects.

Compatibility boundaries are documented in `docs/compatibility.md`. Migration examples live under `examples/migration/` and demonstrate Dib-native APIs with
explicit instances, injected inputs, typed errors, deterministic rendering, and
redaction-safe reports.

## Release Scope

Dib v0 releases are Go module tags. This guidance does not define standalone
artifact publishing, container images, cluster manifests, completion-script
generation, or manpage generation.

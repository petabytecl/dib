# Release Checklist

Use this checklist for each Dib Go module tag. It records release-candidate evidence; it is not a claim that a release is ready until every required item is filled in and reviewed.

## Release Identity

- Go module tag:
- Exact commit:
- Owner:
- Date:
- Reviewer:

## Go Version Alignment

Release is blocked if `go.mod`, `.github/workflows/ci.yml`, release guidance, or user-facing docs disagree about the supported Go version.

- `go.mod` version:
- CI `actions/setup-go` source: `go-version-file: go.mod`
- Release guidance version:
- Documentation version references:
- Drift review result:

## CI Trust Gates

CI failures block tagging. Record the exact command outcome for each required gate.

- `go test ./...`:
- `go vet ./...`:
- `go run ./tools/depgate`:
- Workflow file: `.github/workflows/ci.yml`
- Runner image:
- `actions/checkout` version:
- `actions/setup-go` version:

## Release-Candidate Gates

- `go test -race ./...`:
- Parser fuzz evidence, when parser behavior changed:
  - `go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`:
  - `go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`:
  - `go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`:
- Docs/examples evidence:
- Provenance evidence:
- Compatibility evidence:
- Migration evidence: Story 5.2 example pointers live in `examples/migration/`; Story 5.4 records final release-candidate command outcomes.

## Standard-Library Dependency Evidence

- Root `go.mod` contains no `require`, `replace`, or `toolchain` directives:
- Root `go.sum` absent:
- Dependency gate reviewed:
- Any fixture-local dependency exceptions:

## Waivers

Waivers require owner, reason, expiry, and follow-up tracking. Open-ended waivers block release.

| Item | Owner | Reason | Expiry | Follow-up |
| --- | --- | --- | --- | --- |
| None |  |  |  |  |

## Final Review

- All required evidence captured:
- All waivers approved with expiry:
- Tagging decision:

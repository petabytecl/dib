# Provenance Log

This log records external sources that influence Dib artifacts. Use it before
review when material is copied, adapted, generated from a reference, or
reference-derived. Inspiration-only entries are allowed when recording the
boundary improves auditability.

## Classifications

- `copied`: Material is reproduced verbatim or nearly verbatim. This requires
  explicit source, license or terms, affected artifact, and reviewer approval.
- `adapted`: Material is changed but still materially derived from a source.
  This requires explicit source, license or terms, affected artifact, and
  reviewer approval.
- `inspiration-only`: The source informed behavior understanding or planning,
  but no source, tests, examples, fixtures, internal names, file organization,
  or source-derived structure were copied or adapted.

Generated or reference-derived material uses these same classifications:
`copied` when it reproduces source material, `adapted` when it is materially
derived from a source, and `inspiration-only` when the output is independently
written after behavior-level study.

Classification is an audit label, not approval to use disallowed material. The
clean-room policy still excludes copied or closely derived source, tests,
comments, examples, fixtures, internal names, file organization, README
structure, and source-derived generated content from inspiration projects.
Entries classified as `copied` or `adapted` require reviewer approval only for
material that the clean-room policy and the source terms already permit.

## Entry Template

- Source:
- URL:
- Access date:
- License or terms:
- Affected artifact:
- Classification: copied | adapted | inspiration-only
- Reviewer:
- Approval date:
- Approval notes:
- Notes:

For `copied` or `adapted` entries, prefer immutable URLs such as a commit,
release tag, archive snapshot, or content-addressed artifact. Record the access
date for the source actually reviewed. If an entry is added after the affected
artifact was written, say so in the notes and identify the date or phase the
source influenced when known.

## Story 2.8 Entries

### Go Fuzzing Documentation

- Source: Go fuzzing documentation
- URL: https://go.dev/doc/security/fuzz/
- Access date: 2026-06-11
- License or terms: Go project BSD-style license, https://go.dev/LICENSE
- Affected artifact: `flags/fuzz_test.go` (`FuzzParse` target and seed corpus), `docs/behavior-matrices.md`
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used to understand Go fuzz corpus format (`go test fuzz v1`), `testing.F` API, seed corpus file placement under `testdata/fuzz/FuzzName/`, and the distinction between corpus-based unit runs and `-fuzz=` mutation mode. No source, tests, comments, examples, fixtures, internal names, file organization, or source-derived structure copied from the Go fuzzing documentation into Dib runtime source, tests, seeds, or docs.

### Go testing Package Documentation

- Source: Go `testing` package documentation (`testing.F` section)
- URL: https://pkg.go.dev/testing#hdr-Fuzzing
- Access date: 2026-06-11
- License or terms: Go project BSD-style license, https://go.dev/LICENSE
- Affected artifact: `flags/fuzz_test.go` (`FuzzParse` target)
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used to understand `testing.F.Add` argument type constraints and `testing.F.Fuzz` callback signature. No source, tests, comments, examples, fixtures, internal names, or source-derived structure copied.

## Story 5.1 Entries

### Go flag Package Documentation

- Source: Go `flag` package documentation
- URL: https://pkg.go.dev/flag
- Access date: 2026-06-12
- License or terms: Go project BSD-style license, https://go.dev/LICENSE
- Affected artifact: `docs/compatibility.md`
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used for behavior-boundary context while writing adopter-facing
  compatibility prose. No source, tests, comments, examples, fixtures, internal
  names, file organization, wording, tables, or source-derived structure copied
  into Dib artifacts.

### pflag Package Documentation

- Source: pflag package documentation
- URL: https://pkg.go.dev/github.com/spf13/pflag
- Access date: 2026-06-12
- License or terms: BSD-style license,
  https://github.com/spf13/pflag/blob/master/LICENSE
- Affected artifact: `docs/compatibility.md`
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used for behavior-boundary context while writing adopter-facing
  compatibility prose. No source, tests, comments, examples, fixtures, internal
  names, file organization, wording, tables, or source-derived structure copied
  into Dib artifacts.

### Cobra Flags Documentation

- Source: Cobra flags documentation
- URL: https://cobra.dev/docs/how-to-guides/working-with-flags/
- Access date: 2026-06-12
- License or terms: Apache-2.0 license,
  https://github.com/spf13/cobra/blob/main/LICENSE.txt
- Affected artifact: `docs/compatibility.md`
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used for behavior-boundary context while writing adopter-facing
  compatibility prose. No source, tests, comments, examples, fixtures, internal
  names, file organization, wording, tables, or source-derived structure copied
  into Dib artifacts.

### Viper Package Documentation

- Source: Viper package documentation
- URL: https://pkg.go.dev/github.com/spf13/viper
- Access date: 2026-06-12
- License or terms: MIT license,
  https://github.com/spf13/viper/blob/master/LICENSE
- Affected artifact: `docs/compatibility.md`
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used for behavior-boundary context while writing adopter-facing
  compatibility prose. No source, tests, comments, examples, fixtures, internal
  names, file organization, wording, tables, or source-derived structure copied
  into Dib artifacts.

## Initial Entries

### Go flag Package Documentation

- Source: Go `flag` package documentation
- URL: https://pkg.go.dev/flag
- Access date: 2026-06-11
- License or terms: Go project BSD-style license, https://go.dev/LICENSE
- Affected artifact: PRD, architecture, Story 1.2 clean-room policy and
  provenance guidance
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used to understand public behavior concepts. No source, tests,
  comments, examples, fixtures, internal names, file organization, wording,
  tables, or source-derived structure copied into Dib runtime source, docs,
  story artifacts, fixtures, or examples.

### pflag Package Documentation

- Source: pflag package documentation
- URL: https://pkg.go.dev/github.com/spf13/pflag
- Access date: 2026-06-11
- License or terms: BSD-style license,
  https://github.com/spf13/pflag/blob/master/LICENSE
- Affected artifact: PRD, architecture, Story 1.2 clean-room policy and
  provenance guidance
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used to understand public behavior concepts. No source, tests,
  comments, examples, fixtures, internal names, file organization, wording,
  tables, or source-derived structure copied into Dib runtime source, docs,
  story artifacts, fixtures, or examples.

### Cobra Flags Documentation

- Source: Cobra flags documentation
- URL: https://cobra.dev/docs/how-to-guides/working-with-flags/
- Access date: 2026-06-11
- License or terms: Apache-2.0 license,
  https://github.com/spf13/cobra/blob/main/LICENSE.txt
- Affected artifact: PRD, architecture, Story 1.2 clean-room policy and
  provenance guidance
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used to understand public behavior concepts. No source, tests,
  comments, examples, fixtures, internal names, file organization, wording,
  tables, or source-derived structure copied into Dib runtime source, docs,
  story artifacts, fixtures, or examples.

### Viper Package Documentation

- Source: Viper package documentation
- URL: https://pkg.go.dev/github.com/spf13/viper
- Access date: 2026-06-11
- License or terms: MIT license,
  https://github.com/spf13/viper/blob/master/LICENSE
- Affected artifact: PRD, architecture, Story 1.2 clean-room policy and
  provenance guidance
- Classification: inspiration-only
- Reviewer: Not applicable
- Approval date: Not applicable
- Approval notes: Not applicable for inspiration-only entry
- Notes: Used to understand public behavior concepts. No source, tests,
  comments, examples, fixtures, internal names, file organization, wording,
  tables, or source-derived structure copied into Dib runtime source, docs,
  story artifacts, fixtures, or examples.

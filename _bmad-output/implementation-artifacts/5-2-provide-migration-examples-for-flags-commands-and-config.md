---
baseline_commit: e6eccf4
created: "2026-06-12"
---

# Story 5.2: Provide Migration Examples For Flags, Commands, And Config

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a developer migrating a small CLI,
I want executable examples for familiar flag, command, and config patterns,
so that I can adopt Dib's native API without copying framework-shaped code.

## Requirements Trace

- FR18: migration examples must align with the compatibility boundary table and avoid source-compatible positioning.
- FR19: developers can follow examples from standard `flag`, pflag-style flags, Cobra-style command trees, and Viper-style config resolution.
- FR21: examples, tests, and tooling must remain standard-library-only and pass the dependency gate.
- NFR1: runtime packages, tests, examples, and tooling remain standard-library-only unless architecture changes.
- NFR2: examples must use explicit instances and caller-supplied inputs, not package-global helpers.
- NFR3: examples must show typed errors where migration success depends on inspectable failures.
- NFR5: examples must not call `os.Exit`, mutate process-wide streams, or read ambient process args/env.
- NFR6: examples must be testable with table-driven tests and injected readers, writers, args, and env lookup.
- NFR7: examples must show familiar concepts without implying source compatibility with Go `flag`, pflag, Cobra, or Viper.
- NFR8: config examples must preserve redaction-safe diagnostics and source reports.

## Acceptance Criteria

1. Given a developer knows standard Go `flag`, when they read the migration examples, then an example shows explicit Flag sets, typed errors, and table-driven tests without package-global state, and the example builds through `go test ./...`.
2. Given a developer knows pflag-style shorthand behavior, when they read the migration examples, then an example shows long flags, shorthand flags, grouped shorthands, repeated values, no-option defaults, and `--` behavior, and intentional differences are referenced rather than hidden.
3. Given a developer knows Cobra-style command trees, when they read the migration examples, then an example shows nested command routing, aliases, local/inherited flags, help rendering, and caller-controlled errors, and it does not add a `/cmd` scaffold or process-owning framework shape.
4. Given a developer knows Viper-style config resolution, when they read the migration examples, then an example shows defaults, explicit setters, flag binding, env binding, JSON loading, precedence, typed retrieval, provenance, and redaction, and it uses injected inputs rather than ambient process env or host-specific files.
5. Given examples are executable trust artifacts, when verification runs, then examples compile with the standard library only, trace to relevant FRs, avoid copied source/examples, and pass `go test ./...` plus `go run ./tools/depgate`.

## Tasks / Subtasks

- [x] Confirm preconditions and read UPDATE files before editing (AC: 1-5)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 5 `in-progress` and Story 5.2 `ready-for-dev`.
  - [x] Confirm current `HEAD` is `e6eccf4` (`docs(story-5.1): publish compatibility boundaries`) or intentionally account for newer user changes.
  - [x] Read `docs/clean-room-policy.md`, `docs/provenance-log.md`, `docs/compatibility.md`, `docs/compatibility_test.go`, `docs/behavior-matrices.md`, `docs/config-precedence.md`, `docs/diagnostics-and-errors.md`, and `docs/release-checklist.md` completely before writing migration examples or evidence prose.
  - [x] Read current public API files for `flags/`, `command/`, and `config/` before writing examples: at minimum `doc.go`, constructors/options, snapshot/result APIs, typed error APIs, help/report rendering APIs, and existing workflow tests.
  - [x] Verify `examples/migration/` does not already exist. If it exists by the time implementation starts, preserve user changes and extend the existing package.
  - [x] Do not add runtime package APIs, source-compatible adapters, package-global helper APIs, `/cmd` scaffolding, generated completion/manpage assumptions, new config formats, or process lifecycle ownership.

- [x] Create executable migration examples under `examples/migration/` (AC: 1-5)
  - [x] NEW `examples/migration/standard_flag_concepts_test.go`: show standard-`flag` mental-model migration using `flags.NewSet`, explicit args, `Snapshot.Lookup`, `ValueState.Values`, `ValueState.Explicit`, `RemainingArgs`, and typed failure inspection with `errors.Is` / `errors.As` against `*flags.ParseError`.
  - [x] Include table-driven tests in the standard flag example file for success and failure cases. Cover no package-global state, no `flag.CommandLine`, no `os.Args`, no stdout/stderr scraping, and a failed parse returning a zero-value snapshot.
  - [x] NEW `examples/migration/shorthand_flag_migration_test.go`: show pflag-style concepts through Dib's native parser: long flags, one-rune `flags.Shorthand`, shorthand groups, `flags.Repeatable`, `flags.NoOptionDefault`, custom or built-in value parsing, interspersed positionals, and `--` passthrough.
  - [x] In the shorthand example, explicitly demonstrate the intentional differences that matter to adopters: no source-compatible pflag API, `--no-*` is ordinary unless defined, shorthand lookup is independent from long-name normalization, and `--help` / `-h` remain caller-controlled parse diagnostics when unregistered.
  - [x] NEW `examples/migration/nested_command_migration_test.go`: show Cobra-style command tree migration through `command.NewDefinition`, `command.Children`, `command.Aliases`, `command.InheritedFlags`, `command.LocalFlags`, `Definition.Route`, `Result.PathNames`, `Result.MatchTokens`, `Result.RemainingArgs`, `Result.FlagSnapshot`, `Result.WriteHelp`, and `Definition.RouteBoundary`.
  - [x] In the command example, use caller-owned `context.Context`, `bytes.Buffer` writers, and explicit arg slices. Show that routing/help returns values/errors and does not execute callbacks, decide exit policy, write to process-global streams, or require a `/cmd` scaffold.
  - [x] NEW `examples/migration/config_precedence_migration_test.go`: show Viper-style config migration with `config.NewSet` or `NewNormalizedSet`, defaults, `NewExplicitSnapshot`, `NewFlagSnapshot`, `NewEnvSnapshot` with injected `EnvLookup`, `LoadJSON` with `strings.NewReader` and `JSONReaderLabel`, `Resolve`, typed getters, `IsSet`, `SourceReport`, and `WriteSourceReport`.
  - [x] In the config example, demonstrate the canonical precedence order exactly: `explicit setter > flag binding > env > JSON > default`. Show that flag defaults do not enter the flag binding tier unless `ExplicitlySet` is true.
  - [x] In the config example, mark at least one fake sensitive key with `config.Sensitive()` and prove `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value` do not appear in source reports, rendered diagnostics, or example output.
  - [x] Use `Example...` functions where practical so examples appear as runnable documentation, and use `Test...` functions for structured assertions that should not depend on exact rendered strings. Keep every example and test standard-library-only.

- [x] Add evidence and freshness checks for the new examples (AC: 2, 5)
  - [x] Update `docs/compatibility.md` so it no longer says migration examples are deferred once they exist. Link to `examples/migration/` or the specific example files without claiming final release readiness.
  - [x] Update `docs/compatibility_test.go` to require the migration example evidence and remove or revise any check that expects Story 5.2 to remain deferred.
  - [x] Update `docs/behavior-matrices.md` with a Story 5.2 migration examples row marked `current`, citing exact example/test files and keeping Story 5.3 consolidated adoption evidence and Story 5.4 release readiness deferred.
  - [x] Update `docs/release-checklist.md` only if a narrowly scoped placeholder/link improves evidence traceability. Do not fill final release-candidate command outcomes; Story 5.4 owns completed release evidence.
  - [x] Update `docs/provenance-log.md` if any current public Go `flag`, pflag, Cobra, Viper, Go examples, or testing documentation is used while designing example content. Classify such sources as `inspiration-only` unless material is copied or adapted, which should be avoided.

- [x] Preserve clean-room boundaries (AC: 1-5)
  - [x] Write Dib-owned examples from the current Dib APIs and local tests. Do not copy external examples, fixtures, README layout, test cases, source names, internal names, or file organization from inspiration projects.
  - [x] Examples may mention Go `flag`, pflag-style, Cobra-style, and Viper-style as source mental models, but must not implement compatibility adapters or preserve external framework-shaped APIs.
  - [x] Keep example scenarios small, obviously fake, deterministic, and domain-neutral. Avoid real credentials and use only the architecture-owned fake sensitive corpus when secret-looking values are needed.
  - [x] If external reference docs influence wording or behavior examples, add provenance entries before acceptance and keep them inspiration-only.

- [x] Verify the story implementation (AC: 1-5)
  - [x] `GOCACHE=/tmp/dib-go-build go test ./examples/migration`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./docs`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
  - [x] `git diff --check`
  - [x] `rg -n "(?i)drop-in|source-compatible|clone API|framework compatibility layer|compatible replacement" docs examples/migration`
  - [x] Confirm the search above finds only explicit boundary language, not positive compatibility claims.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` exists.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 5.2 is the primary spec. [Source: `_bmad-output/planning-artifacts/epics.md#Story-5.2-Provide-Migration-Examples-For-Flags-Commands-And-Config`]
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD shards from `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/`.
- No UX artifact discovered; Dib V1 has no browser UI or frontend surface. [Source: `_bmad-output/planning-artifacts/architecture.md#Frontend-Architecture`]
- No `project-context.md` file exists under the project root.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/5-1-publish-compatibility-boundaries-for-familiar-cli-concepts.md`.
- Web research completed 2026-06-12 against official Go pages. Go downloads lists `go1.26.4` as a stable version; Go examples/testing docs confirm `Example...` functions in `_test.go` files are compiled and can be executed by `go test` when they include output comments. Sources: https://go.dev/dl/, https://go.dev/blog/examples, https://pkg.go.dev/testing

### Current Repository State

- Baseline commit at story creation: `e6eccf4` (`docs(story-5.1): publish compatibility boundaries`).
- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no `go.sum` was discovered.
- `examples/migration/` does not exist yet.
- Existing dirty/untracked BMAD installer/config, `.agents/`, `.codex/`, `.idea/`, and story-automator files exist in the worktree. Do not revert or normalize them.
- `docs/compatibility.md` exists and currently states that Story 5.2 owns migration examples. After this story adds examples, update that deferred wording without claiming Story 5.3 or 5.4 is complete.
- `docs/release-checklist.md` has a `Migration evidence` slot, but it is intentionally unfilled release-candidate evidence. Story 5.4 owns final command outcomes and release readiness.

### Current Implementation Evidence To Reuse

- `docs/behavior-matrices.md` is the strongest local evidence source for implemented behavior. It already marks Epic 2 flag parsing, Epic 3 command routing/rendering/boundary behavior, Epic 4 config resolution/provenance behavior, and Story 5.1 compatibility boundaries as `current`.
- `docs/compatibility.md` is the adopter-facing boundary document. Keep migration examples aligned with its supported/narrowed/omitted/intentionally different framing.
- `docs/config-precedence.md` is the canonical precedence authority. The exact order is `explicit setter`, `flag binding`, `env`, `JSON`, `default`. [Source: `docs/config-precedence.md#Precedence-Order`]
- `docs/diagnostics-and-errors.md` is the diagnostic/source-label/redaction vocabulary authority. Config source labels are exactly `default`, `explicit setter`, `flag binding`, `env`, and `JSON`; `JSON` is uppercase. [Source: `docs/diagnostics-and-errors.md#Source-Labels`]
- `tools/depgate/` is already implemented and must remain the local dependency gate. [Source: `_bmad-output/planning-artifacts/architecture.md#Development-Workflow-Integration`]

### Public API Guardrails

- Flags examples should use the current exported API: `flags.String`, `Bool`, `Int`, `StringList`, `Custom`, `Shorthand`, `Repeatable`, `NoOptionDefault`, `Sensitive`, `NewSet`, `NewNormalizedSet`, `NameNormalizer`, `Set.Parse`, `Snapshot.Lookup`, `Snapshot.RemainingArgs`, `ValueState.Values`, `ValueState.Default`, `ValueState.Explicit`, `ValueState.Occurrences`, `ValueOccurrence.Spelling`, and `ValueOccurrence.NormalizedName`. [Source: `flags/definition.go`, `flags/set.go`, `flags/snapshot.go`]
- Flags failures should be asserted with sentinel and typed errors: `flags.ErrUnknownFlag`, `ErrMissingValue`, `ErrDuplicateValue`, `ErrConversion`, `ErrInvalidGroup`, `ErrHelpRequest`, and `*flags.ParseError`. Error strings are diagnostics only. [Source: `flags/errors.go`, `docs/diagnostics-and-errors.md#Diagnostic-Vocabulary`]
- Command examples should use the current exported API: `command.NewDefinition`, `Description`, `Usage`, `Aliases`, `Children`, `LocalFlags`, `InheritedFlags`, `FlagNormalizer`, `Definition.Route`, `Definition.RouteBoundary`, `Result.PathNames`, `Result.MatchTokens`, `Result.RemainingArgs`, `Result.Flags`, `Result.FlagSnapshot`, `Definition.WriteHelp`, `Result.WriteHelp`, `Result.WriteUsage`, and `Boundary` accessors. [Source: `command/definition.go`, `command/route.go`, `command/result.go`, `command/help.go`, `command/boundary.go`]
- Command failures should use `errors.Is` / `errors.As` with `command.ErrUnknownCommand`, `ErrInvalidCommandAlias`, `ErrDuplicateCommandToken`, `ErrFlagComposition`, `*command.UnknownCommandError`, and other typed command errors. Runtime flag parse failures during routing remain `*flags.ParseError` values. [Source: `command/errors.go`, `docs/diagnostics-and-errors.md#Diagnostic-Vocabulary`]
- Config examples should use the current exported API: `config.Define`, `String`, `Bool`, `Int`, `Duration`, `StringList`, `Default`, `Sensitive`, `NewSet`, `NewNormalizedSet`, `NewExplicitSnapshot`, `Assignment`, `BindEnv`, `MapEnv`, `EnvPrefix`, `EnvKeyReplacer`, `NewEnvSnapshot`, `LoadJSON`, `JSONReaderLabel`, `NewFlagSnapshot`, `FlagValue`, `Resolve`, `Snapshot.Lookup`, `Snapshot.IsSet`, typed getters, `SourceReport`, `WriteSourceReport`, `InspectDiagnostic`, and `WriteDiagnostic`. [Source: `config/doc.go`, `config/definition.go`, `config/source.go`, `config/flag.go`, `config/resolve.go`, `config/getter.go`, `config/report.go`]
- Config failure examples should use `errors.Is` / `errors.As` against `config.ErrUnknownSourceKey`, `ErrDuplicateBinding`, `ErrSourceConversion`, `ErrKeyNotFound`, `ErrKeyAbsent`, `ErrGetConversion`, `*config.SourceError`, and `*config.GetError` where relevant. [Source: `config/errors.go`, `docs/diagnostics-and-errors.md#Diagnostic-Vocabulary`]

### Example Design Requirements

- Put the examples in `examples/migration/` as package-level test files. Use a package name such as `migration_test` and import `github.com/petabytecl/dib/command`, `github.com/petabytecl/dib/flags`, and `github.com/petabytecl/dib/config` explicitly.
- Files under `examples/` must contain `Example...` functions where practical, not only ordinary `Test...` functions. Use output comments only for deterministic user-facing output that should be executed by `go test`; otherwise rely on table-driven `Test...` functions for structured API assertions. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Organization`]
- The examples should be adopter-facing and small. They should demonstrate migration concepts through current Dib APIs, not exhaust every package behavior already covered by package tests.
- Avoid introducing an `examples/migration/testdata/` directory unless needed. Inline JSON through `strings.NewReader` is preferable for config migration because AC4 requires injected inputs rather than host-specific files.
- Use `bytes.Buffer` for help/source-report rendering examples. Do not rely on process-global stdout/stderr output except where Go `Example...` output comments intentionally verify a concise final printed summary.
- For config flag binding, translate parsed flag states into `config.FlagValue` values in example code. Only pass `ExplicitlySet: true` when `ValueState.Explicit()` is true; do not let flag defaults override env or JSON.
- Include at least one table-driven test that proves examples preserve explicit inputs despite misleading ambient process state, such as unrelated `os.Args` or injected env values. Do not use `os.Getenv` as the config source.

### Files Expected To Be Created Or Updated

- NEW: `examples/migration/standard_flag_concepts_test.go`
- NEW: `examples/migration/shorthand_flag_migration_test.go`
- NEW: `examples/migration/nested_command_migration_test.go`
- NEW: `examples/migration/config_precedence_migration_test.go`
- UPDATE: `docs/compatibility.md`
- UPDATE: `docs/compatibility_test.go`
- UPDATE: `docs/behavior-matrices.md`
- UPDATE: `docs/provenance-log.md` if public reference docs influence example design or wording.
- OPTIONAL UPDATE: `docs/release-checklist.md` for a narrow migration-evidence pointer only. Do not complete release-candidate evidence in this story.
- Avoid changes under `command/`, `flags/`, `config/`, or `tools/` unless an example exposes a genuine bug that blocks the documented current API.

### Previous Story Intelligence

- Story 5.1 created `docs/compatibility.md`, added `docs/compatibility_test.go`, updated `docs/provenance-log.md`, updated `docs/behavior-matrices.md`, and kept migration examples, consolidated adoption evidence, and release readiness deferred to Stories 5.2, 5.3, and 5.4.
- Story 5.1 senior review tightened ambiguous compatibility-positioning wording. Continue checking `docs/` and `examples/migration/` for prohibited positive compatibility claims.
- Story 5.1 completed with `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, and a manual boundary-term search. Use the same verification rhythm.
- `docs/compatibility_test.go` already contains patterns for evidence link validation and positive-compatibility-claim prevention. Extend those tests rather than creating duplicate docs validation logic.

### Git Intelligence

- Recent commits:
  - `e6eccf4 docs(story-5.1): publish compatibility boundaries`
  - `a718d1d docs: add epic 4 retrospective`
  - `ca11b27 feat(story-4.5): report config provenance`
  - `b974995 feat(story-4.4): add typed config getters`
  - `c231957 feat(story-4.3): resolve config precedence`
- Established implementation rhythm: read all update files first, make the smallest docs/examples change needed, update evidence docs only after examples compile, then run repository gates.

### Architecture Guardrails

- Runtime packages, tests, runnable examples, and tools must remain standard-library-only unless the architecture is updated. [Source: `_bmad-output/planning-artifacts/architecture.md#Enforcement-Guidelines`]
- Dib must not claim source compatibility with Go `flag`, pflag, Cobra, or Viper. It offers a native Dib API with familiar concepts and documented differences. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#API-Contracts-Versioning-And-Dependency-Policy`]
- Compatibility with Go `flag`, pflag, Cobra, and Viper is semantic familiarity only; Dib must not promise source compatibility, drop-in replacement behavior, naming parity, legacy mental-model preservation, or reproduced edge-case bugs. [Source: `_bmad-output/planning-artifacts/architecture.md#Non-Functional-Requirements`]
- Examples must model importable-library usage with explicit inputs and returned results/errors. They must not invent demo apps, hidden process IO, `/cmd` scaffolds, or stdout-only behavior. [Source: `_bmad-output/planning-artifacts/architecture.md#Enforcement-Guidelines`]
- Migration examples belong under `examples/migration/`; architecture names the intended files `standard_flag_concepts_test.go`, `shorthand_flag_migration_test.go`, `nested_command_migration_test.go`, and `config_precedence_migration_test.go`. [Source: `_bmad-output/planning-artifacts/architecture.md#Complete-Project-Directory-Structure`]
- Record provenance for copied/generated/reference-derived artifacts before acceptance. The desired path here is independently written examples plus inspiration-only provenance if external docs are consulted. [Source: `_bmad-output/planning-artifacts/architecture.md#Clean-Room-Provenance-Enforcement`]
- Redaction examples must use only the architecture-owned fake sensitive corpus: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]

### Latest Technical Information

- Official Go downloads listed `go1.26.4` as a stable version on 2026-06-12. Keep `go.mod` at `go 1.26`; do not add a `toolchain` directive unless architecture changes. Source: https://go.dev/dl/
- Official Go examples guidance states that examples in `_test.go` files are compiled and can be executed as part of a package test suite. Source: https://go.dev/blog/examples
- Official `testing` package docs state that `Example...` output comments are compared during `go test`, and examples without output comments are compiled but not executed. Source: https://pkg.go.dev/testing

## Project Structure Notes

- `examples/migration/` is the architecture-owned home for FR19 migration examples. Create this directory rather than placing migration examples under `docs/`, `/cmd`, package-specific test files, or a new scaffold.
- `docs/compatibility.md` remains the compatibility boundary. It may link to migration examples after they exist, but it should not become a long migration guide.
- `docs/behavior-matrices.md` remains the executable-evidence index; package tests and examples remain the executable source of truth.
- `docs/provenance-log.md` owns source, access date, license/terms, affected artifact, and classification for any reference-influenced content.
- No UX files exist and no UI work applies.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- 2026-06-12: Confirmed sprint status preconditions, baseline commit `e6eccf4`, and absence of pre-existing `examples/migration/` before implementation.
- 2026-06-12: Initial shorthand migration example failed focused tests because final grouped value flags consume the next token before no-option defaults and duplicate single-value flags are rejected; adjusted the example to use a non-final no-option default and separate valid long/shorthand cases.
- 2026-06-12: Verification gates passed: `go test ./examples/migration`, `go test ./docs`, `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, compatibility boundary search, and module metadata checks.
- 2026-06-12: Senior review auto-fixes added command `Result.Flags` example coverage and Story 5.2 provenance entries for Go examples/testing docs.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Added executable migration examples for standard flag mental models, pflag-style shorthand behavior, Cobra-style command routing, and Viper-style config precedence using Dib's native APIs.
- Updated compatibility and behavior evidence docs to link Story 5.2 examples while leaving Story 5.3 adoption evidence and Story 5.4 release readiness deferred.
- No runtime package APIs, new dependencies, `/cmd` scaffold, process lifecycle ownership, or external-source-derived examples were added.
- Senior review fixed the remaining task/provenance gaps and re-ran all required verification gates successfully.

### File List

- `examples/migration/standard_flag_concepts_test.go`
- `examples/migration/shorthand_flag_migration_test.go`
- `examples/migration/nested_command_migration_test.go`
- `examples/migration/config_precedence_migration_test.go`
- `docs/compatibility.md`
- `docs/compatibility_test.go`
- `docs/behavior-matrices.md`
- `docs/provenance-log.md`
- `docs/release-checklist.md`
- `_bmad-output/implementation-artifacts/5-2-provide-migration-examples-for-flags-commands-and-config.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

### Change Log

- 2026-06-12: Created standard-library-only executable migration examples for flags, shorthand parsing, command routing, and config precedence/redaction.
- 2026-06-12: Updated compatibility evidence, docs tests, behavior matrix, and release checklist migration pointer for Story 5.2.
- 2026-06-12: Marked Story 5.2 ready for review after all required validation gates passed.
- 2026-06-12: Senior review fixed command composed-flag example coverage and added missing Story 5.2 provenance entries; marked story done after verification.

### Senior Developer Review (AI)

Reviewer: GPT-5 Codex on 2026-06-12

Outcome: Approved after auto-fixes.

Findings fixed:

- [HIGH] `examples/migration/nested_command_migration_test.go` did not exercise `Result.Flags` even though the completed task claimed the command example showed it. Added example/test coverage for the composed inherited/local flag set.
- [MEDIUM] `docs/provenance-log.md` lacked Story 5.2 entries for the Go examples/testing documentation cited in dev notes. Added inspiration-only provenance entries for the migration example artifacts.

Validation:

- `GOCACHE=/tmp/dib-go-build go test ./examples/migration`
- `GOCACHE=/tmp/dib-go-build go test ./docs`
- `GOCACHE=/tmp/dib-go-build go test ./...`
- `GOCACHE=/tmp/dib-go-build go vet ./...`
- `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
- `git diff --check`
- `rg -n "(?i)drop-in|source-compatible|clone API|framework compatibility layer|compatible replacement" docs examples/migration`
- `rg -n "^(require|replace|toolchain)\b" go.mod`
- `test ! -e go.sum`

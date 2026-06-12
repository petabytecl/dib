---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - "_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md"
  - "_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/addendum.md"
  - "_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/review-rubric.md"
  - "_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/reconcile-brief.md"
  - "_bmad-output/planning-artifacts/briefs/brief-dib-2026-06-10/brief.md"
  - "_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12.md"
workflowType: 'architecture'
project_name: 'dib'
user_name: 'Coto'
date: '2026-06-11'
lastStep: 8
status: 'complete'
completedAt: '2026-06-12'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

Dib's value is not "more CLI features." It is lower supply-chain risk, predictable behavior, easier testing, and auditable correctness for Go developers building or maintaining CLIs without hidden global state or runtime dependencies.

Primary users are Go developers building internal tools, platform CLIs, repo-local tools, and distributed binaries. Secondary stakeholders include maintainers reviewing API behavior, security/compliance reviewers validating dependency and clean-room claims, test authors, and downstream users reading help, usage, diagnostics, and source reports.

**Functional Requirements:**
Dib V1 is organized around three composable runtime library surfaces: command routing, flag parsing, and configuration resolution. Each surface must work independently before users compose them. Command requirements cover nested command trees, aliases, local and inherited flags, explicit execution with `context.Context`, deterministic help/usage rendering, and typed command errors. Flag requirements cover explicit `FlagSet` instances, long flags, shorthand flags, shorthand groups, no-option defaults, repeated/custom values, name normalization, `--`, interspersed positional args, and typed parse diagnostics. Config requirements cover registered keys, defaults, explicit setters, lazy flag bindings, env bindings, JSON files/readers, the exact PRD-defined precedence order, typed getters, source reporting, provenance, and sensitive diagnostic redaction.

Documentation, validation, and release evidence are product requirements. Behavior matrices, dependency checks, lint evidence, coverage evidence, migration examples, public usage docs, clean-room audit notes, compatibility documentation, and parser hardening evidence are trust features for adoption, not auxiliary prose. Every behavior matrix, validation artifact, and cross-cutting concern must carry PRD/addendum/rubric IDs where available; gaps must be marked as untraced assumptions.

**Non-Functional Requirements:**
The dominant NFR is zero external runtime dependencies. This serves supply-chain trust, binary portability, auditability, and adoption in dependency-sensitive environments. Runtime packages must import only the Go standard library, while development/test tooling may differ only if isolated and enforcement proves it.

The public API must be explicit-instance-first and testable through returned values/errors plus injected args, readers, writers, env lookup, JSON readers/files, context, and user-provided value parsers. No hidden process state includes package-level registries, mutable singletons, implicit caches, default command/config instances, ambient env/args/stdout access, and `os.Exit` in runtime paths.

Public errors are stable contracts. They must preserve machine-readable category, key/flag/command/source context, wrapping/unwrapping behavior, and redaction guarantees. Determinism includes ordering of commands, flags, aliases, config sources, help text, usage text, diagnostics, source reports, and validation output. Every command, flag, config, diagnostic, provenance, redaction, and determinism contract must be assertable without stdout/stderr scraping or string-only matching.

Compatibility with Go `flag`, pflag, Cobra, and Viper is semantic familiarity only. Dib must support migration confidence without promising source compatibility, drop-in replacement behavior, naming parity, legacy mental-model preservation, or reproduced edge-case bugs.

The initial experimental release targets Go 1.26+ as a v0 API. V0 status limits API-stability promises, but it does not relax correctness, redaction, clean-room evidence, dependency, or release-gate expectations.

**Scale & Complexity:**

- Primary domain: Go developer library for CLI command, flag, and config behavior.
- Complexity level: Low operational scale, high correctness density.
- Estimated architectural components: command routing, flag parsing, config resolution, typed error boundaries, diagnostics/help/usage rendering, documentation/examples, behavior matrices, clean-room evidence, and validation/dependency tooling.

The highest correctness risks are semantic ambiguity in flag parsing and config precedence. Observable contracts include selected command, remaining args, parsed flag values, config value/source, returned typed error, rendered help text, and diagnostics.

### Technical Constraints & Dependencies

This Step 2 analysis does not choose package layout, exported API names, parser algorithms, or validation mechanisms.

Architecture inputs are limited to approved PRD artifacts and the product brief. Public docs and observable behavior are permitted references later, but Step 2 did not perform a fresh public-doc research pass. No UX spec, research document, existing project documentation, `project-context.md`, existing implementation, or repo convention analysis was available, so user pain, migration frequency, adoption priority, and inherited implementation constraints remain PRD-derived assumptions rather than independently validated findings.

Dib must follow a clean-room policy as both a process constraint and an architectural constraint. Public docs and observable behavior are allowed references for behavioral expectations. Copied source, tests, comments, examples, fixtures, names, file organization, implementation details, copied README shapes, AI-generated snippets too closely derived from source projects, and compatibility examples derived too closely from prior libraries are disallowed. Fixtures, behavior matrices, and fuzz seeds require source/provenance records before acceptance.

Config file support is JSON-only. Non-goals include YAML, TOML, HCL, dotenv, INI, properties, remote config, live reload, shell completion generation, man pages, scaffolding, reflection-heavy struct decoding, package-global helpers, and source-compatible Cobra/pflag/Viper APIs. These exclusions protect V1 focus, clean-room confidence, and correctness density.

### Cross-Cutting Concerns Identified

- Runtime dependency enforcement across every runtime package, with automated import/dependency checks as release evidence.
- Explicit instance ownership, repeatability, reset behavior, and no hidden global or ambient process state.
- Public typed error taxonomy with Go error inspection, wrapping/unwrapping, context preservation, and string-free API assertions.
- Parser edge cases: aliases vs command names, inherited vs local flags, long vs shorthand normalization, shorthand groups, `--`, interspersed positionals, repeated values, no-option defaults, unknown flags, and conversion failures.
- Config precedence and provenance across defaults, explicit setters, lazy flag bindings, env, JSON file/reader, typed getters, missing values, conversion failures, source reporting, and redaction.
- Developer debugging journey: users must be able to answer why a command, flag, or config value resolved a certain way.
- Deterministic help, usage, diagnostics, source reports, behavior matrices, and stable ordering rules.
- Clean-room compliance evidence across implementation, generated content, tests, docs, examples, fixtures, migration content, and fuzz seeds.
- Human-facing trust surfaces: help text, usage text, diagnostics, source reports, redaction behavior, examples, and behavior tables.
- Compatibility positioning that shows familiar behavior without promising source-compatible APIs.
- Cross-source sensitive redaction for flags, env, JSON, defaults, setters, conversion errors, diagnostics, source reports, and validation failures while preserving useful key/source identity.
- Terminology consistency for command, flag, config key, source, binding, default, explicit value, inherited flag, and diagnostic.
- Runnable or mechanically checked examples, behavior matrices, and migration snippets where practical.

Success for V1 means a developer can read the docs, run examples, assert behavior without string matching, trace each config value to its source, and trust that runtime packages remain standard-library-only.

## Starter Template Evaluation

### Primary Technology Domain

Dib is a Go importable library/toolkit for CLI command routing, flag parsing, and configuration resolution. It is not a CLI application starter, web app, backend service, or full-stack project.

Existing technical preferences are already defined by the PRD: Go 1.26+, standard-library-only production/runtime packages, explicit-instance APIs, clean-room implementation, deterministic behavior, typed errors, and no package-global helpers. Tests must also prefer the standard library unless a later architecture decision explicitly approves a test-only tool. No `project-context.md`, UX spec, existing implementation, `go.mod`, or source tree was found in the checkout.

### Starter Options Considered

**Standard Go module bootstrap - selected**

Use the official Go module path with no external starter template:

```bash
go mod init github.com/petabytecl/dib
go mod edit -go=1.26
go test ./...
```

This matches Dib's constraints because it introduces no runtime dependencies, no copied template structure, and no third-party framework assumptions. It keeps package layout and public API decisions available for later architecture steps. `go test ./...` is only a bootstrap sanity check until at least one package exists.

**`gonew` templates - not selected**

`gonew` is an experimental Go project-template tool. It is useful for copying predefined module templates, but Dib's clean-room and provenance constraints make template copying a poor default foundation. Any future template-derived content would require explicit provenance review.

**Cobra / urfave/cli starters - not selected**

Cobra and urfave/cli are maintained Go CLI frameworks and provide command/flag/help scaffolding, but Dib is building a clean-room standard-library-only toolkit in the same problem space. Using these as starters would create dependency, provenance, and compatibility-positioning conflicts.

**`golang-standards/project-layout` - not selected**

This repository is not an official Go standard and is broader than Dib needs at initialization. The official Go module layout guidance is a better input for later package-boundary decisions.

### Selected Starter: Standard Go Module Bootstrap

**Rationale for Selection:**

The best starter is no external starter template. Dib needs a small, auditable Go module whose first implementation story can establish `go.mod`, baseline repository metadata, dependency enforcement, initial package documentation, at least one real test, one standard-library-only example, and clean-room provenance tracking without importing framework structure or copied examples.

No external starter protects scope and provenance, but it also means Dib gets no inherited onboarding experience. The first implementation story should include minimal package documentation, one standard-library-only example, and a clear import path so early adopters can verify the library shape without a CLI app scaffold.

**Initialization Command:**

```bash
go mod init github.com/petabytecl/dib
go mod edit -go=1.26
go test ./...
```

Use `go 1.26`. Do not add a `toolchain` directive unless the architecture later chooses patch-level toolchain pinning.

**First Implementation Story Gates:**

```bash
go test ./...
go vet ./...
go list -deps -f '{{if and (not .Standard) (not .Module.Main)}}{{.ImportPath}}{{end}}' ./... | sed '/^$/d'
```

The `go list` dependency command should produce no output for runtime packages. It is POSIX-shell based; if CI must be shell-agnostic later, replace it with a tiny Go verifier or an `awk` equivalent. `go test ./...` should not be treated as meaningful behavioral verification until the first package and package-level `_test.go` file exist.

**Architectural Decisions Provided by Starter:**

**Language & Runtime:**
Go 1.26 module targeting standard-library-only production/runtime packages. Any test-only or tooling dependency needs explicit architecture/provenance approval and must not enter the importable library API or runtime graph.

**Styling Solution:**
Not applicable.

**Build Tooling:**
The Go toolchain is the default build and test foundation. No external build system is introduced by the starter.

**Testing Framework:**
Use the standard `testing` package and `go test`. Unit/table-driven tests are the primary level for routing, flag parsing, config precedence, typed errors, and deterministic behavior. Standard Go fuzzing should be introduced later for parser/config boundary hardening, not as starter scaffolding.

**Code Organization:**
No package layout is selected by this starter. Public package boundaries, internal packages, examples, and command/demo locations remain explicit architecture decisions for later steps. Do not create a `/cmd` application scaffold as part of starter initialization; Dib is library-first and any demo command should be a later explicit architecture decision.

**Development Experience:**
The starter gives a minimal, reproducible module baseline. It supports dependency checks, test-first implementation, clean-room provenance tracking, and small-file growth without committing to a third-party framework or app scaffold.

**Provenance Evidence:**
This starter uses only `go mod init` and no copied template files. Any future generated/copied code requires explicit provenance review before merge.

**Research Evidence:**

- Go downloads page verified Go 1.26.4 as current stable: https://go.dev/dl/
- Go dependency documentation identifies `go mod init` as the module bootstrap command and recommends using the repository location as the module path when publishing: https://go.dev/doc/modules/managing-dependencies
- Go module layout guidance is the official structure reference for later package-boundary decisions: https://go.dev/doc/modules/layout
- `gonew` is described by the Go project as an experimental template tool: https://go.dev/blog/gonew
- Cobra's generator creates application structure and command scaffolding, which is inappropriate for Dib's library-first clean-room starter: https://github.com/spf13/cobra-cli
- urfave/cli is a package for building command-line tools in Go, so it is unsuitable as Dib's dependency-free foundation: https://cli.urfave.org/
- `golang-standards/project-layout` states it is not an official Go standard and can be overkill for simple projects: https://github.com/golang-standards/project-layout

**Note:** Project initialization using this command should be the first implementation story.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions:**
- Reusable definitions plus per-run result snapshots.
- Boundary validation plus redaction-first diagnostics.
- Capability-scoped public API with standard Go error inspection.
- GitHub CI plus Go module release gates.

**Important Decisions:**
- Frontend architecture is not applicable.
- No app deployment, binary publishing, Docker, Kubernetes, or `/cmd` scaffold in V1.
- Go 1.26 is the current baseline; support policy must be explicit at release time.

**Deferred Decisions:**
- Package layout and exported API names.
- Parser/config implementation algorithms.
- Demo command location, if any.
- SHA-pinned CI actions, unless provenance requirements increase.
- Non-stdlib test/tooling dependencies, unless explicitly approved later.

### Data Architecture

**Decision:** Reusable definitions plus per-run result snapshots.

Dib will use caller-observably immutable definition values and explicit per-run result snapshots. Command definitions, flag definitions, flag sets, config definition sets, and config bindings are reusable values. Derived definitions return new values and must avoid exported mutable internals or shallow-copy aliasing.

Command routing, flag parsing, and config resolution return self-contained snapshots containing selected command, set flags, remaining args, config value/source provenance, diagnostics, and typed errors. Snapshots must never write back to definitions and must not depend on live process state, environment variables, readers, or lookups after creation.

**Rationale:** Enables safe reuse, deterministic tests, no hidden state, concurrency safety, redaction, provenance, and composability.

### Authentication & Security

**Decision:** Boundary validation plus redaction-first diagnostics.

Dib will not provide authentication, authorization, encryption, or secret management in V1. It treats CLI args, env values, JSON config, custom parsers, readers, and lookup functions as untrusted boundary inputs.

Setup-time validation must catch invalid definitions, duplicate names, binding collisions, normalization collisions, invalid relationships, and unsupported config combinations where possible.

Sensitivity markers and provenance metadata may guide diagnostics, but they do not provide secret management. Raw sensitive values must never appear in errors, `String` output, debug strings, rendered diagnostics, source reports, or example output.

**Rationale:** Keeps Dib scoped as a library while preventing sensitive value leaks and preserving machine-readable error context.

### API & Communication Patterns

**Decision:** Capability-scoped public API plus standard Go error inspection.

Dib will expose public surfaces for command routing, flag parsing, and config resolution rather than a broad root framework facade. Each surface must be independently usable and composable through explicit values and snapshots.

Public APIs use standard Go conventions: explicit inputs, returned values/errors, `context.Context` where execution crosses a boundary, and `io.Reader` / `io.Writer` where callers provide input/output. Primary APIs must not depend on package globals, hidden process IO, implicit environment reads, or root singletons.

Public errors must support `errors.Is` / `errors.As` compatible inspection. Error strings are diagnostics, not programmatic contracts. Docs and examples are part of the API contract and must be runnable and clean-room.

### Frontend Architecture

Not applicable. Dib has no browser UI, client-side state, routing, bundle, styling, or frontend performance architecture in V1. Developer experience is handled through package docs, examples, diagnostics, and tests.

### Infrastructure & Deployment

**Decision:** GitHub CI plus Go module release gates.

Dib will use GitHub Actions with an explicit runner image such as `ubuntu-24.04`, official actions, and the Go version declared consistently in `go.mod`, docs, CI, and release notes.

Core PR gates:

```bash
go test ./...
go vet ./...
go run ./tools/depgate
```

The dependency gate is `go run ./tools/depgate`. It must fail on unapproved external imports for library, test, example, and tool packages unless this architecture is updated.

Core gates also include a pinned lint command selected by Story 6.1 and a coverage validation command selected by Story 6.2. The lint command must be reproducible and isolated as development or CI tooling. External linter tooling may be downloaded or invoked by CI, but it must not enter Dib runtime package imports or the root module's checked package imports without an approved architecture update.

Coverage validation must use standard Go coverage output where practical and apply package-aware thresholds. Public runtime packages (`command`, `config`, and `flags`) are release-surface packages and must report threshold evidence separately from tooling packages. Tooling packages may carry a documented threshold or exception when critical-path tests cover the tool behavior.

Release-candidate gates additionally include:

```bash
go test -race ./...
```

Dib releases are Go module releases, not binary deployments. v0 tags may include breaking changes, but those changes require release notes, migration guidance, updated examples, provenance notes, and passing CI evidence tied to the exact tagged commit.

Release evidence must record test, vet, dependency-gate, lint, coverage, race-test, docs/examples, runner/action version, provenance, and compatibility/migration status. CI failures block tagging. Waivers require an owner, reason, expiry, and impact.

### Decision Impact Analysis

**Implementation Sequence:**
1. Initialize module and first package/test/example scaffold.
2. Establish CI gates and dependency check.
3. Define immutable definition/result snapshot contracts.
4. Define typed error/redaction contracts.
5. Define public capability surfaces.
6. Implement command, flag, and config behavior with table-driven tests.
7. Add docs/examples/migration evidence as behavior stabilizes.

**Cross-Component Dependencies:**
- Data snapshots shape API contracts, tests, diagnostics, and redaction.
- Error taxonomy spans command, flag, config, and JSON/config source handling.
- CI gates enforce the runtime dependency rule and release evidence.
- Docs/examples must reflect package boundaries, snapshot contracts, and public error behavior.

### Research Evidence

- Go module layout and module initialization guidance: https://go.dev/doc/modules/layout
- Go dependency management and `go mod init`: https://go.dev/doc/modules/managing-dependencies
- Go package naming guidance: https://go.dev/blog/package-names
- Go `errors.Is` / `errors.As`: https://pkg.go.dev/errors
- Go testing examples: https://pkg.go.dev/testing
- Go examples as runnable documentation: https://go.dev/blog/examples
- GitHub Actions Go build/test guidance: https://docs.github.com/actions/automating-builds-and-tests/building-and-testing-go
- GitHub-hosted runner references: https://docs.github.com/en/actions/reference/runners/github-hosted-runners
- Go module publishing and versioning: https://go.dev/blog/publishing-go-modules and https://go.dev/doc/modules/version-numbers

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:**
- Terminology drift across command, flag, config, provenance, source label, binding, snapshot, diagnostic, and error concepts.
- File/test/example placement drift before final package structure is decided.
- Error and diagnostic shape drift that would force string matching.
- Config precedence, provenance, failure-category, and redaction drift across defaults, explicit setters, flags, env, and JSON.
- Test fixture and contract assertion drift across command, flag, and config behavior.
- Dependency-gate drift between runtime packages, tests, examples, and tooling.
- Documentation/provenance drift that could weaken clean-room evidence.
- Scope creep drift that turns V1 from an importable toolkit into a stealth framework.

### Out Of Scope For V1

Dib V1 has no database schema, network API, browser UI, event bus, async event system, static assets, environment-file convention, app deployment, loading-state model, or API response envelope. Agents must not fill these template categories with web, database, UI, or service assumptions.

Dib V1 also has no shell completion, interactive prompting, config watching/reloading, plugin system, subcommand marketplace, binary release workflow, or framework compatibility layer unless the architecture is updated first.

### Naming Patterns

**Canonical Terminology:**

| Concept | Required Term | Avoid |
|---|---|---|
| CLI dispatch | command routing | command execution |
| CLI argument interpretation | flag parsing | option processing |
| config value selection | config resolution | config loading |
| reusable setup | definition | spec, builder state |
| per-run result | snapshot | state, context |
| value origin concept | provenance | provider, backend |
| concrete origin label | source label | source as a vague synonym |

Use `provenance` as the canonical concept. Use `source label` only for the concrete origin label inside provenance.

New domain terms require an architecture-document update before implementation use. Do not add compatibility aliases or migration-oriented terminology from existing CLI/config libraries merely because users may recognize them.

**Code Naming Conventions:**
- Follow standard Go naming conventions: exported identifiers use `PascalCase`; unexported identifiers use `camelCase`; package names are short, lowercase, and not stuttered.
- File names use lowercase words and underscores when needed, especially for tests such as `*_test.go`.
- Test names describe observable behavior, for example `TestFlagParsingRejectsUnknownLongFlag`.
- Avoid names copied from Cobra, pflag, Viper, or other prior libraries unless they are unavoidable generic terms.

### Structure Patterns

**Project Organization:**
- Tests live beside the package under test in `*_test.go` files.
- Runnable examples live in Go example tests where practical so `go test ./...` verifies them.
- Runtime implementation must stay importable as library code and must not introduce a `/cmd` scaffold in V1.
- Shared helpers should stay unexported until at least two concrete call sites prove the abstraction is needed.
- Any internal package boundaries must preserve the three capability surfaces: command routing, flag parsing, and config resolution.

**File Structure Patterns:**
- Keep source files focused and small; split by cohesive behavior instead of broad utility buckets.
- Put package-level documentation in the package it documents.
- Keep clean-room/provenance notes with planning or documentation artifacts, not embedded as noisy comments in runtime code.
- Do not add generated/copied files without explicit provenance review.

### Format Patterns

**Go Library Result Formats:**
- Prefer returned values and errors over stdout/stderr side effects.
- Result snapshots expose machine-readable state for assertions.
- Rendered strings are human-facing diagnostics, not the only programmatic contract.
- Human-facing diagnostics are still part of the product contract: they must be actionable, redaction-safe, and consistent.

**Data Exchange Formats:**
- JSON is the only config file format in V1.
- Config precedence is defined by the PRD and must be followed consistently by resolvers, tests, examples, diagnostics, and source reports.
- Config provenance must distinguish the closed V1 source-label vocabulary: `default`, `explicit setter`, `flag binding`, `env`, and `JSON`.
- Resolution failure categories must distinguish at least `conversion failure` and `source read failure`.
- Failed resolution reports must distinguish attempted source label from failure category when both apply.
- Provenance labels must be stable test-facing concepts, not rendered-message wording.
- Diagnostics that mention value origin must use the same provenance/source-label vocabulary as config resolution.
- Sensitive values must be redacted consistently across errors, debug strings, diagnostics, source reports, and examples.

### Communication Patterns

**State Management Patterns:**
- Definitions are caller-observably immutable.
- Caller-observably immutable means definitions and snapshots must not expose mutable aliases to caller-provided slices, maps, readers, env lookups, or config data after construction or resolution.
- Derived definitions return new values.
- Per-run snapshots do not mutate definitions.
- No package-global registries, default instances, implicit process IO, or ambient env/args reads in primary APIs.
- Any cache-like optimization must be unobservable, deterministic, concurrency-safe, and compatible with definition reuse.

### Process Patterns

**Error Handling Patterns:**
- Public errors must support `errors.Is` / `errors.As`.
- Error strings are diagnostics only and must not be required for programmatic assertions.
- Typed errors preserve category and relevant command/flag/key/provenance context.
- Boundary validation failures must be explicit, typed, deterministic, and redaction-safe.
- Do not silently swallow invalid definitions, parse failures, conversion failures, unknown flags, unsupported config, or source read errors.
- Programmatic behavior must be asserted through typed errors, categories, context, snapshots, and provenance. Rendered messages may be tested only for human-facing wording or redaction.
- Diagnostics should consistently identify the failing boundary, relevant command/flag/key, provenance when applicable, typed category, and redacted value status.

**Runtime Boundary Patterns:**
- Callers provide args, readers, writers, env lookup, JSON readers/files, context, and custom parsers explicitly.
- Snapshots must not depend on live process state, environment variables, readers, or lookup functions after creation.

**Test Fixture Patterns:**
- Fixtures must be small, local to the behavior under test, deterministic, and clean-room.
- Fixtures must not depend on live env, current working directory, wall clock, stdin/stdout, or host files unless the test explicitly models that boundary.
- Shared fixtures require a stable cross-package need; otherwise keep them local.
- Redaction tests must use the architecture-defined sensitive-value corpus so agents do not invent incompatible examples.
- The sensitive-value corpus is owned by this architecture document. Test guidance may reference it but must not redefine it.
- Architecture-defined sensitive-value corpus:
  - `dib_fake_secret_value`
  - `dib_fake_password_value`
  - `dib_fake_token_value`
- Corpus examples must be obviously fake and must not resemble real credentials.

**Contract Assertion Patterns:**
- Tests should assert typed errors, provenance, redaction, determinism, and immutability as separate observable contracts.
- Rendered diagnostics may use golden files only for human-facing formatting. Public behavior must still be asserted through structured state and typed errors.
- Tests for derived definitions must verify the original reusable definition has unchanged observable behavior across later runs.
- At least one test function or table-driven test group should verify reused definitions across repeated and concurrent runs without observed mutation.

### Enforcement Guidelines

**All AI Agents MUST:**
- Preserve the command routing / flag parsing / config resolution terminology.
- Keep runtime code, tests, and runnable examples standard-library-only unless this architecture is updated.
- Use explicit inputs and returned values/errors rather than package globals or ambient process state.
- Add or update tests for observable behavior, typed errors, determinism, provenance, redaction, immutability, and concurrency-safe reuse.
- Keep examples runnable or mechanically checked wherever practical.
- Record provenance for copied/generated/reference-derived artifacts before acceptance.
- Ensure new behavior maps directly to command routing, flag parsing, or config resolution. Convenience features outside those surfaces require an architecture update before implementation.
- Ensure examples model importable-library usage with explicit inputs and returned results/errors. Examples must not invent demo apps, hidden process IO, `/cmd` scaffolds, or stdout-only behavior.

**Pattern Enforcement:**
- `go test ./...` verifies tests and runnable examples.
- `go vet ./...` verifies basic Go correctness.
- The dependency gate is the documented CI/local dependency check; agents must not create alternate local-only dependency checks.
- The dependency gate must fail if any runtime package, test, or runnable example imports a non-standard-library package unless this architecture is updated.
- Explicitly approved non-runtime tooling dependencies may support repository checks or generation, but must not be imported by runtime packages, tests, or runnable examples unless this architecture is updated.
- Reviewers must reject string-only assertions for public error behavior when typed inspection is available.
- Pattern changes must be reflected in this architecture document before implementation agents rely on them.

**Clean-Room Provenance Enforcement:**
Any copied, generated, or reference-derived artifact needs a note recording source, date, license/terms, and whether it was copied, adapted, or only used as inspiration. Inspiration-only references must not contribute copied names, examples, fixtures, or structure unless recorded as adapted or copied.

### Pattern Examples

**Good Examples:**
- A parse test asserts typed error identity with `errors.Is` or `errors.As`.
- A config test asserts the resolved value and provenance without scraping stdout.
- A redaction test verifies sensitive values are absent from errors, debug strings, diagnostics, source reports, and examples.
- A reusable definition is used across multiple parse/resolve runs without observed mutation.
- An example test compiles through `go test ./...`.
- A fixture is local, deterministic, clean-room, and models only the boundary under test.

**Anti-Patterns:**
- Adding a package-global default command, flag set, config registry, or env reader.
- Returning only rendered diagnostic text for behavior that callers need to inspect.
- Copying names, examples, fixtures, or structure from prior CLI libraries without provenance approval.
- Adding a runtime dependency to simplify parsing, config loading, diagnostics, or tests.
- Creating `/cmd` scaffolding or binary-release assumptions during V1 library implementation.
- Letting sensitive values appear in `Error()`, `String()`, debug output, source reports, diagnostics, or examples.
- Adding helper abstractions before two call sites or before behavior is stable.
- Creating alternate dependency checks that differ from the documented CI/local gate.
- Adding compatibility aliases, shell completion, interactive prompting, config reload/watch, plugin behavior, or framework-layer features without an architecture update.

## Project Structure & Boundaries

### Complete Project Directory Structure

```text
dib/
├── go.mod
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
├── .gitignore
├── .github/
│   └── workflows/
│       └── ci.yml
├── command/
│   ├── doc.go
│   ├── command.go
│   ├── route.go
│   ├── result.go
│   ├── flags.go
│   ├── help.go
│   ├── validation.go
│   ├── errors.go
│   ├── command_test.go
│   ├── route_test.go
│   ├── result_test.go
│   ├── flags_test.go
│   ├── help_test.go
│   └── validation_test.go
├── flags/
│   ├── doc.go
│   ├── flag.go
│   ├── set.go
│   ├── parse.go
│   ├── value.go
│   ├── normalize.go
│   ├── diagnostics.go
│   ├── usage.go
│   ├── errors.go
│   ├── set_test.go
│   ├── parse_long_test.go
│   ├── parse_shorthand_test.go
│   ├── repeated_test.go
│   ├── normalize_test.go
│   ├── diagnostics_test.go
│   ├── fuzz_test.go
│   └── testdata/
│       └── fuzz/
│           └── FuzzParse/
│               ├── README.md
│               └── basic.txt
├── config/
│   ├── doc.go
│   ├── key.go
│   ├── definitions.go
│   ├── resolver.go
│   ├── precedence.go
│   ├── binding_flag.go
│   ├── binding_env.go
│   ├── json.go
│   ├── source.go
│   ├── redaction.go
│   ├── errors.go
│   ├── definitions_test.go
│   ├── precedence_test.go
│   ├── binding_flag_test.go
│   ├── binding_env_test.go
│   ├── json_test.go
│   ├── redaction_test.go
│   └── testdata/
│       └── json/
│           ├── valid.json
│           ├── unknown-key.json
│           └── bad-type.json
├── internal/
│   └── README.md
├── examples/
│   ├── multicommand/
│   │   └── example_test.go
│   └── migration/
│       ├── standard_flag_concepts_test.go
│       ├── shorthand_flag_migration_test.go
│       ├── nested_command_migration_test.go
│       └── config_precedence_migration_test.go
├── docs/
│   ├── clean-room-policy.md
│   ├── provenance-log.md
│   ├── compatibility.md
│   ├── behavior-matrices.md
│   ├── diagnostics-and-errors.md
│   ├── config-precedence.md
│   ├── testing.md
│   └── release-checklist.md
└── tools/
    └── depgate/
        ├── main.go
        └── main_test.go
```

### Architectural Boundaries

**API Boundaries:**
- `command/` is the public command routing package. The package name is acceptable because the domain object is a command, but the package owns command routing rather than process lifecycle execution.
- `command/` owns command trees, nested routing, aliases, local/inherited flag attachment, help/usage rendering entry points, and command-specific typed errors.
- Callback handling is deferred. `command/` may model caller-owned callbacks as definition metadata and may return a matched callback in route/result snapshots only if a later architecture/API decision explicitly adds that surface. Dib does not invoke callbacks unless that future invocation surface is explicitly approved.
- `flags/` is the public flag parsing package. It owns explicit flag sets, long/shorthand parsing, shorthand groups, repeated/custom values, no-option defaults, normalization, parse diagnostics, usage metadata, and flag-specific typed errors.
- `config/` is the public config resolution package. It owns caller-owned reusable key definitions, defaults, explicit setters, flag binding inputs, env bindings, JSON readers/files, precedence, typed getters, provenance, redaction, and config-specific typed errors.
- The module root does not provide a broad public facade, package-global helpers, or default singleton API.

**Component Boundaries:**
- `command/` may attach or accept `flags/` definitions and snapshots for command-local and inherited flags.
- Shared flag metadata and flag parsing semantics live in `flags/`, not `command/`.
- `config/` accepts explicit flag binding inputs from callers. A direct `config -> flags` package import is deferred unless API design proves it necessary; if introduced, it must depend only on exported snapshot/value contracts.
- `flags/` must remain fully usable without `command/` or `config/`.
- `command/` must not depend on `config/`; callers compose command, flag, and config behavior explicitly.
- `internal/` is provisional shared support. Internal packages such as `internal/text`, `internal/diagnostic`, or `internal/ordered` may be introduced only after at least two concrete call sites prove the need.
- `tools/depgate/` is repository tooling, not an importable library package. It must remain isolated from `command/`, `flags/`, and `config/`.

**Service Boundaries:**
Not applicable. Dib has no service runtime, server process, network API, deployment service, or app lifecycle owner in V1.

**Data Boundaries:**
- No database schema or persistent app state exists in V1.
- Config data enters through explicit setters, parsed flag snapshots or binding inputs, injected env lookup, JSON path, or JSON reader.
- JSON fixtures stay under `config/testdata/json/`.
- Fuzz corpus data stays under the relevant package `testdata/fuzz/` directory.
- Fuzzing must use standard Go fuzzing support, and seed corpus files must be clean-room and deterministic.

### Requirements To Structure Mapping

**Feature / FR Mapping:**
- FR-1 through FR-4 command routing and help: `command/`, package tests, `docs/behavior-matrices.md`, `examples/multicommand/`.
- FR-5 through FR-10 flag parsing: `flags/`, `flags/testdata/fuzz/`, `docs/behavior-matrices.md`.
- FR-11 through FR-16 config resolution: `config/`, `config/testdata/json/`, `docs/config-precedence.md`, `docs/diagnostics-and-errors.md`.
- FR-17 clean-room policy: `docs/clean-room-policy.md`, `docs/provenance-log.md`, `CONTRIBUTING.md`.
- FR-18 compatibility boundaries: `docs/compatibility.md`.
- FR-19 migration examples: `examples/migration/`.
- FR-20 behavior matrices: co-located package tests plus `docs/behavior-matrices.md`.
- FR-21 dependency rule: `.github/workflows/ci.yml`, `tools/depgate/`, `docs/release-checklist.md`.
- FR-22 parser hardening: `flags/fuzz_test.go`, `flags/testdata/fuzz/FuzzParse/`.

**Cross-Cutting Concerns:**
- Typed errors: `command/errors.go`, `flags/errors.go`, `config/errors.go`, `docs/diagnostics-and-errors.md`.
- Deterministic ordering: package tests first; shared ordering support may move under `internal/` only after repeated need is proven.
- Redaction: `config/redaction.go`, redaction tests, optional internal diagnostics support if repeated need is proven, and the architecture-owned sensitive-value corpus from Step 5.
- Clean-room evidence: `docs/clean-room-policy.md`, `docs/provenance-log.md`, `CONTRIBUTING.md`, release checklist, and provenance notes in review artifacts.
- Runtime dependency enforcement: `tools/depgate/` and CI.

### Integration Points

**Internal Communication:**
- Public package communication happens through explicit values, returned snapshots, and typed errors.
- No package communicates through global registries, process args, process env, stdout/stderr, hidden caches, or default singletons.
- Rendered diagnostics flow through writers supplied by callers.
- Callback handling is deferred; Dib does not invoke callbacks unless a later architecture/API decision explicitly adds an invocation surface.

**External Integrations:**
- `context.Context` may be carried in caller-owned callback metadata or future invocation surfaces only if a later architecture/API decision adds invocation.
- `io.Reader` for JSON config and examples.
- `io.Writer` for help, usage, and diagnostics.
- Env lookup is injected as a function rather than read implicitly from process env.
- Filesystem access is explicit for JSON path loading only.

**Data Flow:**
1. Callers define reusable command, flag, and config definitions.
2. Callers provide args/readers/writers/env lookup/context explicitly.
3. `flags/` returns parse snapshots.
4. `command/` returns route/result snapshots and typed errors.
5. `config/` resolves values using the canonical precedence defined in `docs/config-precedence.md`; the PRD order is explicit setter, parsed flag, environment variable, JSON file, default.
6. Diagnostics and source reports use the shared provenance/source-label vocabulary and redaction rules.

### File Organization Patterns

**Configuration Files:**
- `go.mod` is the module definition.
- `.github/workflows/ci.yml` owns CI execution.
- No `.env`, Docker, Kubernetes, or app deployment config belongs in V1.

**Source Organization:**
- Public source is organized by capability package: `command/`, `flags/`, `config/`.
- Shared code lives under `internal/` only when needed by multiple packages.
- No root facade package is introduced.
- Config definitions are caller-owned reusable values, never package-global state.

**Test Organization:**
- Unit and behavior tests are co-located beside package code.
- Fixtures stay under the package-specific `testdata/`.
- `flags/testdata/fuzz/FuzzParse/basic.txt` is the initial deterministic clean-room fuzz seed placeholder.
- Redaction tests in `config/`, examples, and any future internal diagnostics package must use the Step 5 architecture-owned sensitive-value corpus.
- Examples are runnable Go example tests under `examples/`; files under `examples/` must contain `Example...` functions where practical, not only ordinary `Test...` functions.
- Dependency enforcement is owned by `tools/depgate/` and CI.
- `tools/depgate/` is approved repository tooling, but it must use only the Go standard library unless this architecture is updated.

**Documentation Organization:**
- `README.md` owns public onboarding: install/import guidance, package overview, minimal usage, release status, and links to deeper docs.
- `docs/clean-room-policy.md` owns clean-room policy and evidence requirements.
- `CONTRIBUTING.md` summarizes contributor obligations and links to the clean-room policy.
- `docs/provenance-log.md` records provenance entries with source, access date, license/terms, artifact affected, and classification as copied, adapted, or inspiration-only.
- `docs/behavior-matrices.md` summarizes cross-package expected behavior; package tests remain the executable contract.
- `docs/config-precedence.md` is the canonical precedence authority and must define the exact order used by tests, examples, diagnostics, and source reports.
- `docs/diagnostics-and-errors.md` owns diagnostic shape vocabulary: boundary, command/flag/key, provenance, typed category, and redacted value status.
- `docs/testing.md` owns local verification, lint, coverage, fuzz, race, dependency-gate, and release-candidate validation guidance.
- Migration docs may mention Go `flag`, pflag, Cobra, and Viper as semantic source concepts, but examples must not imply compatibility adapters.

**Asset Organization:**
Not applicable. Dib V1 has no static assets, UI bundles, generated shell completion, man pages, or binary artifacts.

### Development Workflow Integration

**Development Server Structure:**
Not applicable. Dib has no development server.

**Build Process Structure:**
- `go test ./...` is the primary package, test, and example verification path.
- `go vet ./...` verifies basic Go correctness.
- `go run ./tools/depgate` is the intended local/CI dependency-gate entry point once implemented.
- The dependency gate must inspect all non-tool Go packages included by `go test ./...`, including package tests and `examples/` packages, and fail on any non-standard-library import unless this architecture is updated. Tool packages such as `tools/depgate/` must also remain standard-library-only unless this architecture is updated.
- The lint gate must be pinned, reproducible, and isolated from Dib runtime imports. Story 6.1 owns final linter selection and command wiring.
- The coverage gate must generate package-level evidence and enforce package-aware thresholds. Story 6.2 owns final threshold policy and command wiring.
- `go test -race ./...` is a release-candidate gate.

**Deployment Structure:**
- Dib releases are Go module tags, not binary deployments.
- `.github/workflows/ci.yml` records release-gate checks.
- `docs/release-checklist.md` records test, vet, dependency-gate, lint, package-aware coverage, race-test, examples, public usage docs, provenance, compatibility, and migration evidence.

## Architecture Validation Results

### Coherence Validation

**Decision Compatibility:**
The architecture is coherent. Go 1.26, standard-library-only runtime packages, explicit instance ownership, no package-global runtime state, typed errors, deterministic outputs, and clean-room evidence requirements all reinforce the same product goal: a low-risk, auditable Go CLI library.

**Pattern Consistency:**
The implementation patterns support the architectural decisions. Reusable definitions with per-run snapshots, source/provenance labels, redaction rules, typed errors, and package-local tests give implementation agents consistent rules without requiring hidden shared state.

**Structure Alignment:**
The structure supports the architecture through three public capability packages: `command/`, `flags/`, and `config/`. `internal/` is provisional, `tools/depgate/` is isolated tooling, and docs/examples/testdata are mapped to the PRD requirements.

### Requirements Coverage Validation

**Epic/Feature Coverage:**
No epics were provided, so validation used the PRD functional requirements. Command routing, flag parsing, config resolution, clean-room evidence, compatibility positioning, behavior matrices, dependency enforcement, and parser hardening all have architectural homes.

**Functional Requirements Coverage:**
All FR groups are covered by package boundaries, docs, examples, tests, or release evidence paths.

**Non-Functional Requirements Coverage:**
The architecture covers zero external runtime dependencies, explicit state, deterministic behavior, testability, redaction, clean-room evidence, Go version alignment, and release-gate expectations. NFR-9 is covered for implementation start by the Go 1.26 module baseline plus the requirement that `go.mod`, docs, CI, and release notes stay aligned; final supported-version policy must be explicit in release docs.

### Implementation Readiness Validation

**Decision Completeness:**
Critical decisions are documented. Exact exported API identifiers, final public error identity names, exact CI action versions, and final license wording remain deferred implementation decisions, but they do not block scaffolding or story breakdown.

**Structure Completeness:**
The directory structure is specific enough to guide implementation. Public package boundaries, testdata locations, examples, docs, CI, and dependency-gate tooling are all defined.

**Pattern Completeness:**
The architecture defines naming, ownership, validation, diagnostics, redaction, provenance, testing, clean-room, and dependency rules. Callback invocation is explicitly deferred and must not be added by the first implementation story.

### Gap Analysis Results

**Critical Gaps:**
None.

**Important Deferred Decisions:**
- Exact exported API identifiers and public error identities must be finalized during package story work.
- Exact linter selection, pinning mechanism, and command name must be finalized by Story 6.1.
- Exact coverage threshold values and any tooling-package exception policy must be finalized by Story 6.2.
- Callback invocation behavior is deferred. The first implementation story must not add callback invocation behavior; command routing should start with definitions, route snapshots, validation, and typed errors only.
- `tools/depgate/` must be implemented. The first tooling story must make `go run ./tools/depgate` verify all non-tool Go packages included by `go test ./...`, including package tests and `examples/` packages, and verify tool packages remain standard-library-only unless the architecture is updated.
- Final CI action versions, SHA pinning strategy, license wording, and release-doc support policy remain follow-up decisions.

**Nice-to-Have Gaps:**
- Benchmark targets and allocation budgets may be added after behavior stabilizes.
- Additional migration examples can be added after the initial public package APIs settle.

### Validation Issues Addressed

- Party Mode review tightened wording that overstated readiness.
- Technology and performance checklist items were narrowed to match the actual architecture.
- Dependency-gate scope was clarified to include tests and `examples/`.
- Release evidence scope was validated: release checklist must record exact commit, test, vet, dependency-gate, lint, package-aware coverage, race-test, docs/examples, provenance, compatibility, and migration evidence.
- `docs/provenance-log.md` must be created or updated when any copied, adapted, generated, or inspiration-only reference-derived artifact is introduced.
- Correct-course update on 2026-06-12 added lint, package-aware coverage validation, public README/usage docs, and tracker reconciliation as release-hardening scope.

### Architecture Completeness Checklist

**Requirements Analysis**

- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed
- [x] Technical constraints identified
- [x] Cross-cutting concerns mapped

**Architectural Decisions**

- [x] Critical decisions documented with versions
- [x] Technology baseline and constraints specified
- [x] Integration patterns defined
- [x] Performance-sensitive constraints and future evidence needs identified

**Implementation Patterns**

- [x] Naming conventions established
- [x] Structure patterns defined
- [x] Communication patterns specified
- [x] Process patterns documented

**Project Structure**

- [x] Complete directory structure defined
- [x] Component boundaries established
- [x] Integration points mapped
- [x] Requirements to structure mapping complete

### Architecture Readiness Assessment

**Overall Status:** READY WITH MINOR GAPS

**Confidence Level:** high

Readiness means ready for module scaffolding, story breakdown, and first implementation work; it does not mean release readiness or final public API freeze.

**Key Strengths:**
- Strong alignment between product trust goals and technical constraints.
- Clear public package boundaries with no root facade or hidden singleton API.
- Explicit clean-room, provenance, dependency, and release-evidence requirements.
- Testability and deterministic behavior are built into the architecture rather than left to implementation preference.

**Areas for Future Enhancement:**
- Final exported API names and public error identities.
- First implementation of `tools/depgate/`.
- CI action pinning and final release policy details.
- Benchmarks after core behavior stabilizes.

### Implementation Handoff

**AI Agent Guidelines:**

- Follow all architectural decisions exactly as documented.
- Use implementation patterns consistently across all components.
- Respect project structure and package boundaries.
- Do not add callback invocation behavior until a later architecture/API decision explicitly approves it.
- Refer to this document for all architectural questions.

**First Implementation Priority:**
Initialize the module and create the first architecture-proof scaffold:
- `go.mod` with Go 1.26.
- Initial `command/`, `flags/`, and `config/` package docs, or one package selected by the first story.
- One minimal user-facing behavior from a selected package, such as defining and validating a command or flag shape, so examples and tests demonstrate actual toolkit value rather than empty structure.
- At least one standard-library-only package test.
- At least one runnable `Example...` test where practical.
- Initial `tools/depgate/` entry point or documented temporary dependency-gate command covering non-tool packages included by `go test ./...`.
- Initial clean-room policy/provenance docs, including `docs/provenance-log.md`.

Minimum verification for the first scaffold:

```bash
go test ./...
go vet ./...
```

Dependency verification must use one of:

```bash
go run ./tools/depgate
```

or, until `tools/depgate/` exists, the documented temporary `go list` command from this architecture.

Dependency verification must use `go run ./tools/depgate` once `tools/depgate/` exists. The temporary `go list` dependency check is allowed only for the initial scaffold story and must not be used as release-candidate evidence. `tools/depgate/` remains required before any release candidate.

`tools/depgate/` must enforce zero external imports for all library, test, and example packages, and zero external imports for tool packages unless the architecture is updated.

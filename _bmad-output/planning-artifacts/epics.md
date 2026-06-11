---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - "_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md"
  - "_bmad-output/planning-artifacts/architecture.md"
status: complete
completedAt: "2026-06-11"
---

# dib - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for dib, decomposing the requirements from the PRD, UX Design if it exists, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: Define Command trees. Developers can define a root Command and nested child Commands with stable names, descriptions, aliases, and usage metadata.

FR2: Execute Commands explicitly. Developers can execute a Command tree through an explicit API that accepts arguments, output streams, and `context.Context` where execution crosses boundaries.

FR3: Support local and inherited flags. Developers can attach flags to a Command and define inherited flags that apply to descendants.

FR4: Generate deterministic help and usage text. Developers can render help and usage text for a Command tree using supplied writers.

FR5: Define independent Flag sets. Developers can define and parse independent Flag sets without relying on package-level mutable state.

FR6: Parse long and shorthand flags. Developers can parse POSIX/GNU-style long flags and one-character shorthand flags.

FR7: Parse shorthand groups and no-option defaults. Developers can parse shorthand groups and no-option defaults through documented rules.

FR8: Support repeated and custom values. Developers can define repeated flags and custom values using small interfaces.

FR9: Control parse boundaries and diagnostics. Developers can control how Flag parsing treats non-flag arguments, the `--` terminator, and parse errors.

FR10: Normalize names intentionally. Developers can configure name normalization for flags and config bindings where supported.

FR11: Register Config keys and defaults. Developers can register Config keys with default values, type expectations, and optional documentation.

FR12: Resolve values by precedence. Developers can resolve Config values using a stable precedence model.

FR13: Bind flags to Config keys. Developers can bind parsed Flag values to Config keys.

FR14: Bind environment variables to Config keys. Developers can bind environment variables to registered Config keys using explicit names or a configured prefix and replacer.

FR15: Load JSON configuration from paths and readers. Developers can load JSON configuration from a filesystem path or an `io.Reader`.

FR16: Retrieve typed Config values. Developers can retrieve Config values through typed getters and existence checks.

FR17: Publish clean-room source policy. Maintainers can point reviewers to a clean-room policy that defines allowed and disallowed source inputs.

FR18: Document compatibility boundaries. Developers can see which Go `flag`, pflag, Cobra, and Viper behaviors Dib supports, narrows, omits, or intentionally changes.

FR19: Provide migration examples. Developers can follow examples from standard `flag`, pflag-style flags, Cobra-style command trees, and Viper-style config resolution.

FR20: Provide behavior test matrices. Maintainers can validate parser, command, and config behavior through table-driven tests.

FR21: Enforce the runtime dependency rule. Maintainers can verify that runtime packages import only the Go standard library.

FR22: Support fuzz or property-style parser hardening. Maintainers can harden parsers against edge cases without changing the runtime dependency contract.

### NonFunctional Requirements

NFR1: Runtime packages must import only the Go standard library.

NFR2: Primary APIs must operate on explicit instances and caller-supplied inputs/outputs. V1 does not include package-level global command, flag, or config helpers.

NFR3: Public error cases needed by callers must be inspectable without string matching.

NFR4: Help, usage, and diagnostics must be deterministic enough for stable golden or snapshot tests.

NFR5: Library APIs must not call `os.Exit`, mutate process-wide streams, or read `os.Args` unless the caller chooses a convenience path documented to do so.

NFR6: Core behavior must be testable with table-driven unit tests and injected readers, writers, args, and environment lookup.

NFR7: Familiar behavior must be documented as supported, narrowed, omitted, or intentionally different.

NFR8: Error messages must identify bad keys, flags, and sources without dumping sensitive values when a Flag or Config key is marked sensitive.

NFR9: Dib V1 requires Go 1.26 or newer.

NFR10: Public API changes after V1 must follow semantic versioning and include deprecation guidance for at least one minor release before removal where practical.

### Additional Requirements

- Use no external starter template. The implementation foundation is a standard Go module bootstrap for `github.com/petabytecl/dib` with `go 1.26`; do not add a `toolchain` directive unless a later architecture decision chooses patch-level toolchain pinning.
- Initialize the module and repository scaffold before feature work, including Go package documentation, at least one real package behavior, tests, examples where practical, dependency-gate evidence, clean-room policy, and provenance docs.
- Organize runtime library code around three public capability packages: `command/`, `flags/`, and `config/`. Do not introduce a broad root facade, default singleton API, package-global helpers, or `/cmd` application scaffold in V1.
- Keep command routing, flag parsing, and config resolution independently usable. `flags/` must work without `command/` or `config/`; `command/` must not depend on `config/`; callers compose the three surfaces explicitly.
- Treat definitions as caller-observably immutable reusable values. Derived definitions must return new values, avoid exported mutable internals or shallow-copy aliasing, and remain safe to reuse across repeated or concurrent runs.
- Return per-run snapshots for command routing, flag parsing, and config resolution. Snapshots must not mutate definitions and must not depend on live process state, environment variables, readers, or lookup functions after creation.
- Validate invalid definitions, duplicate names, binding collisions, normalization collisions, invalid relationships, and unsupported config combinations at setup boundaries where possible.
- Treat CLI args, env values, JSON config, custom parsers, readers, and lookup functions as untrusted boundary inputs. Errors and diagnostics must be typed, deterministic, and redaction-safe.
- Expose public APIs using standard Go conventions: explicit inputs, returned values and errors, `context.Context` where execution crosses a boundary, and `io.Reader` / `io.Writer` where callers provide input/output.
- Public errors must support `errors.Is` and `errors.As`. Error strings are diagnostics only and must not be the programmatic contract.
- Preserve machine-readable error category, command/flag/key/provenance context, wrapping behavior, and sensitive-value redaction across command, flag, config, JSON, and conversion failures.
- Use the config provenance/source-label vocabulary consistently: `default`, `explicit setter`, `flag binding`, `env`, and `JSON`.
- Follow the PRD config precedence exactly: explicit setter, parsed flag, environment variable, JSON file, default.
- Implement JSON as the only V1 config file format, with strict mode as the documented default for registered-key loads and permissive mode as opt-in.
- Mark sensitive values for diagnostics redaction only; Dib is not a secret manager. Raw sensitive values must not appear in errors, string output, debug strings, rendered diagnostics, source reports, examples, or validation failures.
- Use the architecture-owned fake sensitive-value corpus in tests: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
- Keep help, usage, diagnostics, source reports, behavior matrices, validation output, examples, and ordering rules deterministic and testable.
- Store tests beside the package under test. Keep fixtures small, local, deterministic, clean-room, and under package-specific `testdata/` directories.
- Use standard-library-only tests and runnable examples unless the architecture is updated. Examples should be Go example tests under `examples/` where practical.
- Add parser fuzzing with standard Go fuzzing support and clean-room deterministic seed corpus under `flags/testdata/fuzz/FuzzParse/`.
- Create and maintain clean-room and provenance documentation: `docs/clean-room-policy.md`, `docs/provenance-log.md`, `CONTRIBUTING.md`, compatibility docs, behavior matrices, diagnostics/errors docs, config precedence docs, testing docs, and release checklist.
- Record provenance for copied, generated, adapted, or inspiration-only reference-derived artifacts, including source, access date, license/terms, affected artifact, and classification.
- Implement CI with GitHub Actions using an explicit runner image such as `ubuntu-24.04`, official actions, and Go version alignment across `go.mod`, docs, CI, and release notes.
- Core PR gates must include `go test ./...`, `go vet ./...`, and the dependency gate. Release-candidate gates additionally include `go test -race ./...`.
- Implement `tools/depgate/` as isolated repository tooling. Once it exists, dependency verification must use `go run ./tools/depgate`.
- The dependency gate must verify zero external imports for all library, test, and example packages, and zero external imports for tool packages unless the architecture is updated.
- Until `tools/depgate/` exists, the temporary architecture-approved `go list` dependency check may be used only for the initial scaffold story and not as release-candidate evidence.
- Dib releases are Go module tags, not binary deployments. v0 tags may include breaking changes but require release notes, migration guidance, updated examples, provenance notes, and passing CI evidence tied to the exact tagged commit.
- Do not add callback invocation behavior in the first implementation work. Callback invocation is deferred until a later architecture/API decision explicitly approves that surface.
- Keep source files focused and small. Shared helpers should remain unexported until at least two concrete call sites prove the abstraction is needed.

### UX Design Requirements

No UX design document was discovered in the configured planning artifacts, so no UX-DR requirements were extracted for V1.

### FR Coverage Map

FR1: Epic 3 - Command tree definitions.

FR2: Epic 3 - Explicit command execution inputs and errors.

FR3: Epic 3 - Local and inherited command flags.

FR4: Epic 3 - Deterministic help and usage.

FR5: Epic 2 - Independent Flag sets.

FR6: Epic 2 - Long and shorthand flag parsing.

FR7: Epic 2 - Shorthand groups and no-option defaults.

FR8: Epic 2 - Repeated and custom values.

FR9: Epic 2 - Parse boundaries and diagnostics.

FR10: Epic 2 - Intentional name normalization.

FR11: Epic 4 - Config key registration and defaults.

FR12: Epic 4 - Precedence-based resolution.

FR13: Epic 4 - Flag-to-config bindings.

FR14: Epic 4 - Environment bindings.

FR15: Epic 4 - JSON config loading.

FR16: Epic 4 - Typed Config retrieval.

FR17: Epic 1 - Clean-room policy.

FR18: Epic 5 - Compatibility boundaries.

FR19: Epic 5 - Migration examples.

FR20: Epics 1-5 - Behavior matrices and tests across all surfaces.

FR21: Epics 1 and 5 - Dependency rule enforcement and release evidence.

FR22: Epic 2 - Parser fuzz/property hardening.

## Epic List

### Epic 1: Auditable Toolkit Foundation

Go developers and reviewers can install, inspect, test, and trust the initial Dib module as a standard-library-only Go library with clear clean-room and dependency evidence.

**FRs covered:** FR17, FR20, FR21

### Epic 2: Inspectable Flag Parsing

Developers can define explicit Flag sets and parse familiar long flags, shorthand flags, shorthand groups, repeated/custom values, terminators, and typed parse errors without package-global state.

**FRs covered:** FR5, FR6, FR7, FR8, FR9, FR10, FR20, FR22

### Epic 3: Composable Command Routing

Developers can define nested command trees with aliases, local/inherited flags, explicit execution inputs, deterministic help/usage output, and typed command errors.

**FRs covered:** FR1, FR2, FR3, FR4, FR20

### Epic 4: Provenance-Aware Config Resolution

Developers can register Config keys, resolve values by documented precedence, bind flags/env/JSON sources, retrieve typed values, inspect provenance, and protect sensitive diagnostics.

**FRs covered:** FR11, FR12, FR13, FR14, FR15, FR16, FR20

### Epic 5: Migration, Compatibility, And Release Evidence

Developers and reviewers can understand Dib's compatibility boundaries, follow migration examples, and verify release readiness with behavior matrices, dependency checks, docs, examples, and provenance evidence.

**FRs covered:** FR18, FR19, FR20, FR21

## Epic 1: Auditable Toolkit Foundation

Go developers and reviewers can install, inspect, test, and trust the initial Dib module as a standard-library-only Go library with clear clean-room and dependency evidence.

### Story 1.1: Adopt an Auditable Go Module Baseline

**Requirements:** FR20, FR21

As a Go library adopter,
I want Dib to start as a minimal, standard-library-only Go module with visible package boundaries,
So that I can inspect the foundation before trusting later command, flag, and config behavior.

**Acceptance Criteria:**

**Given** a fresh checkout without an initialized Go module
**When** the module baseline is created
**Then** `go.mod` declares module `github.com/petabytecl/dib` with Go 1.26
**And** no `toolchain` directive is added unless the architecture is updated.

**Given** the architecture-defined package boundaries
**When** the initial source tree is created
**Then** `command/`, `flags/`, and `config/` each contain package documentation or a minimal compilable package file
**And** the repository does not introduce a root facade package, package-global command/flag/config helpers, or a `/cmd` scaffold.

**Given** the baseline module exists
**When** verification runs
**Then** `go test ./...` and `go vet ./...` pass
**And** at least one standard-library-only package test or runnable example proves the module is not only empty structure.

**Given** later stories will add behavior to reusable definitions and snapshots
**When** baseline docs or package comments describe the implementation direction
**Then** they state that Dib favors explicit instances, caller-owned inputs, immutable definitions, per-run snapshots, typed errors, and no ambient process state.

### Story 1.2: Publish the Clean-Room Contribution Contract

**Requirements:** FR17, FR21

As a technical reviewer,
I want Dib's clean-room policy and provenance expectations documented before feature work expands,
So that I can verify the project is not copying source, tests, examples, fixtures, names, or structure from inspiration projects.

**Acceptance Criteria:**

**Given** Dib is inspired by public behavior from Go `flag`, pflag, Cobra, and Viper
**When** the clean-room policy is written
**Then** `docs/clean-room-policy.md` defines allowed inputs as public documentation and observable behavior
**And** it defines disallowed inputs as copied source, tests, comments, examples, internal names, fixtures, file organization, and source-derived generated content.

**Given** contributors need a short operational contract
**When** contribution guidance is written
**Then** `CONTRIBUTING.md` links to the clean-room policy
**And** it states that compatibility examples, tests, docs, and fixtures must be independently written or explicitly recorded with provenance.

**Given** future stories may use public documentation or other references
**When** `docs/provenance-log.md` is created
**Then** it provides a repeatable entry format with source, access date, license or terms, affected artifact, and classification as copied, adapted, or inspiration-only
**And** it includes initial inspiration-only entries for the PRD/architecture-approved public documentation sources where appropriate.

**Given** clean-room compliance is part of the adoption promise
**When** verification runs
**Then** `go test ./...` still passes
**And** the documentation states that provenance gaps block acceptance for copied, generated, adapted, or reference-derived artifacts.

### Story 1.3: Establish Cross-Surface Behavior Contracts

**Requirements:** FR20

As a Dib implementer,
I want the shared behavior contracts for definitions, snapshots, typed errors, provenance, and redaction established early,
So that `command/`, `flags/`, and `config/` do not drift as separate implementations grow.

**Acceptance Criteria:**

**Given** the architecture requires caller-observably immutable definitions
**When** the shared behavior contract is documented or represented in initial package-level tests
**Then** derived definitions must return new values rather than mutating existing values
**And** callers must not observe mutable aliases to slices, maps, readers, env lookups, or config data after construction or resolution.

**Given** the architecture requires per-run snapshots
**When** a package parses, routes, or resolves input in later stories
**Then** the expected contract states that snapshots never write back to definitions
**And** snapshots do not depend on live process state, environment variables, readers, or lookup functions after creation.

**Given** public errors are a cross-package contract
**When** initial error contract guidance or test scaffolding is created
**Then** errors must be inspectable with `errors.Is` or `errors.As` where callers need programmatic handling
**And** string output is documented as diagnostics rather than the programmatic contract.

**Given** config provenance and redaction rules affect diagnostics across packages
**When** shared test guidance or fixtures are created
**Then** the source-label vocabulary is fixed to `default`, `explicit setter`, `flag binding`, `env`, and `JSON`
**And** the fake sensitive-value corpus is fixed to `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.

**Given** behavior contracts are not just prose
**When** verification runs
**Then** at least one initial standard-library-only test or example demonstrates an explicit-instance, no-ambient-state contract
**And** the story leaves clear acceptance hooks for later package stories to add immutability, snapshot, typed-error, provenance, and redaction assertions.

### Story 1.4: Enforce the Standard-Library Dependency Gate

**Requirements:** FR21

As a dependency-policy reviewer,
I want a repeatable repository gate that fails on non-standard-library imports,
So that Dib's zero-runtime-dependency claim is enforced before feature implementation scales.

**Acceptance Criteria:**

**Given** Dib's runtime, tests, examples, and repository tooling must remain standard-library-only unless the architecture changes
**When** `tools/depgate/` is implemented
**Then** `go run ./tools/depgate` inspects all non-tool packages included by `go test ./...`, including package tests and `examples/` packages
**And** it fails when any inspected package imports a non-standard-library package.

**Given** `tools/depgate/` is repository tooling rather than runtime library code
**When** the dependency gate inspects tool packages
**Then** it also verifies tool packages use only the Go standard library unless the architecture is updated
**And** it does not create an import path from runtime packages into `tools/depgate/`.

**Given** dependency failures need to be actionable
**When** the gate finds a non-standard-library import
**Then** the diagnostic identifies the package and offending import path
**And** it exits non-zero without hiding other detected dependency violations.

**Given** the first scaffold may temporarily use the architecture-approved `go list` command
**When** `tools/depgate/` exists
**Then** documentation and CI guidance make `go run ./tools/depgate` the required dependency check
**And** the temporary `go list` command is no longer accepted as release-candidate evidence.

**Given** the gate protects FR21 and NFR1
**When** verification runs
**Then** `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` pass
**And** tests for `tools/depgate/` cover at least a passing stdlib-only fixture and a failing non-stdlib import fixture.

### Story 1.5: Run Trust Gates in CI

**Requirements:** FR20, FR21

As a technical reviewer,
I want Dib's core trust gates to run automatically in CI,
So that standard-library-only dependency enforcement, tests, vet, and clean-room evidence do not rely on manual discipline.

**Acceptance Criteria:**

**Given** Dib uses GitHub Actions for repository verification
**When** `.github/workflows/ci.yml` is created
**Then** it runs on pull requests and pushes to the default development branch
**And** it uses an explicit GitHub-hosted runner image such as `ubuntu-24.04`.

**Given** the project targets Go 1.26
**When** CI config installs Go
**Then** the configured Go version matches `go.mod`, docs, and release guidance
**And** version drift is called out as a release-blocking issue.

**Given** core PR gates are defined by architecture
**When** CI runs
**Then** it executes `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`
**And** failures from any gate block the check.

**Given** release-candidate evidence will be consolidated later
**When** CI and docs are updated
**Then** `docs/release-checklist.md` records that `go test -race ./...` is a release-candidate gate
**And** the checklist has placeholders for exact commit, test, vet, dependency-gate, race-test, docs/examples, provenance, compatibility, and migration evidence.

**Given** CI is part of Dib's adopter trust story
**When** verification runs locally
**Then** the same commands used in CI pass locally
**And** the workflow avoids non-standard-library project dependencies or generated scaffolding that would weaken the dependency claim.

## Epic 2: Inspectable Flag Parsing

Developers can define explicit Flag sets and parse familiar long flags, shorthand flags, shorthand groups, repeated/custom values, terminators, and typed parse errors without package-global state.

### Story 2.1: Define Reusable Flag Sets Without Global State

**Requirements:** FR5, FR8, FR9, FR20

As a Go CLI developer,
I want reusable Flag sets with explicit definitions and value metadata,
So that I can parse CLI input without package-global mutable state or hidden process dependencies.

**Acceptance Criteria:**

**Given** a caller defines a Flag set
**When** flags are registered
**Then** each definition captures name, optional shorthand, default value, usage text, value parser, repeat policy, hidden/deprecated metadata, sensitivity metadata, and no-option default where applicable
**And** built-in value kinds include string, bool, int, int64, uint, uint64, float64, duration, and string list.

**Given** two independent Flag sets use the same flag names
**When** each Flag set is parsed or inspected
**Then** their definitions and parse snapshots remain independent
**And** no package-level global registry, default Flag set, or ambient `os.Args` dependency is introduced.

**Given** later parser behavior depends on typed values and diagnostics
**When** the initial value and diagnostic model is implemented
**Then** value arity, default handling, explicit-set tracking, duplicate detection, and diagnostic categories are represented in machine-readable state
**And** public errors are inspectable through `errors.Is` or `errors.As` where callers need programmatic handling.

**Given** definitions are reusable values
**When** callers derive or extend a Flag set
**Then** the original Flag set keeps unchanged observable behavior
**And** tests prove no caller-observable mutation or slice/map alias leak across repeated parses.

**Given** this story establishes the parser foundation
**When** verification runs
**Then** table-driven tests cover definition validation, duplicate long names, duplicate shorthands, explicit-set tracking, default values, and reusable definitions
**And** `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` pass.

### Story 2.2: Match Flag Names Safely Across Styles

**Requirements:** FR10, FR20

As a Go CLI developer,
I want flag names to match exactly by default and normalize only when I opt in,
So that familiar naming styles do not create silent collisions or surprising parse behavior.

**Acceptance Criteria:**

**Given** no normalizer is configured
**When** flags named `log-level`, `log_level`, and `log.level` are registered or parsed
**Then** exact names are used
**And** no implicit equivalence is applied.

**Given** a caller configures a name normalizer
**When** equivalent names such as `log-level`, `log_level`, and `log.level` normalize to the same key
**Then** the Flag set can resolve those names consistently
**And** the parse snapshot reports the canonical definition rather than the raw spelling alone.

**Given** normalization creates a definition collision
**When** the Flag set is built or derived
**Then** setup fails with a typed deterministic error
**And** the diagnostic identifies the colliding flag names without relying on string-only matching.

**Given** shorthand names are one-character identities
**When** name normalization is configured
**Then** shorthand uniqueness remains independently enforced
**And** long-name normalization never creates hidden shorthand aliases.

**Given** later command and config stories depend on stable flag names
**When** verification runs
**Then** table-driven tests cover exact matching, configured normalization, normalization collisions, shorthand uniqueness, and diagnostic context
**And** no runtime or test dependency outside the Go standard library is introduced.

### Story 2.3: Reject Invalid Long Flags With Inspectable Errors

**Requirements:** FR6, FR9, FR10, FR20

As a Go CLI developer,
I want long flags to parse familiar forms and reject invalid input with typed diagnostics,
So that scripts and tests can handle parser failures without scraping error text.

**Acceptance Criteria:**

**Given** a known long flag accepts a value
**When** input uses `--name=value` or `--name value`
**Then** the parsed snapshot records the explicit value, source spelling, and canonical flag definition
**And** remaining positional arguments preserve their relative order.

**Given** a known boolean long flag
**When** input uses `--name`, `--name=true`, or `--name=false`
**Then** the parsed snapshot records the expected boolean value
**And** invalid boolean text returns a typed conversion parse error.

**Given** an unknown long flag is provided before `--`
**When** parsing runs
**Then** parsing returns a typed unknown-flag diagnostic
**And** the diagnostic includes the flag token and normalized/canonical lookup context where applicable.

**Given** a long flag requires a value
**When** input omits the value or the next argument cannot be consumed as a value
**Then** parsing returns a typed missing-value diagnostic
**And** the failed parse does not mutate the reusable Flag set.

**Given** long flag behavior is a public contract
**When** verification runs
**Then** table-driven tests cover attached values, separate values, booleans, unknown flags, missing values, invalid conversions, duplicate single-value flags, and exact/normalized names
**And** diagnostics are asserted through typed errors and snapshot state, not only rendered strings.

### Story 2.4: Parse Short Flags And Boolean Presence

**Requirements:** FR6, FR9, FR20

As a Go CLI developer,
I want one-character short flags to parse predictably,
So that familiar CLI shorthand behavior works without importing pflag or a larger framework.

**Acceptance Criteria:**

**Given** a known shorthand flag accepts a value
**When** input uses `-n value` or `-n=value`
**Then** the parsed snapshot records the explicit value and canonical flag definition
**And** positional arguments remain in deterministic order.

**Given** a known boolean shorthand flag
**When** input uses `-v`
**Then** the parsed snapshot records the boolean as explicitly set
**And** default values for other flags do not appear as explicit CLI input.

**Given** a shorthand is unknown
**When** parsing reaches that token before `--`
**Then** parsing returns a typed unknown-shorthand diagnostic
**And** the diagnostic identifies the failing shorthand character.

**Given** a shorthand requires a value
**When** the value is missing or invalid
**Then** parsing returns a typed missing-value or conversion diagnostic
**And** no package-global parser state is mutated.

**Given** shorthand behavior must remain independent from long-name normalization
**When** verification runs
**Then** tests cover valid short flags, boolean presence, separate values, equals-attached values, unknown shorthand, missing values, invalid conversions, and shorthand uniqueness
**And** `go test ./...` and `go run ./tools/depgate` pass.

### Story 2.5: Handle Short Flag Groups And Optional Values Predictably

**Requirements:** FR7, FR9, FR20, FR22

As a Go CLI developer,
I want grouped short flags and optional values to follow documented rules,
So that compact CLI input remains testable and failures identify the exact shorthand that failed.

**Acceptance Criteria:**

**Given** a shorthand group contains only boolean flags
**When** input uses a token such as `-abc`
**Then** parsing sets `-a`, `-b`, and `-c` in order
**And** the snapshot records each flag as explicitly set by CLI input.

**Given** a shorthand group contains a non-boolean flag as the final member
**When** input uses forms such as `-ab10` or `-ab 10`
**Then** preceding boolean shorthands are set
**And** the final non-boolean shorthand consumes the attached or next value according to the documented rules.

**Given** a non-boolean shorthand appears before the end of a group
**When** that flag has a no-option default
**Then** parsing applies the no-option default and continues through the group
**And** the snapshot distinguishes no-option default use from ordinary configured defaults.

**Given** a non-boolean shorthand appears before the end of a group without a no-option default
**When** parsing reaches that shorthand
**Then** parsing returns a typed invalid-group diagnostic
**And** the diagnostic identifies the failing shorthand and token.

**Given** grouped shorthand behavior has high ambiguity risk
**When** verification runs
**Then** table-driven tests cover boolean groups, final-value groups, no-option defaults, invalid groups, unknown members, invalid conversions, and partial-failure snapshot behavior
**And** targeted parser fuzz/property tests prove grouped input does not panic and preserves deterministic diagnostics.

### Story 2.6: Accumulate Repeated And Custom Values

**Requirements:** FR8, FR9, FR20

As a Go CLI developer,
I want repeated and custom flag values to be explicit and inspectable,
So that advanced flag behavior remains small, testable, and standard-library-only.

**Acceptance Criteria:**

**Given** a flag is configured for repeated values
**When** CLI input provides the flag multiple times
**Then** the parsed snapshot accumulates values in command-line order
**And** provenance for each value remains available enough for diagnostics and behavior tests.

**Given** a single-value flag is repeated by CLI input
**When** parsing reaches the duplicate value
**Then** parsing returns a typed duplicate-value diagnostic
**And** the diagnostic identifies the flag and duplicate source token.

**Given** a caller provides a custom value parser
**When** parsing succeeds
**Then** the snapshot stores the parsed value through the public value contract
**And** the reusable Flag set remains safe to parse again.

**Given** a custom value parser returns an error
**When** parsing fails
**Then** Dib preserves caller inspection through wrapping or typed context
**And** diagnostics redact sensitive values when the flag definition is marked sensitive.

**Given** repeated and custom values extend the earlier value model
**When** verification runs
**Then** tests cover valid accumulation, duplicate rejection, custom parser success, custom parser failure, redaction, and immutable definition reuse
**And** no third-party parser or assertion library is introduced.

### Story 2.7: Preserve Parse Boundaries And Remaining Args

**Requirements:** FR9, FR20

As a Go CLI developer,
I want parser boundaries and remaining arguments preserved exactly,
So that my application can safely compose flag parsing with positional commands, passthrough arguments, and tests.

**Acceptance Criteria:**

**Given** positional arguments appear before `--`
**When** flags are interspersed with positionals
**Then** flags before `--` remain parseable
**And** positional arguments keep relative order in the remaining-args result.

**Given** input contains the `--` terminator
**When** parsing reaches the terminator
**Then** flag parsing stops
**And** every subsequent argument remains untouched even if it looks like a flag.

**Given** a help request is encountered
**When** parsing identifies it
**Then** Dib returns a typed help-request result or error for caller-controlled rendering and exit policy
**And** no runtime path calls `os.Exit`.

**Given** parse failures happen before `--`
**When** unknown flags, missing values, invalid values, or invalid groups are encountered
**Then** the returned diagnostic is typed and deterministic
**And** the remaining-args behavior is covered by tests for both successful and failed parses.

**Given** parse boundaries feed command routing and config binding
**When** verification runs
**Then** table-driven tests cover interspersed positionals, `--`, passthrough args, help requests, failed parses, and deterministic snapshot state
**And** targeted fuzz/property tests cover boundary tokens and remaining-arg preservation.

### Story 2.8: Prove Flag Parsing Across Matrices And Fuzz Inputs

**Requirements:** FR5, FR6, FR7, FR8, FR9, FR10, FR20, FR22

As a technical reviewer,
I want parser behavior proven across documented matrices and fuzz inputs,
So that flag parsing can be trusted before command routing and config binding depend on it.

**Acceptance Criteria:**

**Given** Stories 2.1 through 2.7 define parser behavior
**When** `docs/behavior-matrices.md` or an equivalent parser matrix artifact is updated
**Then** it covers definitions, normalization, long flags, shorthand flags, shorthand groups, repeated values, custom values, no-option defaults, parse boundaries, help requests, and diagnostics
**And** each matrix row traces back to relevant FRs and tests.

**Given** parser tests are executable evidence
**When** verification runs
**Then** table-driven tests cover valid, invalid, ambiguous, duplicate, boundary, and remaining-arg cases
**And** typed diagnostics are asserted with `errors.Is` or `errors.As` where caller inspection is required.

**Given** parser fuzzing must use only standard Go support
**When** fuzz targets and seed corpus files are added
**Then** they live under the package-specific `testdata/fuzz/` flow
**And** fuzzing proves parser inputs do not panic, mutate reusable definitions, or produce nondeterministic boundary behavior.

**Given** parser evidence must not weaken dependency claims
**When** parser examples, fixtures, fuzz seeds, or docs are added
**Then** clean-room provenance is recorded where required
**And** `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` pass.

## Epic 3: Composable Command Routing

Developers can define nested command trees with aliases, local/inherited flags, explicit execution inputs, deterministic help/usage output, and typed command errors.

### Story 3.1: Route Nested Commands With Inspectable Results

**Requirements:** FR1, FR20

As a Go CLI developer,
I want nested command input to return inspectable routing results,
So that I can build multi-command CLIs without giving a framework control over process lifecycle.

**Acceptance Criteria:**

**Given** a root command has nested children
**When** input such as `deploy apply` is routed
**Then** the route snapshot identifies the canonical matched command path
**And** remaining args are preserved according to the flag parser boundary rules from Epic 2.

**Given** command definitions include names, descriptions, aliases, and usage metadata
**When** command definitions are derived or reused
**Then** original definitions keep unchanged observable behavior
**And** snapshots do not mutate command definitions.

**Given** routing fails to find a command
**When** unknown command input is routed
**Then** Dib returns a typed unknown-command error
**And** no runtime path calls `os.Exit`, writes directly to process streams, or reads `os.Args`.

**Given** route results are public contracts
**When** verification runs
**Then** tests cover root routing, nested routing, remaining args, unknown commands, immutable definition reuse, and deterministic route snapshots
**And** `command/` consumes `flags/` contracts without reimplementing flag syntax.

### Story 3.2: Resolve Aliases And Unknown Commands Predictably

**Requirements:** FR1, FR20

As a Go CLI developer,
I want command aliases and command failures to resolve predictably,
So that users can use shortcuts while tests still assert canonical command behavior.

**Acceptance Criteria:**

**Given** a command defines aliases
**When** input uses an alias
**Then** routing resolves to the intended command
**And** the route snapshot preserves the canonical command name and the raw alias token.

**Given** aliases can collide with command names or other aliases
**When** a command tree is built or derived
**Then** collisions fail during setup with typed deterministic diagnostics
**And** alias-command collisions, alias-alias collisions, and alias cycles are covered by tests.

**Given** unknown command input is supplied near valid commands or aliases
**When** routing fails
**Then** the typed unknown-command diagnostic identifies the failing token and matched parent path
**And** diagnostics remain deterministic without string-only assertions.

**Given** alias support must not introduce hidden state
**When** verification runs
**Then** repeated and concurrent route tests observe stable results from the same reusable command definitions
**And** no root facade, global command registry, or `/cmd` scaffold is introduced.

### Story 3.3: Apply Local And Inherited Flags Predictably During Routing

**Requirements:** FR3, FR20

As a Go CLI developer,
I want command-local and inherited flags to compose predictably during routing,
So that shared CLI options work for child commands without leaking into unrelated siblings.

**Acceptance Criteria:**

**Given** a root command defines inherited flags
**When** a descendant command is routed
**Then** the route snapshot exposes inherited flags available to that descendant
**And** siblings do not receive local flags owned by another command.

**Given** a child command defines local flags
**When** local and inherited flag definitions are combined
**Then** name, shorthand, and normalization conflicts produce deterministic typed setup diagnostics
**And** inherited/local shadowing behavior is explicitly tested.

**Given** command routing consumes parser behavior from Epic 2
**When** flags and positional command tokens are interspersed
**Then** command routing preserves parser boundary behavior and remaining args
**And** `command/` does not reinterpret flag syntax outside exported `flags/` contracts.

**Given** flag composition is the highest-risk command story
**When** verification runs
**Then** tests cover inherited flags, local flags, sibling isolation, conflict diagnostics, normalization collisions, command/flag ambiguity, and immutable route snapshots
**And** `go test ./...` and `go run ./tools/depgate` pass.

### Story 3.4: Render Deterministic Command Help And Usage

**Requirements:** FR4, FR20

As a CLI author,
I want deterministic help and usage output generated from definitions,
So that user-facing text is stable enough for review, examples, and tests.

**Acceptance Criteria:**

**Given** command definitions include names, aliases, descriptions, argument metadata, and visible flags
**When** help or usage is rendered to a caller-supplied writer
**Then** output includes those elements in deterministic order
**And** hidden flags do not appear while remaining parseable when their definitions allow it.

**Given** flags may be deprecated
**When** help or usage includes deprecated visible flags
**Then** a deterministic deprecation note is rendered
**And** the note does not leak sensitive default values.

**Given** rendering is human-facing but still contractual
**When** tests assert help and usage output
**Then** golden tests may verify formatting
**And** structured behavior is still asserted through definitions, snapshots, and typed diagnostics.

**Given** rendering must not execute commands
**When** help or usage is requested
**Then** no callback invocation occurs
**And** no process-global stdout/stderr or `os.Exit` behavior is used.

**Given** documentation and examples depend on stable text
**When** verification runs
**Then** command help/usage tests cover ordering, aliases, hidden flags, deprecated flags, inherited/local flags, and redaction-safe diagnostics
**And** all examples remain standard-library-only.

### Story 3.5: Preserve Caller-Controlled Execution Boundaries

**Requirements:** FR2, FR20

As a Go CLI developer,
I want command routing to preserve caller control over execution inputs and errors,
So that Dib can support execution-oriented CLIs without owning process lifecycle.

**Acceptance Criteria:**

**Given** a caller routes command input with explicit args, writers, and context metadata
**When** routing or execution-boundary APIs are used
**Then** Dib returns route/result snapshots and typed errors to the caller
**And** it does not read `os.Args`, mutate `os.Stdout` or `os.Stderr`, or call `os.Exit`.

**Given** callback invocation remains deferred by the architecture
**When** this story is implemented
**Then** Dib must not invoke command callbacks unless a later architecture/API decision explicitly approves that surface
**And** any callback metadata is returned or modeled only within the approved public contract.

**Given** caller-provided execution functions may eventually return ordinary Go errors
**When** error-boundary guidance is documented or implemented
**Then** ordinary errors are not converted into process exits by default
**And** typed Dib errors remain inspectable separately from caller-owned errors.

**Given** command behavior must stay composable with flags and config
**When** verification runs
**Then** tests cover context propagation boundaries, writer injection boundaries, ordinary error preservation, no process control, and immutable snapshots
**And** the package graph remains `command/` consuming `flags/` without depending on `config/`.

## Epic 4: Provenance-Aware Config Resolution

Developers can register Config keys, resolve values by documented precedence, bind flags/env/JSON sources, retrieve typed values, inspect provenance, and protect sensitive diagnostics.

### Story 4.1: Register Config Keys With Defaults And Type Expectations

**Requirements:** FR11, FR16, FR20

As a Go CLI developer,
I want reusable Config key definitions with defaults, types, and sensitivity metadata,
So that config resolution starts from explicit, inspectable contracts rather than ad hoc map lookups.

**Acceptance Criteria:**

**Given** a caller registers a Config key
**When** the key definition is created
**Then** it captures the stable key name, optional default value, type expectation, documentation metadata, and sensitivity classification
**And** exact key matching is used by default.

**Given** a caller opts into key normalization
**When** normalized keys collide
**Then** setup fails with a typed deterministic error
**And** collisions are detected before resolution.

**Given** no higher-precedence source sets a registered key
**When** the key is resolved
**Then** the default value can be returned with `default` provenance
**And** missing unregistered keys return a documented not-found result rather than panicking.

**Given** redaction must be defined before source diagnostics expand
**When** sensitive metadata is configured
**Then** diagnostics may identify the key and source label
**And** they must not echo `dib_fake_secret_value`, `dib_fake_password_value`, or `dib_fake_token_value`.

**Given** config definitions must be reusable
**When** verification runs
**Then** tests cover defaults, not-found results, type expectations, exact matching, normalization collisions, sensitivity metadata, immutable definition reuse, and redaction corpus false positives/false negatives.

### Story 4.2: Read Config Sources Through Explicit Boundaries

**Requirements:** FR12, FR14, FR15, FR20

As a security-sensitive CLI developer,
I want config values to enter Dib only through explicit setters, injected env lookup, and JSON readers or paths,
So that tests and audits do not depend on ambient process or filesystem state.

**Acceptance Criteria:**

**Given** a caller sets a value explicitly
**When** config resolution reads that source
**Then** the value is tracked with `explicit setter` provenance
**And** repeated valid writes use documented last-writer-wins semantics within the source.

**Given** a caller binds environment variables
**When** an injected environment lookup returns values
**Then** env values are tracked with `env` provenance
**And** empty environment values count as set values.

**Given** a caller loads JSON config from a path or `io.Reader`
**When** JSON loading succeeds
**Then** registered Config keys can be set with `JSON` provenance
**And** strict mode is the documented default for registered-key loads while permissive mode is opt-in.

**Given** source reads can fail
**When** env binding, JSON file, read, decode, unknown-key, or type errors occur
**Then** failures are distinguishable with typed diagnostics
**And** sensitive raw values remain redacted in errors, debug strings, diagnostics, and source reports.

**Given** source boundaries must be testable
**When** verification runs
**Then** tests use injected env lookup, `io.Reader`, package-local JSON fixtures, and fake sensitive values rather than live process env or host files
**And** `go test ./...` and `go run ./tools/depgate` pass.

### Story 4.3: Resolve Config Precedence Only When Values Are Needed

**Requirements:** FR12, FR13, FR20

As a Go CLI developer,
I want config values resolved by the documented precedence only when callers ask for them,
So that flags, env, JSON, explicit setters, and defaults compose without stale copied state.

**Acceptance Criteria:**

**Given** multiple sources provide a value for the same registered key
**When** a caller resolves that key
**Then** the V1 precedence order is explicit setter, parsed flag, environment variable, JSON file, default
**And** the winning value reports the source label that supplied it.

**Given** parsed flags can bind to Config keys
**When** a bound Flag was explicitly set by CLI input
**Then** its value can outrank lower-precedence sources through `flag binding` provenance
**And** a flag's configured default does not accidentally override env or JSON values unless explicitly configured.

**Given** config should not depend on command routing or parser internals
**When** flag binding is implemented
**Then** `config/` accepts exported flag-derived snapshots or a small source interface
**And** it does not import `command/` or depend on unexported `flags/` implementation details.

**Given** same-source ties are valid in some cases
**When** repeated writes or loads occur within the same source
**Then** last-writer-wins behavior is deterministic
**And** binding collisions fail as setup errors rather than becoming runtime precedence ties.

**Given** precedence ambiguity is high risk
**When** verification runs
**Then** tests cover every precedence pair, absent/default/explicit-zero distinctions, flag binding explicit-set behavior, same-source ties, binding collisions, provenance labels, redaction, and immutable snapshots.

### Story 4.4: Retrieve Typed Values And Absence State

**Requirements:** FR11, FR16, FR20

As a Go CLI developer,
I want typed Config retrieval to distinguish absent values, zero values, and conversion failures,
So that config-dependent code can handle decisions explicitly.

**Acceptance Criteria:**

**Given** a registered key has a resolved value
**When** a caller retrieves it through a typed getter
**Then** the getter returns the typed value and provenance
**And** conversion failures return typed errors without panics.

**Given** a key is absent
**When** a caller checks presence
**Then** `IsSet` or equivalent distinguishes absent values from zero values
**And** missing unregistered keys return the documented not-found result.

**Given** a source provides an explicit zero value or empty env value
**When** the value is resolved
**Then** the value counts as set
**And** tests distinguish explicit zero, empty string, default zero, and absent.

**Given** values may be sensitive
**When** typed retrieval fails or reports diagnostics
**Then** diagnostics identify the key, source label, and failure category
**And** sensitive raw values remain redacted.

**Given** typed retrieval is a public caller contract
**When** verification runs
**Then** tests cover string, bool, numeric, duration or supported typed values, conversion failures, absent/default/zero states, source labels, and `errors.Is` or `errors.As` inspection.

### Story 4.5: Report Provenance And Redact Sensitive Diagnostics

**Requirements:** FR16, FR20

As a security-sensitive CLI developer,
I want config diagnostics and source reports to explain why a value won without exposing secrets,
So that audited tools remain debuggable and safe.

**Acceptance Criteria:**

**Given** a config value resolves successfully
**When** a caller asks for provenance or a source report
**Then** the report identifies the key, winning source label, and relevant source metadata
**And** source labels use only the canonical vocabulary: `default`, `explicit setter`, `flag binding`, `env`, and `JSON`.

**Given** resolution attempts fail
**When** diagnostics are rendered or inspected
**Then** diagnostics distinguish attempted source label from failure category where both apply
**And** conversion failure and source read failure are distinguishable.

**Given** a key is marked sensitive
**When** errors, debug strings, diagnostics, source reports, examples, or validation failures are produced
**Then** fake sensitive values from the architecture corpus never appear
**And** the key and source remain identifiable enough for debugging.

**Given** provenance output must be deterministic
**When** tests compare reports
**Then** ordering of keys, attempted sources, failures, and diagnostics is stable
**And** golden tests are limited to human-facing rendering while structured state is asserted separately.

**Given** config provenance is adoption evidence
**When** verification runs
**Then** tests cover success reports, failure reports, redaction false positives/false negatives, deterministic rendering, typed errors, and standard-library-only examples.

## Epic 5: Migration, Compatibility, And Release Evidence

Developers and reviewers can understand Dib's compatibility boundaries, follow migration examples, and verify release readiness with behavior matrices, dependency checks, docs, examples, and provenance evidence.

### Story 5.1: Publish Compatibility Boundaries For Familiar CLI Concepts

**Requirements:** FR17, FR18

As a developer evaluating Dib,
I want a clear compatibility table for Go `flag`, pflag, Cobra, and Viper concepts,
So that I can decide whether Dib fits without assuming source compatibility.

**Acceptance Criteria:**

**Given** Dib supports familiar CLI and config concepts
**When** `docs/compatibility.md` is written
**Then** it documents supported, narrowed, omitted, and intentionally different V1 behavior for Go `flag`, pflag, Cobra, and Viper
**And** it never describes Dib as source-compatible or a drop-in replacement.

**Given** intentional differences can affect migration
**When** a difference is documented
**Then** the docs include the user-facing reason
**And** they link or trace to behavior matrices, examples, or tests where practical.

**Given** clean-room boundaries apply to compatibility prose
**When** compatibility docs reference inspiration projects
**Then** references are classified in `docs/provenance-log.md` where required
**And** copied examples, fixtures, internal names, or file organization are not introduced.

**Given** compatibility docs are adoption evidence
**When** verification runs
**Then** docs are checked for stale claims against implemented behavior where practical
**And** `go test ./...` and `go run ./tools/depgate` pass.

### Story 5.2: Provide Migration Examples For Flags, Commands, And Config

**Requirements:** FR18, FR19, FR21

As a developer migrating a small CLI,
I want executable examples for familiar flag, command, and config patterns,
So that I can adopt Dib's native API without copying framework-shaped code.

**Acceptance Criteria:**

**Given** a developer knows standard Go `flag`
**When** they read the migration examples
**Then** an example shows explicit Flag sets, typed errors, and table-driven tests without package-global state
**And** the example builds through `go test ./...`.

**Given** a developer knows pflag-style shorthand behavior
**When** they read the migration examples
**Then** an example shows long flags, shorthand flags, grouped shorthands, repeated values, no-option defaults, and `--` behavior
**And** intentional differences are referenced rather than hidden.

**Given** a developer knows Cobra-style command trees
**When** they read the migration examples
**Then** an example shows nested command routing, aliases, local/inherited flags, help rendering, and caller-controlled errors
**And** it does not add a `/cmd` scaffold or process-owning framework shape.

**Given** a developer knows Viper-style config resolution
**When** they read the migration examples
**Then** an example shows defaults, explicit setters, flag binding, env binding, JSON loading, precedence, typed retrieval, provenance, and redaction
**And** it uses injected inputs rather than ambient process env or host-specific files.

**Given** examples are executable trust artifacts
**When** verification runs
**Then** examples compile with the standard library only, trace to relevant FRs, avoid copied source/examples, and pass `go test ./...` plus `go run ./tools/depgate`.

### Story 5.3: Consolidate Behavior Matrices Into Adoption Evidence

**Requirements:** FR20, FR21

As a technical reviewer,
I want the implemented behavior matrices consolidated into reviewable documentation,
So that I can audit command, flag, config, diagnostic, redaction, and dependency behavior without reverse-engineering tests.

**Acceptance Criteria:**

**Given** behavior was implemented across Epics 2 through 4
**When** `docs/behavior-matrices.md` is finalized
**Then** it summarizes flag parsing, command routing, config resolution, diagnostics, provenance, redaction, and deterministic rendering behavior
**And** each row traces to FRs, story IDs, and executable tests where practical.

**Given** tests are the executable source of truth
**When** docs and tests disagree
**Then** the mismatch is treated as a blocking issue until docs, tests, or implementation are reconciled
**And** no matrix row claims unsupported behavior.

**Given** human-facing output has deterministic contracts
**When** diagnostics, help, usage, and source reports are documented
**Then** docs distinguish structured assertions from golden rendering tests
**And** sensitive values remain absent from examples and rendered artifacts.

**Given** behavior matrices are release evidence
**When** verification runs
**Then** `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and applicable fuzz/race evidence are recorded or referenced in the release checklist
**And** provenance entries exist for reference-derived matrix content.

### Story 5.4: Prove Release Readiness With Dependency And Provenance Evidence

**Requirements:** FR18, FR19, FR20, FR21

As a release reviewer,
I want a complete release-readiness evidence package,
So that Dib's v0 module tag is backed by tests, dependency checks, compatibility documentation, migration guidance, and clean-room provenance.

**Acceptance Criteria:**

**Given** Dib is preparing a v0 module release
**When** `docs/release-checklist.md` is completed
**Then** it records exact commit, test, vet, dependency-gate, race-test, docs/examples, provenance, compatibility, and migration evidence
**And** unresolved checklist items block release readiness.

**Given** dependency evidence is central to adoption
**When** release checks run
**Then** `go run ./tools/depgate` proves zero external imports for library, test, example, and tool packages unless the architecture has been updated
**And** dependency-gate output is recorded as release evidence.

**Given** v0 may include future breaking changes
**When** release notes are prepared
**Then** they state the v0 experimental API status
**And** they still preserve correctness, redaction, clean-room, dependency, and release-gate expectations.

**Given** compatibility and migration docs are adopter-facing contracts
**When** release readiness is reviewed
**Then** examples, compatibility boundaries, behavior matrices, provenance log, diagnostics/errors docs, and config precedence docs align with implemented behavior
**And** any waivers include owner, reason, expiry, and impact.

**Given** release evidence must be reproducible
**When** a reviewer reruns the documented commands
**Then** `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `go test -race ./...` pass for the tagged commit
**And** no release process assumes binary deployment, Docker, Kubernetes, generated shell completion, or generated man pages.

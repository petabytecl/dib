---
title: "PRD: dib"
status: final
created: "2026-06-10"
updated: "2026-06-12"
---

# PRD: dib

## 0. Document Purpose

This PRD defines the first-version product requirements for Dib, a Go standard-library-only toolkit for command routing, flag parsing, configuration resolution, and explicit composition of those surfaces through an optional CLI invocation package. It is written for downstream architecture, epics, stories, implementation, and review. Functional requirements use stable FR IDs, user journeys use UJ IDs, and resolved compatibility, parser, config, versioning, and public error decisions are captured in section 12.

Primary source input: `_bmad-output/planning-artifacts/briefs/brief-dib-2026-06-10/brief.md`. Supporting rationale lives in `addendum.md`.

## 1. Vision

Dib is the dependency-free CLI foundation for Go projects that need modern command, flag, and configuration ergonomics without adopting a broad framework or transitive runtime dependency graph. It gives infrastructure CLIs, repo-local tools, admin utilities, and security-sensitive build or deployment helpers a small library they can read, audit, test, and keep stable.

The product thesis is constraint discipline: Dib should feel familiar to developers who know the public behavior of Go `flag`, pflag, Cobra, and Viper, while remaining a clean-room, native Go API rather than a source-compatible clone. V1 wins when a team can build a realistic multi-command CLI with long flags, shorthand flags, inherited command flags, help text, environment variables, JSON config, documented precedence, typed errors, and an optional explicit composition path using only the Go standard library at runtime.

This PRD is final for architecture, epic/story creation, and V1 implementation planning. Section 12 closes the compatibility, parser, config, versioning, and public error decisions that previously blocked implementation readiness.

## 2. Target User

### 2.1 Jobs To Be Done

- Build a multi-command Go CLI without pulling Cobra, pflag, Viper, or their transitive runtime dependencies into the shipped module.
- Parse familiar POSIX/GNU-style flags, shorthand flags, grouped boolean shorthands, repeated values, custom values, and `--` terminators with caller-controlled errors.
- Resolve configuration from defaults, explicit setters, flags, environment variables, and JSON files through a clear precedence model.
- Keep CLI code testable by preferring explicit instances, ordinary `io.Reader` / `io.Writer` integration, and inspectable errors over package-global behavior.
- Audit a library dependency quickly enough to use it in internal platform, build, deployment, or security-sensitive tools.
- Migrate a small command surface from standard `flag`, pflag-style flags, Cobra-style command trees, or Viper-style config resolution without treating Dib as a drop-in replacement.

### 2.2 Non-Users (V1)

- Teams that need full Cobra, pflag, or Viper source compatibility.
- Teams that require YAML, TOML, HCL, dotenv, live reload, remote key/value stores, generated shell completion, generated man pages, or scaffolding in the core runtime.
- Teams that want a framework with hidden global state as the default integration path.
- Teams that need reflection-heavy struct decoding in V1.

### 2.3 Key User Journeys

- **UJ-1. Ava builds an internal deployment CLI without a framework.** Ava is a platform engineer building `deployctl` for a small team. She defines a root command, adds `deploy apply` and `deploy status`, attaches shared `--config` and `--output` flags, adds command-local flags, and routes execution with `context.Context`. She confirms help output and parse errors are deterministic enough for tests.

- **UJ-2. Mateo ports a pflag-style parser into a smaller repo tool.** Mateo maintains a build helper that currently expects `--long`, `-s`, `-abc`, repeated tags, no-option defaults, and `--` terminator behavior. He replaces the parser with Dib, runs a table-driven behavior matrix, and finds intentional differences documented before they affect scripts.

- **UJ-3. Lin adds environment and JSON configuration to an audited admin utility.** Lin owns a security-sensitive admin CLI. She binds config keys to defaults, flags, environment variables, and a JSON file loaded from `io.Reader`, then verifies precedence and error handling without adding runtime dependencies or leaking config values in diagnostics.

- **UJ-4. Priya reviews Dib before approving it for internal tools.** Priya is a staff engineer responsible for dependency policy. She checks the module graph, reads the clean-room source policy, reviews compatibility notes, and confirms the public API does not require package-level mutable state.

## 3. Glossary

- **Dib** - The Go library product described by this PRD.
- **Runtime dependency** - Any Go module imported by Dib runtime packages and therefore present in the consumer's build graph.
- **Development dependency** - A tool, generator, linter, or test-only package that does not become a runtime dependency of Dib packages.
- **Command** - A named executable action in a command tree. A Command may own local flags, inherited flags, argument rules, aliases, help text, and an execution function.
- **Command tree** - A hierarchy of Commands used to route CLI input such as `app server start`.
- **Flag** - A command-line option parsed from arguments, such as `--port=8080` or `-p 8080`.
- **Long flag** - A flag invoked with a long name, such as `--config`.
- **Shorthand flag** - A one-character flag invoked with a single dash, such as `-c`.
- **Shorthand group** - A single-dash sequence such as `-abc` that resolves to multiple shorthand flags under documented rules.
- **No-option default** - A value assigned when a flag is present without an explicit value, when the Flag definition permits that behavior.
- **Flag set** - An independent collection of Flag definitions and parse state.
- **Config key** - A stable logical name used to resolve a configuration value across sources.
- **Config source** - One provider of configuration values: defaults, explicit setters, parsed flags, environment variables, or JSON files.
- **Config resolver** - The Dib component that merges Config sources using the documented precedence model.
- **Precedence** - The ranking that decides which Config source wins when multiple sources set the same Config key.
- **Clean-room policy** - The rule that Dib may use public documentation and observable behavior as input, but must not copy source, tests, comments, examples, internal names, or file organization from inspiration projects.
- **Compatibility note** - A documented statement describing whether Dib matches, narrows, or intentionally differs from a behavior in Go `flag`, pflag, Cobra, or Viper.

## 4. Product Constraints And Guardrails

- Dib runtime packages must import only the Go standard library.
- Dib must be a clean-room implementation grounded in public documentation and independently written behavior tests.
- Dib must optimize for explicit instances before package-global convenience.
- Dib must expose errors that callers can inspect without string matching.
- Dib must document intentional differences from Go `flag`, pflag, Cobra, and Viper when a familiar behavior is narrowed or omitted.
- Dib must remain small enough that V1 users can audit the command, flag, and config packages without understanding a framework.

## 5. Features

### 5.1 Command Routing And Execution

**Description:** Dib lets developers define a Command tree with nested Commands, aliases, command-local execution, inherited flags, help output, and caller-controlled execution errors. It realizes UJ-1 and UJ-4.

#### FR-1: Define Command trees

Developers can define a root Command and nested child Commands with stable names, descriptions, aliases, and usage metadata.

**Consequences (testable):**
- A root Command can route `[]string{"deploy", "apply"}` to the `apply` child under the `deploy` parent.
- Unknown Command input returns a typed unknown-command error and does not call `os.Exit`.
- Aliases resolve to the intended Command and preserve the canonical Command name in returned execution metadata.

#### FR-2: Execute Commands explicitly

Developers can execute a Command tree through an explicit API that accepts arguments, output streams, and `context.Context` where execution crosses boundaries.

**Consequences (testable):**
- Command execution can be tested without mutating `os.Args`, `os.Stdout`, or `os.Stderr`.
- A canceled Context is observable by the command execution function.
- Execution functions can return ordinary Go errors that are not wrapped into process exits by default.

#### FR-3: Support local and inherited flags

Developers can attach flags to a Command and define inherited flags that apply to descendants.

**Consequences (testable):**
- A root-level inherited flag is available to child Command execution.
- A child Command can define a local flag without exposing it to siblings.
- Flag name conflicts across local and inherited scopes produce deterministic diagnostics.

#### FR-4: Generate deterministic help and usage text

Developers can render help and usage text for a Command tree using supplied writers.

**Consequences (testable):**
- Help output includes Command names, aliases, descriptions, arguments, and visible flags in deterministic order.
- Hidden flags do not appear in help output but remain parseable if explicitly allowed by their Flag definition.
- Deprecated flags render a deprecation note when configured.

### 5.2 Flag Parsing Foundation

**Description:** Dib provides independent Flag sets with long flags, shorthand flags, shorthand groups, typed values, repeated values, no-option defaults, name normalization, parse terminators, metadata, and typed parse diagnostics. It realizes UJ-2 and supports UJ-1.

#### FR-5: Define independent Flag sets

Developers can define and parse independent Flag sets without relying on package-level mutable state.

**Consequences (testable):**
- Two Flag sets with the same flag names can be parsed independently without shared state.
- Flag definitions include name, optional shorthand, default value, usage text, value parser, repeat policy, hidden/deprecated metadata, and no-option default where applicable.
- V1 does not provide package-level global command, flag, or config helpers. Explicit instances are the only primary API surface.

#### FR-6: Parse long and shorthand flags

Developers can parse POSIX/GNU-style long flags and one-character shorthand flags.

**Consequences (testable):**
- Long flags parse in `--name=value` and `--name value` forms where the Flag type permits a separate value.
- Boolean long flags parse as `--name` and `--name=false`.
- Shorthand flags parse in `-n value`, `-n=value`, and boolean `-v` forms where applicable.
- Shorthand names are unique within the relevant Flag set or Command scope.

#### FR-7: Parse shorthand groups and no-option defaults

Developers can parse shorthand groups and no-option defaults through documented rules.

**Consequences (testable):**
- Boolean shorthand group `-abc` sets `-a`, `-b`, and `-c` when all are boolean or otherwise allowed.
- A non-boolean shorthand may appear in a group only when it is last or when its Flag definition has a no-option default.
- A Flag with a no-option default receives that value when present without an explicit option.
- Invalid shorthand groups return typed errors that identify the failing shorthand.

#### FR-8: Support repeated and custom values

Developers can define repeated flags and custom values using small interfaces.

**Consequences (testable):**
- Repeated flags can accumulate values in command-line order when configured to do so.
- Single-value flags return a typed duplicate-value error when repeated and not configured for accumulation.
- Custom value parsers can return typed or wrapped errors that preserve caller inspection.
- Built-in flag value types include string, bool, int, int64, uint, uint64, float64, duration, and string list because they map cleanly to standard-library parsing primitives.

#### FR-9: Control parse boundaries and diagnostics

Developers can control how Flag parsing treats non-flag arguments, the `--` terminator, and parse errors.

**Consequences (testable):**
- `--` stops Flag parsing and leaves remaining arguments untouched.
- Before `--`, V1 parses flags interspersed with positional arguments and preserves positional argument order in the remaining-args result.
- Unknown flags, missing values, invalid values, duplicate shorthands, invalid groups, and help requests return typed errors.
- Parse errors can render user-facing diagnostics without leaking hidden configuration values.

#### FR-10: Normalize names intentionally

Developers can configure name normalization for flags and config bindings where supported.

**Consequences (testable):**
- A normalization function can map equivalent names such as `log-level`, `log_level`, and `log.level` when configured.
- Normalization collisions are detected at definition time and return deterministic errors.
- If no normalization is configured, exact names are used to avoid silent surprises.

### 5.3 Configuration Resolution

**Description:** Dib resolves Config keys from defaults, explicit setters, parsed flags, environment variables, and JSON files with a documented precedence order. It realizes UJ-3.

#### FR-11: Register Config keys and defaults

Developers can register Config keys with default values, type expectations, and optional documentation.

**Consequences (testable):**
- A registered key can return its default value when no higher-precedence source sets it.
- Missing unregistered keys return a documented not-found result rather than panicking.
- Config keys match exactly by default. Callers may opt into a normalizer; normalization collisions return deterministic setup errors.
- Type mismatches return typed conversion errors.

#### FR-12: Resolve values by precedence

Developers can resolve Config values using a stable precedence model.

**Consequences (testable):**
- The V1 precedence order is explicit setter, parsed flag, environment variable, JSON file, default. Remote key/value stores are excluded from V1 precedence because they are out of scope.
- The winning value can report which Config source supplied it.
- Ties within the same Config source use last-writer-wins semantics where repeated writes or loads are valid. Binding collisions are setup errors rather than precedence ties.

#### FR-13: Bind flags to Config keys

Developers can bind parsed Flag values to Config keys.

**Consequences (testable):**
- A bound Flag overrides lower-precedence sources only when the Flag was explicitly set by CLI input.
- A default Flag value does not accidentally override an environment variable or JSON file value unless explicitly configured.
- Bound Flags are read lazily when Config is resolved, so later parse results are visible without copying values into Config during parse.
- Flag bindings can be inspected for missing or renamed flags during setup.

#### FR-14: Bind environment variables to Config keys

Developers can bind environment variables to registered Config keys using explicit names or a configured prefix and replacer.

**Consequences (testable):**
- `APP_PORT` can bind to Config key `port` when the prefix and key mapping are configured.
- Empty environment values count as set values.
- Environment lookup is injectable for tests and does not require mutating the process environment in unit tests.

#### FR-15: Load JSON configuration from paths and readers

Developers can load JSON configuration from a filesystem path or an `io.Reader`.

**Consequences (testable):**
- JSON objects can set registered Config keys.
- JSON loading exposes strict and permissive modes. Strict mode is the documented default for registered Config loads; permissive mode is opt-in.
- File-not-found, read, decode, and type errors are distinguishable.
- JSON loading does not add non-standard-library parser dependencies.

#### FR-16: Retrieve typed Config values

Developers can retrieve Config values through typed getters and existence checks.

**Consequences (testable):**
- Typed getters return values plus errors where conversion can fail.
- `IsSet` or equivalent can distinguish absent values from zero values.
- Sensitive values can be marked so diagnostics identify the key without echoing the raw value. Sensitive metadata is V1 scope for diagnostics redaction only; Dib is not a secret manager.

### 5.4 Clean-Room Documentation And Migration

**Description:** Dib must be understandable as a clean-room, dependency-free toolkit. Documentation is part of the product because adoption depends on auditability and compatibility clarity. It realizes UJ-2 and UJ-4.

#### FR-17: Publish clean-room source policy

Maintainers can point reviewers to a clean-room policy that defines allowed and disallowed source inputs.

**Consequences (testable):**
- The repository documents that public docs and observable behavior are allowed inputs.
- The repository documents that copied source, tests, comments, examples, internal names, and file organization from inspiration projects are disallowed.
- Compatibility decisions are recorded in PRD, architecture, or package documentation.

#### FR-18: Document compatibility boundaries

Developers can see which Go `flag`, pflag, Cobra, and Viper behaviors Dib supports, narrows, omits, or intentionally changes.

**Consequences (testable):**
- The docs include a compatibility table for in-scope V1 behavior.
- Each intentional difference includes the user-facing reason.
- The docs do not describe Dib V1 as source-compatible with Go `flag`, pflag, Cobra, or Viper.

#### FR-19: Provide migration examples

Developers can follow examples from standard `flag`, pflag-style flags, Cobra-style command trees, and Viper-style config resolution.

**Consequences (testable):**
- Examples build with `go test` or `go test ./...` without network access after module download.
- Examples prefer explicit instances over global mutable state.
- Examples show error handling, output writer injection, and table-driven tests.

### 5.5 Validation And Release Evidence

**Description:** Dib's V1 release must prove behavior and dependency constraints with tests and repeatable checks. It realizes UJ-4 and supports every implementation feature.

#### FR-20: Provide behavior test matrices

Maintainers can validate parser, command, and config behavior through table-driven tests.

**Consequences (testable):**
- Flag tests cover long flags, shorthand flags, shorthand groups, boolean flags, non-boolean values, repeated values, custom values, terminator handling, interspersed args, no-option defaults, and parse diagnostics.
- Command tests cover routing, aliases, unknown commands, local flags, inherited flags, execution errors, and help output.
- Config tests cover defaults, explicit setters, flags, environment variables, JSON files, precedence, missing keys, type mismatches, and source reporting.

#### FR-21: Enforce the runtime dependency rule

Maintainers can verify that runtime packages import only the Go standard library.

**Consequences (testable):**
- A repository check fails when a runtime package imports a non-standard-library module.
- Test-only and tooling dependencies are separated from runtime imports and documented.
- The release checklist includes the dependency check result.

#### FR-22: Support fuzz or property-style parser hardening

Maintainers can harden parsers against edge cases without changing the runtime dependency contract.

**Consequences (testable):**
- Parser fuzz tests run with `go test` using standard Go fuzzing support rather than a third-party fuzzing dependency.
- Fuzz failures produce minimal reproducible inputs stored in the repo's normal testdata flow.
- Parser failures return errors rather than panics except where the public API explicitly documents programmer misuse.

#### FR-23: Run an isolated lint gate

Maintainers can run an automated linter gate in CI without adding runtime dependencies to Dib packages.

**Consequences (testable):**
- CI fails when the configured linter reports issues.
- The linter tool is pinned, reproducible, and isolated as development or CI tooling.
- `tools/depgate` still proves that Dib library, test, example, and approved tool packages do not import external modules unless this PRD and architecture are updated.

#### FR-24: Validate test coverage

Maintainers can validate release-candidate coverage with a documented, package-aware threshold policy.

**Consequences (testable):**
- CI generates coverage evidence with standard Go coverage output.
- Public runtime packages report package-level coverage and fail below the approved release threshold.
- Tooling packages either meet a separate threshold or document targeted exceptions with rationale and critical-path test evidence.

#### FR-25: Provide public usage documentation

Developers can use public documentation to install Dib, understand package roles, and build a small CLI without reading implementation internals.

**Consequences (testable):**
- The repository root includes a public README with import/install guidance, package overview, quickstart usage, release status, and links to deeper docs.
- Usage docs cover command construction, flag parsing, config source precedence, diagnostics, clean-room compatibility boundaries, and release gates.
- Documentation examples are independently written, clean-room compliant, and runnable through `go test ./...` where practical.

#### FR-26: Compose CLI invocation, command routing, flag parsing, and config resolution

Developers can use an optional `cli` package to carry a full process invocation, route commands, translate explicitly-set flags into config bindings, and return typed command, flag, config, and remaining-argument results without handing Dib process lifecycle control.

**Consequences (testable):**
- Callers pass full argv explicitly, for example `cli.FromOSArgs(os.Args)`, and Dib never reads `os.Args` itself.
- `cli.Invocation` exposes program name and user arguments through defensive accessors.
- `cli.Resolve` or equivalent routes command input, composes exported flag/config snapshots, and returns a result containing route, flag, config, and remaining-argument state.
- The composition package does not execute callbacks, call `os.Exit`, mutate streams, read env implicitly, or hide errors behind rendered text.
- The original `command`, `flags`, and `config` packages remain independently usable.

## 6. Cross-Cutting Non-Functional Requirements

- **NFR-1 Runtime dependency ceiling:** Runtime packages must import only the Go standard library.
- **NFR-2 Explicit-instance API:** Primary APIs must operate on explicit instances and caller-supplied inputs/outputs. V1 does not include package-level global command, flag, or config helpers. The optional `cli` package may provide explicit composition helpers only when all process inputs are caller-supplied and no hidden singleton state is introduced.
- **NFR-3 Typed errors:** Public error cases needed by callers must be inspectable without string matching.
- **NFR-4 Deterministic output:** Help, usage, and diagnostics must be deterministic enough for stable golden or snapshot tests.
- **NFR-5 No process control by default:** Library APIs must not call `os.Exit`, mutate process-wide streams, or read `os.Args` unless the caller chooses a convenience path documented to do so.
- **NFR-6 Testability:** Core behavior must be testable with table-driven unit tests and injected readers, writers, args, and environment lookup.
- **NFR-7 Compatibility clarity:** Familiar behavior must be documented as supported, narrowed, omitted, or intentionally different.
- **NFR-8 Security-sensitive diagnostics:** Error messages must identify bad keys, flags, and sources without dumping sensitive values when a Flag or Config key is marked sensitive.
- **NFR-9 Go version policy:** Dib V1 requires Go 1.26 or newer.
- **NFR-10 API stability:** Public API changes after V1 must follow semantic versioning and include deprecation guidance for at least one minor release before removal where practical.
- **NFR-11 Quality gate reproducibility:** Lint and coverage gates must be deterministic enough to run locally and in CI with the same documented commands.
- **NFR-12 Public onboarding clarity:** A new adopter must be able to start from public docs rather than BMAD planning artifacts.

## 7. API Contracts, Versioning, And Dependency Policy

- Runtime packages should be organized around cohesive capabilities: command routing, flag parsing, configuration resolution, and optional explicit CLI composition. Candidate package boundaries are recorded in `addendum.md` and architecture updates.
- Public APIs should prefer small interfaces, explicit constructors, ordinary `io.Reader` and `io.Writer` values, and `context.Context` only where execution crosses a boundary.
- Dib must not promise source compatibility with Go `flag`, pflag, Cobra, or Viper in V1. It offers a native Dib API with familiar concepts and documented differences.
- V1 does not include package-level global command, flag, or config helpers. The explicit instance API is the documented golden path.
- The first usable release is labeled v0 experimental. V0 behavior is intentionally test-covered and documented, but API stability hardens through implementation feedback before a future stable v1.
- Development dependencies are allowed only when they do not enter runtime package imports. Linter and coverage tooling must remain isolated as development or CI concerns unless a future architecture update explicitly approves another model.

## 8. Non-Goals

- Full source compatibility with Cobra, pflag, or Viper.
- Copying source, tests, comments, examples, internal names, or file organization from inspiration projects.
- YAML, TOML, HCL, dotenv, INI, properties, or custom config formats in core V1.
- Remote key/value stores, encrypted remote config, live file watching, or dynamic reload in core V1.
- Generated shell completion, generated man pages, or project scaffolding in core V1.
- Reflection-heavy struct decoding in V1.
- A global singleton as the default configuration or command pattern.
- A process-owning CLI framework, callback runner, source-compatible adapter, or root module facade.
- Shell execution, terminal UI rendering, logging framework integration, or application lifecycle management.

## 9. MVP Scope

### 9.1 In Scope

- Go module for Dib runtime packages.
- Command tree definition, nested routing, aliases, inherited flags, local flags, help/usage output, explicit execution, and typed command errors.
- Flag sets with long flags, shorthand flags, shorthand grouping, boolean flags, non-boolean values, repeated flags, custom values, no-option defaults, normalization, hidden/deprecated metadata, terminator handling, and typed parse diagnostics.
- Config resolver with defaults, explicit setters, flag bindings, environment bindings, JSON file loading, `io.Reader` loading, typed getters, source reporting, and documented precedence.
- Clean-room source policy.
- Compatibility table for Go `flag`, pflag, Cobra, and Viper inspired behavior.
- Migration examples from standard `flag`, pflag-style flags, Cobra-style command trees, and Viper-style config resolution.
- Table-driven behavior tests, runtime dependency check, linter gate, coverage validation, and public usage documentation.
- Optional explicit `cli` composition package for carrying caller-supplied invocations and composing command, flag, and config results.

### 9.2 Out Of Scope For MVP

- Source-compatible clone APIs for Cobra, pflag, or Viper.
- Non-JSON config formats in core.
- Shell completion generation.
- Man page generation.
- CLI project scaffolding generators.
- Process-owning CLI frameworks, callback runners, source-compatible adapters, and root module facades.
- Remote configuration systems.
- Live config watching or reload.
- Reflection-based struct decoding.
- External runtime dependencies.

## 10. Success Metrics

**Primary**

- **SM-1:** A realistic multi-command CLI example can be built with only standard-library runtime dependencies. Target: example includes at least three Commands, inherited flags, local flags, env binding, JSON config, and precedence tests. Validates FR-1 through FR-16 and FR-21.
- **SM-2:** Behavior coverage for the parser and Config resolver is explicit and repeatable. Target: table-driven tests cover every consequence under FR-6 through FR-16. Validates FR-20.
- **SM-3:** Runtime dependency gate passes for every release candidate. Target: no non-standard-library imports from runtime packages. Validates FR-21 and NFR-1.

**Secondary**

- **SM-4:** Migration examples cover the intended mental models. Target: examples exist for standard `flag`, pflag-style flags, Cobra-style command trees, and Viper-style config resolution. Validates FR-18 and FR-19.
- **SM-5:** Public error handling is inspectable. Target: documented typed errors exist for every error family listed in FR-9, FR-15, and FR-16. Validates NFR-3.
- **SM-6:** Release quality gates fail closed. Target: CI fails when lint or package-aware coverage validation fails. Validates FR-23, FR-24, and NFR-11.
- **SM-7:** Public onboarding works without planning artifacts. Target: a new adopter can follow the README and usage docs to build a minimal multi-command CLI using Dib. Validates FR-25 and NFR-12.
- **SM-8:** Three-package composition is ergonomic without hidden behavior. Target: a new adopter can use `cli` to compose `command`, `flags`, and `config` without manually slicing `os.Args[1:]` or manually converting `flags.Snapshot` values into `config.FlagValue` entries. Validates FR-26, NFR-2, NFR-5, and NFR-6.

**Counter-Metrics**

- **SM-C1:** Do not optimize for feature parity count with Cobra, pflag, or Viper. Higher parity is not success if it expands V1 beyond the dependency-free core.
- **SM-C2:** Do not optimize for fewer files or fewer packages. Dib should stay easy to audit through cohesive, small files rather than one large framework module.
- **SM-C3:** Do not optimize for package-global convenience at the expense of explicit, testable instances.

## 11. Risks And Mitigations

- **Risk: Clean-room drift.** Contributors may accidentally copy examples, internal names, or test fixtures from inspiration projects. **Mitigation:** keep the clean-room policy visible, require independently written tests, and record compatibility decisions.
- **Risk: Compatibility ambiguity.** Users may assume Dib is a drop-in Cobra/pflag/Viper replacement. **Mitigation:** publish compatibility notes and migration examples that show familiar concepts but native APIs.
- **Risk: Config resolver scope creep.** Config support can expand quickly into formats, watchers, remote stores, and struct decoding. **Mitigation:** keep V1 to defaults, explicit setters, flags, env, JSON, and reader loading.
- **Risk: Hidden global state.** Convenience helpers could undermine the auditability thesis. **Mitigation:** exclude package-level global helpers from V1 and make explicit instances the only primary API surface.
- **Risk: Parser edge cases dominate implementation time.** GNU/POSIX-style parsing has many corner cases. **Mitigation:** implement against the parser behavior matrix in section 12 and require table-driven tests before release.
- **Risk: Linter tooling weakens dependency claims.** External lint tooling could be confused with runtime dependency policy. **Mitigation:** isolate the linter as development/CI tooling, document the isolation model, and keep `tools/depgate` authoritative for root module import policy.
- **Risk: Coverage validation becomes a vanity metric.** A single aggregate threshold can hide weak public package coverage or unfairly penalize tooling packages. **Mitigation:** use package-aware thresholds and document any tooling-package exception with critical-path test evidence.
- **Risk: Public docs overstate compatibility.** New onboarding docs could imply drop-in compatibility with familiar CLI libraries. **Mitigation:** reuse compatibility boundaries, keep claims behavior-scoped, and verify examples with `go test ./...`.
- **Risk: CLI composition drifts into a framework.** The optional `cli` package could start owning process lifecycle, callbacks, streams, env reads, or file loading. **Mitigation:** require caller-supplied inputs, preserve returned typed results, keep callbacks/application execution caller-owned, and verify no hidden process reads or exits in tests.

## 12. Closed Compatibility And Behavior Decisions

All phase-blocking PRD questions are resolved for V1. Dib is a native API that supports familiar command, flag, and config concepts without promising source compatibility with Go `flag`, pflag, Cobra, or Viper.

### 12.1 Compatibility Boundary Table

| Inspiration | Supported in V1 | Narrowed in V1 | Omitted in V1 | Intentionally different in V1 |
| --- | --- | --- | --- | --- |
| Go `flag` | Explicit Flag sets, typed values, defaults, custom value parsing concepts, parse errors, and testable argument parsing. | Dib documents GNU-style `--long` and `-s` forms as the primary CLI syntax rather than Go `flag` source-level behavior. | Source-compatible `flag.*` APIs and package-global `CommandLine` helpers. | Flags can be parsed interspersed with positional args before `--`; explicit instances are the only primary API. |
| pflag | Long flags, one-character shorthands, shorthand grouping, no-option defaults, repeated values, custom values, `--` terminator behavior, and pflag-style interspersed args. | Built-ins are limited to string, bool, int, int64, uint, uint64, float64, duration, and string list for V1. Name normalization is opt-in. | Source-compatible pflag APIs and broader pflag extension surface beyond the V1 parser contract. | No automatic `--no-*` boolean aliases. Booleans use `--flag`, `--flag=true`, and `--flag=false`. |
| Cobra | Nested command routing, aliases, command-local flags, inherited flags, deterministic help, and explicit execution errors. | Command execution is instance-oriented and test-oriented; no API path may require process-global streams or exits. | Source-compatible `*cobra.Command` APIs, generated shell completion, generated man pages, and scaffolding. | Dib treats command routing as a library operation that returns typed errors rather than owning process lifecycle. |
| Viper | Defaults, explicit setters, flag bindings, env bindings, JSON config, precedence, typed getters, and source reporting. | Registered Config keys, exact key matching by default, JSON-only file loading, and separate strict/permissive JSON modes. | Source-compatible Viper APIs, package-global singleton config, YAML/TOML/HCL/dotenv/INI/properties, remote stores, live reload, and reflection-heavy struct decoding. | Bound Flags are read lazily by Config; empty env values count as set; binding collisions fail during setup. |

### 12.2 Parser Behavior Matrix

| Case | V1 behavior | Error contract |
| --- | --- | --- |
| `--name=value` | Parses a known long flag with an attached value when the Flag type accepts values. | Unknown long flag or invalid conversion returns a typed parse error. |
| `--name value` | Parses a known long flag with the next argument as value when the Flag type permits a separate value. | Missing value returns a typed missing-value parse error. |
| `--flag`, `--flag=true`, `--flag=false` | Boolean long flags accept present-without-value and explicit true/false values. | Invalid boolean text returns a typed conversion parse error. |
| `--no-flag` | Not generated or recognized automatically. It is treated as an ordinary long flag name only if the caller explicitly registers `no-flag`. | Unregistered `--no-*` input returns a typed unknown-flag parse error. |
| `-n value`, `-n=value`, `-v` | Shorthand flags accept separate, equals-attached, and boolean present forms where the Flag definition permits them. | Unknown shorthand, missing value, or invalid conversion returns a typed parse error. |
| `-abc` where all are boolean | Sets `-a`, `-b`, and `-c` in order. | The failing shorthand is identified if any member is invalid. |
| `-ab10` where `a` is boolean and `b` is non-boolean | Sets `-a`; `-b` consumes `10` because the non-boolean shorthand is final in the group. | Invalid conversion returns a typed parse error for `b`. |
| `-ab 10` where `a` is boolean and `b` is non-boolean | Sets `-a`; `-b` consumes the next argument because `b` is final in the group. | Missing next argument returns a typed missing-value parse error for `b`. |
| Non-boolean shorthand before the end of a group | Allowed only when that Flag definition has a no-option default; it receives the no-option value and parsing continues through the group. | Without a no-option default, returns a typed invalid-group parse error. |
| Repeated flag | Accumulates in command-line order only when configured for repetition. | Single-value flags repeated by CLI input return a typed duplicate-value parse error. |
| Positional args before `--` | Flags remain parseable when interspersed with positional args; positional args keep relative order in the remaining-args result. | Invalid flags encountered before `--` still return typed parse errors. |
| `--` terminator | Stops flag parsing and leaves every remaining argument untouched, including strings that look like flags. | No parse errors are emitted for arguments after `--`. |
| Help request | Returns a typed help-request result/error for caller-controlled rendering and exit policy. | Library APIs do not call `os.Exit`. |

### 12.3 Config Semantics Table

| Aspect | V1 decision |
| --- | --- |
| Config key matching | Exact Config keys by default. Callers may opt into a normalizer; normalized key collisions return setup errors. |
| Env bindings | Env values bind to registered Config keys through explicit env names or an optional prefix and replacer. Dib does not import arbitrary unregistered environment variables. |
| Empty env values | Empty environment values count as set values. |
| Flag bindings | Bound Flags are read lazily when Config is resolved. Only flags explicitly set by CLI input outrank lower-precedence sources by default. |
| Precedence | Explicit setter, parsed flag, environment variable, JSON file, default. |
| Same-source ties | Last writer wins for valid repeated writes or loads within the same source. Binding collisions fail at setup instead of becoming runtime precedence ties. |
| JSON load modes | Strict mode is the default registered-key load mode and errors on unknown keys. Permissive mode is an explicit opt-in that ignores unknown keys. |
| JSON source errors | File-not-found, read, decode, unknown-key, and conversion failures are distinguishable. |
| Sensitive metadata | V1 includes sensitive metadata for diagnostics redaction only. Diagnostics may name the key, flag, or source but must not echo the sensitive raw value. |

### 12.4 Error, Version, And Release Decisions

| Topic | V1 decision |
| --- | --- |
| Public error boundary | Export stable typed or sentinel boundary errors for command lookup, parse, config source, and conversion failures. Private internals remain private and may be wrapped behind these boundaries. |
| Minimum Go version | Go 1.26 or newer. The Go downloads page lists Go 1.26.4 as a stable release at the time of this PRD update. |
| First release label | The first usable release is v0 experimental, not stable v1 and not only an internal milestone. |
| Runtime dependency policy | Runtime packages import only the Go standard library. Development and test tooling may use external dependencies only when they do not enter runtime imports. |
| Release quality gates | Release candidates must include test, vet, dependency, lint, coverage, race, docs/examples, provenance, compatibility, and migration evidence. Lint tooling remains isolated development/CI tooling, and coverage policy is package-aware. |

## 13. Deferred Questions

No phase-blocking open questions remain for V1 architecture, epic/story creation, or implementation planning. These non-blocking topics are explicitly deferred until after v0 feedback:

- Whether to add source-compatible adapter packages for specific Go `flag`, pflag, Cobra, or Viper use cases.
- Whether package-level convenience helpers are worth adding after explicit-instance APIs settle.
- Whether to add non-JSON config formats, remote config, live reload, generated shell completion, generated man pages, scaffolding, or reflection-heavy struct decoding outside core V1.
- What evidence and compatibility threshold should promote Dib from v0 experimental to a future stable v1.

## 14. Source Grounding

- Product brief: `_bmad-output/planning-artifacts/briefs/brief-dib-2026-06-10/brief.md`
- Go `flag` package documentation: https://pkg.go.dev/flag
- pflag package documentation: https://pkg.go.dev/github.com/spf13/pflag
- Cobra flags documentation: https://cobra.dev/docs/how-to-guides/working-with-flags/
- Viper package documentation: https://pkg.go.dev/github.com/spf13/viper
- Go downloads page, verified 2026-06-11 for Go 1.26.4 stable release: https://go.dev/dl/

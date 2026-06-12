# Behavior Matrices

This is the initial shared behavior contract matrix for Dib. Package tests
remain the executable truth. This document records the cross-surface hooks that
keep `command/`, `flags/`, and `config/` from drifting as each surface grows.

## Shared Contracts

| Contract | Applies To | Required Behavior | Initial Evidence | Later Hook |
| --- | --- | --- | --- | --- |
| Immutable definitions | `command`, `flags`, `config` | Definitions and derived definitions return new values instead of mutating existing values. Exported APIs do not expose mutable internals. | Story 1.1 `command.NewDefinition` returns a value with unexported fields; Story 2.1 `flags.Set.With` returns a derived set without mutating the original. | Add derivation assertions for command routing and config keys when those APIs exist. |
| No mutable aliases | `command`, `flags`, `config` | Constructors and derivation methods defensively handle caller-owned slices, maps, readers, environment lookup functions, and config data. Callers must not observe mutation through retained aliases after construction or resolution. | Story 2.1 tests prove `flags.StringList` defaults, `Set.Definitions`, and default snapshots do not leak caller-observable mutable storage. | Add table tests for each new caller-owned input as the command and config APIs are introduced. |
| Per-run snapshots | `command`, `flags`, `config` | Route, parse, and config resolution snapshots never write back to definitions. A snapshot copies the values it needs and does not depend on live process state, environment variables, readers, lookup functions, or caller-owned mutable inputs after creation. | Story 2.1 `flags.Set.DefaultSnapshot` tests prove default value state is self-contained and ignores misleading `os.Args` and environment state. | Extend snapshot assertions when full flag parsing, command routing, and config resolution exist. |
| Explicit instances | `command`, `flags`, `config` | Primary APIs accept caller-provided values and instances. They do not read package globals, default singletons, `os.Args`, environment variables, stdin, stdout, or stderr unless that behavior is explicit in the API under test. | `command/contract_test.go` demonstrates explicit command definition construction; Story 2.1 flags tests demonstrate independent `flags.Set` construction despite misleading process state. | Extend the same pattern to config when that surface adds constructors. |
| Explicit flag name normalization | `flags` | Long flag names match exactly unless a caller supplies a `flags.NameNormalizer`. Normalized sets resolve equivalent long-name spellings to the canonical definition name, reject normalized collisions with typed setup errors, and keep shorthand lookup independent from long-name normalization. | Story 2.2 tests cover exact matching, configured normalization, normalized collision diagnostics, immutable derivation with `WithNormalizer`, and shorthand separation. | Extend parsing tests to reuse the same lookup contract once flag-token parsing exists. |
| Long flag parsing | `flags` | `flags.Set.Parse(args)` accepts caller-supplied arguments, parses long flags in attached and separate value forms, treats boolean presence as true, preserves remaining positional args, records source occurrences, and returns typed diagnostics for unknown, missing, conversion, and duplicate-value failures. | Story 2.3 tests cover `--name=value`, `--name value`, booleans, `--no-*` as an ordinary name, exact and normalized lookup, duplicate single-value flags, minimal `--` terminator behavior, source metadata, defensive snapshots, and redaction-safe parse errors. | Extend the parser matrix for shorthand, short groups, repeat/custom accumulation, and complete boundary behavior in later Epic 2 stories. |
| Public error inspection | `command`, `flags`, `config` | Errors that callers need to handle programmatically are typed or sentinel errors inspectable with `errors.Is` or `errors.As`. Error strings are diagnostics, not the programmatic contract. | Story 1.1 uses `errors.As` for `*command.NameError`; Story 2.1 uses sentinel errors with `*flags.DefinitionError` and `*flags.ValueError`. | Add typed error assertions for command routing and config resolution. |
| Diagnostic vocabulary | `command`, `flags`, `config` | Human-facing diagnostics identify the package surface, command/flag/key name when applicable, typed category, source label when applicable, and redaction status. | `docs/diagnostics-and-errors.md` defines the vocabulary shape. | Add source report and rendered diagnostic tests when diagnostics are implemented. |
| Source labels | `config` | Config provenance uses exactly `default`, `explicit setter`, `flag binding`, `env`, and `JSON`. | `docs/diagnostics-and-errors.md` fixes the vocabulary before config resolution exists. | Add provenance matrix tests in config precedence stories. |
| Redaction corpus | `command`, `flags`, `config` | The fake sensitive values are exactly `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`. Raw sensitive values must not appear in public diagnostics once redaction behavior exists. | `docs/diagnostics-and-errors.md` fixes the corpus and redaction scope; Story 2.1 proves sensitive flag conversion errors omit `dib_fake_secret_value`. | Add redaction assertions for debug strings, examples, and source reports as those surfaces exist. |

## Current Scope Boundaries

This matrix is not the final V1 behavior matrix. It intentionally does not
claim that flag parsing, command routing, config precedence, source reports, or
redaction rendering exist today.

Do not use this story to add dependency tooling, CI, examples, compatibility
tables, release checklists, parser fuzz seeds, root facade packages, `/cmd`
scaffolding, or shared `internal/` packages. Later stories own those surfaces
when their APIs and executable behavior exist.

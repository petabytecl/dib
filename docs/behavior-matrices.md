# Behavior Matrices

This is the initial shared behavior contract matrix for Dib. Package tests
remain the executable truth. This document records the cross-surface hooks that
keep `command/`, `flags/`, and `config/` from drifting as each surface grows.

## Shared Contracts

| Contract | Applies To | Required Behavior | Initial Evidence | Later Hook |
| --- | --- | --- | --- | --- |
| Immutable definitions | `command`, `flags`, `config` | Definitions and derived definitions return new values instead of mutating existing values. Exported APIs do not expose mutable internals. | Story 1.1 `command.NewDefinition` returns a value with unexported fields. | Add derivation assertions when flag sets, command routing, and config keys exist. |
| No mutable aliases | `command`, `flags`, `config` | Constructors and derivation methods defensively handle caller-owned slices, maps, readers, environment lookup functions, and config data. Callers must not observe mutation through retained aliases after construction or resolution. | Story 1.3 records the contract before alias-bearing APIs exist. | Add table tests for each caller-owned input as the APIs are introduced. |
| Per-run snapshots | `command`, `flags`, `config` | Route, parse, and config resolution snapshots never write back to definitions. A snapshot copies the values it needs and does not depend on live process state, environment variables, readers, lookup functions, or caller-owned mutable inputs after creation. | Story 1.3 records the expected snapshot contract. | Add executable assertions in the routing, parsing, and config resolution stories. |
| Explicit instances | `command`, `flags`, `config` | Primary APIs accept caller-provided values and instances. They do not read package globals, default singletons, `os.Args`, environment variables, stdin, stdout, or stderr unless that behavior is explicit in the API under test. | `command/contract_test.go` demonstrates explicit command definition construction despite misleading process state. | Extend the same pattern to flags and config when those surfaces add constructors. |
| Public error inspection | `command`, `flags`, `config` | Errors that callers need to handle programmatically are typed or sentinel errors inspectable with `errors.Is` or `errors.As`. Error strings are diagnostics, not the programmatic contract. | Story 1.1 uses `errors.As` for `*command.NameError`. | Add typed error assertions for flag parsing, routing, and config resolution. |
| Diagnostic vocabulary | `command`, `flags`, `config` | Human-facing diagnostics identify the package surface, command/flag/key name when applicable, typed category, source label when applicable, and redaction status. | `docs/diagnostics-and-errors.md` defines the vocabulary shape. | Add source report and rendered diagnostic tests when diagnostics are implemented. |
| Source labels | `config` | Config provenance uses exactly `default`, `explicit setter`, `flag binding`, `env`, and `JSON`. | `docs/diagnostics-and-errors.md` fixes the vocabulary before config resolution exists. | Add provenance matrix tests in config precedence stories. |
| Redaction corpus | `command`, `flags`, `config` | The fake sensitive values are exactly `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`. Raw sensitive values must not appear in public diagnostics once redaction behavior exists. | `docs/diagnostics-and-errors.md` fixes the corpus and redaction scope. | Add redaction assertions for errors, debug strings, examples, and source reports. |

## Current Scope Boundaries

This matrix is not the final V1 behavior matrix. It intentionally does not
claim that flag parsing, command routing, config precedence, source reports, or
redaction rendering exist today.

Do not use this story to add dependency tooling, CI, examples, compatibility
tables, release checklists, parser fuzz seeds, root facade packages, `/cmd`
scaffolding, or shared `internal/` packages. Later stories own those surfaces
when their APIs and executable behavior exist.

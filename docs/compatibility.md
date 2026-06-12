# Compatibility Boundaries

Dib is a clean-room native Go API with familiar CLI and configuration concepts.
Dib is not a source-compatible clone, not a drop-in replacement, and not a framework compatibility layer.
This boundary applies to Go `flag`, pflag, Cobra, Viper, and comparable projects.

This document describes V1 behavior boundaries for adopters. It does not provide
migration walkthroughs; Story 5.2 owns migration examples. It does not claim
that final adoption matrices or release readiness are complete; Story 5.3 owns
consolidated behavior evidence, and Story 5.4 owns release-readiness evidence.

## How To Read This Table

- **Supported** means Dib implements the familiar behavior through Dib-owned
  APIs and tests.
- **Narrowed** means Dib supports the core behavior with a smaller or more
  explicit scope.
- **Omitted** means Dib intentionally has no V1 surface for that behavior.
- **Intentionally different** means Dib chooses a different user-facing contract
  to preserve explicit instances, typed errors, deterministic output,
  clean-room provenance, or caller-owned process lifecycle.

## Compatibility Boundary Table

| Inspiration surface | Supported familiar concepts | Narrowed behavior | Omitted behavior | Intentionally different behavior | User-facing reason | Dib evidence |
| --- | --- | --- | --- | --- | --- | --- |
| Go `flag` | Explicit flag sets, typed flag values, default snapshots, custom parsing concepts, parse errors, and testable argument parsing. | Parsing is scoped to caller-created `flags.Set` values and caller-supplied args. Defaults are represented in Dib snapshots rather than process-global state. | Dib omits `flag.*` source compatibility, package-global `CommandLine` helpers, and implicit reads of `os.Args` or process output streams. | Parse failures are typed `*flags.ParseError` values with sentinel categories; diagnostic strings are not the programmatic contract. | Adopters can test parsing without ambient process state, inspect failures with `errors.Is` and `errors.As`, and keep exit/rendering policy in application code. | [Flag parser evidence](behavior-matrices.md#flag-parser-evidence-map), [Explicit instances](behavior-matrices.md#shared-contracts), [Public error inspection](behavior-matrices.md#shared-contracts), [Diagnostic vocabulary](diagnostics-and-errors.md#programmatic-error-contract). |
| pflag | Long flags, one-character shorthands, shorthand groups, no-option defaults, repeated values, custom values, `--` boundaries, interspersed args, and opt-in long-name normalization. | Normalization is caller-supplied and limited to long-name lookup. Shorthand lookup remains independent. No-option defaults apply only where Dib parser rules document them. | Dib omits source-compatible pflag APIs, broader extension hooks, deprecation and hidden-flag policy beyond Dib's implemented rendering scope, and automatic `--no-*` aliases. | `--no-*` is an ordinary long name unless the caller defines it. Failed parses return zero-value snapshots so partial state is not accidentally used. | The parser stays deterministic, auditable, and explicit about which spellings exist while still covering common shorthand and repeated-value workflows. | [Long flag parsing](behavior-matrices.md#flag-parser-evidence-map), [Short flag parsing](behavior-matrices.md#flag-parser-evidence-map), [Shorthand groups](behavior-matrices.md#flag-parser-evidence-map), [No-option defaults](behavior-matrices.md#flag-parser-evidence-map), [Parse boundaries](behavior-matrices.md#flag-parser-evidence-map), [Redaction](behavior-matrices.md#flag-parser-evidence-map). |
| Cobra | Nested command routing, aliases, local and inherited flags, deterministic help and usage rendering, typed routing failures, and caller-controlled execution-boundary metadata. | Routing returns snapshots and boundary metadata; it does not own handler dispatch or process lifecycle. Help and usage render only through caller-supplied writers. | Dib omits `*cobra.Command` compatibility, shell completion, manpage generation, scaffolding, suggestions, callback invocation, and process lifecycle ownership. | Help requests, unknown commands, flag failures, writer failures, context cancellation, stdout/stderr selection, and exit codes remain under caller control. | Applications can embed Dib routing without hidden exits, process-global writes, or generated assets that would expand the clean-room surface. | [Command route snapshots](behavior-matrices.md#flag-parser-evidence-map), [Command aliases](behavior-matrices.md#flag-parser-evidence-map), [Command local and inherited flag composition](behavior-matrices.md#flag-parser-evidence-map), [Command help and usage rendering](behavior-matrices.md#flag-parser-evidence-map), [Command execution-boundary metadata](behavior-matrices.md#flag-parser-evidence-map), [Story 3.5 diagnostics](diagnostics-and-errors.md#current-scope). Deferred areas are stated in [Current Scope Boundaries](behavior-matrices.md#current-scope-boundaries). |
| Viper | Defaults, explicit setters, flag bindings, env bindings, JSON readers and paths, precedence, typed getters, source reports, and rendered diagnostics. | V1 source ingestion is explicit: env uses caller-injected lookup, file input is JSON only, and flag binding accepts caller-supplied parsed flag values. Precedence is exactly `explicit setter > flag binding > env > JSON > default`. | Dib omits package-global singleton config, non-JSON formats, remote stores, live reload, aliases, reflection-heavy struct decoding, and ambient config discovery. | Config resolution occurs only when the caller invokes `config.Resolve`; source reports are value-free, and typed getters are the value retrieval API. | Adopters get reproducible source precedence, deterministic provenance, and redaction-safe diagnostics without hidden I/O or mutable global configuration. | [Config source ingestion](behavior-matrices.md#flag-parser-evidence-map), [Config precedence resolution](behavior-matrices.md#flag-parser-evidence-map), [Typed config retrieval](behavior-matrices.md#flag-parser-evidence-map), [Config provenance and diagnostics](behavior-matrices.md#flag-parser-evidence-map), [canonical precedence](config-precedence.md#precedence-order), [source labels](diagnostics-and-errors.md#source-labels), [release checklist dependency evidence placeholder](release-checklist.md#standard-library-dependency-evidence). |

## Evidence And Deferred Areas

Compatibility claims in this document trace to current Dib behavior evidence in
`docs/behavior-matrices.md`, package tests named by that matrix, the config
precedence authority in `docs/config-precedence.md`, and diagnostic/source-label
vocabulary in `docs/diagnostics-and-errors.md`. The dependency claim remains
guarded by `go run ./tools/depgate` and the standard-library dependency section
of `docs/release-checklist.md`.

The provenance boundary for public Go `flag`, pflag, Cobra, and Viper
documentation is recorded in `docs/provenance-log.md`. Those references are
inspiration-only. Dib compatibility prose is independently written and does not
copy external examples, fixtures, source structure, internal names, or document
layout.

The following areas are deferred rather than implied as supported:

- Migration examples before Story 5.2.
- Consolidated release adoption evidence before Story 5.3.
- Release readiness before Story 5.4.
- Callback invocation, shell completion, manpages, scaffolding, and generated
  command assets.
- Additional config formats, remote stores, live reload, aliases, and
  reflection-heavy struct decoding.

Rendered diagnostic strings are human-facing evidence only. Programmatic
contracts are the structured tests, typed errors, source reports, snapshots, and
behavior matrix rows referenced above.

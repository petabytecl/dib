# Diagnostics And Errors

This document defines the initial cross-surface contract for public errors,
diagnostics, provenance labels, and redaction vocabulary. It is implementation
guidance for Dib-owned code, not a copied pattern from another CLI library.

## Programmatic Error Contract

Errors that callers need to handle programmatically must be inspectable through
Go error inspection. Use typed errors, sentinel errors, or documented wrapping
behavior so callers can use `errors.Is` or `errors.As`.

Diagnostic strings are human-facing output. Tests and downstream callers must
not treat exact error strings as the only programmatic contract. When a surface
exposes typed or sentinel errors, tests should assert that contract directly and
only check strings for deliberate user-facing diagnostic behavior.

Do not document wrapping behavior unless the implementation actually provides
it. A typed error returned directly can satisfy an `errors.As` contract without
also being an `errors.Is` or wrapping contract.

## Diagnostic Vocabulary

Diagnostics should be structured around these concepts when the relevant
surface exists:

- Package surface: `command`, `flags`, or `config`.
- Name: command name, flag name, or config key when applicable.
- Source label: provenance source when config or binding behavior applies.
- Typed category: the public typed or sentinel error category.
- Redaction status: whether sensitive input was omitted or replaced.

The vocabulary describes what diagnostics communicate. It does not require a
shared runtime diagnostic type in this story.

Story 2.1 implements the first concrete `flags` error categories for definition
validation and value conversion. Callers can inspect them with `errors.Is` and
`errors.As`; diagnostic strings remain non-contractual and must not echo raw
sensitive values.

## Source Labels

Config provenance source labels are fixed to these exact spellings:

- `default`
- `explicit setter`
- `flag binding`
- `env`
- `JSON`

Use `JSON` in uppercase. Do not add synonyms such as `json`, `environment`, or
`explicit` without an explicit compatibility decision.

## Redaction Corpus

The fake sensitive-value corpus is fixed to these exact values:

- `dib_fake_secret_value`
- `dib_fake_password_value`
- `dib_fake_token_value`

Once redaction behavior exists, raw sensitive values must not appear in errors,
`String` output, debug strings, rendered diagnostics, source reports, examples,
or validation failures.

## Current Scope

Story 1.3 established the shared contract language. Story 2.1 adds initial
`flags` definition and conversion categories. Later stories own source-report
structures, rendered diagnostics, config provenance, and the remaining concrete
error categories for each package surface.

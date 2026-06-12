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

Story 2.2 adds `flags.ErrDuplicateNormalizedName` for caller-configured long-name
normalization collisions. The error is exposed through `*flags.DefinitionError`;
callers can inspect both raw names with `Name` and `CollidingName`, plus the
shared normalized lookup key with `NormalizedName`.

Story 2.3 adds long-flag parse diagnostics through `*flags.ParseError`.
Unknown long flags use `flags.ErrUnknownFlag`, omitted required values use
`flags.ErrMissingValue`, duplicate single-value flags reuse
`flags.ErrDuplicateValue`, and conversion failures continue to satisfy
`flags.ErrConversion` and expose `*flags.ValueError`. Parse errors expose the
source token without attached raw values, the raw long name, the normalized
lookup key, and the canonical definition when one was resolved.

Story 2.4 widens the same parse diagnostic contract to one-rune shorthand
tokens. Unknown shorthand uses `flags.ErrUnknownFlag`; omitted required
shorthand values use `flags.ErrMissingValue`; duplicate values across long and
short spellings reuse `flags.ErrDuplicateValue`; and shorthand conversion
failures continue to satisfy `flags.ErrConversion` and expose
`*flags.ValueError`. For shorthand parse errors, `ParseError.Token` exposes the
source token without any attached raw value, `ParseError.Name` exposes the
failing shorthand text, and `ParseError.NormalizedName` exposes the canonical
definition name when a definition was resolved. Short flag diagnostics must not
run through long-name normalization and must not echo sensitive raw values.

Story 2.5 adds grouped shorthand diagnostics for single-dash tokens with
multiple shorthand members and no `=`. Unknown group members reuse
`flags.ErrUnknownFlag`; missing final values reuse `flags.ErrMissingValue`;
conversion failures continue to satisfy `flags.ErrConversion` and expose
`*flags.ValueError`; and duplicate values reuse `flags.ErrDuplicateValue`.
Invalid non-final required-value members without no-option defaults use
`flags.ErrInvalidGroup`. For grouped parse errors, `ParseError.Token` exposes
the group prefix through the failing shorthand, omitting any attached value
suffix, `ParseError.Name` exposes the failing one-rune shorthand, and
`ParseError.Definition` is present when that shorthand resolved to a definition.
Successful grouped occurrences record the matched shorthand member as the source
spelling while keeping the canonical definition name as the occurrence lookup
key. If a grouped value-taking shorthand has a no-option default and no explicit
attached or separate value is available, the no-option value is recorded as
explicit CLI input. Long-name normalization still does not create shorthand
aliases.

Story 2.6 locks repeated and custom value diagnostics into the same parse
contract. Single-value duplicates are detected before converting the duplicate
token, so invalid duplicate values reuse `flags.ErrDuplicateValue` and do not
also expose `flags.ErrConversion`. Custom parser failures reuse
`flags.ErrConversion`, expose `*flags.ParseError`, and expose
`*flags.ValueError`. Non-sensitive custom parser failures preserve the caller's
parser cause through Go error inspection. Sensitive custom parser failures keep
the typed Dib context but do not expose raw sensitive values or the caller cause
that may contain them.

Story 2.7 adds `flags.ErrHelpRequest` as a new parse diagnostic category. When
`--help` is parsed and no `help` long flag is registered, or when `-h` is parsed
and no `h` shorthand is registered, `flags.Set.Parse` returns a `*flags.ParseError`
whose category is `flags.ErrHelpRequest`. The error is not an unknown-flag
error: `errors.Is(err, flags.ErrUnknownFlag)` returns false, and
`errors.Is(err, flags.ErrHelpRequest)` returns true. `ParseError.Token()` returns
`--help` regardless of whether `-h` or `--help` was the source token; this
gives callers a consistent token for rendering. When `--help` or `-h` ARE
registered as definitions by the caller, they parse through the normal flag path
and do not produce `ErrHelpRequest`. Help-request detection never calls
`os.Exit`, writes to stdout or stderr, or renders usage text; all of that
remains the caller's responsibility.

Story 2.8 adds parser hardening evidence without adding new public diagnostic
categories. The broad parser fuzz target asserts that parse failures remain
typed `*flags.ParseError` values, failed parses return zero-value snapshots,
successful snapshots expose defensive copies, reusable definitions are not
mutated, and sensitive conversion values remain redacted. These tests reinforce
the contracts above rather than changing the diagnostic vocabulary.

Story 3.1 adds the first concrete command routing diagnostic category. Unknown
command tokens return `command.ErrUnknownCommand` through
`*command.UnknownCommandError`, so callers can use both
`errors.Is(err, command.ErrUnknownCommand)` and
`errors.As` with `*command.UnknownCommandError`. The typed error exposes the
unmatched token with `Token()` and the matched parent command path with
`ParentPath()`. Unknown-command routing failures return a zero-value route
result so callers cannot accidentally use partial routing state. Routing
diagnostics do not render help, call `os.Exit`, read `os.Args`, or write to
stdout or stderr.

Story 3.2 adds command alias setup diagnostics. Blank aliases and aliases that
match their own command name return `command.ErrInvalidCommandAlias` through
`*command.AliasError`; callers can inspect the command name, alias token, and
parent path when one is available. Ambiguous lookup tokens return
`command.ErrDuplicateCommandToken` through `*command.TokenConflictError`;
callers can inspect the parent path, conflicting token, first canonical child
command, and colliding canonical child command. This category covers duplicate
child names, duplicate aliases, alias-vs-child-name collisions, and cross-alias
cycles. Diagnostic strings remain non-contractual; callers should use
`errors.Is`, `errors.As`, and accessors.

Story 3.3 adds command flag-composition setup diagnostics while preserving the
underlying `flags` package contracts. Long-name, shorthand, and normalized-name
collisions produced while composing inherited flags root-to-leaf plus the final
command's local flags return `command.ErrFlagComposition` through
`*command.FlagCompositionError`. Callers can inspect the canonical command path
with `Path()` and the composition scope with `Scope()`. The same error unwraps
the underlying `*flags.DefinitionError`, so callers can still use
`errors.Is` with `flags.ErrDuplicateName`, `flags.ErrDuplicateShorthand`, or
`flags.ErrDuplicateNormalizedName`, and `errors.As` with
`*flags.DefinitionError`. Runtime flag parse failures during routing continue
to return the typed `*flags.ParseError` values from `flags.Set.Parse`; command
routing does not convert help requests, unknown flags, missing values,
conversion failures, or duplicate values into command diagnostics.

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

Story 1.3 established the shared contract language. Stories 2.1 through 2.8
add the initial `flags` definition, normalization, conversion, long-parse,
shorthand-parse, shorthand-group, repeated-value, custom-parser, and
help-request categories, plus parser fuzz evidence that those categories remain
inspectable under arbitrary input. Story 3.1 adds the first `command` routing
diagnostic for unknown commands. Story 3.2 adds command alias setup diagnostics
for invalid aliases and duplicate lookup tokens. Story 3.3 adds command
flag-composition setup diagnostics and routes runtime flag parse failures
through the existing `flags` parse diagnostics. Later stories own source-report
structures, rendered diagnostics, config provenance, and the remaining concrete
error categories for each package surface.

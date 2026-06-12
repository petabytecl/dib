# Config Precedence

This is the canonical V1 precedence authority for Dib configuration resolution.

## Precedence Order

Values are resolved from highest to lowest priority:

1. **explicit setter** — caller-supplied `config.Assignment` values via `NewExplicitSnapshot`
2. **flag binding** — explicitly-set CLI flag values via `NewFlagSnapshot` (ExplicitlySet=true only)
3. **env** — injected environment variable values via `NewEnvSnapshot`
4. **JSON** — values from JSON readers or paths via `LoadJSON` / `LoadJSONFile`
5. **default** — registered definition defaults from `Set.DefaultSnapshot`

The first tier that provides a non-absent value for a key wins. A zero-value `Snapshot{}` for any tier is safe and equivalent to that tier having no values.

## Source Labels

The provenance label on a resolved value reports which tier provided it. Labels are fixed:

| Label | Constant | Tier |
|---|---|---|
| `default` | `config.SourceDefault` | definition default |
| `explicit setter` | `config.SourceExplicit` | explicit setter |
| `flag binding` | `config.SourceFlagBinding` | CLI flag binding |
| `env` | `config.SourceEnv` | environment variable |
| `JSON` | `config.SourceJSON` | JSON file or reader |

An absent key with no default returns an empty provenance label.

## Flag Default Exclusion

A flag's configured default does **not** enter the flag binding tier. Only flags with `ExplicitlySet: true` in their `FlagValue` binding contribute a value. This preserves the invariant that env and JSON values rank above flag defaults.

## Immutability

Source snapshots supplied to `Resolve` are never mutated. The returned snapshot is a new self-contained value. Callers may reuse the same source snapshots across multiple `Resolve` calls.

## Source Reports

`Snapshot.SourceReport` exposes the resolved winning source for every registered
key in definition order, including absent keys with an empty source label and
`IsSet() == false`. Report entries identify the canonical key, kind, source label,
redaction status, env name, JSON path, and JSON reader label where applicable.

`Snapshot.WriteSourceReport` renders the same report to a caller-supplied writer.
Reports are value-free: they never include raw config values, including
non-sensitive values. Use typed getters to retrieve values.

## Scope

Story 4.3 implements `NewFlagSnapshot` and `Resolve`. Story 4.4 implements typed
getters. Story 4.5 implements source reports and rendered diagnostics. Epic 5
owns compatibility tables, migration examples, and release-readiness evidence.

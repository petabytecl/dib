# Addendum: PRD dib

## Workflow Notes

- The required question tool was unavailable in Default mode during PRD creation, so the initial PRD was drafted in Fast path from the existing product brief. Those initial unresolved choices were later closed in `prd.md` section 12.
- No prior in-progress PRD workspace was found under `_bmad-output/planning-artifacts/prds`.
- The source brief is `_bmad-output/planning-artifacts/briefs/brief-dib-2026-06-10/brief.md`.

## Source Notes

- Go `flag` provides the dependency-free baseline: top-level helpers, `FlagSet`, typed flag helpers, custom `Value`, parse errors, `Args`, `Arg`, `NArg`, usage rendering, and `--` termination.
- pflag public docs position it as a drop-in replacement for Go `flag` with POSIX/GNU-style `--flags`, shorthand flags, shorthand grouping, no-option defaults, interspersed args before `--`, and support for importing Go flag sets.
- Cobra public docs frame CLI apps as Command trees with local and persistent flags, nested subcommands, aliases, help, suggestions, generated assets, and optional Viper integration.
- Viper public docs describe configuration precedence across explicit setters, flags, environment variables, config files, external key/value stores, and defaults; they also document broad format and remote-store support that Dib intentionally excludes from V1.

## Clean-Room Guardrails

- Allowed inputs: public documentation, package docs, observable behavior, independently written user stories, and independently written tests.
- Disallowed inputs: copied source, comments, examples, tests, fixtures, internal names, file organization, or non-public implementation details from Go `flag`, pflag, Cobra, Viper, or comparable projects.
- Compatibility claims should be behavior-scoped and test-backed. Dib should not claim full source compatibility with Cobra, pflag, or Viper in V1.

## Candidate Package Boundaries

- `dib/command`: Command tree, execution, args, help, usage, aliases, command-local validation.
- `dib/flag`: Flag sets, parsing, typed values, shorthand behavior, normalization, usage rendering.
- `dib/config`: Config key registry, source binding, precedence, env/file readers, typed getters.
- `dib/internal/text`: wrapping, alignment, usage formatting helpers.
- `dib/internal/errors`: shared typed errors if cross-package errors need stable inspection.

These package names are provisional. The architecture workflow should test whether public package names should be shorter, flatter, or more idiomatic before implementation.

## Deferred Technical Design Questions

- Whether command execution receives a `context.Context` at every command or only at the root execution boundary.
- Which concrete exported error types or sentinel values implement the public boundary error families.
- Whether JSON loading should preserve source location for diagnostics.
- Whether usage rendering should be stable enough for golden tests across Go versions, or whether tests should assert structured output before rendering.

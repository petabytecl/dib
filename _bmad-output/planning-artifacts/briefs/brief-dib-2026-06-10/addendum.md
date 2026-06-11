# Addendum: Dib Product Brief

## Source Notes

The product direction is grounded in public documentation and observable API positioning, not source copying.

- pflag describes itself as a drop-in replacement for Go's `flag` package with POSIX/GNU-style `--flags`. It adds shorthand helpers, shorthand grouping rules, no-option defaults, flag name normalization, hidden/deprecated metadata, disabled sorting, and support for importing Go `flag` sets.
- Cobra describes modern CLI applications as command, argument, and flag structures. Its public feature set includes nested commands, persistent and local flags, automatic help, suggestions, aliases, generated completion, generated man pages, and optional Viper integration.
- Viper describes configuration as a registry that can read defaults, explicit values, config files, environment variables, command-line flags, buffers, remote systems, live watching, and aliases. Its documented precedence is explicit values, flags, environment variables, config files, external stores, then defaults.
- Go's standard `flag` package provides the dependency-free baseline: basic flag definitions, `Parse`, `Args`, `Arg`, `NArg`, top-level functions, `FlagSet`, `Value`, and parse termination behavior.

## Clean-Room Guardrails

- Use public documentation, user-visible behavior, and independently written tests.
- Do not copy source, comments, examples, test cases, internal names, or file organization from Cobra, pflag, or Viper.
- Avoid claiming source compatibility unless the PRD explicitly narrows and tests that claim.
- Prefer native Dib API names where a copied name would imply implementation lineage.
- Record every intentional compatibility decision in the PRD or architecture artifact.

## Runtime Dependency Rule

The core runtime contract is "standard library only." This allows:

- `go test`, `go vet`, fuzzing, coverage, generators, linters, and local development tools to use external dependencies when isolated from runtime imports.
- Optional command-line tools in separate modules if they do not become dependencies of the library packages.
- Build tags or examples that demonstrate external integration only if the default module remains dependency-free.

This disallows in core:

- YAML/TOML parsers.
- File watching libraries.
- Shell completion frameworks.
- Remote key/value clients.
- Reflection decode libraries.
- Assertion libraries in runtime packages.

## PRD Questions To Resolve

- Does Dib need source-compatible function names for pflag-like APIs, or only familiar concepts?
- Should global command/flag state exist at all, or should it be a compatibility layer over explicit instances?
- Should config support case-insensitive keys, normalized keys, or exact keys only?
- What is the minimum file format set for version one: JSON only, or JSON plus a simple standard-library key/value format?
- How should environment variables map to nested config keys?
- Should flag values be resolved lazily when config is read, or eagerly during parse?
- What errors must be typed for callers: unknown flag, missing value, invalid value, duplicate shorthand, unknown command, argument validation failure, config file not found, config decode failure?
- What exact parser behavior is required around `--`, interspersed args, repeated flags, shorthand grouping, and boolean negation?

## Candidate Package Boundaries

- `dib/command`: command tree, execution, args, help, usage, aliases, command-local validation.
- `dib/flag`: flag sets, parsing, typed values, shorthand behavior, normalization, usage rendering.
- `dib/config`: value registry, source binding, precedence, env/file readers, typed getters.
- `dib/internal/text`: wrapping, alignment, usage formatting helpers.
- `dib/internal/errors`: shared typed errors if cross-package errors need stable inspection.

These package names are provisional and should be tested against the architecture workflow.

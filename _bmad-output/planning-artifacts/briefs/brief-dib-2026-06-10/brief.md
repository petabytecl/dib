---
title: "Dib: Standard-Library CLI, Flag, and Configuration Toolkit for Go"
status: draft
created: 2026-06-10
updated: 2026-06-10
---

# Product Brief: Dib

## Executive Summary

Dib is a clean-room Go toolkit for building command-line applications with three tightly integrated pieces: command routing, flag parsing, and configuration resolution. The goal is to give developers the ergonomic foundation they expect from the public behavior of Cobra, pflag, and Viper while keeping the runtime dependency graph at zero external packages.

The sharp constraint is the product: Dib must be useful precisely because it is boring to adopt. A team should be able to import it into infrastructure tools, internal CLIs, build tools, or security-sensitive projects without pulling in transitive dependencies. External dependencies are allowed only for development tooling, tests, or generators when isolated from the shipped runtime, matching the repo pattern used in adjacent Petabyte projects.

The first version should not try to clone every feature. It should establish a small, trustworthy core: POSIX/GNU-style flags, nested commands, help output, defaults, environment variables, JSON config files, clear precedence, and explicit error handling. Richer formats, live reload, shell completion, and code generation can come later if they do not violate the standard-library runtime contract.

## The Problem

Go's standard `flag` package is stable and dependency-free, but it is intentionally small. It supports basic flag definition and `FlagSet` usage, yet it does not provide modern CLI application structure, shorthand flag behavior, cascading command flags, or integrated configuration precedence.

Popular libraries fill those gaps. pflag adds POSIX/GNU-style long flags and shorthand behavior. Cobra adds command trees, help, aliases, suggestions, and generated assets. Viper adds defaults, explicit values, files, environment variables, flags, and precedence. Those libraries are productive, but adopting them means adopting their APIs, design assumptions, and dependency graph.

Dib serves teams that want the operational simplicity of the standard library and the developer experience of a modern CLI foundation. The pain is not that existing tools are bad. The pain is that they are often more than a small infrastructure CLI wants to carry.

## The Solution

Dib provides a compact set of packages that work together but can be used independently:

- A command package for nested commands, arguments, command-local execution, inherited flags, and help text.
- A flag package with standard-library-compatible concepts plus long flags, shorthand flags, boolean shorthand grouping, no-option defaults, normalized names, and parse errors that callers control.
- A config package that resolves values from defaults, explicit setters, flags, environment variables, and JSON files through a documented precedence model.

The product should feel familiar to Go developers. APIs should prefer small interfaces, explicit constructors, ordinary `io.Reader` and `io.Writer` integration, `context.Context` where execution crosses boundaries, and errors that can be inspected without string matching.

## What Makes This Different

The differentiator is not novelty. It is constraint discipline.

Dib is built for projects where zero runtime dependencies are a feature. That means no YAML/TOML parser in the core, no file watcher dependency, no shell-completion generator dependency, no reflection-heavy decode stack unless the standard library already provides the primitives, and no hidden global state as the default path.

The clean-room rule matters. Dib should be inspired by public docs, observable behavior, and user expectations, not copied implementation. Compatibility should mean "familiar and predictable," not "source-compatible replacement for Cobra, pflag, or Viper." When exact compatibility would inflate the design, the brief favors a smaller native API with clear migration notes.

## Who This Serves

Primary users are Go developers building operational CLIs: internal platform tools, deploy helpers, admin utilities, build wrappers, small daemons with command-line control surfaces, and repo-local tools.

Secondary users are maintainers who care about dependency auditability, supply-chain risk, reproducible builds, and long-lived maintenance. Success for them is not a flashy framework. It is a library they can understand, vendor if necessary, test thoroughly, and keep stable.

## Success Criteria

- Runtime dependency graph is limited to the Go standard library.
- A realistic multi-command CLI can be built without package-level global state.
- Flag parsing covers long flags, shorthand flags, shorthand grouping, boolean flags, non-boolean values, repeated flags, custom values, terminator handling, and useful parse diagnostics.
- Config resolution supports defaults, explicit values, flags, environment variables, and JSON files with a documented precedence order.
- The first PRD defines which pflag/Cobra/Viper behaviors are in scope, out of scope, and intentionally different.
- The implementation can be validated with standard `go test` and table-driven behavior tests without network access.
- Documentation includes clean-room source policy and migration examples from standard `flag`, pflag-style flags, Cobra-style command trees, and Viper-style config resolution.

## First-Version Scope

In scope:

- Go module for the Dib library.
- Runtime code using only the standard library.
- Command tree with nested commands, aliases, arguments, help, usage, and explicit execution errors.
- Flag parser with independent flag sets, top-level helpers only if they do not force global-state-first usage, shorthand support, no-option defaults, name normalization, hidden/deprecated metadata, and stable usage rendering.
- Config resolver with defaults, explicit setters, env binding, flag binding, JSON file loading, `io.Reader` loading, and precedence.
- Developer tooling may use external dependencies only when isolated from runtime imports.

Out of scope for the first version:

- Full source compatibility with Cobra, pflag, or Viper.
- YAML, TOML, HCL, dotenv, remote key/value stores, and live file watching in core.
- Generated shell completion, man pages, or project scaffolding.
- Reflection-heavy struct decoding unless justified later by the PRD and architecture.

## Key Assumptions

- [ASSUMPTION] The repeated pflag URL in the prompt meant pflag plus the adjacent Cobra/Viper problem space: CLI manager, flags parser, and config manager.
- [ASSUMPTION] The language target is Go because the cited inspirations are Go libraries and the zero-dependency constraint maps directly to Go's standard library.
- [ASSUMPTION] "Clean room clone" means independently designed behavior-compatible inspiration from public documentation, not copied source, test fixtures, comments, names, or internal structure.
- [ASSUMPTION] External dependencies are acceptable for repo-local tools only when they are not imported by runtime packages.

## Source Grounding

- [pflag](https://github.com/spf13/pflag) positions itself as a replacement for Go's `flag` package with POSIX/GNU-style flags and shorthand behavior.
- [Cobra](https://github.com/spf13/cobra) frames CLI applications around commands, arguments, and flags, with nested command support.
- [Viper](https://github.com/spf13/viper) models configuration as merged sources with precedence across explicit values, flags, environment variables, files, external stores, and defaults.
- [Go flag](https://pkg.go.dev/flag) is the standard-library baseline for basic flag parsing and `FlagSet` behavior.

## Vision

Dib becomes the dependency-free CLI foundation for Go projects that value explicitness, auditability, and boring maintenance. It should be small enough to read, strict enough to trust, and familiar enough that experienced Go developers do not need to learn a framework before shipping a reliable CLI.

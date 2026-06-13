# Behavior Matrices

This document is the consolidated adoption-evidence matrix for Dib V1 behavior.
Package tests, docs tests, migration examples, and the dependency gate remain
the executable source of truth. The matrix gives reviewers a single audit path
for command, flag, config, diagnostic, redaction, compatibility, migration, and
dependency behavior without requiring them to reverse-engineer every test first.

## How To Audit This Matrix

- **Story coverage** identifies the implementation story or evidence story that
  owns the behavior.
- **FR/NFR trace** maps the row to product requirements and quality constraints.
- **Expected behavior** states only behavior that is supported by executable
  evidence today, or explicitly marks later work as deferred.
- **Executable evidence** names local files and Go `Test...`, `Fuzz...`, or
  `Example...` functions where practical. Function names are cited only when
  they exist in the repository.
- **Status** is `current` for implemented behavior or captured evidence,
  `deferred` for future scope, and never means release-candidate approval.
  Story 5.4 owns final release evidence for this v0 candidate.

Dib is a clean-room native Go API. It is not a source-compatible clone, not a
drop-in replacement, and not a framework compatibility layer for Go `flag`,
pflag, Cobra, Viper, or comparable projects. Behavior-scoped familiarity is
documented as supported, narrowed, omitted, or intentionally different in
`docs/compatibility.md`.

## Consolidated Adoption Evidence

| Behavior family | Story coverage | FR/NFR trace | Expected behavior | Executable evidence | Status |
| --- | --- | --- | --- | --- | --- |
| Shared immutability, snapshots, and explicit instances | Story 2.1, Story 3.1, Story 3.5, Story 4.1, Story 4.2 | FR20, NFR4 | Command definitions, flag sets, config sets, route results, parse snapshots, config source snapshots, and boundary metadata are caller-owned values. Constructors and derivation APIs avoid package-global mutable state, clone caller-owned slices and values, and do not read ambient process state unless explicit in the API. | `flags/state_atdd_test.go` `TestIndependentFlagSetsIgnoreAmbientProcessState`; `flags/set_test.go` `TestSetWithReturnsIndependentSet`; `command/contract_test.go` `TestRoutingUsesExplicitInputsAndReturnedValues`; `command/boundary_test.go` `TestRouteBoundaryPackagesExplicitCallerInputs`; `config/snapshot_test.go` `TestDefaultSnapshotIsReusableAndDefensive`; `config/source_test.go` `TestExplicitSnapshotLastWriterWinsAndDefensiveValues` | current |
| Flag definitions and parsing | Story 2.1, Story 2.2, Story 2.3, Story 2.4, Story 2.5, Story 2.6, Story 2.8 | FR5, FR6, FR7, FR8, FR10, FR22, NFR3, NFR4, NFR8 | Flag sets expose immutable definitions; long names match exactly unless a caller supplies a normalizer; shorthands remain independent from long-name normalization; long flags, short flags, shorthand groups, no-option defaults, repeated values, custom values, typed parse failures, and sensitive-value redaction are implemented through caller-supplied args. | `flags/set_atdd_test.go` `TestDefinitionsExposeMetadata`; `flags/normalize_test.go` `TestNameNormalizationLookupModes`; `flags/parse_long_test.go` `TestParseLongFlagsCoversContractMatrix`; `flags/parse_shorthand_test.go` `TestParseShorthandValuesAndBooleanPresence`; `flags/parse_group_test.go` `TestParseShortGroupBooleanMembers`; `flags/repeated_test.go` `TestParseRepeatableValuesAccumulateAcrossSpellings`; `flags/repeated_test.go` `TestParseSensitiveCustomFailureRedactsRawValueAndCause` | current |
| Parser boundaries | Story 2.7 | FR9, NFR3, NFR4 | Interspersed positionals stay in relative order, `--` terminates flag parsing and preserves later tokens verbatim, unregistered `--help` and `-h` return `ErrHelpRequest` for caller-controlled rendering, and failed parses return zero-value snapshots with inspectable `*flags.ParseError` diagnostics. | `flags/parse_boundary_test.go` `TestParseBoundaryInterspersedPositionals`; `flags/parse_boundary_test.go` `TestParseBoundaryTerminator`; `flags/parse_boundary_test.go` `TestParseBoundaryFailedParses`; `flags/parse_boundary_test.go` `TestParseHelpLongUnregistered`; `flags/parse_boundary_test.go` `TestParseHelpRequestZeroValueSnapshot` | current |
| Fuzz and property hardening | Story 2.5, Story 2.7, Story 2.8 | FR22, NFR2, NFR4 | Standard-library fuzz targets assert parser invariants under arbitrary input: no panic, deterministic repeated parsing, no mutation of reusable definitions, defensive snapshot values, zero-value snapshots on failure, and typed `*flags.ParseError` failures. Corpus files are independently written clean-room inputs. | `flags/fuzz_test.go` `FuzzParse`; `flags/fuzz_test.go` `FuzzParseBoundary`; `flags/parse_group_test.go` `FuzzParseShortGroups`; `flags/testdata/fuzz/FuzzParse/`; `flags/testdata/fuzz/FuzzParseBoundary/` | current |
| Command routing and aliases | Story 3.1, Story 3.2 | FR1, FR20, NFR3, NFR4 | Command routing accepts explicit args, returns deterministic route snapshots with canonical path names, raw match tokens, and remaining args, resolves direct child aliases, and reports unknown commands or ambiguous alias setup through typed errors instead of string matching. | `command/route_test.go` `TestRouteRootAndNestedCommands`; `command/route_test.go` `TestRouteAliasesExposeCanonicalPathAndRawMatchTokens`; `command/route_test.go` `TestRouteUnknownCommandErrorsAreInspectable`; `command/definition_test.go` `TestDefinitionRejectsAliasTokenCollisions`; `command/errors_test.go` `TestTokenConflictErrorParentPathIsDefensive` | current |
| Local and inherited command flags | Story 3.3 | FR3, FR20, NFR3, NFR4 | Command definitions compose inherited flags root-to-leaf and final-command local flags last. Sibling local flags and ancestor local flags remain isolated, route results expose the composed flag set and parsed snapshot, and long-name, shorthand, or normalized conflicts return inspectable composition diagnostics. | `command/flags_test.go` `TestRouteComposesInheritedAndLocalFlags`; `command/flags_test.go` `TestRouteKeepsSiblingAndAncestorLocalFlagsIsolated`; `command/flags_test.go` `TestRouteFlagCompositionConflictsAreInspectable`; `command/flags_test.go` `TestRouteFlagSnapshotsAreDefensiveAndReusable` | current |
| Help and usage rendering | Story 3.4 | FR4, FR20, NFR4, NFR8 | Command definitions and route results render deterministic usage/help text only to caller-supplied writers. Structured tests assert route and metadata behavior; rendering tests intentionally compare deterministic output fragments and writer-error behavior. Hidden flags are omitted from output and sensitive defaults or parse values are not rendered. | `command/help_test.go` `TestWriteHelpRendersDefinitionMetadataDeterministically`; `command/help_test.go` `TestWriteHelpRendersRoutedCommandFlagsAndUsage`; `command/help_test.go` `TestWriteHelpOmitsHiddenAndSensitiveValues`; `command/help_test.go` `TestWriteHelpUsesSuppliedWriterOnly`; `command/help_qa_test.go` `TestWriteUsagePropagatesWriterFailuresAndRejectsInvalidTargets` | current |
| Execution-boundary metadata | Story 3.5 | FR2, FR20, NFR3, NFR4 | `command.Boundary` packages route results with caller-supplied context, args, stdout, and stderr metadata. It does not execute callbacks, write diagnostics, call `os.Exit`, decide exit policy, convert ordinary caller errors, or duplicate mutable route state. Zero-value boundaries report absent state through boolean accessors. | `command/boundary_qa_test.go` `TestRouteBoundaryPassesHelpRequestsWithoutRendering`; `command/boundary_qa_test.go` `TestRouteBoundaryLeavesRenderingUnderCallerControl`; `command/boundary_test.go` `TestBoundaryRetainsWritersWithoutWriting`; `command/boundary_test.go` `TestBoundaryKeepsOrdinaryCallerErrorsSeparate`; `command/boundary_test.go` `TestBoundaryZeroValueIsAbsentState` | current |
| Config definitions and source ingestion | Story 4.1, Story 4.2 | FR11, FR12, FR14, FR15, FR20, NFR3, NFR4, NFR8 | Config definitions expose key metadata, type expectations, defaults, usage text, and sensitivity metadata. Exact key matching is default; configured normalization rejects collisions. Explicit assignments, injected env lookup, and explicit JSON readers/paths produce source snapshots without resolving cross-source precedence and report typed source diagnostics. | `config/definition_test.go` `TestDefinitionsExposeMetadataAndDefensiveDefaults`; `config/set_test.go` `TestSetExactLookupAndDefinitionsAreDeterministic`; `config/set_test.go` `TestNormalizedSetLookupAndCollisions`; `config/env_test.go` `TestEnvSnapshotUsesInjectedLookupAndTracksMetadata`; `config/json_test.go` `TestJSONReaderSnapshotStrictAndPermissiveModes`; `config/source_test.go` `TestExplicitSnapshotValidationErrorsAreTypedAndNoPartialSnapshot` | current |
| Config precedence and typed getters | Story 4.3, Story 4.4 | FR12, FR13, FR16, FR20, NFR3, NFR4, NFR8 | `config.Resolve` applies the documented order `explicit setter > flag binding > env > JSON > default`. Flag binding contributes only explicitly-set CLI values. Typed getters return values or inspectable `*config.GetError` failures; `IsSet` distinguishes absent state from explicit zero, empty string, and empty list values. | `docs/config-precedence.md`; `config/resolve_test.go` `TestResolvePrecedenceAdjacentPairs`; `config/resolve_test.go` `TestResolveFlagDefaultDoesNotOverrideEnv`; `config/resolve_test.go` `TestSourceFlagBindingConstant`; `config/getter_test.go` `TestGetStringReturnedValueAndProvenance`; `config/getter_test.go` `TestGetErrorSentinelCategories`; `config/getter_test.go` `TestGetStringListReturnsDefensiveCopy` | current |
| Source reports, diagnostics, and redaction | Story 4.5 | FR16, FR20, NFR3, NFR4, NFR8 | Source reports are deterministic, value-free, definition-ordered, and include absent registered keys. Rendered diagnostics are deterministic, value-free, caller-writer-bound, and distinguish source labels from error categories. The redaction corpus is named only as fake sensitive corpus: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`; rendered examples and value-bearing matrix prose must not expose those raw values as output claims. | `docs/diagnostics-and-errors.md`; `config/report_test.go` `TestSourceReportCoversWinningSourcesAndAbsentKeys`; `config/report_test.go` `TestSourceReportRenderingIsDeterministicValueFreeAndWriterBound`; `config/report_test.go` `TestInspectDiagnosticClassifiesConfigErrors`; `config/report_test.go` `TestWriteDiagnosticIsDeterministicValueFreeAndWriterBound`; `config/report_test.go` `TestDiagnosticFalsePositiveRedactionCoverage`; `config/qa_e2e_test.go` `TestQAConfigProvenanceRenderingRedactsSensitiveCorpus` | current |
| Compatibility boundaries | Story 5.1 | FR17, FR18, FR20, FR21, NFR1, NFR7 | Familiar Go `flag`, pflag, Cobra, and Viper concepts are documented as supported, narrowed, omitted, or intentionally different. Dib does not claim source compatibility, a compatible replacement, a clone API, or a framework compatibility layer. | `docs/compatibility.md`; `docs/compatibility_test.go` `TestCompatibilityDocumentBoundaries`; `docs/compatibility_test.go` `TestCompatibilityDocumentDoesNotMakePositiveCompatibilityClaims`; `docs/compatibility_test.go` `TestCompatibilityTableCoversStorySurfaces`; `docs/provenance-log.md` | current |
| Migration examples | Story 5.2 | FR18, FR19, FR21, NFR1, NFR2, NFR3, NFR5, NFR6, NFR7, NFR8 | Executable migration examples demonstrate familiar standard `flag`, pflag-style, Cobra-style, and Viper-style concepts through Dib's native API. Examples use explicit instances, caller-supplied args/readers/writers/env lookup, typed errors, deterministic help, and redaction-safe source reports. | `examples/migration/standard_flag_concepts_test.go` `Example_standardFlagConcepts` `TestStandardFlagConceptsInspectParseErrors`; `examples/migration/shorthand_flag_migration_test.go` `Example_shorthandFlagMigration` `TestShorthandFlagMigrationIntentionalDifferences`; `examples/migration/nested_command_migration_test.go` `Example_nestedCommandMigration` `TestNestedCommandMigrationRoutesAliasesFlagsAndHelp`; `examples/migration/config_precedence_migration_test.go` `Example_configPrecedenceMigration` `TestConfigPrecedenceMigrationReportsAreValueFreeAndRedacted`; `go test ./examples/migration` | current |
| Dependency gate evidence | Story 5.3, Story 6.1, Story 6.2, Story 7.4 | FR21, FR23, FR24, NFR1 | Dependency, lint, and coverage behavior are adoption evidence. Runtime packages, tests, examples, and tools stay standard-library-only unless architecture changes. The local dependency gate checks package imports, CI runs the same dependency gate after tests and vet, Story 6.1 adds a lint gate (now `golangci-lint`, configured by `.golangci.yml`) for deterministic formatting and static analysis, Story 6.2 adds a package-aware coverage gate, and Story 7.4 extends that gate to the four public runtime packages (`cli`, `command`, `config`, `flags`). The root module has no `require`, `replace`, or `toolchain` directive and no root go sum file. | `go.mod`; `.github/workflows/ci.yml`; `docs/testing.md`; `tools/depgate/main_test.go` `TestDepgateFixtures`; `tools/depgate/main_test.go` `TestDepgateReportsEveryViolationDeterministically`; `tools/depgate/main_test.go` `TestDepgateDisablesWorkspaceMode`; `tools/coverage/main_test.go` `TestCoveragePassesPackagesMeetingThreshold`; `tools/coverage/main_test.go` `TestCoverageFailsPackagesBelowThreshold`; `tools/coverage/main_test.go` `TestCoverageCommandRunsFromRepositoryRoot`; `go run ./tools/depgate`; `golangci-lint run`; `go run ./tools/coverage`; `docs/release-checklist.md` | current |
| Release-candidate evidence | Story 5.4 | FR21, NFR1, NFR4 | The v0 evidence package records exact commit, tag candidate, Go version alignment, CI runner/action versions, test/vet/dependency-gate/race outcomes, parser mutation-fuzz outcomes, docs/examples evidence, final provenance review, compatibility review, migration review, dependency evidence, and waiver status. It is evidence for human release review, not tag approval. | `docs/release-checklist.md`; `docs/release-notes-v0.md`; `docs/release_checklist_test.go` | current |
| Public usage documentation | Story 6.3 | FR25, NFR12 | `README.md` provides install/import guidance, package roles, quickstart flag/command/config/`cli` composition usage, distributed command tree registration, application-owned dispatch guidance, v0 experimental API status, and links to compatibility, behavior matrix, diagnostics, config precedence, testing, and release evidence docs. Does not imply source compatibility with Go `flag`, pflag, Cobra, or Viper. | `docs/readme_test.go` `TestREADMEExistsAndCoversAdoptionOnboarding`; `docs/readme_test.go` `TestREADMEDoesNotImplySourceCompatibility`; `README.md` | current |
| Release hardening reconciliation | Story 6.4 | FR23, FR24, FR25, NFR11 | Epic 6 scope (lint gate, coverage validation, public usage documentation) is reconciled as formal release gates in `docs/release-notes-v0.md` and `docs/release-checklist.md`. Sprint-status.yaml records all Epic 6 stories. GitHub issues for Epic 6 align with local sprint status. Any accepted gate waiver records owner, reason, expiry, and impact. | `docs/release-checklist.md`; `docs/release_checklist_test.go` `TestReleaseChecklistRecordsReleaseCandidateEvidence` | current |
| CLI composition ergonomics | Story 7.1, Story 7.2, Story 7.3, Story 7.4, Issue #52 | FR26, FR25, FR20, NFR2, NFR5 | `cli.Invocation` names the invocation boundary; `cli.Plan` carries root command, config set, source snapshots, and flag bindings; `cli.Resolve` routes, builds flag-tier snapshot, resolves config, and returns `cli.Result` as the low-level inspectable path. `cli.New` builds a distributed command tree, `cli.Command.Run` resolves the matched route with route-scoped bindings, and invokes exactly one `func(cli.CommandContext) error` handler without calling `os.Exit`, reading env implicitly, loading files, or writing streams. | `cli/resolve_test.go` `TestResolveSuccessPathFlagBindingWinsOverLowerPrecedence`; `cli/resolve_qa_test.go` `TestQAResolveFullPrecedenceChainWithAllSourceTiers`; `cli/command_run_test.go` `TestCommandRunDispatchesDistributedSubcommand`; `cli/command_run_test.go` `TestCommandRunReturnsNoHandlerErrorForMatchedCommandWithoutHandler`; `examples/multicommand/example_test.go` `Example_composedCLI`; `examples/multicommand/example_test.go` `Example_dispatchStartStop`; `examples/multicommand/example_test.go` `Example_lowLevelDispatch` | current |

## Flag Parser Evidence Map

The consolidated table above is the authoritative Story 5.3 matrix. This
section preserves the historical anchor used by compatibility links and points
reviewers to the Epic 2 parser rows in the consolidated matrix:

- Shared immutability, snapshots, and explicit instances.
- Flag definitions and parsing.
- Parser boundaries.
- Fuzz and property hardening.
- Source reports, diagnostics, and redaction.
- Compatibility boundaries and migration examples.

For parser-specific audit, start with `flags/parse_long_test.go`,
`flags/parse_shorthand_test.go`, `flags/parse_group_test.go`,
`flags/repeated_test.go`, `flags/parse_boundary_test.go`, and
`flags/fuzz_test.go`. Package tests are executable evidence; this document is
the review index.

## Rendering And Diagnostic Evidence

Human-facing output has two evidence categories:

- **Structured assertions** verify machine-readable contracts such as typed
  errors, route snapshots, config source reports, source labels, redaction
  flags, and writer ownership. Examples include
  `TestRouteUnknownCommandErrorsAreInspectable`,
  `TestParseBoundaryFailedParses`,
  `TestInspectDiagnosticClassifiesConfigErrors`, and
  `TestQAConfigProvenanceReportsExplainWinningSourcesWithoutValues`.
- **Deterministic rendering assertions** intentionally inspect rendered help,
  usage, source reports, diagnostics, or example output. Examples include
  `TestWriteHelpRendersDefinitionMetadataDeterministically`,
  `TestWriteHelpRendersRoutedCommandFlagsAndUsage`,
  `TestSourceReportRenderingIsDeterministicValueFreeAndWriterBound`,
  `TestWriteDiagnosticIsDeterministicValueFreeAndWriterBound`,
  `Example_standardFlagConcepts`, `Example_nestedCommandMigration`, and
  `Example_configPrecedenceMigration`.

Rendered diagnostic strings are human-facing evidence only. Programmatic
contracts are the structured tests, typed errors, snapshots, source reports, and
matrix rows referenced above. Sensitive values remain absent from rendered
artifacts; the fake sensitive corpus is named only for redaction auditing.

## Dependency And Release Evidence

Adoption evidence for dependency behavior started in Story 5.3 and is current
through the Epic 7 CLI composition evidence:

- `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`.
- The root repository has no go sum file.
- Root `go.mod` has no `require`, `replace`, or `toolchain` directives.
- `.github/workflows/ci.yml` runs `golangci-lint` via
  `golangci/golangci-lint-action@v7` (pinned to `v2.10.1`), `go test ./...`,
  `go vet ./...`, `go run ./tools/coverage`, and `go run ./tools/depgate` on
  `ubuntu-24.04` with the Go version read from `go.mod`.
- `tools/depgate/` remains the executable local dependency gate.
- The lint gate runs `golangci-lint` (an external, pinned CI binary, never
  imported) configured by `.golangci.yml`; the `depguard` linter keeps the
  module dependency-free.
- `tools/coverage/` remains the executable local coverage gate, imports only
  the Go standard library, and enforces thresholds for `cli`, `command`,
  `config`, and `flags`.
- `docs/release-checklist.md` records the consolidated behavior matrix and the
  Story 5.4 release-candidate outcomes for human release review.

Story 5.4 records exact release-candidate command results, race-test evidence,
mutation-fuzz evidence, final provenance review, exact commit, waiver status,
and the tag candidate. The final tag action remains a human release-review
decision.

## Current Scope Boundaries

This matrix covers implemented behavior from Epics 2 through 4, adoption,
release, and public-onboarding evidence from Stories 5.1 through 6.4, and CLI
composition evidence from Epic 7. It does not add source-compatible adapters,
package-global helpers, `/cmd` scaffolding, generated assets, new config
formats, process-owning behavior, or release-ready claims.

Shell completion, manpages, scaffolding, generated command assets, additional
config formats, remote stores, live reload, config aliases, reflection-heavy
struct decoding, framework callback models, and final release readiness remain
deferred unless a future story or architecture decision explicitly approves
that scope.

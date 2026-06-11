# Clean-Room Policy

Dib is a clean-room Go toolkit. It may provide familiar behavior, but it is not
a source-compatible clone, drop-in replacement, or compatibility layer for Go
`flag`, pflag, Cobra, Viper, or comparable projects.

## Allowed Inputs

Contributors may use:

- Public documentation and package documentation for behavior understanding and
  factual metadata. Do not copy wording, examples, names, layout, or structure
  from those documents.
- Independently observed runtime behavior.
- Independently written user stories, requirements, and tests.
- Project-authored planning artifacts, architecture notes, and review findings
  whose external-source provenance is already recorded or whose inputs are
  project-owned.
- Small factual metadata such as source name, URL, access date, and license or
  terms when recording provenance.

## Disallowed Inputs

Contributors must not copy or closely derive from:

- Source code, tests, comments, examples, fixtures, or generated output from
  inspiration projects.
- Internal names, private implementation details, file organization, or README
  structure from inspiration projects.
- Non-public material or material whose terms do not permit the intended use.
- AI-generated content that is too closely derived from third-party source,
  examples, fixtures, names, comments, or structure.

## Provenance Requirement

Use `docs/provenance-log.md` before acceptance when an artifact is copied,
adapted, generated from a reference, or materially derived from a reference.
Inspiration-only references must not contribute copied names, examples,
fixtures, file layout, or source-derived structure. When the boundary is
unclear, record the source and classification before review.

Missing provenance blocks acceptance for copied, generated, adapted, or
reference-derived artifacts until the gap is resolved.

## Contributor Workflow

1. Classify every external input before writing.
2. Prefer behavior observations and project-owned requirements over third-party
   implementation details.
3. Write Dib code, tests, examples, fixtures, and docs independently.
4. Record provenance for anything copied, adapted, generated from, or materially
   derived from a reference.
5. Keep compatibility language behavior-scoped and test-backed where practical.

Acceptable language includes "familiar behavior", "behavior-scoped", and
"native Dib API". Avoid language such as "source-compatible", "drop-in
replacement", "clone API", or "framework compatibility layer".

See `CONTRIBUTING.md` for the short contributor checklist.

package config

import (
	"errors"
	"fmt"
	"io"
)

// SourceReportEntry is a value-free provenance row for one registered config key.
type SourceReportEntry struct {
	key             string
	kind            Kind
	set             bool
	sourceLabel     string
	envName         string
	jsonPath        string
	jsonReaderLabel string
	redacted        bool
}

// Key returns the canonical registered config key for this report entry.
func (e SourceReportEntry) Key() string {
	return e.key
}

// Kind returns the registered value kind for this report entry.
func (e SourceReportEntry) Kind() Kind {
	return e.kind
}

// IsSet reports whether the resolved snapshot has a value for this key.
func (e SourceReportEntry) IsSet() bool {
	return e.set
}

// SourceLabel returns the winning provenance label, or empty for absent keys.
func (e SourceReportEntry) SourceLabel() string {
	return e.sourceLabel
}

// EnvName returns the environment variable name for env-sourced values.
func (e SourceReportEntry) EnvName() string {
	return e.envName
}

// JSONPath returns the caller-supplied JSON file path for file-sourced values.
func (e SourceReportEntry) JSONPath() string {
	return e.jsonPath
}

// JSONReaderLabel returns the caller-supplied JSON reader label for reader-sourced values.
func (e SourceReportEntry) JSONReaderLabel() string {
	return e.jsonReaderLabel
}

// Redacted reports whether raw values for this key are sensitive.
func (e SourceReportEntry) Redacted() bool {
	return e.redacted
}

func (e SourceReportEntry) String() string {
	return fmt.Sprintf("config.SourceReportEntry{key:%q kind:%s set:%t source:%q redacted:%t}", e.key, e.kind, e.set, e.sourceLabel, e.redacted)
}

// GoString renders the entry in Go syntax for the %#v verb.
func (e SourceReportEntry) GoString() string {
	return e.String()
}

// SourceReport returns value-free provenance entries in snapshot definition order.
func (s Snapshot) SourceReport() []SourceReportEntry {
	report := make([]SourceReportEntry, 0, len(s.values))
	for _, value := range s.values {
		def, ok := value.Definition()
		if !ok {
			continue
		}
		_, hasValue := value.Value()
		source := value.Source()
		label := ""
		if hasValue {
			label = canonicalSourceLabel(value.Provenance())
		}
		report = append(report, SourceReportEntry{
			key:             def.Name(),
			kind:            def.Kind(),
			set:             hasValue,
			sourceLabel:     label,
			envName:         source.EnvName(),
			jsonPath:        source.JSONPath(),
			jsonReaderLabel: source.JSONReaderLabel(),
			redacted:        def.Sensitive(),
		})
	}
	return report
}

// WriteSourceReport renders value-free source provenance to the caller-supplied writer.
func (s Snapshot) WriteSourceReport(w io.Writer) error {
	if w == nil {
		return errors.New("config: nil source report writer")
	}
	for _, entry := range s.SourceReport() {
		if _, err := fmt.Fprintf(
			w,
			"key=%q kind=%s set=%t source=%q redacted=%t env=%q json_path=%q json_reader=%q\n",
			entry.Key(),
			entry.Kind(),
			entry.IsSet(),
			entry.SourceLabel(),
			entry.Redacted(),
			entry.EnvName(),
			entry.JSONPath(),
			entry.JSONReaderLabel(),
		); err != nil {
			return err
		}
	}
	return nil
}

// Diagnostic is a structured, value-free view of public config diagnostics.
type Diagnostic struct {
	category        error
	key             string
	kind            Kind
	wantKind        Kind
	sourceLabel     string
	envName         string
	jsonPath        string
	jsonReaderLabel string
	redacted        bool
	hasSafeCause    bool
}

// Category returns the sentinel category for the diagnostic.
func (d Diagnostic) Category() error {
	return d.category
}

// Key returns the config key associated with the diagnostic.
func (d Diagnostic) Key() string {
	return d.key
}

// Kind returns the actual or expected registered kind associated with the diagnostic.
func (d Diagnostic) Kind() Kind {
	return d.kind
}

// WantKind returns the getter-requested kind for retrieval conversion diagnostics.
func (d Diagnostic) WantKind() Kind {
	return d.wantKind
}

// SourceLabel returns the attempted or winning source label when available.
func (d Diagnostic) SourceLabel() string {
	return d.sourceLabel
}

// EnvName returns env metadata associated with the diagnostic.
func (d Diagnostic) EnvName() string {
	return d.envName
}

// JSONPath returns JSON file path metadata associated with the diagnostic.
func (d Diagnostic) JSONPath() string {
	return d.jsonPath
}

// JSONReaderLabel returns JSON reader label metadata associated with the diagnostic.
func (d Diagnostic) JSONReaderLabel() string {
	return d.jsonReaderLabel
}

// Redacted reports whether raw values for this diagnostic are sensitive.
func (d Diagnostic) Redacted() bool {
	return d.redacted
}

// HasSafeCause reports whether the original error exposes an underlying cause.
func (d Diagnostic) HasSafeCause() bool {
	return d.hasSafeCause
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("config.Diagnostic{category:%q key:%q kind:%s want:%s source:%q redacted:%t safe_cause:%t}", diagnosticCategoryString(d.category), d.key, d.kind, d.wantKind, d.sourceLabel, d.redacted, d.hasSafeCause)
}

// GoString renders the diagnostic in Go syntax for the %#v verb.
func (d Diagnostic) GoString() string {
	return d.String()
}

// InspectDiagnostic extracts structured config diagnostic state from public config errors.
func InspectDiagnostic(err error) (Diagnostic, bool) {
	if err == nil {
		return Diagnostic{}, false
	}

	var definitionErr *DefinitionError
	if errors.As(err, &definitionErr) {
		return Diagnostic{
			category:    definitionErr.Category(),
			key:         definitionErr.Key(),
			kind:        definitionErr.Kind(),
			sourceLabel: canonicalSourceLabel(definitionErr.Provenance()),
			redacted:    definitionErr.Redacted(),
		}, true
	}

	var sourceErr *SourceError
	if errors.As(err, &sourceErr) {
		return Diagnostic{
			category:        sourceErr.Category(),
			key:             sourceErr.Key(),
			kind:            sourceErr.Kind(),
			sourceLabel:     canonicalSourceLabel(sourceErr.Source()),
			envName:         sourceErr.EnvName(),
			jsonPath:        sourceErr.JSONPath(),
			jsonReaderLabel: sourceErr.JSONReaderLabel(),
			redacted:        sourceErr.Redacted(),
			hasSafeCause:    sourceErr.Cause() != nil,
		}, true
	}

	var getErr *GetError
	if errors.As(err, &getErr) {
		return Diagnostic{
			category:    getErr.Category(),
			key:         getErr.Key(),
			kind:        getErr.Kind(),
			wantKind:    getErr.WantKind(),
			sourceLabel: canonicalSourceLabel(getErr.SourceLabel()),
			redacted:    getErr.Redacted(),
		}, true
	}

	return Diagnostic{}, false
}

// WriteDiagnostic renders a deterministic value-free diagnostic to the caller-supplied writer.
func WriteDiagnostic(w io.Writer, err error) error {
	if w == nil {
		return errors.New("config: nil diagnostic writer")
	}
	diagnostic, ok := InspectDiagnostic(err)
	if !ok {
		_, writeErr := fmt.Fprintln(w, "config diagnostic: unsupported")
		return writeErr
	}
	_, writeErr := fmt.Fprintf(
		w,
		"config diagnostic: category=%q key=%q kind=%s want_kind=%s source=%q redacted=%t safe_cause=%t env=%q json_path=%q json_reader=%q\n",
		diagnosticCategoryString(diagnostic.Category()),
		diagnostic.Key(),
		diagnostic.Kind(),
		diagnostic.WantKind(),
		diagnostic.SourceLabel(),
		diagnostic.Redacted(),
		diagnostic.HasSafeCause(),
		diagnostic.EnvName(),
		diagnostic.JSONPath(),
		diagnostic.JSONReaderLabel(),
	)
	return writeErr
}

func canonicalSourceLabel(label string) string {
	switch label {
	case SourceDefault, SourceExplicit, SourceFlagBinding, SourceEnv, SourceJSON:
		return label
	default:
		return ""
	}
}

func diagnosticCategoryString(category error) string {
	if category == nil {
		return ""
	}
	return category.Error()
}

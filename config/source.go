package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Assignment is one ordered explicit config source write.
type Assignment struct {
	Key   string
	Value any
}

// NewExplicitSnapshot ingests caller-supplied key/value assignments.
func NewExplicitSnapshot(set Set, assignments ...Assignment) (Snapshot, error) {
	snapshot := newSourceSnapshot(set)
	for _, assignment := range assignments {
		def, index, ok := lookupSourceDefinition(set, assignment.Key)
		if !ok {
			return Snapshot{}, newSourceError(assignment.Key, SourceExplicit, "", "", "", Kind(-1), "", false, ErrUnknownSourceKey, nil)
		}
		value, err := explicitValue(def, assignment.Value)
		if err != nil {
			return Snapshot{}, err
		}
		snapshot.values[index] = newSourceValue(def, value, Source{label: SourceExplicit})
	}
	return snapshot, nil
}

// EnvLookup is an injected environment lookup function.
type EnvLookup func(string) (string, bool)

// EnvBinding binds a registered config key to an environment variable name.
type EnvBinding struct {
	key      string
	envName  string
	explicit bool
}

// BindEnv binds key to the explicit environment variable name.
func BindEnv(key, envName string) EnvBinding {
	return EnvBinding{key: key, envName: envName, explicit: true}
}

// MapEnv binds key to a derived environment name using EnvPrefix and EnvKeyReplacer.
func MapEnv(key string) EnvBinding {
	return EnvBinding{key: key}
}

type envOptions struct {
	prefix   string
	replacer *strings.Replacer
}

// EnvOption configures derived env binding names.
type EnvOption func(*envOptions)

// EnvPrefix prefixes derived environment variable names.
func EnvPrefix(prefix string) EnvOption {
	return func(options *envOptions) {
		options.prefix = prefix
	}
}

// EnvKeyReplacer replaces key text before uppercasing and applying EnvPrefix.
func EnvKeyReplacer(replacer *strings.Replacer) EnvOption {
	return func(options *envOptions) {
		options.replacer = replacer
	}
}

// NewEnvSnapshot ingests config values from an injected environment lookup.
func NewEnvSnapshot(set Set, lookup EnvLookup, bindings []EnvBinding, opts ...EnvOption) (Snapshot, error) {
	options := envOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	if lookup == nil {
		return Snapshot{}, newSourceError("", SourceEnv, "", "", "", Kind(-1), "", false, ErrInvalidSource, nil)
	}

	snapshot := newSourceSnapshot(set)
	for _, binding := range bindings {
		def, index, ok := lookupSourceDefinition(set, binding.key)
		envName := binding.envName
		if !binding.explicit {
			envName = deriveEnvName(binding.key, options)
		}
		if !ok {
			return Snapshot{}, newSourceError(binding.key, SourceEnv, envName, "", "", Kind(-1), "", false, ErrUnknownSourceKey, nil)
		}
		if envName == "" {
			return Snapshot{}, newSourceError(def.key, SourceEnv, envName, "", "", def.kind, "", def.sensitive, ErrInvalidSource, nil)
		}
		raw, present := lookup(envName)
		if !present {
			continue
		}
		value, err := convertStringValue(def, raw, SourceEnv, envName, "", "")
		if err != nil {
			return Snapshot{}, err
		}
		snapshot.values[index] = newSourceValue(def, value, Source{label: SourceEnv, envName: envName})
	}
	return snapshot, nil
}

func deriveEnvName(key string, options envOptions) string {
	name := key
	if options.replacer != nil {
		name = options.replacer.Replace(name)
	}
	return options.prefix + strings.ToUpper(name)
}

type jsonOptions struct {
	permissive  bool
	readerLabel string
}

// JSONOption configures JSON source ingestion.
type JSONOption func(*jsonOptions)

// JSONPermissive ignores unknown JSON object keys while validating registered keys.
func JSONPermissive() JSONOption {
	return func(options *jsonOptions) {
		options.permissive = true
	}
}

// JSONReaderLabel records a caller-supplied label for JSON reader source values.
func JSONReaderLabel(label string) JSONOption {
	return func(options *jsonOptions) {
		options.readerLabel = label
	}
}

// LoadJSON ingests config values from a caller-supplied JSON reader.
func LoadJSON(set Set, reader io.Reader, opts ...JSONOption) (Snapshot, error) {
	options := resolveJSONOptions(opts)
	if reader == nil {
		return Snapshot{}, newSourceError("", SourceJSON, "", "", options.readerLabel, Kind(-1), "", false, ErrInvalidSource, nil)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return Snapshot{}, newSourceError("", SourceJSON, "", "", options.readerLabel, Kind(-1), "", false, ErrSourceRead, err)
	}
	return loadJSONBytes(set, data, "", options)
}

// LoadJSONFile ingests config values from a caller-supplied JSON file path.
func LoadJSONFile(set Set, path string, opts ...JSONOption) (Snapshot, error) {
	options := resolveJSONOptions(opts)
	if path == "" {
		return Snapshot{}, newSourceError("", SourceJSON, "", path, "", Kind(-1), "", false, ErrInvalidSource, nil)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, newSourceError("", SourceJSON, "", path, "", Kind(-1), "", false, ErrSourceRead, err)
	}
	return loadJSONBytes(set, data, path, options)
}

func resolveJSONOptions(opts []JSONOption) jsonOptions {
	options := jsonOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func loadJSONBytes(set Set, data []byte, path string, options jsonOptions) (Snapshot, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		return Snapshot{}, newSourceError("", SourceJSON, "", path, options.readerLabel, Kind(-1), "", false, ErrJSONDecode, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("extra JSON data after object")
		}
		return Snapshot{}, newSourceError("", SourceJSON, "", path, options.readerLabel, Kind(-1), "", false, ErrJSONDecode, err)
	}

	object, ok := document.(map[string]any)
	if !ok {
		return Snapshot{}, newSourceError("", SourceJSON, "", path, options.readerLabel, Kind(-1), "", false, ErrJSONDecode, nil)
	}

	snapshot := newSourceSnapshot(set)
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		raw := object[key]
		def, index, found := lookupSourceDefinition(set, key)
		if !found {
			if options.permissive {
				continue
			}
			return Snapshot{}, newSourceError(key, SourceJSON, "", path, options.readerLabel, Kind(-1), "", false, ErrUnknownSourceKey, nil)
		}
		value, err := convertJSONValue(def, raw, path, options.readerLabel)
		if err != nil {
			return Snapshot{}, err
		}
		snapshot.values[index] = newSourceValue(def, value, Source{label: SourceJSON, jsonPath: path, jsonReaderLabel: options.readerLabel})
	}
	return snapshot, nil
}

func newSourceSnapshot(set Set) Snapshot {
	values := make([]Value, 0, len(set.definitions))
	for _, def := range set.definitions {
		values = append(values, newAbsentSourceValue(def))
	}
	return Snapshot{
		values:     values,
		byKey:      cloneIndex(set.byKey),
		normalizer: set.normalizer,
	}
}

func lookupSourceDefinition(set Set, key string) (Definition, int, bool) {
	if invalidKey(key) {
		return Definition{}, -1, false
	}
	normalizedKey := normalizeName(set.normalizer, key)
	index, ok := set.byKey[normalizedKey]
	if !ok || index < 0 || index >= len(set.definitions) {
		return Definition{}, -1, false
	}
	return set.definitions[index], index, true
}

func explicitValue(def Definition, value any) (any, error) {
	if !valueMatchesKind(def.kind, value) {
		return nil, newSourceError(def.key, SourceExplicit, "", "", "", def.kind, fmt.Sprintf("%v", value), def.sensitive, ErrSourceConversion, nil)
	}
	return clonePublicValue(value), nil
}

func convertStringValue(def Definition, raw, source, envName, jsonPath, jsonReaderLabel string) (any, error) {
	switch def.kind {
	case KindString:
		return raw, nil
	case KindBool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	case KindInt:
		value, err := strconv.ParseInt(raw, 10, 0)
		if err != nil {
			return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return int(value), nil
	case KindInt64:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	case KindUint:
		value, err := strconv.ParseUint(raw, 10, 0)
		if err != nil {
			return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return uint(value), nil
	case KindUint64:
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	case KindFloat64:
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	case KindDuration:
		value, err := time.ParseDuration(raw)
		if err != nil {
			return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	case KindStringList:
		if raw == "" {
			return []string{}, nil
		}
		return strings.Split(raw, ","), nil
	default:
		return nil, newSourceError(def.key, source, envName, jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, nil)
	}
}

func convertJSONValue(def Definition, raw any, jsonPath, jsonReaderLabel string) (any, error) {
	switch value := raw.(type) {
	case string:
		if def.kind == KindString {
			return value, nil
		}
		return convertStringValue(def, value, SourceJSON, "", jsonPath, jsonReaderLabel)
	case bool:
		if def.kind == KindBool {
			return value, nil
		}
	case json.Number:
		return convertJSONNumber(def, value, jsonPath, jsonReaderLabel)
	case []any:
		if def.kind == KindStringList {
			values := make([]string, 0, len(value))
			for _, item := range value {
				text, ok := item.(string)
				if !ok {
					return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, fmt.Sprintf("%v", raw), def.sensitive, ErrSourceConversion, nil)
				}
				values = append(values, text)
			}
			return values, nil
		}
	}
	return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, fmt.Sprintf("%v", raw), def.sensitive, ErrSourceConversion, nil)
}

func convertJSONNumber(def Definition, number json.Number, jsonPath, jsonReaderLabel string) (any, error) {
	raw := number.String()
	switch def.kind {
	case KindInt:
		value, err := strconv.ParseInt(raw, 10, 0)
		if err != nil {
			return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return int(value), nil
	case KindInt64:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	case KindUint:
		value, err := strconv.ParseUint(raw, 10, 0)
		if err != nil {
			return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return uint(value), nil
	case KindUint64:
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	case KindFloat64:
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, safeCause(def, err))
		}
		return value, nil
	default:
		return nil, newSourceError(def.key, SourceJSON, "", jsonPath, jsonReaderLabel, def.kind, raw, def.sensitive, ErrSourceConversion, nil)
	}
}

func safeCause(def Definition, err error) error {
	if def.sensitive {
		return nil
	}
	return err
}

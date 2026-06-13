package cli

import (
	"errors"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
)

// FlagBinding maps a routed flag name to a registered config key.
type FlagBinding struct {
	FlagName  string
	ConfigKey string
}

// BindFlag returns a flag-to-config binding value.
func BindFlag(flagName, configKey string) FlagBinding {
	return FlagBinding{FlagName: flagName, ConfigKey: configKey}
}

// NewFlagSnapshot translates explicit routed flag values into config flag bindings.
func NewFlagSnapshot(set config.Set, route command.Result, bindings []FlagBinding) (config.Snapshot, error) {
	if len(bindings) == 0 {
		return config.NewFlagSnapshot(set, nil)
	}

	flagSnapshot, ok := route.FlagSnapshot()
	if !ok {
		return config.Snapshot{}, newBindingError("", "", ErrInvalidBinding, nil)
	}

	values := make([]config.FlagValue, 0, len(bindings))
	for _, binding := range bindings {
		state, exists := flagSnapshot.Lookup(binding.FlagName)
		if !exists {
			return config.Snapshot{}, newBindingError(binding.FlagName, binding.ConfigKey, ErrUnknownFlagBinding, nil)
		}

		flagValue := config.FlagValue{
			ConfigKey:     binding.ConfigKey,
			ExplicitlySet: state.Explicit(),
		}
		if state.Explicit() {
			flagValue.Value = bindingValue(set, binding.ConfigKey, state.Values())
		}
		values = append(values, flagValue)
	}

	snapshot, err := config.NewFlagSnapshot(set, values)
	if err != nil {
		flagName, configKey := bindingContextForCause(set, bindings, err)
		return config.Snapshot{}, newBindingError(flagName, configKey, ErrInvalidBinding, err)
	}
	return snapshot, nil
}

func bindingContextForCause(set config.Set, bindings []FlagBinding, cause error) (string, string) {
	var sourceErr *config.SourceError
	if !errors.As(cause, &sourceErr) {
		return "", ""
	}

	key := sourceErr.Key()
	if errors.Is(cause, config.ErrDuplicateBinding) {
		if flagName, configKey, ok := duplicateBindingContext(set, bindings, key); ok {
			return flagName, configKey
		}
	}

	for _, binding := range bindings {
		if binding.ConfigKey == key {
			return binding.FlagName, binding.ConfigKey
		}
	}

	def, ok := set.Lookup(key)
	if !ok {
		return "", key
	}
	for _, binding := range bindings {
		bindingDef, found := set.Lookup(binding.ConfigKey)
		if found && bindingDef.Name() == def.Name() {
			return binding.FlagName, binding.ConfigKey
		}
	}
	return "", key
}

func duplicateBindingContext(set config.Set, bindings []FlagBinding, key string) (string, string, bool) {
	target, ok := set.Lookup(key)
	if !ok {
		return "", key, false
	}

	seen := false
	for _, binding := range bindings {
		def, found := set.Lookup(binding.ConfigKey)
		if !found || def.Name() != target.Name() {
			continue
		}
		if seen {
			return binding.FlagName, binding.ConfigKey, true
		}
		seen = true
	}
	return "", key, false
}

func bindingValue(set config.Set, configKey string, values []any) any {
	if def, ok := set.Lookup(configKey); ok && def.Kind() == config.KindStringList {
		strings := make([]string, 0, len(values))
		for _, value := range values {
			s, isString := value.(string)
			if !isString {
				return append([]any(nil), values...)
			}
			strings = append(strings, s)
		}
		return strings
	}

	switch len(values) {
	case 0:
		return nil
	case 1:
		return cloneBindingValue(values[0])
	default:
		return append([]any(nil), values...)
	}
}

func cloneBindingValue(value any) any {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		return append([]any(nil), typed...)
	default:
		return typed
	}
}

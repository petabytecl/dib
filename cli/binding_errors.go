package cli

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidBinding identifies malformed or unresolvable flag-to-config bindings.
	ErrInvalidBinding = errors.New("invalid cli flag binding")
	// ErrUnknownFlagBinding identifies a binding whose route flag name is unknown.
	ErrUnknownFlagBinding = errors.New("unknown cli flag binding")
)

// BindingError reports an inspectable flag-to-config binding failure.
type BindingError struct {
	flagName  string
	configKey string
	category  error
	cause     error
}

func newBindingError(flagName string, configKey string, category error, cause error) *BindingError {
	return &BindingError{
		flagName:  flagName,
		configKey: configKey,
		category:  category,
		cause:     cause,
	}
}

func (e *BindingError) Error() string {
	if e == nil {
		return "cli: binding error"
	}
	message := fmt.Sprintf("cli: %v", e.category)
	if e.flagName != "" {
		message += fmt.Sprintf(" flag %q", e.flagName)
	}
	if e.configKey != "" {
		message += fmt.Sprintf(" config key %q", e.configKey)
	}
	if e.cause != nil {
		message += ": " + bindingCauseSummary(e.cause)
	}
	return message
}

func bindingCauseSummary(cause error) string {
	var categorized interface {
		Category() error
	}
	if !errors.As(cause, &categorized) || categorized.Category() == nil {
		return "wrapped cause"
	}

	message := fmt.Sprintf("cause %v", categorized.Category())
	var sourced interface {
		Source() string
	}
	if errors.As(cause, &sourced) && sourced.Source() != "" {
		message += fmt.Sprintf(" from %s", sourced.Source())
	}
	return message
}

// Is exposes both the cli binding category and wrapped cause categories.
func (e *BindingError) Is(target error) bool {
	if e == nil {
		return false
	}
	return target == e.category || errors.Is(e.cause, target)
}

// Unwrap exposes the wrapped lower-level cause for errors.Is and errors.As.
func (e *BindingError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// FlagName returns the route flag name associated with the binding failure.
func (e *BindingError) FlagName() string {
	if e == nil {
		return ""
	}
	return e.flagName
}

// ConfigKey returns the config key associated with the binding failure.
func (e *BindingError) ConfigKey() string {
	if e == nil {
		return ""
	}
	return e.configKey
}

// Category returns the cli binding diagnostic category.
func (e *BindingError) Category() error {
	if e == nil {
		return nil
	}
	return e.category
}

// Cause returns the wrapped lower-level cause, if any.
func (e *BindingError) Cause() error {
	if e == nil {
		return nil
	}
	return e.cause
}

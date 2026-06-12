package command

import (
	"strings"
)

// NameError reports command name validation failure.
type NameError struct{}

// Error returns a stable diagnostic without echoing rejected input.
func (*NameError) Error() string {
	return "command name must not be empty or whitespace"
}

// Definition is an immutable command definition.
//
// Use NewDefinition to create validated command definitions. The zero value is
// not a validated command definition.
type Definition struct {
	name        string
	description string
	aliases     []string
	usage       string
	children    []Definition
}

// Option configures a command definition during construction.
type Option func(*Definition) error

// NewDefinition returns a command definition with a stable name.
//
// Invalid names return *NameError so callers can inspect failures with
// errors.As.
func NewDefinition(name string, options ...Option) (Definition, error) {
	if strings.TrimSpace(name) == "" {
		return Definition{}, &NameError{}
	}

	definition := Definition{name: name}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&definition); err != nil {
			return Definition{}, err
		}
	}

	return definition, nil
}

// Name returns the command definition name supplied at construction.
func (d Definition) Name() string {
	return d.name
}

// Description returns the command description.
func (d Definition) Description() string {
	return d.description
}

// Aliases returns the command alias metadata.
func (d Definition) Aliases() []string {
	return append([]string(nil), d.aliases...)
}

// Usage returns the command usage metadata.
func (d Definition) Usage() string {
	return d.usage
}

// Children returns the nested child command definitions.
func (d Definition) Children() []Definition {
	return cloneDefinitions(d.children)
}

// Description records command description metadata.
func Description(description string) Option {
	return func(d *Definition) error {
		d.description = description
		return nil
	}
}

// Aliases records command alias metadata.
func Aliases(aliases ...string) Option {
	return func(d *Definition) error {
		d.aliases = append([]string(nil), aliases...)
		return nil
	}
}

// Usage records command usage metadata.
func Usage(usage string) Option {
	return func(d *Definition) error {
		d.usage = usage
		return nil
	}
}

// Children records nested child command definitions.
func Children(children ...Definition) Option {
	return func(d *Definition) error {
		if err := validateChildren(children); err != nil {
			return err
		}
		d.children = cloneDefinitions(children)
		return nil
	}
}

// WithDescription returns a definition derived with new description metadata.
func (d Definition) WithDescription(description string) Definition {
	d.description = description
	return d
}

// WithAliases returns a definition derived with new alias metadata.
func (d Definition) WithAliases(aliases ...string) Definition {
	d.aliases = append([]string(nil), aliases...)
	return d
}

// WithUsage returns a definition derived with new usage metadata.
func (d Definition) WithUsage(usage string) Definition {
	d.usage = usage
	return d
}

// WithChildren returns a definition derived with new child command definitions.
func (d Definition) WithChildren(children ...Definition) (Definition, error) {
	if err := validateChildren(children); err != nil {
		return Definition{}, err
	}
	d.children = cloneDefinitions(children)
	return d, nil
}

func validateChildren(children []Definition) error {
	for _, child := range children {
		if strings.TrimSpace(child.name) == "" {
			return &NameError{}
		}
	}
	return nil
}

func cloneDefinitions(definitions []Definition) []Definition {
	if definitions == nil {
		return nil
	}
	copied := make([]Definition, len(definitions))
	for i, definition := range definitions {
		copied[i] = cloneDefinition(definition)
	}
	return copied
}

func cloneDefinition(definition Definition) Definition {
	definition.aliases = append([]string(nil), definition.aliases...)
	definition.children = cloneDefinitions(definition.children)
	return definition
}

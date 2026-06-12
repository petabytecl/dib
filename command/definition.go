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
	if err := validateAliases(nil, definition.name, definition.aliases); err != nil {
		return Definition{}, err
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
		if err := validateAliases(nil, d.name, aliases); err != nil {
			return err
		}
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
		if err := validateChildren([]string{d.name}, children); err != nil {
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
func (d Definition) WithAliases(aliases ...string) (Definition, error) {
	if err := validateAliases(nil, d.name, aliases); err != nil {
		return Definition{}, err
	}
	d.aliases = append([]string(nil), aliases...)
	return d, nil
}

// WithUsage returns a definition derived with new usage metadata.
func (d Definition) WithUsage(usage string) Definition {
	d.usage = usage
	return d
}

// WithChildren returns a definition derived with new child command definitions.
func (d Definition) WithChildren(children ...Definition) (Definition, error) {
	if err := validateChildren([]string{d.name}, children); err != nil {
		return Definition{}, err
	}
	d.children = cloneDefinitions(children)
	return d, nil
}

func validateAliases(parentPath []string, command string, aliases []string) error {
	seen := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		if strings.TrimSpace(alias) == "" {
			return newAliasError(parentPath, command, alias)
		}
		if alias == command {
			return newAliasError(parentPath, command, alias)
		}
		if _, ok := seen[alias]; ok {
			return newTokenConflictError(parentPath, alias, command, command)
		}
		seen[alias] = struct{}{}
	}
	return nil
}

func validateChildren(parentPath []string, children []Definition) error {
	for _, child := range children {
		if strings.TrimSpace(child.name) == "" {
			return &NameError{}
		}
		if err := validateAliases(parentPath, child.name, child.aliases); err != nil {
			return err
		}
		if err := validateChildren(appendPath(parentPath, child.name), child.children); err != nil {
			return err
		}
	}

	owners := make(map[string]string, len(children))
	for _, child := range children {
		if first, ok := owners[child.name]; ok {
			return newTokenConflictError(parentPath, child.name, first, child.name)
		}
		owners[child.name] = child.name
	}
	for _, child := range children {
		for _, alias := range child.aliases {
			if first, ok := owners[alias]; ok {
				return newTokenConflictError(parentPath, alias, first, child.name)
			}
			owners[alias] = child.name
		}
	}
	return nil
}

func appendPath(path []string, name string) []string {
	next := append([]string(nil), path...)
	return append(next, name)
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

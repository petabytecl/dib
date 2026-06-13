package cli

import (
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

// Handler is the canonical high-level command handler signature.
type Handler func(CommandContext) error

// Command is a builder for a command tree that can be registered across files
// before being converted into immutable command and config values at Run time.
type Command struct {
	name           string
	description    string
	aliases        []string
	usage          string
	children       []*Command
	localFlags     []flags.Definition
	inheritedFlags []flags.Definition
	flagNormalizer flags.NameNormalizer
	configDefs     []config.Definition
	bindings       []FlagBinding
	handler        Handler
}

// Option configures a high-level command builder.
type Option func(*Command)

// New returns a root command builder.
func New(name string, options ...Option) *Command {
	root := &Command{name: name}
	root.Apply(options...)
	return root
}

// Command appends a child command and returns it so callers can register a
// subtree from another package or file.
func (c *Command) Command(name string, options ...Option) *Command {
	child := &Command{name: name}
	child.Apply(options...)
	if c != nil {
		c.children = append(c.children, child)
	}
	return child
}

// Apply applies options to an existing command builder and returns the command
// so distributed registration helpers can extend an existing node fluently.
func (c *Command) Apply(options ...Option) *Command {
	if c == nil {
		return c
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		option(c)
	}
	return c
}

// Name returns the command builder name.
func (c *Command) Name() string {
	if c == nil {
		return ""
	}
	return c.name
}

// Description records command description metadata.
func Description(description string) Option {
	return func(c *Command) {
		c.description = description
	}
}

// Aliases records command alias metadata.
func Aliases(aliases ...string) Option {
	return func(c *Command) {
		c.aliases = append([]string(nil), aliases...)
	}
}

// Usage records command usage metadata.
func Usage(usage string) Option {
	return func(c *Command) {
		c.usage = usage
	}
}

// Flags records flag definitions scoped to this command as the final route.
func Flags(definitions ...flags.Definition) Option {
	return func(c *Command) {
		c.localFlags = append([]flags.Definition(nil), definitions...)
	}
}

// InheritedFlags records flag definitions inherited by descendant routes.
func InheritedFlags(definitions ...flags.Definition) Option {
	return func(c *Command) {
		c.inheritedFlags = append([]flags.Definition(nil), definitions...)
	}
}

// FlagNormalizer records the long-name normalizer used when composing command flags.
func FlagNormalizer(normalizer flags.NameNormalizer) Option {
	return func(c *Command) {
		c.flagNormalizer = normalizer
	}
}

// Config records config definitions contributed by this command subtree.
func Config(definitions ...config.Definition) Option {
	return func(c *Command) {
		c.configDefs = append([]config.Definition(nil), definitions...)
	}
}

// Bindings records flag-to-config bindings contributed by this command subtree.
func Bindings(bindings ...FlagBinding) Option {
	return func(c *Command) {
		c.bindings = append([]FlagBinding(nil), bindings...)
	}
}

// Handle records the handler invoked when this command is the final routed command.
func Handle(handler Handler) Option {
	return func(c *Command) {
		c.handler = handler
	}
}

// Definition builds an immutable command definition from the builder tree.
func (c *Command) Definition() (command.Definition, error) {
	if c == nil {
		return command.Definition{}, &command.NameError{}
	}

	children := make([]command.Definition, 0, len(c.children))
	for _, child := range c.children {
		definition, err := child.Definition()
		if err != nil {
			return command.Definition{}, err
		}
		children = append(children, definition)
	}

	options := []command.Option{
		command.Description(c.description),
		command.Usage(c.usage),
		command.Aliases(c.aliases...),
		command.LocalFlags(c.localFlags...),
		command.InheritedFlags(c.inheritedFlags...),
		command.FlagNormalizer(c.flagNormalizer),
		command.Children(children...),
	}
	return command.NewDefinition(c.name, options...)
}

// Plan builds the low-level composition plan used by Resolve.
func (c *Command) Plan() (Plan, error) {
	return c.planWithBindings(c.flagBindings())
}

func (c *Command) planWithBindings(bindings []FlagBinding) (Plan, error) {
	root, err := c.Definition()
	if err != nil {
		return Plan{}, err
	}
	set, err := config.NewSet(c.configDefinitions()...)
	if err != nil {
		return Plan{}, err
	}
	return NewPlan(root, set).WithBindings(bindings), nil
}

func (c *Command) configDefinitions() []config.Definition {
	if c == nil {
		return nil
	}
	definitions := append([]config.Definition(nil), c.configDefs...)
	for _, child := range c.children {
		definitions = append(definitions, child.configDefinitions()...)
	}
	return definitions
}

func (c *Command) flagBindings() []FlagBinding {
	if c == nil {
		return nil
	}
	bindings := append([]FlagBinding(nil), c.bindings...)
	for _, child := range c.children {
		bindings = append(bindings, child.flagBindings()...)
	}
	return bindings
}

func (c *Command) flagBindingsForPath(path []string) []FlagBinding {
	nodes := c.nodesForPath(path)
	if len(nodes) == 0 {
		return nil
	}
	var bindings []FlagBinding
	for _, node := range nodes {
		bindings = append(bindings, node.bindings...)
	}
	return append([]FlagBinding(nil), bindings...)
}

func (c *Command) nodesForPath(path []string) []*Command {
	if c == nil || len(path) == 0 || c.name != path[0] {
		return nil
	}
	nodes := []*Command{c}
	current := c
	for _, name := range path[1:] {
		next, ok := current.childByName(name)
		if !ok {
			return nil
		}
		nodes = append(nodes, next)
		current = next
	}
	return nodes
}

func (c *Command) routedCommand(path []string) (*Command, bool) {
	nodes := c.nodesForPath(path)
	if len(nodes) == 0 {
		return nil, false
	}
	return nodes[len(nodes)-1], true
}

func (c *Command) childByName(name string) (*Command, bool) {
	for _, child := range c.children {
		if child.name == name {
			return child, true
		}
	}
	return nil, false
}

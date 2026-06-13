package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/petabytecl/dib/flags"
)

// WriteUsage renders deterministic usage text for this command definition to w.
func (d Definition) WriteUsage(w io.Writer) error {
	if strings.TrimSpace(d.name) == "" {
		return &NameError{}
	}
	return writeUsage(w, []Definition{cloneDefinition(d)})
}

// WriteHelp renders deterministic help text for this command definition to w.
func (d Definition) WriteHelp(w io.Writer) error {
	if strings.TrimSpace(d.name) == "" {
		return &NameError{}
	}
	composition, err := composeFlags([]Definition{cloneDefinition(d)})
	if err != nil {
		return err
	}
	return writeHelp(w, []Definition{cloneDefinition(d)}, composition.set.Definitions())
}

// WriteUsage renders deterministic usage text for the routed command to w.
func (r Result) WriteUsage(w io.Writer) error {
	if len(r.path) == 0 {
		return &NameError{}
	}
	return writeUsage(w, r.Path())
}

// WriteHelp renders deterministic help text for the routed command to w.
func (r Result) WriteHelp(w io.Writer) error {
	if len(r.path) == 0 {
		return &NameError{}
	}
	definitions := []flags.Definition(nil)
	if set, ok := r.Flags(); ok {
		definitions = set.Definitions()
	}
	return writeHelp(w, r.Path(), definitions)
}

func writeHelp(w io.Writer, path []Definition, definitions []flags.Definition) error {
	if err := writeUsage(w, path); err != nil {
		return err
	}
	command := path[len(path)-1]
	if len(command.aliases) > 0 {
		if err := writeSection(w, "Aliases", strings.Join(command.aliases, ", ")); err != nil {
			return err
		}
	}
	if command.description != "" {
		if err := writeSection(w, "Description", command.description); err != nil {
			return err
		}
	}
	if len(command.children) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Commands:"); err != nil {
			return err
		}
		width := commandNameWidth(command.children)
		for _, child := range command.children {
			line := fmt.Sprintf("  %-*s", width, child.name)
			if child.description != "" {
				line += "  " + child.description
			}
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	visibleFlags := visibleFlagDefinitions(definitions)
	if len(visibleFlags) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Flags:"); err != nil {
			return err
		}
		for _, definition := range visibleFlags {
			spelling := flagSpelling(definition)
			line := "  " + spelling
			if definition.Usage() != "" {
				line += "  " + definition.Usage()
			}
			if deprecated := definition.Deprecated(); deprecated != "" {
				line += " (deprecated: " + deprecated + ")"
			}
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeUsage(w io.Writer, path []Definition) error {
	command := path[len(path)-1]
	usage := strings.Join(pathNames(path), " ")
	if command.usage != "" {
		usage += " " + command.usage
	}
	if _, err := fmt.Fprintln(w, "Usage:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "  "+usage); err != nil {
		return err
	}
	return nil
}

func writeSection(w io.Writer, title, body string) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, title+":"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "  "+body); err != nil {
		return err
	}
	return nil
}

func commandNameWidth(definitions []Definition) int {
	width := 0
	for _, definition := range definitions {
		if len(definition.name) > width {
			width = len(definition.name)
		}
	}
	return width
}

func visibleFlagDefinitions(definitions []flags.Definition) []flags.Definition {
	visible := make([]flags.Definition, 0, len(definitions))
	for _, definition := range definitions {
		if definition.Hidden() {
			continue
		}
		visible = append(visible, definition)
	}
	return visible
}

func flagSpelling(definition flags.Definition) string {
	spelling := "--" + definition.Name()
	if shorthand, ok := definition.Shorthand(); ok {
		spelling += ", -" + shorthand
	}
	if definition.Arity() == flags.ArityRequired {
		spelling += " <" + definition.Kind().String() + ">"
	}
	return spelling
}

package command

import "strings"

// Route matches caller-supplied arguments against the command tree.
func (d Definition) Route(args []string) (Result, error) {
	if strings.TrimSpace(d.name) == "" {
		return Result{}, &NameError{}
	}

	current := cloneDefinition(d)
	path := []Definition{current}

	for i := 0; i < len(args); i++ {
		token := args[i]
		if token == "--" {
			return newResult(path, args[i+1:]), nil
		}

		if len(current.children) == 0 {
			return newResult(path, args[i:]), nil
		}

		child, ok := current.childByName(token)
		if !ok {
			return Result{}, newUnknownCommandError(token, path)
		}

		current = child
		path = append(path, current)
	}

	return newResult(path, nil), nil
}

func (d Definition) childByName(name string) (Definition, bool) {
	for _, child := range d.children {
		if child.name == name {
			return child, true
		}
	}
	return Definition{}, false
}

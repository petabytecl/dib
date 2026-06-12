package command

import "strings"

// Route matches caller-supplied arguments against the command tree.
func (d Definition) Route(args []string) (Result, error) {
	if strings.TrimSpace(d.name) == "" {
		return Result{}, &NameError{}
	}

	current := cloneDefinition(d)
	path := []Definition{current}
	matchTokens := []string{current.name}

	for i := 0; i < len(args); i++ {
		token := args[i]
		if token == "--" {
			return newResult(path, matchTokens, args[i+1:]), nil
		}

		if len(current.children) == 0 {
			return newResult(path, matchTokens, args[i:]), nil
		}

		child, ok := current.childByToken(token)
		if !ok {
			return Result{}, newUnknownCommandError(token, path)
		}

		current = child
		path = append(path, current)
		matchTokens = append(matchTokens, token)
	}

	return newResult(path, matchTokens, nil), nil
}

func (d Definition) childByToken(token string) (Definition, bool) {
	for _, child := range d.children {
		if child.name == token {
			return child, true
		}
	}
	for _, child := range d.children {
		for _, alias := range child.aliases {
			if alias == token {
				return child, true
			}
		}
	}
	return Definition{}, false
}

package command

// Result is a self-contained snapshot of one command routing run.
type Result struct {
	path        []Definition
	matchTokens []string
	remaining   []string
}

func newResult(path []Definition, matchTokens []string, remaining []string) Result {
	return Result{
		path:        cloneDefinitions(path),
		matchTokens: append([]string(nil), matchTokens...),
		remaining:   append([]string(nil), remaining...),
	}
}

// Path returns the canonical matched command definition path.
func (r Result) Path() []Definition {
	return cloneDefinitions(r.path)
}

// PathNames returns the canonical matched command name path.
func (r Result) PathNames() []string {
	return pathNames(r.path)
}

// MatchTokens returns the raw input tokens that matched the canonical path.
func (r Result) MatchTokens() []string {
	return append([]string(nil), r.matchTokens...)
}

// Command returns the final matched command definition.
func (r Result) Command() (Definition, bool) {
	if len(r.path) == 0 {
		return Definition{}, false
	}
	return cloneDefinition(r.path[len(r.path)-1]), true
}

// RemainingArgs returns the caller arguments left after routing.
func (r Result) RemainingArgs() []string {
	return append([]string(nil), r.remaining...)
}

func pathNames(path []Definition) []string {
	if path == nil {
		return nil
	}
	names := make([]string, len(path))
	for i, definition := range path {
		names[i] = definition.name
	}
	return names
}

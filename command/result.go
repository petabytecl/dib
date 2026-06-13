package command

import "github.com/petabytecl/dib/flags"

// Result is a self-contained snapshot of one command routing run.
type Result struct {
	path            []Definition
	matchTokens     []string
	remaining       []string
	flags           flags.Set
	hasFlags        bool
	flagSnapshot    flags.Snapshot
	hasFlagSnapshot bool
}

func newResult(path []Definition, matchTokens, remaining []string) Result {
	return Result{
		path:        cloneDefinitions(path),
		matchTokens: append([]string(nil), matchTokens...),
		remaining:   append([]string(nil), remaining...),
	}
}

func newFlagResult(path []Definition, matchTokens, remaining []string, set flags.Set, snapshot flags.Snapshot) Result {
	return Result{
		path:            cloneDefinitions(path),
		matchTokens:     append([]string(nil), matchTokens...),
		remaining:       append([]string(nil), remaining...),
		flags:           set,
		hasFlags:        true,
		flagSnapshot:    snapshot,
		hasFlagSnapshot: true,
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

// Flags returns the composed flag definitions available to the matched command.
func (r Result) Flags() (flags.Set, bool) {
	if !r.hasFlags {
		return flags.Set{}, false
	}
	return r.flags, true
}

// FlagSnapshot returns the parsed flag snapshot for this route.
func (r Result) FlagSnapshot() (flags.Snapshot, bool) {
	if !r.hasFlagSnapshot {
		return flags.Snapshot{}, false
	}
	return r.flagSnapshot, true
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

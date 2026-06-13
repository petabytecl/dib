package command

import (
	"strings"

	"github.com/petabytecl/dib/flags"
)

// Route matches caller-supplied arguments against the command tree.
func (d Definition) Route(args []string) (Result, error) {
	if strings.TrimSpace(d.name) == "" {
		return Result{}, &NameError{}
	}

	root := cloneDefinition(d)
	if !hasFlagDefinitions(root) {
		return routeWithoutFlags(root, args)
	}

	candidates, preferredParseErr, preferredParseDepth, compositionErr := routeCandidates(root, args)
	if compositionErr != nil {
		return Result{}, compositionErr
	}
	if len(candidates) == 0 {
		if preferredParseErr != nil {
			return Result{}, preferredParseErr
		}
		return Result{}, newUnknownCommandError("", []Definition{root})
	}

	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if len(candidate.path) > len(best.path) {
			best = candidate
		}
	}
	if preferredParseErr != nil && preferredParseDepth > len(best.path) {
		return Result{}, preferredParseErr
	}
	if len(best.preCommandRemaining) > 0 && len(best.path[len(best.path)-1].children) > 0 {
		return Result{}, newUnknownCommandError(best.preCommandRemaining[0], best.path)
	}

	return newFlagResult(best.path, best.matchTokens, append(best.preCommandRemaining, best.afterTerminator...), best.set, best.snapshot), nil
}

func routeWithoutFlags(root Definition, args []string) (Result, error) {
	current := root
	path := []Definition{current}
	matchTokens := []string{current.name}

	for i := range args {
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

type routeCandidate struct {
	path                []Definition
	matchTokens         []string
	preCommandRemaining []string
	afterTerminator     []string
	set                 flags.Set
	snapshot            flags.Snapshot
}

func routeCandidates(root Definition, args []string) ([]routeCandidate, error, int, error) {
	beforeTerminator, afterTerminator := splitTerminator(args)
	var candidates []routeCandidate
	var preferredParseErr error
	preferredParseDepth := -1

	err := walkFlagPaths([]Definition{root}, func(path []Definition) error {
		composition, err := composeFlags(path)
		if err != nil {
			return err
		}

		snapshot, err := composition.set.Parse(beforeTerminator)
		if err != nil {
			if commandPathAppears(beforeTerminator, path) && len(path) > preferredParseDepth {
				preferredParseDepth = len(path)
				preferredParseErr = err
			}
			return nil
		}

		remaining := snapshot.RemainingArgs()
		matchTokens, unmatched, ok := matchPathTokens(path, remaining)
		if !ok {
			return nil
		}
		candidates = append(candidates, routeCandidate{
			path:                cloneDefinitions(path),
			matchTokens:         matchTokens,
			preCommandRemaining: unmatched,
			afterTerminator:     append([]string(nil), afterTerminator...),
			set:                 composition.set,
			snapshot:            snapshot,
		})
		return nil
	})
	if err != nil {
		return nil, nil, -1, err
	}
	return candidates, preferredParseErr, preferredParseDepth, nil
}

func splitTerminator(args []string) ([]string, []string) {
	for i, arg := range args {
		if arg == "--" {
			return append([]string(nil), args[:i]...), append([]string(nil), args[i+1:]...)
		}
	}
	return append([]string(nil), args...), nil
}

func matchPathTokens(path []Definition, remaining []string) ([]string, []string, bool) {
	if len(path) == 0 {
		return nil, nil, false
	}
	if len(remaining) < len(path)-1 {
		return nil, nil, false
	}

	matchTokens := []string{path[0].name}
	for i := 1; i < len(path); i++ {
		token := remaining[i-1]
		if !matchesDefinitionToken(path[i], token) {
			return nil, nil, false
		}
		matchTokens = append(matchTokens, token)
	}
	return matchTokens, append([]string(nil), remaining[len(path)-1:]...), true
}

func commandPathAppears(args []string, path []Definition) bool {
	if len(path) <= 1 {
		return true
	}
	index := 0
	for i := 1; i < len(path); i++ {
		found := false
		for index < len(args) {
			if matchesDefinitionToken(path[i], args[index]) {
				found = true
				index++
				break
			}
			index++
		}
		if !found {
			return false
		}
	}
	return true
}

func matchesDefinitionToken(definition Definition, token string) bool {
	if definition.name == token {
		return true
	}
	for _, alias := range definition.aliases {
		if alias == token {
			return true
		}
	}
	return false
}

func hasFlagDefinitions(definition Definition) bool {
	if len(definition.localFlags) > 0 || len(definition.inheritedFlags) > 0 {
		return true
	}
	for _, child := range definition.children {
		if hasFlagDefinitions(child) {
			return true
		}
	}
	return false
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

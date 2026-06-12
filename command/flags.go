package command

import "github.com/petabytecl/dib/flags"

type flagComposition struct {
	set flags.Set
}

func composeFlags(path []Definition) (flagComposition, error) {
	definitions := make([]flags.Definition, 0)
	var normalizer flags.NameNormalizer
	for _, definition := range path {
		if definition.flagNormalizer != nil {
			normalizer = definition.flagNormalizer
		}
		definitions = append(definitions, definition.inheritedFlags...)
	}
	if len(path) > 0 {
		definitions = append(definitions, path[len(path)-1].localFlags...)
	}

	set, err := newFlagSet(normalizer, definitions)
	if err != nil {
		return flagComposition{}, newFlagCompositionError(path, "available", err)
	}
	return flagComposition{set: set}, nil
}

func newFlagSet(normalizer flags.NameNormalizer, definitions []flags.Definition) (flags.Set, error) {
	if normalizer != nil {
		return flags.NewNormalizedSet(normalizer, definitions...)
	}
	return flags.NewSet(definitions...)
}

func validateFlagCompositionTree(root Definition) error {
	if root.name == "" {
		return nil
	}
	return walkFlagPaths([]Definition{cloneDefinition(root)}, func(path []Definition) error {
		_, err := composeFlags(path)
		return err
	})
}

func walkFlagPaths(path []Definition, visit func([]Definition) error) error {
	if err := visit(path); err != nil {
		return err
	}
	current := path[len(path)-1]
	for _, child := range current.children {
		next := append(cloneDefinitions(path), cloneDefinition(child))
		if err := walkFlagPaths(next, visit); err != nil {
			return err
		}
	}
	return nil
}

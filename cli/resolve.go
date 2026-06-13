package cli

import (
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
)

// Plan carries the caller-supplied composition inputs for cli.Resolve.
type Plan struct {
	root     command.Definition
	set      config.Set
	explicit config.Snapshot
	env      config.Snapshot
	jsonSrc  config.Snapshot
	bindings []FlagBinding
}

// NewPlan returns a Plan with the required root command and config set.
// Optional source snapshots and flag bindings are added via With* methods.
func NewPlan(root command.Definition, set config.Set) Plan {
	return Plan{root: root, set: set}
}

// WithExplicit returns a new Plan with the given explicit-setter snapshot.
func (p Plan) WithExplicit(s config.Snapshot) Plan {
	p.explicit = s
	return p
}

// WithEnv returns a new Plan with the given env snapshot.
func (p Plan) WithEnv(s config.Snapshot) Plan {
	p.env = s
	return p
}

// WithJSON returns a new Plan with the given JSON snapshot.
func (p Plan) WithJSON(s config.Snapshot) Plan {
	p.jsonSrc = s
	return p
}

// WithBindings returns a new Plan with the given flag bindings.
func (p Plan) WithBindings(b []FlagBinding) Plan {
	p.bindings = append([]FlagBinding(nil), b...)
	return p
}

// Root returns the root command definition.
func (p Plan) Root() command.Definition { return p.root }

// ConfigSet returns the config definition set.
func (p Plan) ConfigSet() config.Set { return p.set }

// ExplicitSnapshot returns the explicit-setter source snapshot.
func (p Plan) ExplicitSnapshot() config.Snapshot { return p.explicit }

// EnvSnapshot returns the env source snapshot.
func (p Plan) EnvSnapshot() config.Snapshot { return p.env }

// JSONSnapshot returns the JSON source snapshot.
func (p Plan) JSONSnapshot() config.Snapshot { return p.jsonSrc }

// Bindings returns a defensive copy of the flag bindings.
func (p Plan) Bindings() []FlagBinding { return append([]FlagBinding(nil), p.bindings...) }

// Resolve routes the invocation, builds the flag-tier config snapshot, resolves
// config by precedence, and returns a Result with all composed outputs.
//
// Resolve does not invoke callbacks, call os.Exit, write to any writer, read
// os.Args or env, or load files. All process inputs are caller-supplied.
func Resolve(inv Invocation, plan Plan) (Result, error) {
	route, err := plan.Root().Route(inv.Args())
	if err != nil {
		return Result{}, err
	}

	flagSnap, err := NewFlagSnapshot(plan.ConfigSet(), route, plan.Bindings())
	if err != nil {
		return Result{}, err
	}

	resolved := config.Resolve(
		plan.ConfigSet(),
		plan.ExplicitSnapshot(),
		flagSnap,
		plan.EnvSnapshot(),
		plan.JSONSnapshot(),
	)

	snap, hasFlagSnap := route.FlagSnapshot()
	if hasFlagSnap {
		return newFlagResult(inv, route, snap, resolved), nil
	}
	return newResult(inv, route, resolved), nil
}

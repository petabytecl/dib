package cli

import (
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

// Result is an immutable snapshot of a successful cli.Resolve call.
type Result struct {
	invocation  Invocation
	route       command.Result
	flagSnap    flags.Snapshot
	hasFlagSnap bool
	cfg         config.Snapshot
}

func newResult(inv Invocation, route command.Result, cfg config.Snapshot) Result {
	return Result{invocation: inv, route: route, cfg: cfg}
}

func newFlagResult(inv Invocation, route command.Result, snap flags.Snapshot, cfg config.Snapshot) Result {
	return Result{invocation: inv, route: route, flagSnap: snap, hasFlagSnap: true, cfg: cfg}
}

// Invocation returns the caller-supplied invocation used for this resolve.
func (r Result) Invocation() Invocation { return r.invocation }

// Route returns the command routing result.
func (r Result) Route() command.Result { return r.route }

// FlagSnapshot returns the parsed flag snapshot and whether one is present.
func (r Result) FlagSnapshot() (flags.Snapshot, bool) { return r.flagSnap, r.hasFlagSnap }

// Config returns the fully resolved config snapshot.
func (r Result) Config() config.Snapshot { return r.cfg }

// RemainingArgs returns the arguments left after routing and flag parsing.
func (r Result) RemainingArgs() []string {
	return append([]string(nil), r.route.RemainingArgs()...)
}

package cli

import (
	"context"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

// CommandContext is the single handler argument passed to high-level command handlers.
type CommandContext struct {
	ctx    context.Context
	result Result
}

func newCommandContext(ctx context.Context, result Result) CommandContext {
	if ctx == nil {
		ctx = context.Background()
	}
	return CommandContext{ctx: ctx, result: result}
}

// Context returns the caller-supplied execution context.
func (c CommandContext) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// Result returns the full resolved CLI result.
func (c CommandContext) Result() Result { return c.result }

// Invocation returns the caller-supplied invocation.
func (c CommandContext) Invocation() Invocation { return c.result.Invocation() }

// Route returns the matched command route.
func (c CommandContext) Route() command.Result { return c.result.Route() }

// Config returns the fully resolved config snapshot.
func (c CommandContext) Config() config.Snapshot { return c.result.Config() }

// FlagSnapshot returns the parsed flag snapshot and whether one is present.
func (c CommandContext) FlagSnapshot() (flags.Snapshot, bool) {
	return c.result.FlagSnapshot()
}

// RemainingArgs returns arguments left after routing and flag parsing.
func (c CommandContext) RemainingArgs() []string {
	return c.result.RemainingArgs()
}

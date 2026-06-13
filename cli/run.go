package cli

import "context"

// Run builds a plan from the command tree, resolves a full argv slice, and
// invokes the matched command handler.
func (c *Command) Run(ctx context.Context, argv []string) (Result, error) {
	inv, err := FromOSArgs(argv)
	if err != nil {
		return Result{}, err
	}
	return c.RunInvocation(ctx, inv)
}

// RunArgs builds a plan from the command tree, resolves already-stripped user
// args, and invokes the matched command handler.
func (c *Command) RunArgs(ctx context.Context, args []string) (Result, error) {
	return c.RunInvocation(ctx, FromArgs(c.Name(), args))
}

// RunInvocation builds a plan from the command tree, resolves the invocation,
// and invokes exactly one matched command handler.
func (c *Command) RunInvocation(ctx context.Context, inv Invocation) (Result, error) {
	routePlan, err := c.planWithBindings(nil)
	if err != nil {
		return Result{}, err
	}

	initial, err := Resolve(inv, routePlan)
	if err != nil {
		return Result{}, err
	}

	bindings := c.flagBindingsForPath(initial.Route().PathNames())
	result, err := Resolve(inv, routePlan.WithBindings(bindings))
	if err != nil {
		return Result{}, err
	}

	routed, ok := c.routedCommand(result.Route().PathNames())
	if !ok || routed.handler == nil {
		return result, newDispatchError(result.Route().PathNames(), ErrNoHandler)
	}
	return result, routed.handler(newCommandContext(ctx, result))
}

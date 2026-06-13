package command

import (
	"context"
	"io"
)

// Boundary captures caller-owned execution-boundary metadata for a routed command.
//
// Boundary does not execute callbacks, write to streams, read process state, or
// decide exit policy. It keeps the route snapshot and explicit caller inputs
// together so callers can make those decisions outside Dib.
type Boundary struct {
	ctx       context.Context
	result    Result
	hasResult bool
	args      []string
	stdout    io.Writer
	stderr    io.Writer
}

// NewBoundary returns immutable boundary metadata for an existing route result.
func NewBoundary(ctx context.Context, result Result, args []string, stdout, stderr io.Writer) Boundary {
	return Boundary{
		ctx:       ctx,
		result:    result,
		hasResult: len(result.path) > 0,
		args:      append([]string(nil), args...),
		stdout:    stdout,
		stderr:    stderr,
	}
}

// RouteBoundary routes explicit caller args and packages the resulting route
// snapshot with caller-owned context and writer metadata.
func (d Definition) RouteBoundary(ctx context.Context, args []string, stdout, stderr io.Writer) (Boundary, error) {
	result, err := d.Route(args)
	if err != nil {
		return Boundary{}, err
	}
	return NewBoundary(ctx, result, args, stdout, stderr), nil
}

// Context returns the caller-supplied context.
func (b Boundary) Context() context.Context {
	return b.ctx
}

// Args returns the caller args captured for this boundary.
func (b Boundary) Args() []string {
	return append([]string(nil), b.args...)
}

// Result returns the captured route result.
func (b Boundary) Result() (Result, bool) {
	if !b.hasResult {
		return Result{}, false
	}
	return b.result, true
}

// Stdout returns the caller-supplied stdout writer metadata.
func (b Boundary) Stdout() (io.Writer, bool) {
	if b.stdout == nil {
		return nil, false
	}
	return b.stdout, true
}

// Stderr returns the caller-supplied stderr writer metadata.
func (b Boundary) Stderr() (io.Writer, bool) {
	if b.stderr == nil {
		return nil, false
	}
	return b.stderr, true
}

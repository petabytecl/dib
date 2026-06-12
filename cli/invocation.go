package cli

// Invocation is an immutable snapshot of a caller-supplied CLI invocation.
type Invocation struct {
	program string
	args    []string
}

// FromOSArgs constructs an invocation from a full argv slice.
//
// The first element is treated as the program name and the remaining elements
// as user arguments. Callers remain responsible for deciding whether this input
// came from os.Args or another source.
func FromOSArgs(argv []string) (Invocation, error) {
	if len(argv) == 0 {
		return Invocation{}, newInvocationError(ErrInvalidInvocation, "missing program")
	}
	return FromArgs(argv[0], argv[1:]), nil
}

// FromArgs constructs an invocation from an explicit program name and already
// stripped user arguments.
func FromArgs(program string, args []string) Invocation {
	return Invocation{
		program: program,
		args:    append([]string(nil), args...),
	}
}

// Program returns the caller-supplied program name.
func (i Invocation) Program() string {
	return i.program
}

// Args returns the caller-supplied user arguments.
func (i Invocation) Args() []string {
	return append([]string(nil), i.args...)
}

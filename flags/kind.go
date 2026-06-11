package flags

// Kind identifies the value type accepted by a flag definition.
type Kind int

const (
	KindString Kind = iota
	KindBool
	KindInt
	KindInt64
	KindUint
	KindUint64
	KindFloat64
	KindDuration
	KindStringList
)

func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindInt64:
		return "int64"
	case KindUint:
		return "uint"
	case KindUint64:
		return "uint64"
	case KindFloat64:
		return "float64"
	case KindDuration:
		return "duration"
	case KindStringList:
		return "string-list"
	default:
		return "unknown"
	}
}

// Arity identifies how a flag consumes command-line values.
type Arity int

const (
	ArityNone Arity = iota
	ArityOptional
	ArityRequired
)

// RepeatPolicy identifies how repeated occurrences are represented.
type RepeatPolicy int

const (
	RepeatLast RepeatPolicy = iota
	RepeatAccumulated
)

package flags

import (
	"strconv"
	"time"
)

// Parser converts one raw CLI token into a typed flag value.
type Parser interface {
	ParseFlagValue(raw string) (any, error)
}

// ParserFunc adapts a function into a Parser.
type ParserFunc func(raw string) (any, error)

func (f ParserFunc) ParseFlagValue(raw string) (any, error) {
	return f(raw)
}

func stringParser(raw string) (any, error) {
	return raw, nil
}

func boolParser(raw string) (any, error) {
	return strconv.ParseBool(raw)
}

func intParser(raw string) (any, error) {
	value, err := strconv.ParseInt(raw, 0, 0)
	if err != nil {
		return nil, err
	}
	return int(value), nil
}

func int64Parser(raw string) (any, error) {
	return strconv.ParseInt(raw, 0, 64)
}

func uintParser(raw string) (any, error) {
	value, err := strconv.ParseUint(raw, 0, 0)
	if err != nil {
		return nil, err
	}
	return uint(value), nil
}

func uint64Parser(raw string) (any, error) {
	return strconv.ParseUint(raw, 0, 64)
}

func float64Parser(raw string) (any, error) {
	return strconv.ParseFloat(raw, 64)
}

func durationParser(raw string) (any, error) {
	return time.ParseDuration(raw)
}

func stringListParser(raw string) (any, error) {
	return raw, nil
}

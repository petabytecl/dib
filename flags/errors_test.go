package flags_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestSensitiveParseErrorDoesNotExposeRawValue(t *testing.T) {
	parserErr := errors.New("parser rejected value")
	def := flags.Custom("token", flags.KindString, "", "token", flags.ParserFunc(func(raw string) (any, error) {
		return nil, parserErr
	}), flags.Sensitive())

	_, err := def.Parse("dib_fake_secret_value")
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("errors.Is(err, ErrConversion) = false; err=%v", err)
	}
	if errors.Is(err, parserErr) {
		t.Fatalf("sensitive error exposes parser cause: %v", err)
	}
	if strings.Contains(err.Error(), "dib_fake_secret_value") {
		t.Fatalf("sensitive raw value leaked in error: %v", err)
	}
}

func TestNonSensitiveParseErrorWrapsParserCause(t *testing.T) {
	parserErr := errors.New("parser rejected value")
	def := flags.Custom("endpoint", flags.KindString, "http://localhost", "endpoint", flags.ParserFunc(func(raw string) (any, error) {
		return nil, parserErr
	}))

	_, err := def.Parse("http://example.invalid")
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("errors.Is(err, ErrConversion) = false; err=%v", err)
	}
	if !errors.Is(err, parserErr) {
		t.Fatalf("non-sensitive error does not expose parser cause: %v", err)
	}
}

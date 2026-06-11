package command_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/petabytecl/dib/command"
)

func TestNewDefinitionAcceptsStableName(t *testing.T) {
	definition, err := command.NewDefinition("serve")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	if got := definition.Name(); got != "serve" {
		t.Fatalf("Name() = %q, want %q", got, "serve")
	}
}

func TestNewDefinitionRejectsBlankNames(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "spaces", in: "   \t\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := command.NewDefinition(tt.in)
			if err == nil {
				t.Fatal("NewDefinition returned nil error")
			}
			var nameErr *command.NameError
			if !errors.As(err, &nameErr) {
				t.Fatalf("error does not expose NameError: %T", err)
			}
		})
	}
}

func ExampleNewDefinition() {
	definition, err := command.NewDefinition("serve")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(definition.Name())
	// Output: serve
}

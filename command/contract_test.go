package command_test

import (
	"os"
	"testing"

	"github.com/petabytecl/dib/command"
)

func TestDefinitionConstructionUsesExplicitInputs(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"dib", "ambient-command"}
	t.Setenv("DIB_COMMAND_NAME", "ambient-command")

	tests := []struct {
		name string
		want string
	}{
		{name: "serve", want: "serve"},
		{name: "status", want: "status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition, err := command.NewDefinition(tt.name)
			if err != nil {
				t.Fatalf("NewDefinition(%q) returned unexpected error: %v", tt.name, err)
			}

			if got := definition.Name(); got != tt.want {
				t.Fatalf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

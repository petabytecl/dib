package command_test

import (
	"os"
	"reflect"
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

func TestRoutingUsesExplicitInputsAndReturnedValues(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	t.Cleanup(func() {
		os.Args = originalArgs
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatalf("CreateTemp stdout returned unexpected error: %v", err)
	}
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("CreateTemp stderr returned unexpected error: %v", err)
	}
	os.Stdout = stdoutFile
	os.Stderr = stderrFile
	os.Args = []string{"dib", "ambient", "command"}
	t.Setenv("DIB_COMMAND_NAME", "ambient-command")

	apply, err := command.NewDefinition("apply")
	if err != nil {
		t.Fatalf("NewDefinition(apply) returned unexpected error: %v", err)
	}
	deploy, err := command.NewDefinition("deploy", command.Children(apply))
	if err != nil {
		t.Fatalf("NewDefinition(deploy) returned unexpected error: %v", err)
	}
	root, err := command.NewDefinition("dib", command.Children(deploy))
	if err != nil {
		t.Fatalf("NewDefinition(dib) returned unexpected error: %v", err)
	}

	result, err := root.Route([]string{"deploy", "apply", "manifest.yaml"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"manifest.yaml"}) {
		t.Fatalf("RemainingArgs() = %q, want %q", got, []string{"manifest.yaml"})
	}

	if err := stdoutFile.Sync(); err != nil {
		t.Fatalf("stdout Sync returned unexpected error: %v", err)
	}
	if err := stderrFile.Sync(); err != nil {
		t.Fatalf("stderr Sync returned unexpected error: %v", err)
	}
	assertEmptyFile(t, stdoutFile)
	assertEmptyFile(t, stderrFile)
}

func assertEmptyFile(t *testing.T, file *os.File) {
	t.Helper()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Stat(%s) returned unexpected error: %v", file.Name(), err)
	}
	if info.Size() != 0 {
		t.Fatalf("%s size = %d, want 0", file.Name(), info.Size())
	}
}

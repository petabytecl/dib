package flags_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runConsumerContract(t *testing.T, name string, source string) {
	t.Helper()

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}

	dir := t.TempDir()
	modulePath := filepath.ToSlash(repoRoot)
	goMod := fmt.Sprintf(`module flagsconsumer

go 1.26

require github.com/petabytecl/dib v0.0.0

replace github.com/petabytecl/dib => %s
`, modulePath)

	writeTempFile(t, dir, "go.mod", goMod)
	writeTempFile(t, dir, "consumer_test.go", source)

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("consumer contract %q failed:\n%s", name, output)
	}
}

func writeTempFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

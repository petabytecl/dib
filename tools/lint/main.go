package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	lintViolationExit    = 1
	executionFailureExit = 2
)

func main() {
	os.Exit(run(context.Background(), ".", os.Stdout, os.Stderr))
}

func run(ctx context.Context, dir string, stdout io.Writer, stderr io.Writer) int {
	diagnostics, err := findDiagnostics(ctx, dir)
	if err != nil {
		fmt.Fprintf(stderr, "lint execution error: %v\n", err)
		return executionFailureExit
	}

	if len(diagnostics) == 0 {
		return 0
	}

	for _, diagnostic := range diagnostics {
		fmt.Fprintf(stdout, "lint: %s: %s\n", diagnostic.Path, diagnostic.Message)
	}
	return lintViolationExit
}

func findDiagnostics(ctx context.Context, root string) ([]diagnostic, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	var diagnostics []diagnostic
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}

		fileDiagnostics, err := lintGoFile(root, path)
		if err != nil {
			return err
		}
		diagnostics = append(diagnostics, fileDiagnostics...)
		return nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}

	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Path == diagnostics[j].Path {
			return diagnostics[i].Message < diagnostics[j].Message
		}
		return diagnostics[i].Path < diagnostics[j].Path
	})
	return diagnostics, nil
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".agents", ".codex", ".git", "_bmad", "_bmad-output":
		return true
	}
	return false
}

func lintGoFile(root string, path string) ([]diagnostic, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	formatted, err := format.Source(data)
	if err != nil {
		return nil, fmt.Errorf("format %s: %w", path, err)
	}
	if bytes.Equal(data, formatted) {
		return nil, nil
	}

	relativePath, err := filepath.Rel(root, path)
	if err != nil {
		return nil, fmt.Errorf("rel %s: %w", path, err)
	}
	return []diagnostic{{
		Path:    filepath.ToSlash(relativePath),
		Message: "not gofmt-formatted",
	}}, nil
}

type diagnostic struct {
	Path    string
	Message string
}

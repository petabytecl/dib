package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"time"
)

const (
	dependencyViolationExit = 1
	executionFailureExit    = 2
	goListTimeout           = 2 * time.Minute
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), goListTimeout)
	defer cancel()

	os.Exit(run(ctx, ".", os.Stdout, os.Stderr))
}

func run(ctx context.Context, dir string, stdout io.Writer, stderr io.Writer) int {
	violations, err := findViolations(ctx, dir)
	if err != nil {
		fmt.Fprintf(stderr, "depgate execution error: %v\n", err)
		return executionFailureExit
	}

	if len(violations) == 0 {
		return 0
	}

	for _, violation := range violations {
		fmt.Fprintf(stdout, "non-standard import: package=%s import=%s\n", violation.Package, violation.Import)
	}
	return dependencyViolationExit
}

func findViolations(ctx context.Context, dir string) ([]violation, error) {
	output, err := runGoList(ctx, dir)
	if err != nil {
		return nil, err
	}

	packages, err := decodePackages(output)
	if err != nil {
		return nil, err
	}

	return collectViolations(packages), nil
}

func runGoList(ctx context.Context, dir string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-deps", "-test", "-e", "-json", "-buildvcs=false", "./...")
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("go list timed out after %s", goListTimeout)
		}
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("go list failed: %w: %s", err, stderr.String())
		}
		return nil, fmt.Errorf("go list failed: %w", err)
	}

	return stdout.Bytes(), nil
}

func decodePackages(data []byte) ([]listedPackage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var packages []listedPackage

	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				return packages, nil
			}
			return nil, fmt.Errorf("decode go list JSON: %w", err)
		}
		packages = append(packages, pkg)
	}
}

func collectViolations(packages []listedPackage) []violation {
	seen := make(map[violation]struct{})
	for _, pkg := range packages {
		if pkg.ImportPath == "" || allowedPackage(pkg) {
			continue
		}

		v := violation{
			Package: packageContext(pkg),
			Import:  pkg.ImportPath,
		}
		seen[v] = struct{}{}
	}

	violations := make([]violation, 0, len(seen))
	for v := range seen {
		violations = append(violations, v)
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Package == violations[j].Package {
			return violations[i].Import < violations[j].Import
		}
		return violations[i].Package < violations[j].Package
	})

	return violations
}

func allowedPackage(pkg listedPackage) bool {
	if pkg.Standard {
		return true
	}
	return pkg.Module != nil && pkg.Module.Main
}

func packageContext(pkg listedPackage) string {
	if pkg.Error != nil && len(pkg.Error.ImportStack) > 0 && pkg.Error.ImportStack[0] != "" {
		return pkg.Error.ImportStack[0]
	}
	return pkg.ImportPath
}

type listedPackage struct {
	ImportPath string         `json:"ImportPath"`
	Standard   bool           `json:"Standard"`
	Module     *listedModule  `json:"Module"`
	Error      *listedError   `json:"Error"`
	DepsErrors []*listedError `json:"DepsErrors"`
}

type listedModule struct {
	Main bool `json:"Main"`
}

type listedError struct {
	ImportStack []string `json:"ImportStack"`
	Pos         string   `json:"Pos"`
	Err         string   `json:"Err"`
}

type violation struct {
	Package string
	Import  string
}

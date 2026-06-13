// Command coverage enforces per-package test-coverage thresholds for dib's
// public runtime packages.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	thresholdViolationExit = 1
	executionFailureExit   = 2

	minCoveragePct = 85.0
)

type threshold struct {
	pkg    string
	minPct float64
}

//nolint:gochecknoglobals // Threshold table is the tool's configuration and is referenced directly by tests.
var thresholds = []threshold{
	{pkg: "github.com/petabytecl/dib/command", minPct: minCoveragePct},
	{pkg: "github.com/petabytecl/dib/config", minPct: minCoveragePct},
	{pkg: "github.com/petabytecl/dib/flags", minPct: minCoveragePct},
	{pkg: "github.com/petabytecl/dib/cli", minPct: minCoveragePct},
}

var coverageRE = regexp.MustCompile(`coverage: (\d+\.\d+)%`)

func main() {
	os.Exit(run(os.Stdout, os.Stderr))
}

func run(stdout, stderr io.Writer) int {
	cmd := exec.CommandContext(context.Background(), "go", "test", "-cover", "./command", "./config", "./flags", "./cli")
	out, err := cmd.CombinedOutput()
	if err != nil && len(strings.TrimSpace(string(out))) == 0 {
		_, _ = fmt.Fprintf(stderr, "coverage execution error: %v\n", err)
		return executionFailureExit
	}
	return check(out, stdout, stderr)
}

func check(goTestOutput []byte, stdout, stderr io.Writer) int {
	coverages := make(map[string]float64)
	for _, line := range strings.Split(string(goTestOutput), "\n") {
		for _, t := range thresholds {
			if strings.Contains(line, t.pkg) {
				m := coverageRE.FindStringSubmatch(line)
				if m != nil {
					pct, _ := strconv.ParseFloat(m[1], 64)
					coverages[t.pkg] = pct
				}
			}
		}
	}

	for _, t := range thresholds {
		if _, ok := coverages[t.pkg]; !ok {
			_, _ = fmt.Fprintf(stderr, "coverage execution error: no coverage data for %s\n", t.pkg)
			return executionFailureExit
		}
	}

	failed := false
	for _, t := range thresholds {
		pct := coverages[t.pkg]
		status := "PASS"
		if pct < t.minPct {
			status = "FAIL"
			failed = true
		}
		_, _ = fmt.Fprintf(stdout, "coverage: %s: %.1f%% (threshold %.0f%%) %s\n", t.pkg, pct, t.minPct, status)
	}

	if failed {
		return thresholdViolationExit
	}
	return 0
}

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoveragePassesPackagesMeetingThreshold(t *testing.T) {
	out := fakeCoverageOutput(90.0, 90.0, 90.0, 90.0)
	var stdout, stderr bytes.Buffer
	code := check([]byte(out), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("check() = %d, want 0; stderr=%q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("check() wrote unexpected stderr=%q", stderr.String())
	}
	for _, t2 := range thresholds {
		if !strings.Contains(stdout.String(), t2.pkg) {
			t.Fatalf("stdout missing %q:\n%s", t2.pkg, stdout.String())
		}
	}
	if !strings.Contains(stdout.String(), "PASS") {
		t.Fatalf("stdout missing PASS:\n%s", stdout.String())
	}
}

func TestCoverageFailsPackagesBelowThreshold(t *testing.T) {
	out := fakeCoverageOutput(50.0, 90.0, 90.0, 90.0)
	var stdout, stderr bytes.Buffer
	code := check([]byte(out), &stdout, &stderr)
	if code != thresholdViolationExit {
		t.Fatalf("check() = %d, want %d (threshold violation)", code, thresholdViolationExit)
	}
	if !strings.Contains(stdout.String(), "FAIL") {
		t.Fatalf("stdout missing FAIL:\n%s", stdout.String())
	}
}

func TestCoverageSeparatesExecutionFailuresFromThresholdViolations(t *testing.T) {
	badOutput := "FAIL\tgithub.com/petabytecl/dib/command [build failed]"
	var stdout, stderr bytes.Buffer
	code := check([]byte(badOutput), &stdout, &stderr)
	if code != executionFailureExit {
		t.Fatalf("check() = %d, want %d (execution failure)", code, executionFailureExit)
	}
	if stdout.Len() != 0 {
		t.Fatalf("check() wrote stdout on execution failure: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "coverage execution error:") {
		t.Fatalf("stderr = %q, want coverage execution error message", stderr.String())
	}
}

func TestCoverageReportsDeterministicOutput(t *testing.T) {
	out := fakeCoverageOutput(90.0, 90.0, 90.0, 90.0)
	var stdout bytes.Buffer
	check([]byte(out), &stdout, &bytes.Buffer{})

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != len(thresholds) {
		t.Fatalf("output has %d lines, want %d", len(lines), len(thresholds))
	}
	for i, th := range thresholds {
		if !strings.Contains(lines[i], th.pkg) {
			t.Fatalf("output line %d = %q, want to contain %q", i, lines[i], th.pkg)
		}
	}
}

func TestCoverageCommandRunsFromRepositoryRoot(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	cmd := exec.Command("go", "run", "./tools/coverage")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "go-build"))

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go run ./tools/coverage failed: %v\nstdout: %s", err, string(out))
	}
	output := string(out)
	for _, th := range thresholds {
		if !strings.Contains(output, th.pkg) {
			t.Fatalf("output missing %q:\n%s", th.pkg, output)
		}
	}
}

func TestCoveragePassesAtExactThreshold(t *testing.T) {
	// Package at exactly the threshold (85.0%) should PASS, not FAIL.
	out := fakeCoverageOutput(85.0, 85.0, 85.0, 85.0)
	var stdout, stderr bytes.Buffer
	code := check([]byte(out), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("check() = %d, want 0 for coverage at exact threshold; stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "PASS") {
		t.Fatalf("stdout missing PASS for coverage at exact threshold:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "FAIL") {
		t.Fatalf("stdout contains FAIL for coverage at exact threshold:\n%s", stdout.String())
	}
}

func TestCoverageFailsAllPackagesBelowThreshold(t *testing.T) {
	out := fakeCoverageOutput(50.0, 60.0, 70.0, 80.0)
	var stdout, stderr bytes.Buffer
	code := check([]byte(out), &stdout, &stderr)
	if code != thresholdViolationExit {
		t.Fatalf("check() = %d, want %d when all packages fail", code, thresholdViolationExit)
	}
	output := stdout.String()
	failCount := strings.Count(output, "FAIL")
	if failCount != len(thresholds) {
		t.Fatalf("stdout has %d FAIL markers, want %d (one per package):\n%s", failCount, len(thresholds), output)
	}
	if strings.Contains(output, "PASS") {
		t.Fatalf("stdout contains unexpected PASS when all packages fail:\n%s", output)
	}
}

func TestCoverageOutputFormatVerifiesAllFields(t *testing.T) {
	// Verify the exact output format: "coverage: <pkg>: X.X% (threshold Y%) PASS"
	out := fakeCoverageOutput(90.0, 90.0, 90.0, 90.0)
	var stdout bytes.Buffer
	check([]byte(out), &stdout, &bytes.Buffer{})

	for _, th := range thresholds {
		expected := fmt.Sprintf("coverage: %s: 90.0%% (threshold 85%%) PASS", th.pkg)
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("stdout missing expected output line %q:\n%s", expected, stdout.String())
		}
	}
}

func TestCoverageExecutionFailureWhenConfigMissing(t *testing.T) {
	// Only command data present; config and flags missing → execution failure.
	partial := fmt.Sprintf("ok  \tgithub.com/petabytecl/dib/command\t0.006s\tcoverage: 90.0%% of statements")
	var stdout, stderr bytes.Buffer
	code := check([]byte(partial), &stdout, &stderr)
	if code != executionFailureExit {
		t.Fatalf("check() = %d, want %d when config data is absent", code, executionFailureExit)
	}
	if stdout.Len() != 0 {
		t.Fatalf("check() wrote stdout on execution failure: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "coverage execution error:") {
		t.Fatalf("stderr = %q, want coverage execution error message", stderr.String())
	}
}

func fakeCoverageOutput(commandPct, configPct, flagsPct, cliPct float64) string {
	return strings.Join([]string{
		fmt.Sprintf("ok  \tgithub.com/petabytecl/dib/command\t0.006s\tcoverage: %.1f%% of statements", commandPct),
		fmt.Sprintf("ok  \tgithub.com/petabytecl/dib/config\t0.004s\tcoverage: %.1f%% of statements", configPct),
		fmt.Sprintf("ok  \tgithub.com/petabytecl/dib/flags\t2.227s\tcoverage: %.1f%% of statements", flagsPct),
		fmt.Sprintf("ok  \tgithub.com/petabytecl/dib/cli\t0.013s\tcoverage: %.1f%% of statements", cliPct),
	}, "\n")
}

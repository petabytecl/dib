package flags_test

import (
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

// FuzzParseBoundary verifies parse boundary invariants against arbitrary inputs:
// no panic, RemainingArgs non-nil when args follow --, all-positional input succeeds,
// and the Set is reusable across parse runs.
func FuzzParseBoundary(f *testing.F) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose"),
		flags.String("name", "", "name"),
	)
	if err != nil {
		f.Fatalf("NewSet: %v", err)
	}

	// Each seed is a newline-separated list of CLI args.
	f.Add("")                                     // empty string slice
	f.Add("--")                                   // -- alone
	f.Add("--\n--verbose\n-v")                    // -- followed by flag-like args
	f.Add("pos-a\n--verbose\npos-b\n--name\nfoo") // interspersed positionals and flags
	f.Add("--help")                               // --help alone (unregistered)
	f.Add("--name")                               // flag followed by missing value
	f.Add("--name=attached")                      // flag with attached value

	f.Fuzz(func(t *testing.T, raw string) {
		var args []string
		if raw != "" {
			args = strings.Split(raw, "\n")
		}

		snapshot, err := set.Parse(args)

		if err == nil {
			// When args follow --, RemainingArgs must not be nil.
			for i, a := range args {
				if a == "--" && i+1 < len(args) {
					if snapshot.RemainingArgs() == nil {
						t.Error("RemainingArgs returned nil when args follow --")
					}
					break
				}
			}
		}

		// Parse must succeed when every arg is a positional (no leading -).
		allPositional := true
		for _, a := range args {
			if strings.HasPrefix(a, "-") {
				allPositional = false
				break
			}
		}
		if allPositional && err != nil {
			t.Errorf("all-positional args produced unexpected error: %v", err)
		}

		// Set is reusable: a second parse of the same args must agree on success/failure.
		_, err2 := set.Parse(args)
		if (err == nil) != (err2 == nil) {
			t.Errorf("re-parse produced different error outcome: first=%v, second=%v", err, err2)
		}
	})
}

# Executable Examples

These examples are standalone `main` packages. Run them from the repository
root with Go 1.26 or newer. They import only the Go standard library and Dib.

| Example | Use case | Run |
|---|---|---|
| `examples/basic` | High-level handler dispatch with typed flag-to-config bindings. | `go run ./examples/basic hello --name Ada --shout` |
| `examples/config` | Low-level `Resolve` with flag, env, JSON, and default precedence. | `DIB_REGION=eu-south go run ./examples/config deploy --workers 4` |
| `examples/inspection` | Inspect routing, remaining args, and resolved config without invoking a handler. | `go run ./examples/inspection events list --tenant acme --limit 25 severity=error` |
| `examples/errors` | Convert typed Dib errors into deterministic user-facing diagnostics. | `go run ./examples/errors run --workers many` |
| `examples/multicommand` | Go example tests for distributed command registration and low-level dispatch. | `go test ./examples/multicommand` |
| `examples/migration` | Go example tests for migration-style flag and config behavior. | `go test ./examples/migration` |

Expected output for the standalone examples:

```sh
$ go run ./examples/basic hello --name Ada --shout
HELLO, ADA

$ DIB_REGION=eu-south go run ./examples/config deploy --workers 4
route=deployctl deploy
region=eu-south workers=4 format=json
sources region=env workers=flag binding format=JSON

$ go run ./examples/inspection events list --tenant acme --limit 25 severity=error
path=auditctl events list
remaining=severity=error
tenant=acme limit=25

$ go run ./examples/errors run --workers many
error=flags: flag value conversion failed at "--workers" for "workers"
parse_error token="--workers" name="workers" category="flag value conversion failed"
exit status 2
```

The final `exit status 2` line is emitted by `go run` because the example
program returns the same non-zero exit code a real CLI wrapper would return.

The same behavior is covered by normal Go tests, so `go test ./...` verifies
that the examples still compile and match their documented output.

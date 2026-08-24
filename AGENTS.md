# Agent Rules — edgeworkers-cli

Rules for AI agents and contributors working in this repository.

## Formatting

Run `gofmt -w <file>` on every `.go` file modified in a session before finishing.
Alternatively, run `make format` from the repo root to format all Go files at once.
Unformatted code must not be committed.

```
make format
```

The `make lint` target runs `go vet ./...` and should also pass cleanly.

## Go Documentation

Follow [Go doc comments](https://go.dev/doc/comment). Documentation is part of
the API contract, not optional prose.

- Every non-test package needs a package comment immediately before `package`.
  Start library comments with `Package <name>` and command comments with
  `Command <name>`; state purpose and, for substantial packages, key API areas.
- Every exported declaration needs a doc comment: packages, types, aliases,
  functions, methods, constants, variables, struct fields, and interface methods.
  Begin with its identifier, or with `Package` or `Command` for package docs.
- Explain material behavior and contracts: formats, units, valid values, ownership,
  mutation, errors, concurrency, security, and lifecycle constraints when relevant.
- Keep comments concise, complete sentences. Do not restate syntax or narrate
  self-evident code. Document intent when it explains a non-obvious decision.
- Verify rendered documentation with `go doc`; `go vet` does not check completeness.
- Unexported helpers only need a comment when the reason is non-obvious.

Good:
```go
// parseEdgeRc reads a section from an .edgerc INI file.
func parseEdgeRc(path, section string) (*Credentials, error) {
```

Bad:
```go
// This function parses the edgerc file and returns credentials by reading INI sections.
func parseEdgeRc(path, section string) (*Credentials, error) {
```

Do not add comments to close braces, blank lines, or section separators.

## Test Layout

Go distinguishes two test scopes. Use each in the right place.

### Unit tests — co-located

Unit tests live alongside the source file they test, in the same package directory.
File naming: `<source>_test.go` (e.g., `client_test.go` next to `client.go`).

Use the package under test, such as `package edgeworkers` or `package edgekv`, when a test needs unexported symbols. Use the matching `_test` package when testing only the public surface.

This is the standard Go approach and matches existing test files.

### Integration and acceptance tests — `/tests`

Tests that cross package boundaries, hit external services (even mocked), require
fixtures, or validate end-to-end flows go in `/tests`.

```
tests/
  integration/       # tests that call multiple packages together
  testdata/          # shared fixtures, golden files, sample .edgerc files
```

Create subdirectories as needed. Keep test binaries out of `/tests` — only `*_test.go`
files and data files belong here.

### Running tests

```bash
# All tests (unit + integration)
go test ./...

# Verbose
make test

# Single package
go test ./internal/edgekv/... -v
```

## Command Groups

Commands are organized into groups by API resource or concern — not by HTTP method.
Each group maps to one file, one parent cobra command, and one API resource or concern.

Examples: `ids` owns EdgeWorker ID CRUD; `versions` owns bundle upload/download/validate;
`activations` owns network activation lifecycle.

**When to add to an existing group:** if the new command operates on the same resource
(same API path prefix) as an existing group, add it there.

**When to create a new group:** if the new command targets a distinct resource or concern
with no existing group. New group = new file + new `new<Domain>Cmd()` + registration in
`root.go`.

The parent command itself carries no `RunE` — it exists only to namespace the
subcommands. This keeps `--help` output clean and makes intent obvious.

### Adding a new command to an existing group

1. Add the subcommand constructor inside the owning CLI package: `internal/<product>/cli/<domain>.go`.
2. Register it with `cmd.AddCommand(...)` inside `new<Domain>Cmd()`.
3. Add unit tests in `<domain>_test.go`.
4. Run `gofmt -w $(rg --files -g '*.go')` before finishing.

### Adding a new command group

1. Create `internal/<product>/cli/<domain>.go` with the Cobra commands and add remote behavior to the matching `api` package.
2. Create co-located `*_test.go` files in the affected `cli` and `api` packages.
3. Register with `rootCmd.AddCommand(new<Domain>Cmd())` in that CLI's `root.go`.
4. Run `gofmt -w $(rg --files -g '*.go')` before finishing.

Each CLI domain file owns its Cobra constructors (`new<Domain>Cmd()`); the matching API
domain file owns remote request types and methods on `*Client`.

## Module and Package Rules

- Module: `github.com/akamai/edgeworkers-cli`
- Application code lives under `internal/`. Product-specific remote APIs belong in `internal/<product>/api`, Cobra adapters in `internal/<product>/cli`, and MCP adapters in `internal/<product>/mcp`; share only code with the same invariant in a focused internal package such as `internal/apiclient/`.
- `cmd/edgeworkers/main.go` and `cmd/edgekv/main.go` own release-version injection and process exit, then call their CLI package's `Execute` function.
- Do not add new top-level packages without a clear architectural reason.

## Release Artifacts

- `make build` writes current-platform binaries to `dist/edgeworkers/` and `dist/edgekv/`.
- `make release` writes each CLI's six Akamai binaries, six Spin archives, and Spin manifest to its own directory under `dist/`.
- Artifact names are `akamai-<cli>-<os>-<arch>` and `<cli>-<version>-<spin-os>-<spin-arch>.tar.gz`.
- Each command's `cli.json` version is its sole version source. EdgeWorkers and EdgeKV release independently.
- Keep local release directories separate. GitHub Release assets are flat, so external URLs use unique artifact filenames without CLI directory segments.

## Documentation

- Keep root `README.md` concise: product overview, installation, authentication, build, artifact layout, and links to detailed docs.
- Keep `docs/edgeworkers.md` and `docs/edgekv.md` complete for their respective public command surfaces.
- Update the owning detailed doc and root README when changing user-facing commands, flags, output, authentication, packaging, or installation behavior.

## Dependencies

Add dependencies only when necessary. Prefer stdlib.
After adding a dependency run `go mod tidy`.

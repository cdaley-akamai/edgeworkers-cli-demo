# Contributing

Contributions improve EdgeWorkers and EdgeKV command behavior, tests, and documentation.

## Before Changing Code

1. Open an issue for substantial behavior changes, public command changes, or design questions.
2. Keep each change focused on one problem.
3. Preserve existing command groups. A command for an existing API resource belongs in that resource's group; add a group only for a distinct API resource or concern.

## Development Setup

Install Go 1.26.6 or later, then clone the repository and run:

```bash
make test
make lint
```

`make lint` runs `go vet ./...`. No Akamai credentials are needed for unit tests. Do not commit `.edgerc` files, tokens, or other credentials.

## Code and Tests

- Put product code under `internal/edgeworkers/` or `internal/edgekv/`.
- Keep shared API-client behavior in focused `internal/` packages only when both CLIs share its invariant.
- Add unit tests beside the code under test using Go's standard `*_test.go` naming.
- Put integration or acceptance tests that cross packages or use fixtures in `tests/`.
- Add Go doc comments for exported declarations and packages.
- Format all modified Go files before submitting:

```bash
gofmt -w $(rg --files -g '*.go')
```

## Documentation

Update the relevant CLI guide in `docs/` when a user-visible command, flag, output, authentication flow, packaging behavior, or installation behavior changes. Update `README.md` when a project-level setup or release workflow changes.

## Pull Requests

Describe behavior change, tests run, and documentation updates. Include sanitized command output for user-facing changes. Keep pull requests reviewable and do not mix unrelated refactors.

## Reporting Bugs

Use [GitHub Issues](https://github.com/akamai/cli-edgeworkers/issues) for bugs and feature requests. Include command, CLI version, operating system, expected behavior, actual behavior, and reproducible steps. Remove credentials, access tokens, customer data, and other sensitive values before posting.

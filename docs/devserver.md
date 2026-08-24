# EdgeWorkersDevServer

EdgeWorkersDevServer is a placeholder that stands in for a real EdgeWorkers dev server. It
demonstrates the lifecycle a real one would need — a separate, independently versioned
executable that the `edgeworkers` CLI installs, updates, and runs as a subprocess for
`dev run`, `dev serve`, and `dev playground`. It does not yet execute real EdgeWorker code
or implement a request/response schema. See [Placeholder behavior](#placeholder-behavior)
for exactly what is and is not implemented today.

## Subcommands

EdgeWorkersDevServer's subcommands, invoked by the parent `edgeworkers` CLI:

| Subcommand | Behavior |
| --- | --- |
| `run` | Reads one JSON request from stdin, writes the JSON response to stdout, exits. |
| `serve --port --debug` | Starts a long-lived HTTP server. `POST /run` accepts a JSON body and returns a JSON response synchronously; many requests can be handled in sequence. `--debug` opens a second placeholder listener. |
| `playground --port --no-browser` | Same HTTP server as `serve`, plus a `GET /` page with a code editor, request panel, and response panel (see [Playground page](#playground-page)). Opens the page in the OS default browser once the server is listening; pass `--no-browser` to skip this. |

Run `edgeworkersdevserver playground` directly instead of through `dev playground` —
behavior is identical since the server is fully standalone.

All startup, readiness, and error messages are logged to stderr, never stdout, so `run`'s
stdout stays a clean JSON pipe.

## Placeholder behavior

These are deliberate, documented gaps — not bugs:

- **No real EdgeWorker execution.** `run`/`serve`/`playground` all echo the request JSON
  body back unchanged as the response. There is no bundle loading, no JS runtime, and no
  request/response JSON schema yet.
- **No debug protocol.** The `--debug` port on `serve` accepts connections and logs them,
  but responds `501 Not Implemented` to everything. The protocol it will eventually speak
  is not yet defined.
- **No code signing.** Downloaded binaries are verified with a sha256 checksum from the
  manifest only.

## Playground page

`playground` serves a single static page (`internal/edgeworkers/devserver/playground.html`,
embedded via `go:embed`) with three panels:

1. **Code** — a mock EdgeWorker handler; editing it has no effect yet.
2. **Request** — the JSON body to send.
3. **Response** — read-only, populated after clicking **Run**.

Clicking **Run** mock-saves the code (no persistence), `POST`s the request panel's JSON to
`/run` on the same server, shows a brief "thinking" state, then renders the response panel.

## Installing and updating

Install or update the local `edgeworkersdevserver` binary with the `dev` commands below:

```bash
akamai edgeworkers dev update              # install the manifest's current version
akamai edgeworkers dev install <version>   # install a specific version
akamai edgeworkers dev list                # list versions known to the manifest
```

`dev run`, `dev serve`, and `dev playground` accept `--server-version` to use a version
other than current.

## Manifest

`devserver-manifest.json` is what `dev install`, `dev update`, and `dev list` read to
know which EdgeWorkersDevServer versions exist and where to download them:

```json
{
  "current": {
    "version": "0.1.0",
    "releaseDate": "2026-09-10",
    "packages": []
  },
  "previous": []
}
```

Each `packages` entry is `{os, arch, url, sha256}`, using Go's `runtime.GOOS`/`GOARCH`
values directly. At runtime the CLI fetches this file from
`https://raw.githubusercontent.com/<repository>/main/devserver-manifest.json`, overridable
with the `AKAMAI_EDGEWORKERS_DEVSERVER_MANIFEST_URL` environment variable.

## Releasing

See the [Release Walkthrough](release.md) for the build and publish steps: bump
`DEVSERVER_VERSION` in `versions.mk`, then `make release-devserver` builds six platform
binaries and rewrites `devserver-manifest.json`, moving the previous `current` release
into `previous`.

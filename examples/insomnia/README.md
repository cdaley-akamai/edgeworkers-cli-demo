# EdgeWorkersDevServer Insomnia collection

`edgeworkersdevserver.insomnia.json` is an Insomnia v4 export with nine
requests against the EdgeWorkersDevServer `POST /run` endpoint (see
[docs/devserver.md](../../docs/devserver.md)). It's meant to be run with the
`inso` CLI — there are no desktop-app/import instructions here.

## Prerequisites

- Install `inso`: `npm install -g insomnia-inso` (or `brew install inso`).
- Start the dev server: `akamai edgeworkers dev serve` (listens on port 9998
  by default, matching this collection's `base_url` environment variable).

## Running

From the repo root. `-w/--workingDir` points `inso` at the export file (its
data source); the positional identifier is the collection's workspace id,
`wrk_edgeworkersdevserver`.

```bash
# Run every request in the collection
inso run collection wrk_edgeworkersdevserver -w examples/insomnia/edgeworkersdevserver.insomnia.json

# Run a single scenario by name (regex match)
inso run collection wrk_edgeworkersdevserver -w examples/insomnia/edgeworkersdevserver.insomnia.json -t "02 -"

# Run a single scenario by request id
inso run collection wrk_edgeworkersdevserver -w examples/insomnia/edgeworkersdevserver.insomnia.json -i req_02_geo_block_canada

# Point at a non-default host/port without editing the file
inso run collection wrk_edgeworkersdevserver -w examples/insomnia/edgeworkersdevserver.insomnia.json --env-var base_url=http://localhost:9001

# Fail fast in CI, with compact output
inso run collection wrk_edgeworkersdevserver -w examples/insomnia/edgeworkersdevserver.insomnia.json --ci -b -r min
```

`-r/--reporter` accepts `dot`, `list`, `min`, `progress`, `spec`, or `tap` for
different output formats; `-b/--bail` stops after the first non-2xx response.

## Placeholder behavior

EdgeWorkersDevServer doesn't execute real EdgeWorker code yet — `/run`
validates the request body and echoes it back unchanged (see "Placeholder
behavior" in [docs/devserver.md](../../docs/devserver.md)). The "expected"
outcomes below are aspirational: these request bodies are shaped to
demonstrate that behavior once real execution is implemented.

## Scenarios

| # | Name | Scenario | Expected once execution exists |
| --- | --- | --- | --- |
| 01 | Minimal GET | Bare `onClientRequest`, no headers or optional fields | 200 baseline |
| 02 | Geo-block Canada | `features.userLocation.country = "CA"` | 403 from a geo-blocking handler |
| 03 | Missing Authorization header | No `Authorization` header on `/account` | 401 from an auth-gate handler |
| 04 | Malformed Authorization header | `Authorization: Bearer` with no token | 401 from auth parsing |
| 05 | Bot signal present | `features.botScore` with a high score | Block/challenge from bot mitigation |
| 06 | A/B test cookie missing | No `Cookie` header (no `testGroup`) | New-visitor A/B group assignment |
| 07 | Mobile device request | `request.device.type = "mobile"` | Device-based routing/redirect |
| 08 | Malformed Accept-Language header | `Accept-Language: xx-yy;q=abc` | Graceful fallback, not an error |
| 09 | onOriginResponse missing Cache-Control | Origin response has no `Cache-Control` | Caching logic supplies a default |

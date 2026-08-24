# EdgeWorkers CLI

`edgeworkers` manages Akamai EdgeWorkers. Run it through Akamai CLI as `akamai edgeworkers`, or use alias `akamai ew`.

## Authentication

Credentials default to `~/.edgerc`, section `default`. All commands accept:

| Flag | Description |
| --- | --- |
| `--edgerc PATH` | Credentials file path. |
| `--edgerc-section NAME` | `.edgerc` section. |
| `--cli-config-section NAME` | CLI configuration section for stored defaults. |
| `--account-switch-key KEY` | Account context. |
| `--timeout SECONDS` | HTTP timeout. |
| `--debug` | EdgeGrid HTTP tracing. |
| `--json` | JSON output. |

Use `edgeworkers config set` to store defaults in `~/.akamai-cli/ew-config.json`.

## Command Groups

Commands are grouped by EdgeWorkers API resource or operational concern. Start with `akamai edgeworkers <group> --help` to see a group's commands, then run `akamai edgeworkers <group> <command> --help` for arguments and flags. Use `akamai ew` instead of `akamai edgeworkers` if preferred.

| Group | Use it for | Find commands |
| --- | --- | --- |
| `ids` | Create, inspect, update, clone, and remove EdgeWorker IDs and manage resource tiers. | `akamai edgeworkers ids --help` |
| `versions` | Upload, validate, download, list, and delete EdgeWorker code bundle versions. | `akamai edgeworkers versions --help` |
| `activations` | List, activate, and deactivate versions on staging or production. | `akamai edgeworkers activations --help` |
| `revisions` | Inspect revision history, compare revisions, activate or pin revisions, and download revision data. | `akamai edgeworkers revisions --help` |
| `account` | List account groups, contracts, properties, resource tiers, and limits. | `akamai edgeworkers account --help` |
| `reports` | List and retrieve usage reports and active-customer data. | `akamai edgeworkers reports --help` |
| `logging` | Read and update EdgeWorkers logging configuration. | `akamai edgeworkers logging --help` |
| `debug` | Obtain debug secrets and authorization tokens. | `akamai edgeworkers debug --help` |
| `config` | Save and manage persistent CLI defaults. | `akamai edgeworkers config --help` |
| `dev` | Run, serve, or preview EdgeWorkers locally via EdgeWorkersDevServer, and manage installed EdgeWorkersDevServer versions. | `akamai edgeworkers dev --help` |
| `mcp` | Run the EdgeWorkers MCP server over stdio. | `akamai edgeworkers mcp` |

```text
edgeworkers ids list|create|update|delete|clone|resource-tier
edgeworkers versions list|upload|download|delete|validate
edgeworkers activations list|activate|deactivate
edgeworkers revisions list|get|compare|activate|pin|unpin|bom|download|activations
edgeworkers account groups|contracts|properties|limits
edgeworkers account resource-tiers <contract-id>
edgeworkers reports list|get|active-customers
edgeworkers logging get|set
edgeworkers debug secret|auth-token
edgeworkers config list|get|set|unset|save
edgeworkers dev run|update|install|list
edgeworkers mcp
```

## Persistent Configuration

The shared configuration file is JSON5. Top-level keys are profiles selected by `--cli-config-section`; each profile can contain independent `edgeworkers` and `edgekv` objects.

```json5
{
	default: {
		edgeworkers: {
			section: "default",
			accountSwitchKey: "1-ABCDE",
			timeout: 120,
			mcp: {
				tools: ["list*", "get*", "hello_world"],
			},
		},
	},
}
```

Supported keys are `edgerc`, `section`, `accountSwitchKey`, `timeout`, and `mcp.tools`. Explicit command flags override the selected profile. An MCP tool's `accountSwitchKey` input overrides both for that request.

Manage values through the CLI; quote the tool array so the shell passes it unchanged:

```bash
akamai edgeworkers config set section production
akamai edgeworkers config set mcp.tools '["list*", "get*"]'
akamai edgeworkers config get mcp.tools
akamai edgeworkers config unset accountSwitchKey
```

`config save` stores only the persistent flags explicitly supplied with that invocation. Writes are atomic and set the configuration file to owner-only permissions.

## MCP Server

Configure an MCP client to launch the installed command. For example, clients using an `mcpServers` object can launch the Akamai CLI package form as follows:

```json
{
	"mcpServers": {
		"akamai-edgeworkers": {
			"command": "akamai",
			"args": ["edgeworkers", "mcp"]
		}
	}
}
```

The Spin plugin command is `spin edgeworkers mcp`. The server communicates only through newline-delimited MCP JSON on stdout; diagnostics go to stderr. It stops cleanly on `SIGINT` or `SIGTERM`. Initialization, tool listing, and `hello_world` do not load credentials. Each product API tool invocation creates a fresh client from the selected `.edgerc` credentials.

`mcp.tools` accepts case-sensitive portable globs using `*`, `?`, character classes such as `[A-Z]`, and alternatives such as `{list,get}*`. The default `['*']` enables every product tool. An empty array disables every product tool. `hello_world` is always available regardless of filters, and invalid patterns stop server startup.

The EdgeWorkers server exposes these 42 product tools plus `hello_world`:

```text
listEdgeworkers                getEdgeworker                   createEdgeworker
updateEdgeworker               deleteEdgeworker                cloneEdgeworker
getEdgeworkerResourceTier      listVersions                    getVersion
createVersion                  deleteVersion                   downloadVersionContent
listActivations                getActivation                   activateEdgeworker
cancelActivation               rollbackEdgeworker              listDeactivations
getDeactivation                deactivateEdgeworker             listProperties
listGroups                     getGroup                        listContracts
listResourceTiers              listLimits                      listReports
getReport                      listLoggingOverrides             getLoggingOverride
createLoggingOverride          createSecureToken               validateCodeBundle
listRevisions                  getRevision                     getRevisionBom
downloadRevisionContent        compareRevisions                listRevisionActivations
activateRevision               pinRevision                     unpinRevision
```

Use the MCP client's tool inspector for each tool's input schema and description. Destructive tools require their documented `confirm` input; declining confirmation returns a normal abort result and makes no API call.

## Local Development Commands

`dev` commands exec **EdgeWorkersDevServer**, a separate executable installed locally at `~/.akamai/edgeworkers/devserver/<version>/`. Install it with `dev install <version>` or `dev update` before using `run`, `serve`, or `playground`.

| Command | Behavior |
| --- | --- |
| `dev run` | Pipes one JSON request from stdin to EdgeWorkersDevServer and writes its JSON response to stdout. Accepts `--server-version` to pick an installed version other than the current one. |
| `dev serve` | Starts EdgeWorkersDevServer as a long-lived HTTP server accepting `POST /run` with a JSON body, returning JSON responses synchronously. Accepts `--port` (default `9998`) and `--debug` (default `9999`, reserved for a future debug protocol). |
| `dev playground` | Starts EdgeWorkersDevServer with the same `/run` endpoint plus an interactive playground page, then opens it in the default browser. Accepts `--port` (default `8888`) and `--no-browser` to skip the automatic browser launch. Works identically when run directly as `edgeworkersdevserver playground`, bypassing this CLI. |
| `dev update` | Installs the manifest's current EdgeWorkersDevServer version if a different version is installed locally. |
| `dev install <version>` | Installs a specific EdgeWorkersDevServer version listed in the manifest. |
| `dev list` | Lists EdgeWorkersDevServer versions from the manifest (release date, version, and a marker for the manifest's current version). Accepts `--search` for a fuzzy filter. |

EdgeWorkersDevServer's own request/response JSON schema is not yet defined; the current release echoes the request body back as the response. The manifest listing available versions is fetched from `devserver-manifest.json`, overridable with the `AKAMAI_EDGEWORKERS_DEVSERVER_MANIFEST_URL` environment variable.

### Examples

```bash
akamai edgeworkers ids list --group-id 12345
akamai edgeworkers versions upload 12345 --bundle bundle.tgz
akamai edgeworkers activations activate 12345 staging 3
```

Run `akamai edgeworkers <group> --help` for command flags and arguments.

# EdgeKV CLI

`edgekv` manages Akamai EdgeKV. Run it through Akamai CLI as `akamai edgekv`, or use alias `akamai ekv`.

## Authentication

Credentials default to `~/.edgerc`, section `default`. All commands accept `--edgerc`, `--edgerc-section`, `--cli-config-section`, `--account-switch-key`, `--timeout`, `--debug`, and `--json`. Use `edgekv config set` to store defaults in `~/.akamai-cli/ew-config.json`.

`network` arguments are `staging` or `production`.

## Command Groups

Commands are grouped by EdgeKV resource or operational concern. Start with `akamai edgekv <group> --help` to see a group's commands, then run `akamai edgekv <group> <command> --help` for arguments and flags. Use `akamai ekv` instead of `akamai edgekv` if preferred.

| Group | Use it for | Find commands |
| --- | --- | --- |
| `database` | Initialize EdgeKV, check initialization status, and set the default data-access policy. | `akamai edgekv database --help` |
| `namespaces` | Create, list, get, and update namespaces; manage namespace data groups and access groups. | `akamai edgekv namespaces --help` |
| `items` | List, read, create or update, and delete data items in a namespace group. | `akamai edgekv items --help` |
| `tokens` | Create, list, download, refresh, and revoke EdgeKV access tokens. | `akamai edgekv tokens --help` |
| `config` | Save and manage persistent CLI defaults. | `akamai edgekv config --help` |
| `mcp` | Run the EdgeKV MCP server over stdio. | `akamai edgekv mcp` |

```text
edgekv database initialize [--data-access-policy POLICY]
edgekv database status
edgekv database set-policy <data-access-policy>

edgekv namespaces list <network> [--details]
edgekv namespaces get <network> <namespace>
edgekv namespaces create <network> <namespace> --retention-days DAYS [flags]
edgekv namespaces update <network> <namespace> --retention-days DAYS
edgekv namespaces groups list <network> <namespace>
edgekv namespaces permissions set <namespace> <group-id>

edgekv items list <network> <namespace> <group-id> [--max-items N] [--sandbox-id ID]
edgekv items get <network> <namespace> <group-id> <item-id> [--sandbox-id ID]
edgekv items put <text|jsonfile> <network> <namespace> <group-id> <item-id> <value>
edgekv items delete <network> <namespace> <group-id> <item-id> [--sandbox-id ID]

edgekv tokens list [--include-expired]
edgekv tokens create <name> --staging allow|deny --production allow|deny --namespaces NAMESPACES
edgekv tokens download <name>
edgekv tokens refresh <name>
edgekv tokens revoke <name>

edgekv config list|get|set|unset|save
edgekv mcp
```

## Persistent Configuration

The shared configuration file is JSON5. Top-level keys are profiles selected by `--cli-config-section`; each profile can contain independent `edgeworkers` and `edgekv` objects.

```json5
{
  default: {
    edgekv: {
      section: "default",
      accountSwitchKey: "1-ABCDE",
      timeout: 120,
      mcp: {
        tools: ["listEdgeKV*", "getEdgeKV*"],
      },
    },
  },
}
```

Supported keys are `edgerc`, `section`, `accountSwitchKey`, `timeout`, and `mcp.tools`. Explicit command flags override the selected profile. An MCP tool's `accountSwitchKey` input overrides both for that request.

```bash
akamai edgekv config set section production
akamai edgekv config set mcp.tools '["listEdgeKV*", "getEdgeKV*"]'
akamai edgekv config get mcp.tools
akamai edgekv config unset accountSwitchKey
```

`config save` stores only the persistent flags explicitly supplied with that invocation. Writes are atomic and set the configuration file to owner-only permissions.

## MCP Server

Configure an MCP client to launch the installed command. For example, clients using an `mcpServers` object can launch the Akamai CLI package form as follows:

```json
{
  "mcpServers": {
    "akamai-edgekv": {
      "command": "akamai",
      "args": ["edgekv", "mcp"]
    }
  }
}
```

The Spin plugin command is `spin edgekv mcp`. The server communicates only through newline-delimited MCP JSON on stdout; diagnostics go to stderr. It stops cleanly on `SIGINT` or `SIGTERM`. Initialization, tool listing, and `hello_world` do not load credentials. Each product API tool invocation creates a fresh client from the selected `.edgerc` credentials.

`mcp.tools` accepts case-sensitive portable globs using `*`, `?`, character classes such as `[A-Z]`, and alternatives such as `{list,get}*`. The default `['*']` enables every product tool. An empty array disables every product tool. `hello_world` is always available regardless of filters, and invalid patterns stop server startup.

The EdgeKV server exposes these 24 product tools plus `hello_world`:

```text
getEdgeKVInitializeStatus       initializeEdgeKV
listEdgeKVNamespaces           getEdgeKVNamespace
createEdgeKVNamespace          updateEdgeKVNamespace
deleteEdgeKVNamespace          listEdgeKVNamespaceGroups
reauthorizeEdgeKVNamespace     updateEdgeKVDataAccessPolicy
getEdgeKVScheduledDelete       rescheduleEdgeKVNamespaceDelete
cancelEdgeKVNamespaceDelete    listEdgeKVItems
getEdgeKVItem                  writeEdgeKVItem
deleteEdgeKVItem               listEdgeKVTokens
getEdgeKVToken                 createEdgeKVToken
deleteEdgeKVToken              refreshEdgeKVToken
listEdgeKVGroups               getEdgeKVGroup
```

Use the MCP client's tool inspector for each tool's input schema and description. Destructive tools require their documented `confirm` input; declining confirmation returns a normal abort result and makes no API call.

## Examples

Initialize a database and create a namespace:

```bash
akamai edgekv database initialize \
  --data-access-policy restrictDataAccess=true,allowNamespacePolicyOverride=false
akamai edgekv namespaces create staging products --retention-days 30 --group-id 12345
```

Write and retrieve an item:

```bash
akamai edgekv items put text staging products catalog item-1 '{"sku":"123"}'
akamai edgekv items get staging products catalog item-1
```

Create a token limited to namespace access:

```bash
akamai edgekv tokens create catalog-reader \
  --staging allow --production deny --namespaces products+r
```

Use `--save-path` on token create or download to write or update `edgekv_tokens.js` in an existing directory or `.tgz` bundle. Existing namespace entries require `--overwrite`. Use `--output FILE` to save the raw API token response to a file with mode `0600`.

## Related Documentation

- [EdgeKV user guide](https://techdocs.akamai.com/edgekv/docs)
- [EdgeKV API reference](https://techdocs.akamai.com/edgekv/reference/api)
- [EdgeKV data model and namespace behavior](https://techdocs.akamai.com/edgekv/docs/edgekv-data-model#namespace)
- [EdgeKV limits](https://techdocs.akamai.com/edgekv/docs/limits)
- [Known issues](https://techdocs.akamai.com/edgekv/docs/known-issues)
- [EdgeGrid credential setup](https://techdocs.akamai.com/developer/docs/edgegrid)

Item writes and deletes are eventually consistent. A newly written or deleted item can take time to become visible on EdgeKV networks.

Run `akamai edgekv <group> --help` for command flags and arguments.

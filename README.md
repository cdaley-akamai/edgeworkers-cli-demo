# Akamai EdgeWorkers and EdgeKV CLI

This package provides two Go CLIs for Akamai services:

| CLI | Purpose | Detailed reference |
| --- | --- | --- |
| `edgeworkers` (`akamai ew`) | Manage EdgeWorkers IDs, bundles, deployments, revisions, logging, and reports, or run the EdgeWorkers MCP server. | [EdgeWorkers CLI guide](docs/edgeworkers.md) |
| `edgekv` (`akamai ekv`) | Manage EdgeKV databases, namespaces, items, and access tokens, or run the EdgeKV MCP server. | [EdgeKV CLI guide](docs/edgekv.md) |

Both CLIs use EdgeGrid credentials and can run as Akamai CLI package commands or as [Spin](https://spinframework.dev/) plugins.

## Technical Setup

### Requirements

- An Akamai account with access to EdgeWorkers or EdgeKV APIs.
- EdgeGrid API credentials. Follow the [EdgeGrid setup guide](https://techdocs.akamai.com/developer/docs/edgegrid) to create and download credentials.
- One command host:
  - [Akamai CLI](https://techdocs.akamai.com/developer/docs/cli), for `akamai edgeworkers` and `akamai edgekv`.
  - [Spin](https://spinframework.dev/v3/install), for `spin edgeworkers` and `spin edgekv` plugins.

Install Akamai CLI with Homebrew on macOS:

```bash
brew install akamai
```

Or download an OS-specific binary from [Akamai CLI releases](https://github.com/akamai/cli/releases). See the [Akamai CLI installation instructions](https://techdocs.akamai.com/developer/docs/cli) for supported installation methods.

Install Spin using its [installation guide](https://spinframework.dev/v3/install).

### Authentication

Save the EdgeGrid credentials downloaded from Control Center in `~/.edgerc`. Both CLIs use its `default` section unless `--edgerc` or `--edgerc-section` overrides it. See the [`.edgerc` format reference](https://techdocs.akamai.com/developer/docs/edgegrid#basic-client).

```ini
[default]
host = akaa-xxxxxxxxxxxxxxxx.luna.akamaiapis.net
client_token = akaa-xxxxxxxxxxxxxxxx
client_secret = xxxxxxxxxxxxxxxxxxxx
access_token = akaa-xxxxxxxxxxxxxxxx
```

### MCP Servers

Each binary includes a product-specific [Model Context Protocol](https://modelcontextprotocol.io/) server over stdio:

```bash
akamai edgeworkers mcp
akamai edgekv mcp
```

Configure an MCP client to launch one of those commands, or use the Spin plugin's `spin edgeworkers mcp` and `spin edgekv mcp` forms. Product tools use the same `.edgerc` credentials as the CLI commands — no separate setup. See the product guides for the complete tool catalogs and filtering configuration.

## Install

### Akamai CLI Package

Install both commands with one package. Since this is a custom package, point `akamai install` at the repository URL rather than a package name:

```bash
akamai install https://github.com/cdaley-akamai/edgeworkers-cli-demo.git
```

Update an existing installation:

```bash
akamai update edgeworkers-cli-demo
```

Run either CLI:

```bash
akamai edgeworkers --help
akamai edgekv --help
```

### Spin Plugins

Install Spin first, then install each plugin from a release manifest. Replace version with published release values.

```bash
spin plugins install --url https://github.com/cdaley-akamai/edgeworkers-cli-demo/releases/download/edgeworkers-2.0.0/edgeworkers.json
spin plugins install --url https://github.com/cdaley-akamai/edgeworkers-cli-demo/releases/download/edgekv-2.0.0/edgekv.json
```

After installation, run `spin edgeworkers --help` or `spin edgekv --help`.

## Shell Completion

Both CLIs generate completion scripts for Bash, Zsh, Fish, and PowerShell. Generate the script through the command form you installed, then load it using your shell's normal completion setup.

```bash
# Akamai CLI package
akamai edgeworkers completion zsh > _edgeworkers
akamai edgekv completion zsh > _edgekv

# Spin plugin
spin edgeworkers completion bash > edgeworkers.bash
spin edgekv completion bash > edgekv.bash
```

Run `completion --help` to list supported shells, then `completion <shell> --help` for shell-specific loading instructions.

## Documentation

Each CLI guide explains authentication, command groups, examples, and how to locate full command help:

- [EdgeWorkers CLI guide](docs/edgeworkers.md)
- [EdgeKV CLI guide](docs/edgekv.md)
- [EdgeWorkersDevServer guide](docs/devserver.md)
- [EdgeWorkers documentation](https://techdocs.akamai.com/edgeworkers/docs)
- [EdgeKV documentation](https://techdocs.akamai.com/edgekv/docs)
- [EdgeKV API reference](https://techdocs.akamai.com/edgekv/reference/api)

## Build

Building requires Go 1.26.6 or later.

```bash
make build
```

Current-platform binaries are written to:

```text
dist/edgeworkers/akamai-edgeworkers
dist/edgekv/akamai-edgekv
```

## Release

Full commands, artifact layout, and verification steps: [release walkthrough](docs/release.md).

## Contributing

Contribution workflow, code standards, tests, and documentation expectations: [contributing guide](CONTRIBUTING.md).

## Reporting Issues

Report bugs, regressions, or documentation gaps in [GitHub Issues](https://github.com/akamai/cli-edgeworkers/issues). Include command, version, sanitized output, operating system, and reproduction steps. Never include `.edgerc` credentials or access tokens.

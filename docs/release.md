# Release Walkthrough

EdgeWorkers, EdgeKV, and EdgeWorkersDevServer each release independently. Follow the
section for the one product you're releasing; the other two are unaffected.

## Prerequisites

- Go 1.26.6 or later.
- Write permission for the target GitHub repository and release.
- [GitHub CLI](https://cli.github.com/) or access to GitHub Releases for artifact upload.
- [Spin](https://spinframework.dev/v3/install), only if verifying an EdgeWorkers/EdgeKV
  install (last step of each of those two sections).
- [Node.js](https://nodejs.org/), only if verifying an EdgeWorkers/EdgeKV MCP server.
- [Akamai CLI](https://techdocs.akamai.com/developer/docs/install-and-configure-akamai-cli),
  only if verifying an install (last step of each section).

`versions.mk`, at the repo root, holds all three versions:

```make
EDGEWORKERS_VERSION=2.0.1
EDGEKV_VERSION=2.0.0
DEVSERVER_VERSION=5.1.4
```

## EdgeWorkers

1. Bump `EDGEWORKERS_VERSION` in `versions.mk`.
2. Validate:

   ```bash
   make validate-edgeworkers
   ```

3. Build:

   ```bash
   make release-edgeworkers
   ```

4. Publish:

   ```bash
   make gh-release-edgeworkers
   ```

5. Verify the published install:

   ```bash
   spin plugins install --url https://github.com/cdaley-akamai/edgeworkers-cli-demo/releases/download/edgeworkers-<version>/edgeworkers.json
   spin edgeworkers --help
   akamai install https://github.com/cdaley-akamai/edgeworkers-cli-demo.git
   akamai edgeworkers --help
   ```

## EdgeKV

1. Bump `EDGEKV_VERSION` in `versions.mk`.
2. Validate:

   ```bash
   make validate-edgekv
   ```

3. Build:

   ```bash
   make release-edgekv
   ```

4. Publish:

   ```bash
   make gh-release-edgekv
   ```

5. Verify the published install:

   ```bash
   spin plugins install --url https://github.com/cdaley-akamai/edgeworkers-cli-demo/releases/download/edgekv-<version>/edgekv.json
   spin edgekv --help
   akamai install https://github.com/cdaley-akamai/edgeworkers-cli-demo.git
   akamai edgekv --help
   ```

## EdgeWorkersDevServer

See [devserver.md](devserver.md) for what this product is. It has no Spin packaging.

1. Bump `DEVSERVER_VERSION` in `versions.mk`.
2. Validate:

   ```bash
   make validate-devserver
   ```

3. Build:

   ```bash
   make release-devserver
   ```

4. Publish, then commit the updated `devserver-manifest.json`:

   ```bash
   make gh-release-devserver
   ```

5. Verify the published install:

   ```bash
   akamai edgeworkers dev update
   akamai edgeworkers dev install <version>
   akamai edgeworkers dev list
   ```


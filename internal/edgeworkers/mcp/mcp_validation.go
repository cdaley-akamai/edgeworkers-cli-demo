package mcp

import (
	"context"
	"os"

	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type validateCodeBundleInput struct {
	BundlePath string `json:"bundlePath" jsonschema:"Absolute path to the local GZIP code bundle file to validate. Must be a .tgz or .gz file containing main.js and bundle.json."`
	mcpAccountInput
}

func registerMCPValidationTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "validateCodeBundle",
		Description: "Validate an EdgeWorker code bundle before uploading it as a new version. " +
			"Returns a list of errors and warnings. The bundle must be a GZIP-compressed tarball.",
	}, func(ctx context.Context, input validateCodeBundleInput) (*mcp.CallToolResult, error) {
		bundle, err := os.Open(input.BundlePath)
		if err != nil {
			return nil, err
		}
		defer bundle.Close()
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ValidateCodeBundle(ctx, bundle)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

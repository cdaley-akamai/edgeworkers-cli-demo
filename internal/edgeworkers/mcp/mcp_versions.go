package mcp

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getVersionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	Version      string  `json:"version" jsonschema:"The version identifier (e.g. '1', '2')"`
	mcpAccountInput
}

type createVersionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID to create a version for"`
	BundlePath   string  `json:"bundlePath" jsonschema:"Absolute path to the local GZIP code bundle file (application/gzip)"`
	mcpAccountInput
}

type deleteVersionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	Version      string  `json:"version" jsonschema:"The version identifier to delete"`
	Confirm      bool    `json:"confirm" jsonschema:"Must be true to confirm this destructive operation. Set to true to proceed."`
	mcpAccountInput
}

type downloadVersionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	Version      string  `json:"version" jsonschema:"The version identifier to download"`
	OutputPath   string  `json:"outputPath" jsonschema:"Absolute local file path where the GZIP bundle will be saved"`
	mcpAccountInput
}

func registerMCPVersionTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listVersions",
		Description: "List all versions of an EdgeWorker.",
	}, func(ctx context.Context, input getEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListVersions(ctx, input.EdgeWorkerID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getVersion",
		Description: "Get details for a specific version of an EdgeWorker.",
	}, func(ctx context.Context, input getVersionInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetVersion(ctx, input.EdgeWorkerID, input.Version)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "createVersion",
		Description: "Upload a new version of an EdgeWorker code bundle. " +
			"The bundle must be a GZIP-compressed tarball (.tgz) at the specified local file path.",
	}, func(ctx context.Context, input createVersionInput) (*mcp.CallToolResult, error) {
		bundle, err := os.Open(input.BundlePath)
		if err != nil {
			return nil, err
		}
		defer bundle.Close()
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.CreateVersion(ctx, input.EdgeWorkerID, bundle)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "deleteVersion",
		Description: "Permanently delete a specific version of an EdgeWorker. Requires confirm: true.",
	}, func(ctx context.Context, input deleteVersionInput) (*mcp.CallToolResult, error) {
		if !input.Confirm {
			return mcpserver.AbortResult("deleteVersion"), nil
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		if err := client.DeleteVersion(ctx, input.EdgeWorkerID, input.Version); err != nil {
			return nil, err
		}
		id := strconv.FormatFloat(input.EdgeWorkerID, 'f', -1, 64)
		return mcpserver.JSONResult(map[string]string{
			"message": "Version " + input.Version + " of EdgeWorker " + id + " deleted successfully.",
		})
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "downloadVersionContent",
		Description: "Download the code bundle (GZIP) for a specific EdgeWorker version to a local file.",
	}, func(ctx context.Context, input downloadVersionInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		data, err := client.DownloadVersionContent(ctx, input.EdgeWorkerID, input.Version)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(input.OutputPath, data, 0o600); err != nil {
			return nil, fmt.Errorf("write bundle to %s: %w", input.OutputPath, err)
		}
		return mcpserver.JSONResult(map[string]string{"message": "Bundle saved to " + input.OutputPath})
	})
}

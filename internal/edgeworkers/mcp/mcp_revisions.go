package mcp

import (
	"context"
	"fmt"
	"os"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getRevisionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	RevisionID   string  `json:"revisionId" jsonschema:"The revision ID (e.g. '5-1')"`
	mcpAccountInput
}

type downloadRevisionContentInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	RevisionID   string  `json:"revisionId" jsonschema:"The revision ID (e.g. '5-1')"`
	OutputPath   string  `json:"outputPath" jsonschema:"Absolute local file path where the combined GZIP bundle will be saved"`
	mcpAccountInput
}

type compareRevisionsInput struct {
	EdgeWorkerID      float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	RevisionID        string  `json:"revisionId" jsonschema:"The revision ID to compare against (e.g. '5-1')"`
	CompareRevisionID string  `json:"compareRevisionId" jsonschema:"The other revision ID to compare with (e.g. '4-1')"`
	mcpAccountInput
}

type activateRevisionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	RevisionID   string  `json:"revisionId" jsonschema:"The revision ID to activate as fallback (e.g. '5-1')"`
	mcpAccountInput
}

type pinRevisionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	RevisionID   string  `json:"revisionId" jsonschema:"The revision ID to pin (e.g. '5-1')"`
	PinNote      string  `json:"pinNote,omitempty" jsonschema:"Optional note explaining why this revision is pinned"`
	mcpAccountInput
}

type unpinRevisionInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	RevisionID   string  `json:"revisionId" jsonschema:"The revision ID to unpin (e.g. '5-1')"`
	mcpAccountInput
}

func registerMCPRevisionTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "listRevisions",
		Description: "List revisions for an EdgeWorker. Revisions are used with Flexible Composition " +
			"to track changes across parent and sub-worker dependencies.",
	}, func(ctx context.Context, input getEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListRevisions(ctx, input.EdgeWorkerID, edgeworkersapi.RevisionListOptions{})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getRevision",
		Description: "Get details for a specific revision of an EdgeWorker.",
	}, func(ctx context.Context, input getRevisionInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetRevision(ctx, input.EdgeWorkerID, input.RevisionID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getRevisionBom",
		Description: "Get the Bill of Materials (BOM) for a revision — lists all sub-worker dependencies and their versions.",
	}, func(ctx context.Context, input getRevisionInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetRevisionBOM(ctx, input.EdgeWorkerID, input.RevisionID, edgeworkersapi.RevisionBOMOptions{})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "downloadRevisionContent",
		Description: "Download the combined code bundle (GZIP) for a specific revision to a local file. " +
			"This includes the parent EdgeWorker and all sub-worker dependencies.",
	}, func(ctx context.Context, input downloadRevisionContentInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		data, err := client.DownloadRevisionContent(ctx, input.EdgeWorkerID, input.RevisionID)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(input.OutputPath, data, 0o600); err != nil {
			return nil, fmt.Errorf("write revision bundle to %s: %w", input.OutputPath, err)
		}
		return mcpserver.JSONResult(map[string]string{"message": "Revision bundle saved to " + input.OutputPath})
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "compareRevisions",
		Description: "Compare two revisions of an EdgeWorker to see dependency differences between them.",
	}, func(ctx context.Context, input compareRevisionsInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.CompareRevisions(ctx, input.EdgeWorkerID, input.RevisionID, input.CompareRevisionID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listRevisionActivations",
		Description: "List revision activations for an EdgeWorker (Flexible Composition feature).",
	}, func(ctx context.Context, input getEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListRevisionActivations(ctx, input.EdgeWorkerID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "activateRevision",
		Description: "Activate a fallback revision for an EdgeWorker (Flexible Composition feature). " +
			"Used to activate a specific dependency snapshot.",
	}, func(ctx context.Context, input activateRevisionInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ActivateRevision(ctx, input.EdgeWorkerID, edgeworkersapi.RevisionActivationRequest{RevisionID: input.RevisionID})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "pinRevision",
		Description: "Pin a revision to prevent it from being automatically deactivated during deployments " +
			"(Flexible Composition feature).",
	}, func(ctx context.Context, input pinRevisionInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.PinRevision(ctx, input.EdgeWorkerID, input.RevisionID, edgeworkersapi.RevisionPinRequest{PinNote: optionalString(input.PinNote)})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "unpinRevision",
		Description: "Unpin a revision, allowing it to be automatically managed again (Flexible Composition feature).",
	}, func(ctx context.Context, input unpinRevisionInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.UnpinRevision(ctx, input.EdgeWorkerID, input.RevisionID, edgeworkersapi.RevisionUnpinRequest{})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

package mcp

import (
	"context"
	"strconv"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listEdgeWorkersInput struct {
	GroupID        float64 `json:"groupId,omitempty" jsonschema:"Filter by group ID"`
	ResourceTierID float64 `json:"resourceTierId,omitempty" jsonschema:"Filter by resource tier ID"`
	mcpAccountInput
}

type getEdgeWorkerInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	mcpAccountInput
}

type createEdgeWorkerInput struct {
	GroupID        float64 `json:"groupId" jsonschema:"The group ID to associate the EdgeWorker with"`
	Name           string  `json:"name" jsonschema:"A human-readable name for the EdgeWorker"`
	ResourceTierID float64 `json:"resourceTierId" jsonschema:"The resource tier ID for the EdgeWorker"`
	Description    string  `json:"description,omitempty" jsonschema:"Optional description"`
	mcpAccountInput
}

type updateEdgeWorkerInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID to update"`
	GroupID      float64 `json:"groupId" jsonschema:"The group ID to associate the EdgeWorker with"`
	Name         string  `json:"name" jsonschema:"Updated human-readable name for the EdgeWorker"`
	Description  string  `json:"description,omitempty" jsonschema:"Optional updated description"`
	mcpAccountInput
}

type deleteEdgeWorkerInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID to delete"`
	Confirm      bool    `json:"confirm" jsonschema:"Must be true to confirm this destructive operation. Set to true to proceed."`
	mcpAccountInput
}

type cloneEdgeWorkerInput struct {
	EdgeWorkerID   float64 `json:"edgeWorkerId" jsonschema:"The source EdgeWorker ID to clone"`
	GroupID        float64 `json:"groupId" jsonschema:"The group ID for the cloned EdgeWorker"`
	Name           string  `json:"name" jsonschema:"Name for the cloned EdgeWorker"`
	ResourceTierID float64 `json:"resourceTierId" jsonschema:"The resource tier ID for the clone (can differ from source)"`
	Description    string  `json:"description,omitempty" jsonschema:"Optional description for the clone"`
	mcpAccountInput
}

func registerMCPIDTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listEdgeworkers",
		Description: "List all EdgeWorker IDs for the account. Optionally filter by groupId or resourceTierId.",
	}, func(ctx context.Context, input listEdgeWorkersInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListEdgeWorkers(ctx, optionalFloat64(input.GroupID), optionalFloat64(input.ResourceTierID))
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getEdgeworker",
		Description: "Get details for a specific EdgeWorker ID.",
	}, func(ctx context.Context, input getEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetEdgeWorker(ctx, input.EdgeWorkerID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "createEdgeworker",
		Description: "Create a new EdgeWorker ID.",
	}, func(ctx context.Context, input createEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.CreateEdgeWorker(ctx, edgeworkersapi.EdgeWorkerMutation{
			GroupID:        input.GroupID,
			Name:           input.Name,
			ResourceTierID: &input.ResourceTierID,
			Description:    optionalString(input.Description),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "updateEdgeworker",
		Description: "Update the name, group, or description of an existing EdgeWorker ID.",
	}, func(ctx context.Context, input updateEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.UpdateEdgeWorker(ctx, input.EdgeWorkerID, edgeworkersapi.EdgeWorkerMutation{
			GroupID:     input.GroupID,
			Name:        input.Name,
			Description: optionalString(input.Description),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "deleteEdgeworker",
		Description: "Permanently delete an EdgeWorker ID and all its associated data. Requires confirm: true.",
	}, func(ctx context.Context, input deleteEdgeWorkerInput) (*mcp.CallToolResult, error) {
		if !input.Confirm {
			return mcpserver.AbortResult("deleteEdgeworker"), nil
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		if err := client.DeleteEdgeWorker(ctx, input.EdgeWorkerID); err != nil {
			return nil, err
		}
		id := strconv.FormatFloat(input.EdgeWorkerID, 'f', -1, 64)
		return mcpserver.JSONResult(map[string]string{"message": "EdgeWorker " + id + " deleted successfully."})
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "cloneEdgeworker",
		Description: "Clone an existing EdgeWorker ID. Useful for moving an EdgeWorker to a different resource tier.",
	}, func(ctx context.Context, input cloneEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.CloneEdgeWorker(ctx, input.EdgeWorkerID, edgeworkersapi.EdgeWorkerMutation{
			GroupID:        input.GroupID,
			Name:           input.Name,
			ResourceTierID: &input.ResourceTierID,
			Description:    optionalString(input.Description),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getEdgeworkerResourceTier",
		Description: "Get the resource tier associated with a specific EdgeWorker ID.",
	}, func(ctx context.Context, input getEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetEdgeWorkerResourceTier(ctx, input.EdgeWorkerID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

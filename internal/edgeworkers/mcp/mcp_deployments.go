package mcp

import (
	"context"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listActivationsInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	Version      string  `json:"version,omitempty" jsonschema:"Filter activations by version"`
	Network      string  `json:"network,omitempty" jsonschema:"Filter by network: staging or production,enum=staging,enum=production"`
	mcpAccountInput
}

type getActivationInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	ActivationID float64 `json:"activationId" jsonschema:"The activation ID"`
	mcpAccountInput
}

type activateEdgeWorkerInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID to activate"`
	Network      string  `json:"network" jsonschema:"Network to activate on: staging or production,enum=staging,enum=production"`
	Version      string  `json:"version" jsonschema:"The version to activate"`
	Note         string  `json:"note,omitempty" jsonschema:"Optional note about this activation"`
	mcpAccountInput
}

type cancelActivationInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	ActivationID float64 `json:"activationId" jsonschema:"The activation ID to cancel"`
	Confirm      bool    `json:"confirm" jsonschema:"Must be true to confirm cancellation of this activation. Set to true to proceed."`
	mcpAccountInput
}

type rollbackEdgeWorkerInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID to roll back"`
	Network      string  `json:"network" jsonschema:"Network to roll back on: staging or production,enum=staging,enum=production"`
	Note         string  `json:"note,omitempty" jsonschema:"Optional note about this rollback"`
	mcpAccountInput
}

type listDeactivationsInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	Version      string  `json:"version,omitempty" jsonschema:"Filter deactivations by version"`
	Network      string  `json:"network,omitempty" jsonschema:"Filter by network: staging or production,enum=staging,enum=production"`
	mcpAccountInput
}

type getDeactivationInput struct {
	EdgeWorkerID   float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	DeactivationID float64 `json:"deactivationId" jsonschema:"The deactivation ID"`
	mcpAccountInput
}

type deactivateEdgeWorkerInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID to deactivate"`
	Network      string  `json:"network" jsonschema:"Network to deactivate on: staging or production,enum=staging,enum=production"`
	Version      string  `json:"version" jsonschema:"The version to deactivate"`
	Note         string  `json:"note,omitempty" jsonschema:"Optional note about this deactivation"`
	mcpAccountInput
}

type listPropertiesInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	ActiveOnly   bool    `json:"activeOnly" jsonschema:"If true, return only active property associations"`
	Details      bool    `json:"details" jsonschema:"If true, return full property details (hostname, rule tree path, etc.)"`
	mcpAccountInput
}

func registerMCPDeploymentTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listActivations",
		Description: "List activations for an EdgeWorker, optionally filtered by version or network.",
	}, func(ctx context.Context, input listActivationsInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListActivations(ctx, input.EdgeWorkerID, optionalString(input.Version), optionalString(input.Network))
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getActivation",
		Description: "Get details for a specific activation of an EdgeWorker.",
	}, func(ctx context.Context, input getActivationInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetActivation(ctx, input.EdgeWorkerID, input.ActivationID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "activateEdgeworker",
		Description: "Activate a specific version of an EdgeWorker on STAGING or PRODUCTION.",
	}, func(ctx context.Context, input activateEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ActivateEdgeWorker(ctx, input.EdgeWorkerID, edgeworkersapi.DeploymentRequest{
			Network: input.Network,
			Version: input.Version,
			Note:    optionalString(input.Note),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "cancelActivation",
		Description: "Cancel a pending activation for an EdgeWorker. Only works on activations in PENDING or " +
			"IN_PROGRESS status. Requires confirm: true.",
	}, func(ctx context.Context, input cancelActivationInput) (*mcp.CallToolResult, error) {
		if !input.Confirm {
			return mcpserver.AbortResult("cancelActivation"), nil
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.CancelActivation(ctx, input.EdgeWorkerID, input.ActivationID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "rollbackEdgeworker",
		Description: "Roll back an EdgeWorker to its previously active version on the specified network.",
	}, func(ctx context.Context, input rollbackEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.RollbackEdgeWorker(ctx, input.EdgeWorkerID, edgeworkersapi.DeploymentRequest{
			Network: input.Network,
			Note:    optionalString(input.Note),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listDeactivations",
		Description: "List deactivations for an EdgeWorker, optionally filtered by version or network.",
	}, func(ctx context.Context, input listDeactivationsInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListDeactivations(ctx, input.EdgeWorkerID, optionalString(input.Version), optionalString(input.Network))
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getDeactivation",
		Description: "Get details for a specific deactivation of an EdgeWorker.",
	}, func(ctx context.Context, input getDeactivationInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetDeactivation(ctx, input.EdgeWorkerID, input.DeactivationID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "deactivateEdgeworker",
		Description: "Deactivate a specific version of an EdgeWorker on STAGING or PRODUCTION. " +
			"This removes the EdgeWorker from the network but does not delete it.",
	}, func(ctx context.Context, input deactivateEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.DeactivateEdgeWorker(ctx, input.EdgeWorkerID, edgeworkersapi.DeploymentRequest{
			Network: input.Network,
			Version: input.Version,
			Note:    optionalString(input.Note),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listProperties",
		Description: "List the properties (Akamai delivery configurations) that reference a specific EdgeWorker.",
	}, func(ctx context.Context, input listPropertiesInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListProperties(ctx, input.EdgeWorkerID, input.ActiveOnly, input.Details)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

package mcp

import (
	"context"

	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getGroupInput struct {
	GroupID float64 `json:"groupId" jsonschema:"The group ID"`
	mcpAccountInput
}

type listResourceTiersInput struct {
	ContractID string `json:"contractId,omitempty" jsonschema:"Filter resource tiers by contract ID"`
	mcpAccountInput
}

func registerMCPAccountTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listGroups",
		Description: "List all permission groups available in the account that can be used with EdgeWorkers.",
	}, func(ctx context.Context, input mcpAccountInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListGroups(ctx)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getGroup",
		Description: "Get details for a specific permission group.",
	}, func(ctx context.Context, input getGroupInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetGroup(ctx, input.GroupID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listContracts",
		Description: "List all contract IDs associated with the account that support EdgeWorkers.",
	}, func(ctx context.Context, input mcpAccountInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListContracts(ctx)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "listResourceTiers",
		Description: "List available EdgeWorker resource tiers for the account. " +
			"Resource tiers determine the CPU time, memory, and execution limits available to an EdgeWorker.",
	}, func(ctx context.Context, input listResourceTiersInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListResourceTiers(ctx, optionalString(input.ContractID))
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listLimits",
		Description: "List the EdgeWorker resource limits for the account (max EdgeWorker IDs, versions, activations, etc.).",
	}, func(ctx context.Context, input mcpAccountInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListLimits(ctx)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

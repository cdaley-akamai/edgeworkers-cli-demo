package mcp

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getEdgeKVGroupInput struct {
	GroupID float64 `json:"groupId" jsonschema:"The permission group ID"`
	mcpAccountInput
}

func registerMCPGroupTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "listEdgeKVGroups",
		Description: "List all permission groups that have EdgeKV capabilities assigned, along with their " +
			"specific capabilities (read, write, delete data; manage namespaces and tokens).",
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
		Name:        "getEdgeKVGroup",
		Description: "Get details for a specific permission group, including its EdgeKV capabilities.",
	}, func(ctx context.Context, input getEdgeKVGroupInput) (*mcp.CallToolResult, error) {
		groupID, err := formatMCPGroupID(input.GroupID)
		if err != nil {
			return nil, err
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetGroup(ctx, groupID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

func formatMCPGroupID(groupID float64) (string, error) {
	if math.IsNaN(groupID) || math.IsInf(groupID, 0) {
		return "", fmt.Errorf("groupId must be a finite number")
	}
	return strconv.FormatFloat(groupID, 'f', -1, 64), nil
}

package mcp

import (
	"context"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getLoggingOverrideInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	LoggingID    float64 `json:"loggingId" jsonschema:"The logging override ID"`
	mcpAccountInput
}

type createLoggingOverrideInput struct {
	EdgeWorkerID float64 `json:"edgeWorkerId" jsonschema:"The EdgeWorker ID"`
	Level        string  `json:"level" jsonschema:"Log level override (default is ERROR without an override),enum=TRACE,enum=DEBUG,enum=INFO,enum=WARN,enum=ERROR"`
	Network      string  `json:"network" jsonschema:"Network to apply the logging override on: staging or production,enum=staging,enum=production"`
	Timeout      string  `json:"timeout" jsonschema:"ISO 8601 timestamp when the override expires (e.g. 2023-10-23T16:37:32Z)"`
	DS2ID        float64 `json:"ds2Id,omitempty" jsonschema:"DataStream 2 stream ID for log delivery"`
	mcpAccountInput
}

func registerMCPLoggingTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listLoggingOverrides",
		Description: "List all logging overrides for an EdgeWorker ID.",
	}, func(ctx context.Context, input getEdgeWorkerInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListLoggingOverrides(ctx, input.EdgeWorkerID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getLoggingOverride",
		Description: "Get status details for a specific logging override.",
	}, func(ctx context.Context, input getLoggingOverrideInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetLoggingOverride(ctx, input.EdgeWorkerID, input.LoggingID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "createLoggingOverride",
		Description: "Create a logging override for an EdgeWorker ID to change the JavaScript log level " +
			"temporarily for debugging. Overrides expire at the specified timeout.",
	}, func(ctx context.Context, input createLoggingOverrideInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		schema := "v1"
		result, err := client.CreateLoggingOverride(ctx, input.EdgeWorkerID, edgeworkersapi.LoggingOverrideRequest{
			Level:   input.Level,
			Network: input.Network,
			Timeout: &input.Timeout,
			Schema:  &schema,
			DS2ID:   optionalFloat64(input.DS2ID),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

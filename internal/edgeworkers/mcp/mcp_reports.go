package mcp

import (
	"context"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getReportInput struct {
	ReportID     float64 `json:"reportId" jsonschema:"Report ID. Use listReports to see available IDs. 1=overview, 3=execution statuses, 5=execution times with percentiles, 6=memory usage with percentiles, 7=sub-requests."`
	Start        string  `json:"start" jsonschema:"ISO 8601 start timestamp (e.g. 2022-10-18T14:10:50Z)"`
	End          string  `json:"end" jsonschema:"ISO 8601 end timestamp (e.g. 2022-10-19T14:10:50Z)"`
	EdgeWorker   string  `json:"edgeWorker" jsonschema:"Filter by EdgeWorker ID or ID-version (e.g. '42' or '42-1.0'). For multiple, call the tool multiple times or use comma-separation if supported."`
	Status       string  `json:"status" jsonschema:"Filter by execution status,enum=success,enum=genericError,enum=unknownEdgeWorkerId,enum=runtimeError,enum=executionError,enum=timeoutError,enum=resourceLimitHit,enum=cpuTimeoutError,enum=wallTimeoutError,enum=initCpuTimeoutError,enum=initWallTimeoutError,enum=subworkerNotEnabled,enum=subworkersLimitHit"`
	EventHandler string  `json:"eventHandler,omitempty" jsonschema:"Filter by event handler,enum=onClientRequest,enum=onOriginRequest,enum=onOriginResponse,enum=onClientResponse,enum=responseProvider,enum=onBotSegmentAvailable"`
	Network      string  `json:"network,omitempty" jsonschema:"Filter by network: staging or production,enum=staging,enum=production"`
	RevisionID   string  `json:"revisionId,omitempty" jsonschema:"Filter by revision ID (e.g. '3-1')"`
	mcpAccountInput
}

func registerMCPReportTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listReports",
		Description: "List available EdgeWorker reports. Note: reports 2 and 4 are deprecated.",
	}, func(ctx context.Context, input mcpAccountInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListReports(ctx)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getReport",
		Description: "Run an EdgeWorker report for a specific time range. Use listReports first to find valid report IDs.",
	}, func(ctx context.Context, input getReportInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetReport(ctx, edgeworkersapi.ReportRequest{
			ReportID:     input.ReportID,
			Start:        input.Start,
			End:          input.End,
			EdgeWorker:   input.EdgeWorker,
			Status:       input.Status,
			EventHandler: optionalString(input.EventHandler),
			Network:      optionalString(input.Network),
			RevisionID:   optionalString(input.RevisionID),
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

package mcp

import (
	"context"
	"fmt"
	"strings"

	edgekvapi "github.com/akamai/edgeworkers-cli/internal/edgekv/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpNetworkInput struct {
	Network string `json:"network" jsonschema:"Network to query: staging or production,enum=staging,enum=production"`
}

type mcpNamespaceDataAccessPolicyInput struct {
	RestrictDataAccess bool   `json:"restrictDataAccess" jsonschema:"Whether to restrict data access for this namespace"`
	PolicyType         string `json:"policyType,omitempty" jsonschema:"Policy type"`
}

type listEdgeKVNamespacesInput struct {
	mcpNetworkInput
	mcpAccountInput
}

type getEdgeKVNamespaceInput struct {
	mcpNetworkInput
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID"`
	mcpAccountInput
}

type createEdgeKVNamespaceInput struct {
	Network            string                             `json:"network" jsonschema:"Network to create the namespace on: staging or production,enum=staging,enum=production"`
	Namespace          string                             `json:"namespace" jsonschema:"The namespace ID (name)"`
	RetentionInSeconds float64                            `json:"retentionInSeconds" jsonschema:"Data retention period in seconds. Required: 0 (indefinite retention) or 86400-315360000."`
	GeoLocation        string                             `json:"geoLocation,omitempty" jsonschema:"Data storage location. Write-once — cannot be changed after creation.,enum=US,enum=EU,enum=JP,enum=GLOBAL"`
	GroupID            float64                            `json:"groupId" jsonschema:"Access control group ID for this namespace. Required: 0 makes the namespace available to all access groups with EdgeKV capabilities."`
	DataAccessPolicy   *mcpNamespaceDataAccessPolicyInput `json:"dataAccessPolicy,omitempty" jsonschema:"Data access policy for this namespace"`
	mcpAccountInput
}

type updateEdgeKVNamespaceInput struct {
	Network            string  `json:"network" jsonschema:"Network to update the namespace on: staging or production,enum=staging,enum=production"`
	NamespaceID        string  `json:"namespaceId" jsonschema:"The namespace ID to update"`
	RetentionInSeconds float64 `json:"retentionInSeconds" jsonschema:"Data retention period in seconds. 0 means indefinite retention."`
	mcpAccountInput
}

type deleteEdgeKVNamespaceInput struct {
	Network     string `json:"network" jsonschema:"Network to delete the namespace from: staging or production,enum=staging,enum=production"`
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID to delete"`
	Confirm     bool   `json:"confirm" jsonschema:"Must be true to confirm this destructive operation. Schedules deletion of the namespace and all its data. Set to true to proceed."`
	mcpAccountInput
}

type listEdgeKVNamespaceGroupsInput struct {
	mcpNetworkInput
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID"`
	mcpAccountInput
}

type reauthorizeEdgeKVNamespaceInput struct {
	NamespaceID string  `json:"namespaceId" jsonschema:"The namespace ID to reauthorize"`
	GroupID     float64 `json:"groupId" jsonschema:"The new access control group ID for this namespace"`
	mcpAccountInput
}

type updateEdgeKVDataAccessPolicyInput struct {
	RestrictDataAccess           bool `json:"restrictDataAccess" jsonschema:"Whether to restrict data access globally across namespaces"`
	AllowNamespacePolicyOverride bool `json:"allowNamespacePolicyOverride" jsonschema:"Whether individual namespaces can override the default data access policy"`
	mcpAccountInput
}

type getEdgeKVScheduledDeleteInput struct {
	mcpNetworkInput
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID"`
	mcpAccountInput
}

type rescheduleEdgeKVNamespaceDeleteInput struct {
	Network             string `json:"network" jsonschema:"Network: staging or production,enum=staging,enum=production"`
	NamespaceID         string `json:"namespaceId" jsonschema:"The namespace ID"`
	ScheduledDeleteTime string `json:"scheduledDeleteTime" jsonschema:"ISO 8601 timestamp for when to delete the namespace (e.g. 2024-10-23T16:37:32Z)"`
	mcpAccountInput
}

type cancelEdgeKVNamespaceDeleteInput struct {
	Network     string `json:"network" jsonschema:"Network: staging or production,enum=staging,enum=production"`
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID"`
	Confirm     bool   `json:"confirm" jsonschema:"Must be true to confirm cancelling the scheduled deletion. Set to true to proceed."`
	mcpAccountInput
}

func registerMCPNamespaceTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listEdgeKVNamespaces",
		Description: "List all EdgeKV namespaces on the specified network.",
	}, func(ctx context.Context, input listEdgeKVNamespacesInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		namespaces, err := client.ListNamespaces(ctx, network, false)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(map[string]any{"namespaces": namespaces})
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getEdgeKVNamespace",
		Description: "Get details for a specific EdgeKV namespace.",
	}, func(ctx context.Context, input getEdgeKVNamespaceInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetNamespace(ctx, network, input.NamespaceID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "createEdgeKVNamespace",
		Description: "Create a new EdgeKV namespace on the specified network.",
	}, func(ctx context.Context, input createEdgeKVNamespaceInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.Namespace == "" {
			return nil, fmt.Errorf("namespace is required")
		}
		if err := requireGeoLocation(input.GeoLocation); err != nil {
			return nil, err
		}
		if err := requireRetention(input.RetentionInSeconds); err != nil {
			return nil, err
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		body := map[string]any{
			"namespace":          input.Namespace,
			"retentionInSeconds": input.RetentionInSeconds,
			"groupId":            input.GroupID,
		}
		if input.GeoLocation != "" {
			body["geoLocation"] = input.GeoLocation
		}
		if input.DataAccessPolicy != nil {
			policy := map[string]any{"restrictDataAccess": input.DataAccessPolicy.RestrictDataAccess}
			if input.DataAccessPolicy.PolicyType != "" {
				policy["policyType"] = input.DataAccessPolicy.PolicyType
			}
			body["dataAccessPolicy"] = policy
		}
		result, err := client.CreateNamespace(ctx, network, body)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "updateEdgeKVNamespace",
		Description: "Update an EdgeKV namespace's retention period. Use updateEdgeKVDataAccessPolicy to change the default data access policy for future namespaces — an existing namespace's data access policy can't be changed after creation.",
	}, func(ctx context.Context, input updateEdgeKVNamespaceInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if err := requireRetention(input.RetentionInSeconds); err != nil {
			return nil, err
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		body := map[string]any{"retentionInSeconds": input.RetentionInSeconds}
		result, err := client.UpdateNamespace(ctx, network, input.NamespaceID, body)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "deleteEdgeKVNamespace",
		Description: "Schedule deletion of an EdgeKV namespace and all its data. The deletion is asynchronous " +
			"(returns 202). Requires confirm: true.",
	}, func(ctx context.Context, input deleteEdgeKVNamespaceInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if !input.Confirm {
			return mcpserver.AbortResult("deleteEdgeKVNamespace"), nil
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.DeleteNamespace(ctx, network, input.NamespaceID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listEdgeKVNamespaceGroups",
		Description: "List all groups (item key prefixes) within an EdgeKV namespace.",
	}, func(ctx context.Context, input listEdgeKVNamespaceGroupsInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListNamespaceGroups(ctx, network, input.NamespaceID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "reauthorizeEdgeKVNamespace",
		Description: "Reauthorize an EdgeKV namespace by moving it to a different access group.",
	}, func(ctx context.Context, input reauthorizeEdgeKVNamespaceInput) (*mcp.CallToolResult, error) {
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ReauthorizeNamespace(ctx, input.NamespaceID, input.GroupID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "updateEdgeKVDataAccessPolicy",
		Description: "Modify the default data access policy for EdgeKV. Controls whether data access is " +
			"restricted globally and whether per-namespace overrides are permitted.",
	}, func(ctx context.Context, input updateEdgeKVDataAccessPolicyInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.UpdateDatabasePolicy(ctx, edgekvapi.DataAccessPolicy{
			RestrictDataAccess:           input.RestrictDataAccess,
			AllowNamespacePolicyOverride: input.AllowNamespacePolicyOverride,
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getEdgeKVScheduledDelete",
		Description: "Get the scheduled deletion time for an EdgeKV namespace.",
	}, func(ctx context.Context, input getEdgeKVScheduledDeleteInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetScheduledDelete(ctx, network, input.NamespaceID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "rescheduleEdgeKVNamespaceDelete",
		Description: "Reschedule the deletion time for a namespace that is already pending deletion.",
	}, func(ctx context.Context, input rescheduleEdgeKVNamespaceDeleteInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if input.ScheduledDeleteTime == "" {
			return nil, fmt.Errorf("scheduledDeleteTime is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.RescheduleNamespaceDelete(ctx, network, input.NamespaceID, input.ScheduledDeleteTime)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "cancelEdgeKVNamespaceDelete",
		Description: "Cancel a scheduled deletion for an EdgeKV namespace, restoring it to active status. " +
			"Requires confirm: true.",
	}, func(ctx context.Context, input cancelEdgeKVNamespaceDeleteInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if !input.Confirm {
			return mcpserver.AbortResult("cancelEdgeKVNamespaceDelete"), nil
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.CancelNamespaceDelete(ctx, network, input.NamespaceID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

func requireNetwork(value string) (string, error) {
	network := strings.ToLower(strings.TrimSpace(value))
	if network != "staging" && network != "production" {
		return "", fmt.Errorf("network must be staging or production")
	}
	return network, nil
}

func requireGeoLocation(value string) error {
	if value == "" {
		return nil
	}
	switch value {
	case "US", "EU", "JP", "GLOBAL":
		return nil
	default:
		return fmt.Errorf("geoLocation must be one of US, EU, JP, GLOBAL")
	}
}

func requireRetention(value float64) error {
	if value == 0 {
		return nil
	}
	if value < 86400 || value > 315360000 {
		return fmt.Errorf("retentionInSeconds must be 0 (indefinite) or between 86400 and 315360000")
	}
	return nil
}

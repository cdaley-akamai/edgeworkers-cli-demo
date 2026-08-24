package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listEdgeKVItemsInput struct {
	Network     string  `json:"network" jsonschema:"Network to query: staging or production,enum=staging,enum=production"`
	NamespaceID string  `json:"namespaceId" jsonschema:"The namespace ID"`
	GroupID     string  `json:"groupId" jsonschema:"The group ID (item key prefix)"`
	MaxItems    float64 `json:"maxItems,omitempty" jsonschema:"Maximum number of item IDs to return (API limit: 100)"`
	SandboxID   string  `json:"sandboxId,omitempty" jsonschema:"Sandbox ID for sandbox-scoped reads"`
	mcpAccountInput
}

type getEdgeKVItemInput struct {
	Network     string `json:"network" jsonschema:"Network to query: staging or production,enum=staging,enum=production"`
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID"`
	GroupID     string `json:"groupId" jsonschema:"The group ID"`
	ItemID      string `json:"itemId" jsonschema:"The item ID (key)"`
	SandboxID   string `json:"sandboxId,omitempty" jsonschema:"Sandbox ID for sandbox-scoped reads"`
	mcpAccountInput
}

type writeEdgeKVItemInput struct {
	Network     string `json:"network" jsonschema:"Network to write to: staging or production,enum=staging,enum=production"`
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID"`
	GroupID     string `json:"groupId" jsonschema:"The group ID"`
	ItemID      string `json:"itemId" jsonschema:"The item ID (key)"`
	Value       any    `json:"value" jsonschema:"Item value. Strings are sent as text/plain; objects are serialized and sent as application/json."`
	SandboxID   string `json:"sandboxId,omitempty" jsonschema:"Sandbox ID for sandbox-scoped writes"`
	mcpAccountInput
}

type deleteEdgeKVItemInput struct {
	Network     string `json:"network" jsonschema:"Network to delete from: staging or production,enum=staging,enum=production"`
	NamespaceID string `json:"namespaceId" jsonschema:"The namespace ID"`
	GroupID     string `json:"groupId" jsonschema:"The group ID"`
	ItemID      string `json:"itemId" jsonschema:"The item ID (key) to delete"`
	SandboxID   string `json:"sandboxId,omitempty" jsonschema:"Sandbox ID for sandbox-scoped deletes"`
	Confirm     bool   `json:"confirm" jsonschema:"Must be true to confirm this destructive operation. Set to true to proceed."`
	mcpAccountInput
}

func registerMCPItemTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "listEdgeKVItems",
		Description: "List item IDs within an EdgeKV group. Returns up to 100 item IDs. " +
			"Note: reads are eventually consistent — changes may take up to 10 seconds to appear.",
	}, func(ctx context.Context, input listEdgeKVItemsInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if input.GroupID == "" {
			return nil, fmt.Errorf("groupId is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListItems(ctx, network, input.NamespaceID, input.GroupID, optionalFloat64(input.MaxItems), input.SandboxID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "getEdgeKVItem",
		Description: "Read the value of a single EdgeKV item. Returns the raw value (text or JSON). " +
			"Note: reads are eventually consistent — changes may take up to 10 seconds to appear.",
	}, func(ctx context.Context, input getEdgeKVItemInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if input.GroupID == "" {
			return nil, fmt.Errorf("groupId is required")
		}
		if input.ItemID == "" {
			return nil, fmt.Errorf("itemId is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetItem(ctx, network, input.NamespaceID, input.GroupID, input.ItemID, input.SandboxID, "application/json")
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(decodeRawResult(result))
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "writeEdgeKVItem",
		Description: "Write or upsert an EdgeKV item. Accepts a string value (text/plain) or an object " +
			"(application/json). Creates the item if it does not exist, updates it if it does.",
	}, func(ctx context.Context, input writeEdgeKVItemInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if input.GroupID == "" {
			return nil, fmt.Errorf("groupId is required")
		}
		if input.ItemID == "" {
			return nil, fmt.Errorf("itemId is required")
		}
		if !isMCPItemValue(input.Value) {
			return nil, fmt.Errorf("value must be a string or an object")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.WriteItem(ctx, network, input.NamespaceID, input.GroupID, input.ItemID, input.Value, "", input.SandboxID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "deleteEdgeKVItem",
		Description: "Mark an EdgeKV item for deletion. The deletion is eventually consistent. Requires confirm: true.",
	}, func(ctx context.Context, input deleteEdgeKVItemInput) (*mcp.CallToolResult, error) {
		network, err := requireNetwork(input.Network)
		if err != nil {
			return nil, err
		}
		if input.NamespaceID == "" {
			return nil, fmt.Errorf("namespaceId is required")
		}
		if input.GroupID == "" {
			return nil, fmt.Errorf("groupId is required")
		}
		if input.ItemID == "" {
			return nil, fmt.Errorf("itemId is required")
		}
		if !input.Confirm {
			return mcpserver.AbortResult("deleteEdgeKVItem"), nil
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.DeleteItem(ctx, network, input.NamespaceID, input.GroupID, input.ItemID, input.SandboxID)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

func isMCPItemValue(value any) bool {
	if _, ok := value.(string); ok {
		return true
	}
	_, ok := value.(map[string]any)
	return ok
}

func decodeRawResult(data []byte) any {
	if len(data) == 0 {
		return map[string]any{}
	}
	var result any
	if err := json.Unmarshal(data, &result); err == nil {
		return result
	}
	return string(data)
}

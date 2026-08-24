package mcp

import (
	"context"
	"fmt"

	edgekvapi "github.com/akamai/edgeworkers-cli/internal/edgekv/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listEdgeKVTokensInput struct {
	IncludeExpired bool `json:"includeExpired,omitempty" jsonschema:"Set to true to include expired tokens in the results"`
	mcpAccountInput
}

type getEdgeKVTokenInput struct {
	TokenName string `json:"tokenName" jsonschema:"The token name"`
	mcpAccountInput
}

type createEdgeKVTokenInput struct {
	Name                    string              `json:"name" jsonschema:"Token name (1–32 characters). Used to identify and retrieve the token."`
	AllowOnProduction       bool                `json:"allowOnProduction" jsonschema:"Allow this token to be used on production network"`
	AllowOnStaging          bool                `json:"allowOnStaging" jsonschema:"Allow this token to be used on staging network"`
	Expiry                  string              `json:"expiry" jsonschema:"ISO 8601 expiry date for the token (e.g. 2025-12-31T00:00:00Z)"`
	NamespacePermissions    map[string][]string `json:"namespacePermissions" jsonschema:"Namespace permission map. Keys are namespace names; values are arrays of permissions: 'r' (read), 'w' (write), 'd' (delete)."`
	RestrictToEdgeWorkerIDs []string            `json:"restrictToEdgeWorkerIds,omitempty" jsonschema:"Restrict token usage to specific EdgeWorker IDs"`
	mcpAccountInput
}

type deleteEdgeKVTokenInput struct {
	TokenName string `json:"tokenName" jsonschema:"The token name to revoke"`
	Confirm   bool   `json:"confirm" jsonschema:"Must be true to confirm this irreversible operation. Revoking a token cannot be undone. Set to true to proceed."`
	mcpAccountInput
}

type refreshEdgeKVTokenInput struct {
	TokenName string `json:"tokenName" jsonschema:"The token name to refresh"`
	mcpAccountInput
}

func registerMCPTokenTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "listEdgeKVTokens",
		Description: "List EdgeKV access tokens for the account.",
	}, func(ctx context.Context, input listEdgeKVTokensInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.ListTokens(ctx, optionalBool(input.IncludeExpired))
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getEdgeKVToken",
		Description: "Download a specific EdgeKV access token by name.",
	}, func(ctx context.Context, input getEdgeKVTokenInput) (*mcp.CallToolResult, error) {
		if input.TokenName == "" {
			return nil, fmt.Errorf("tokenName is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetToken(ctx, input.TokenName)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "createEdgeKVToken",
		Description: "Create a new EdgeKV access token. Tokens grant EdgeWorkers scoped access to namespaces. " +
			"Token activation is asynchronous — check tokenActivationStatus in the response.",
	}, func(ctx context.Context, input createEdgeKVTokenInput) (*mcp.CallToolResult, error) {
		if input.Name == "" {
			return nil, fmt.Errorf("name is required")
		}
		if len(input.Name) > 32 {
			return nil, fmt.Errorf("name must be 32 characters or fewer")
		}
		if input.Expiry == "" {
			return nil, fmt.Errorf("expiry is required")
		}
		if input.NamespacePermissions == nil {
			return nil, fmt.Errorf("namespacePermissions is required")
		}
		if err := validateNamespacePermissions(input.NamespacePermissions); err != nil {
			return nil, err
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.CreateToken(ctx, edgekvapi.CreateTokenRequest{
			Name:                    input.Name,
			AllowOnProduction:       input.AllowOnProduction,
			AllowOnStaging:          input.AllowOnStaging,
			Expiry:                  input.Expiry,
			NamespacePermissions:    input.NamespacePermissions,
			RestrictToEdgeWorkerIDs: input.RestrictToEdgeWorkerIDs,
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "deleteEdgeKVToken",
		Description: "Permanently revoke an EdgeKV access token. This operation is irreversible. " +
			"Requires confirm: true.",
	}, func(ctx context.Context, input deleteEdgeKVTokenInput) (*mcp.CallToolResult, error) {
		if input.TokenName == "" {
			return nil, fmt.Errorf("tokenName is required")
		}
		if !input.Confirm {
			return mcpserver.AbortResult("deleteEdgeKVToken"), nil
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.DeleteToken(ctx, input.TokenName)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "refreshEdgeKVToken",
		Description: "Manually trigger an early refresh of an EdgeKV access token before its scheduled " +
			"refresh date.",
	}, func(ctx context.Context, input refreshEdgeKVTokenInput) (*mcp.CallToolResult, error) {
		if input.TokenName == "" {
			return nil, fmt.Errorf("tokenName is required")
		}
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.RefreshToken(ctx, input.TokenName)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

func validateNamespacePermissions(permissions map[string][]string) error {
	for namespace, values := range permissions {
		if namespace == "" {
			return fmt.Errorf("namespacePermissions keys must be non-empty")
		}
		for _, permission := range values {
			switch permission {
			case "r", "w", "d":
			default:
				return fmt.Errorf("invalid permission %q for namespace %q; use r, w, or d", permission, namespace)
			}
		}
	}
	return nil
}

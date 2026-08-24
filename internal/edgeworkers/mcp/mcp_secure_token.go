package mcp

import (
	"context"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createSecureTokenInput struct {
	PropertyID string   `json:"propertyId" jsonschema:"The property ID to generate the secure token for"`
	Hostname   string   `json:"hostname" jsonschema:"The primary hostname for the token"`
	Expiry     float64  `json:"expiry" jsonschema:"Token expiry time in minutes from now"`
	Hostnames  []string `json:"hostnames,omitempty" jsonschema:"Additional hostnames to include in the token"`
	mcpAccountInput
}

func registerMCPSecureTokenTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name: "createSecureToken",
		Description: "Create a JWT authentication token for securing EdgeWorker JavaScript logging. " +
			"Used with DataStream 2 log delivery.",
	}, func(ctx context.Context, input createSecureTokenInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		var hostnames *[]string
		if len(input.Hostnames) > 0 {
			hostnames = &input.Hostnames
		}
		result, err := client.CreateSecureToken(ctx, edgeworkersapi.SecureTokenRequest{
			PropertyID: &input.PropertyID,
			Hostname:   &input.Hostname,
			Expiry:     &input.Expiry,
			Hostnames:  hostnames,
		})
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

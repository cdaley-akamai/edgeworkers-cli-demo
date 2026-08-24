// Package mcp implements the Akamai EdgeKV MCP server.
package mcp

import (
	"context"
	"io"
	"net/http"
	"time"

	edgekvapi "github.com/akamai/edgeworkers-cli/internal/edgekv/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config configures an EdgeKV MCP server and its per-call API clients.
type Config struct {
	// Version is the enclosing EdgeKV CLI release version.
	Version string
	// Tools contains glob patterns selecting enabled tools.
	Tools []string
	// LogOutput receives MCP diagnostics.
	LogOutput io.Writer
	// EdgeRc is the credentials file path.
	EdgeRc string
	// Section is the credentials file section.
	Section string
	// AccountSwitchKey is the default account context.
	AccountSwitchKey string
	// Timeout bounds each API client's HTTP requests.
	Timeout time.Duration
	// Debug enables HTTP tracing.
	Debug bool
	// HTTPClient optionally supplies a custom transport while retaining fresh API clients.
	HTTPClient *http.Client
}

type mcpAccountInput struct {
	AccountSwitchKey string `json:"accountSwitchKey,omitempty" jsonschema:"Account switch key for this request. Overrides the configured default."`
}

// optionalFloat64 converts an omitempty MCP input field back to the pointer form API clients expect for "unset".
func optionalFloat64(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}

// optionalBool converts an omitempty MCP input field back to the pointer form API clients expect for "unset".
func optionalBool(value bool) *bool {
	if !value {
		return nil
	}
	return &value
}

// NewServer creates an EdgeKV MCP server.
func NewServer(config Config) (*mcpserver.Server, error) {
	runtime, err := mcpserver.New(mcpserver.Config{
		Name:         "akamai-edgekv-mcp",
		Version:      config.Version,
		ProductName:  "Akamai EdgeKV MCP",
		ToolPatterns: config.Tools,
		LogOutput:    config.LogOutput,
		Debug:        config.Debug,
	})
	if err != nil {
		return nil, err
	}
	registerMCPStatusTools(runtime, config)
	registerMCPNamespaceTools(runtime, config)
	registerMCPItemTools(runtime, config)
	registerMCPTokenTools(runtime, config)
	registerMCPGroupTools(runtime, config)
	return runtime, nil
}

func newEdgeKVMCPServer(version string, patterns []string, logOutput io.Writer, debug bool, config Config) (*mcpserver.Server, error) {
	config.Version = version
	config.Tools = patterns
	config.LogOutput = logOutput
	config.Debug = debug
	return NewServer(config)
}

func newClient(config Config, accountSwitchKey string) (*edgekvapi.Client, error) {
	if accountSwitchKey == "" {
		accountSwitchKey = config.AccountSwitchKey
	}
	return edgekvapi.NewClientWithHTTPClient(config.EdgeRc, config.Section, accountSwitchKey, config.Timeout, config.Debug, config.HTTPClient)
}

func registerMCPStatusTools(runtime *mcpserver.Server, config Config) {
	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "getEdgeKVInitializeStatus",
		Description: "Get the EdgeKV initialization status for the account. Shows whether EdgeKV has been initialized, along with CP code, production/staging status, and data access policy.",
	}, func(ctx context.Context, input mcpAccountInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.GetDatabaseStatus(ctx)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})

	mcpserver.AddTool(runtime, &mcp.Tool{
		Name:        "initializeEdgeKV",
		Description: "Initialize EdgeKV for the account. This is a one-time operation that creates the default namespace and CP code. Safe to call if already initialized — returns current status.",
	}, func(ctx context.Context, input mcpAccountInput) (*mcp.CallToolResult, error) {
		client, err := newClient(config, input.AccountSwitchKey)
		if err != nil {
			return nil, err
		}
		result, err := client.InitializeDatabase(ctx, nil)
		if err != nil {
			return nil, err
		}
		return mcpserver.JSONResult(result)
	})
}

// Package mcp implements the Akamai EdgeWorkers MCP server.
package mcp

import (
	"io"
	"net/http"
	"time"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
)

// Config configures an EdgeWorkers MCP server and its per-call API clients.
type Config struct {
	// Version is the enclosing EdgeWorkers CLI release version.
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

// NewServer creates an EdgeWorkers MCP server.
func NewServer(config Config) (*mcpserver.Server, error) {
	runtime, err := mcpserver.New(mcpserver.Config{
		Name:         "akamai-edgeworkers-mcp",
		Version:      config.Version,
		ProductName:  "Akamai EdgeWorkers MCP",
		ToolPatterns: config.Tools,
		LogOutput:    config.LogOutput,
		Debug:        config.Debug,
	})
	if err != nil {
		return nil, err
	}
	registerMCPIDTools(runtime, config)
	registerMCPVersionTools(runtime, config)
	registerMCPDeploymentTools(runtime, config)
	registerMCPAccountTools(runtime, config)
	registerMCPReportTools(runtime, config)
	registerMCPSecureTokenTools(runtime, config)
	registerMCPLoggingTools(runtime, config)
	registerMCPValidationTools(runtime, config)
	registerMCPRevisionTools(runtime, config)
	return runtime, nil
}

func newEdgeWorkersMCPServer(version string, patterns []string, logOutput io.Writer, debug bool, config Config) (*mcpserver.Server, error) {
	config.Version = version
	config.Tools = patterns
	config.LogOutput = logOutput
	config.Debug = debug
	return NewServer(config)
}

func newClient(config Config, accountSwitchKey string) (*edgeworkersapi.Client, error) {
	if accountSwitchKey == "" {
		accountSwitchKey = config.AccountSwitchKey
	}
	return edgeworkersapi.NewClientWithHTTPClient(config.EdgeRc, config.Section, accountSwitchKey, config.Timeout, config.Debug, config.HTTPClient)
}

// optionalString converts an omitempty MCP input field back to the pointer form API clients expect for "unset".
func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// optionalFloat64 converts an omitempty MCP input field back to the pointer form API clients expect for "unset".
func optionalFloat64(value float64) *float64 {
	if value == 0 {
		return nil
	}
	return &value
}

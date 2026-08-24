package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akamai/edgeworkers-cli/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeKVMCPStatusToolUsesSharedServiceAndAccountOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method)
		require.Equal(t, "/edgekv/v1/initialize", request.URL.Path)
		require.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"accountStatus":"INITIALIZED","cpcode":"1234"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"getEdgeKVInitializeStatus"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "getEdgeKVInitializeStatus",
		Arguments: map[string]any{"accountSwitchKey": "A-CCT123"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Equal(t, "{\n  \"accountStatus\": \"INITIALIZED\",\n  \"cpcode\": \"1234\"\n}", result.Content[0].(*mcp.TextContent).Text)

	tools, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 2)
	require.ElementsMatch(t, []string{"hello_world", "getEdgeKVInitializeStatus"}, []string{tools.Tools[0].Name, tools.Tools[1].Name})
	require.Equal(t, "akamai-edgekv-mcp", clientSession.InitializeResult().ServerInfo.Name)
	require.Equal(t, "9.8.7", clientSession.InitializeResult().ServerInfo.Version)
}

func TestEdgeKVMCPServerRegistersExactCatalog(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"*"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	require.ElementsMatch(t, []string{
		"hello_world",
		"getEdgeKVInitializeStatus",
		"initializeEdgeKV",
		"listEdgeKVNamespaces",
		"getEdgeKVNamespace",
		"createEdgeKVNamespace",
		"updateEdgeKVNamespace",
		"deleteEdgeKVNamespace",
		"listEdgeKVNamespaceGroups",
		"reauthorizeEdgeKVNamespace",
		"updateEdgeKVDataAccessPolicy",
		"getEdgeKVScheduledDelete",
		"rescheduleEdgeKVNamespaceDelete",
		"cancelEdgeKVNamespaceDelete",
		"listEdgeKVItems",
		"getEdgeKVItem",
		"writeEdgeKVItem",
		"deleteEdgeKVItem",
		"listEdgeKVTokens",
		"getEdgeKVToken",
		"createEdgeKVToken",
		"deleteEdgeKVToken",
		"refreshEdgeKVToken",
		"listEdgeKVGroups",
		"getEdgeKVGroup",
	}, names)
	for name, description := range edgeKVNodeDescriptions() {
		assert.Equal(t, description, mustFindToolByName(t, result.Tools, name).Description, name)
	}
}

func edgeKVNodeDescriptions() map[string]string {
	return map[string]string{
		"getEdgeKVInitializeStatus":       "Get the EdgeKV initialization status for the account. Shows whether EdgeKV has been initialized, along with CP code, production/staging status, and data access policy.",
		"initializeEdgeKV":                "Initialize EdgeKV for the account. This is a one-time operation that creates the default namespace and CP code. Safe to call if already initialized — returns current status.",
		"listEdgeKVNamespaces":            "List all EdgeKV namespaces on the specified network.",
		"getEdgeKVNamespace":              "Get details for a specific EdgeKV namespace.",
		"createEdgeKVNamespace":           "Create a new EdgeKV namespace on the specified network.",
		"updateEdgeKVNamespace":           "Update an EdgeKV namespace's retention period. Use updateEdgeKVDataAccessPolicy to change the default data access policy for future namespaces — an existing namespace's data access policy can't be changed after creation.",
		"deleteEdgeKVNamespace":           "Schedule deletion of an EdgeKV namespace and all its data. The deletion is asynchronous (returns 202). Requires confirm: true.",
		"listEdgeKVNamespaceGroups":       "List all groups (item key prefixes) within an EdgeKV namespace.",
		"reauthorizeEdgeKVNamespace":      "Reauthorize an EdgeKV namespace by moving it to a different access group.",
		"updateEdgeKVDataAccessPolicy":    "Modify the default data access policy for EdgeKV. Controls whether data access is restricted globally and whether per-namespace overrides are permitted.",
		"getEdgeKVScheduledDelete":        "Get the scheduled deletion time for an EdgeKV namespace.",
		"rescheduleEdgeKVNamespaceDelete": "Reschedule the deletion time for a namespace that is already pending deletion.",
		"cancelEdgeKVNamespaceDelete":     "Cancel a scheduled deletion for an EdgeKV namespace, restoring it to active status. Requires confirm: true.",
		"listEdgeKVItems":                 "List item IDs within an EdgeKV group. Returns up to 100 item IDs. Note: reads are eventually consistent — changes may take up to 10 seconds to appear.",
		"getEdgeKVItem":                   "Read the value of a single EdgeKV item. Returns the raw value (text or JSON). Note: reads are eventually consistent — changes may take up to 10 seconds to appear.",
		"writeEdgeKVItem":                 "Write or upsert an EdgeKV item. Accepts a string value (text/plain) or an object (application/json). Creates the item if it does not exist, updates it if it does.",
		"deleteEdgeKVItem":                "Mark an EdgeKV item for deletion. The deletion is eventually consistent. Requires confirm: true.",
		"listEdgeKVTokens":                "List EdgeKV access tokens for the account.",
		"getEdgeKVToken":                  "Download a specific EdgeKV access token by name.",
		"createEdgeKVToken":               "Create a new EdgeKV access token. Tokens grant EdgeWorkers scoped access to namespaces. Token activation is asynchronous — check tokenActivationStatus in the response.",
		"deleteEdgeKVToken":               "Permanently revoke an EdgeKV access token. This operation is irreversible. Requires confirm: true.",
		"refreshEdgeKVToken":              "Manually trigger an early refresh of an EdgeKV access token before its scheduled refresh date.",
		"listEdgeKVGroups":                "List all permission groups that have EdgeKV capabilities assigned, along with their specific capabilities (read, write, delete data; manage namespaces and tokens).",
		"getEdgeKVGroup":                  "Get details for a specific permission group, including its EdgeKV capabilities.",
	}
}

func connectEdgeKVMCPClient(t *testing.T, runtime *mcpserver.Server) *mcp.ClientSession {
	t.Helper()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := runtime.Connect(context.Background(), serverTransport)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, &mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{}})
	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })
	return clientSession
}

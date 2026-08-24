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

func TestEdgeWorkersMCPGetEdgeworkerUsesSharedServiceAndAccountOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"edgeWorkerId":123,"description":"kept"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"getEdgeworker"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "getEdgeworker",
		Arguments: map[string]any{"edgeWorkerId": 123, "accountSwitchKey": "A-CCT123"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"description\": \"kept\",\n  \"edgeWorkerId\": 123\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeWorkersMCPDeleteEdgeworkerAbortSkipsClientCreation(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"deleteEdgeworker"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "deleteEdgeworker",
		Arguments: map[string]any{"edgeWorkerId": 123, "confirm": false},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Aborted: deleteEdgeworker requires confirm to be true. Set confirm: true to proceed.", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeWorkersMCPRegistersIDCatalog(t *testing.T) {
	config := Config{}
	patterns := []string{
		"listEdgeworkers", "getEdgeworker", "createEdgeworker", "updateEdgeworker",
		"deleteEdgeworker", "cloneEdgeworker", "getEdgeworkerResourceTier",
	}
	runtime, err := newEdgeWorkersMCPServer("9.8.7", patterns, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	assert.ElementsMatch(t, append([]string{"hello_world"}, patterns...), names)
	assert.Equal(t, "akamai-edgeworkers-mcp", clientSession.InitializeResult().ServerInfo.Name)
	assert.Equal(t, "9.8.7", clientSession.InitializeResult().ServerInfo.Version)
}

func connectEdgeWorkersMCPClient(t *testing.T, runtime *mcpserver.Server) *mcp.ClientSession {
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

func mustFindEdgeWorkersTool(t *testing.T, tools []*mcp.Tool, name string) *mcp.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool not found: %s", name)
	return nil
}

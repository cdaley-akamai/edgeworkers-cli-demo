package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeKVMCPListGroupsUsesExpectedPathAndAccountOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/auth/groups", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `[{"groupId":123,"groupName":"team-a"}]`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"listEdgeKVGroups"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "listEdgeKVGroups",
		Arguments: map[string]any{"accountSwitchKey": "A-CCT123"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "[\n  {\n    \"groupId\": 123,\n    \"groupName\": \"team-a\"\n  }\n]", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPGetGroupUsesExpectedPathAndAccountOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/auth/groups/456", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"groupId":456,"groupName":"team-b"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"getEdgeKVGroup"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "getEdgeKVGroup",
		Arguments: map[string]any{
			"groupId":          456,
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"groupId\": 456,\n  \"groupName\": \"team-b\"\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPGetGroupPreservesNumericGroupID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgekv/v1/auth/groups/12.5", request.URL.Path)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"groupId":12.5}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"getEdgeKVGroup"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "getEdgeKVGroup",
		Arguments: map[string]any{"groupId": 12.5},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeKVMCPGroupToolsExposeNodeParityMetadata(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"listEdgeKVGroups", "getEdgeKVGroup"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	tools, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)

	listTool := mustFindToolByName(t, tools.Tools, "listEdgeKVGroups")
	assert.Equal(t,
		"List all permission groups that have EdgeKV capabilities assigned, along with their specific capabilities (read, write, delete data; manage namespaces and tokens).",
		listTool.Description,
	)

	getTool := mustFindToolByName(t, tools.Tools, "getEdgeKVGroup")
	assert.Equal(t, "Get details for a specific permission group, including its EdgeKV capabilities.", getTool.Description)
}

func mustFindToolByName(t *testing.T, tools []*mcp.Tool, name string) *mcp.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool not found: %s", name)
	return nil
}

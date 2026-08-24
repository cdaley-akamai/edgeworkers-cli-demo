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

func TestEdgeWorkersMCPActivateEdgeworkerPreservesNodeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123/activations", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"activationId":456}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"activateEdgeworker"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "activateEdgeworker",
		Arguments: map[string]any{
			"edgeWorkerId":     123,
			"network":          "staging",
			"version":          "2",
			"note":             "ship it",
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"activationId\": 456\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeWorkersMCPCancelActivationAbortSkipsClientCreation(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"cancelActivation"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "cancelActivation",
		Arguments: map[string]any{"edgeWorkerId": 123, "activationId": 456, "confirm": false},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Aborted: cancelActivation requires confirm to be true. Set confirm: true to proceed.", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeWorkersMCPListPropertiesPreservesExplicitFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "false", request.URL.Query().Get("activeOnly"))
		assert.Equal(t, "false", request.URL.Query().Get("details"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"properties":[]}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"listProperties"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "listProperties",
		Arguments: map[string]any{"edgeWorkerId": 123, "activeOnly": false, "details": false},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeWorkersMCPDeploymentCatalog(t *testing.T) {
	config := Config{}
	patterns := []string{
		"listActivations", "getActivation", "activateEdgeworker", "cancelActivation", "rollbackEdgeworker",
		"listDeactivations", "getDeactivation", "deactivateEdgeworker", "listProperties",
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

	activate := mustFindEdgeWorkersTool(t, result.Tools, "activateEdgeworker")
	assert.Equal(t, "Activate a specific version of an EdgeWorker on STAGING or PRODUCTION.", activate.Description)
}

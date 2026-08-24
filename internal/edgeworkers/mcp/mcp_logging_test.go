package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeWorkersMCPCreateLoggingOverridePreservesNodeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/ids/123/loggings", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"level":   "DEBUG",
			"network": "staging",
			"timeout": "2023-10-23T16:37:32Z",
			"schema":  "v1",
			"ds2Id":   float64(789),
		}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"loggingId":456}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"createLoggingOverride"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createLoggingOverride",
		Arguments: map[string]any{
			"edgeWorkerId":     123,
			"level":            "DEBUG",
			"network":          "staging",
			"timeout":          "2023-10-23T16:37:32Z",
			"ds2Id":            789,
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeWorkersMCPLoggingCatalog(t *testing.T) {
	config := Config{}
	patterns := []string{"listLoggingOverrides", "getLoggingOverride", "createLoggingOverride"}
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
	assert.Equal(t, "Get status details for a specific logging override.", mustFindEdgeWorkersTool(t, result.Tools, "getLoggingOverride").Description)
}

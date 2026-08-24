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

func TestEdgeWorkersMCPSecureTokenCatalogAndRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/secure-token", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"propertyId": "prp_123",
			"hostname":   "www.example.com",
			"expiry":     float64(15),
			"hostnames":  []any{"a.example.com"},
		}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"akamaiEwTrace":"jwt"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"createSecureToken"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	tools, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	tool := mustFindEdgeWorkersTool(t, tools.Tools, "createSecureToken")
	assert.Equal(t, "Create a JWT authentication token for securing EdgeWorker JavaScript logging. Used with DataStream 2 log delivery.", tool.Description)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createSecureToken",
		Arguments: map[string]any{
			"propertyId":       "prp_123",
			"hostname":         "www.example.com",
			"expiry":           15,
			"hostnames":        []any{"a.example.com"},
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

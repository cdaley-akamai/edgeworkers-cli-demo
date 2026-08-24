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

func TestEdgeKVMCPRescheduleNamespaceDeleteUsesExpectedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "/edgekv/v1/networks/production/namespaces/sample/status/scheduled-delete", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		var body map[string]string
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, "2026-09-10T12:34:56Z", body["scheduledDeleteTime"])
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"status":"UPDATED"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"rescheduleEdgeKVNamespaceDelete"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "rescheduleEdgeKVNamespaceDelete",
		Arguments: map[string]any{
			"network":             "production",
			"namespaceId":         "sample",
			"scheduledDeleteTime": "2026-09-10T12:34:56Z",
			"accountSwitchKey":    "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"status\": \"UPDATED\"\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPDeleteNamespaceAbortSkipsClientCreation(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"deleteEdgeKVNamespace"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "deleteEdgeKVNamespace",
		Arguments: map[string]any{
			"network":     "staging",
			"namespaceId": "sample",
			"confirm":     false,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Aborted: deleteEdgeKVNamespace requires confirm to be true. Set confirm: true to proceed.", result.Content[0].(*mcp.TextContent).Text)
}

// TestEdgeKVMCPDeleteNamespaceRejectsMissingConfirm proves the SDK's own required-field
// validation rejects a call omitting confirm before the handler runs, making a manual
// nil-check in the handler redundant now that Confirm is a plain (non-pointer) bool.
func TestEdgeKVMCPDeleteNamespaceRejectsMissingConfirm(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"deleteEdgeKVNamespace"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "deleteEdgeKVNamespace",
		Arguments: map[string]any{"network": "staging", "namespaceId": "sample"},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestEdgeKVMCPCancelNamespaceDeleteAbortSkipsClientCreation(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"cancelEdgeKVNamespaceDelete"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "cancelEdgeKVNamespaceDelete",
		Arguments: map[string]any{
			"network":     "production",
			"namespaceId": "sample",
			"confirm":     false,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Aborted: cancelEdgeKVNamespaceDelete requires confirm to be true. Set confirm: true to proceed.", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPReauthorizeNamespacePreservesNumericGroupID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "/edgekv/v1/auth/namespaces/sample", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, 12.5, body["groupId"])
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"status":"AUTHORIZED"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"reauthorizeEdgeKVNamespace"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "reauthorizeEdgeKVNamespace",
		Arguments: map[string]any{
			"namespaceId": "sample",
			"groupId":     12.5,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

// TestEdgeKVMCPCreateNamespaceRejectsMissingRetention and
// TestEdgeKVMCPCreateNamespaceRejectsMissingGroupID prove the SDK's own required-field
// validation rejects a call omitting retentionInSeconds/groupId before the handler runs,
// now that both fields are plain (non-pointer) required values instead of optional pointers.
func TestEdgeKVMCPCreateNamespaceRejectsMissingRetention(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"createEdgeKVNamespace"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createEdgeKVNamespace",
		Arguments: map[string]any{
			"network":   "staging",
			"namespace": "sample",
			"groupId":   0,
		},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestEdgeKVMCPCreateNamespaceRejectsMissingGroupID(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"createEdgeKVNamespace"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createEdgeKVNamespace",
		Arguments: map[string]any{
			"network":            "staging",
			"namespace":          "sample",
			"retentionInSeconds": 0,
		},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// TestEdgeKVMCPCreateNamespaceRejectsOutOfRangeRetention proves the handler's client-side
// range validation rejects an out-of-range retentionInSeconds with a clear MCP error.
func TestEdgeKVMCPCreateNamespaceRejectsOutOfRangeRetention(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"createEdgeKVNamespace"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createEdgeKVNamespace",
		Arguments: map[string]any{
			"network":            "staging",
			"namespace":          "sample",
			"retentionInSeconds": 60,
			"groupId":            0,
		},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestEdgeKVMCPUpdateNamespaceRejectsMissingRetention(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"updateEdgeKVNamespace"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "updateEdgeKVNamespace",
		Arguments: map[string]any{
			"network":     "staging",
			"namespaceId": "sample",
		},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

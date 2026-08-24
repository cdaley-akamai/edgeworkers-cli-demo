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

func TestEdgeKVMCPListItemsUsesExpectedQueryAndAccountOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/networks/staging/namespaces/default/groups/g1", request.URL.Path)
		assert.Equal(t, "25", request.URL.Query().Get("maxItems"))
		assert.Equal(t, "sbx-123", request.URL.Query().Get("sandboxId"))
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `["item-a","item-b"]`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"listEdgeKVItems"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "listEdgeKVItems",
		Arguments: map[string]any{
			"network":          "staging",
			"namespaceId":      "default",
			"groupId":          "g1",
			"maxItems":         25,
			"sandboxId":        "sbx-123",
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "[\n  \"item-a\",\n  \"item-b\"\n]", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPListItemsPreservesNumericMaxItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "25.5", request.URL.Query().Get("maxItems"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `[]`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"listEdgeKVItems"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "listEdgeKVItems",
		Arguments: map[string]any{
			"network":     "staging",
			"namespaceId": "default",
			"groupId":     "g1",
			"maxItems":    25.5,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeKVMCPWriteItemObjectUsesJSONContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "/edgekv/v1/networks/production/namespaces/default/groups/g1/items/item-1", request.URL.Path)
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", request.Header.Get("Accept"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, "v", body["k"])
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"message":"updated"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"writeEdgeKVItem"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "writeEdgeKVItem",
		Arguments: map[string]any{
			"network":     "production",
			"namespaceId": "default",
			"groupId":     "g1",
			"itemId":      "item-1",
			"value":       map[string]any{"k": "v"},
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"message\": \"updated\"\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPWriteItemTextUsesPlainTextContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "text/plain", request.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", request.Header.Get("Accept"))
		data, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"message":"updated"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"writeEdgeKVItem"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "writeEdgeKVItem",
		Arguments: map[string]any{
			"network":     "production",
			"namespaceId": "default",
			"groupId":     "g1",
			"itemId":      "item-1",
			"value":       "hello world",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeKVMCPWriteItemRejectsInvalidValuesBeforeClientCreation(t *testing.T) {
	for name, value := range map[string]any{
		"array":  []any{"not", "an", "object"},
		"number": 42,
	} {
		t.Run(name, func(t *testing.T) {
			config := Config{}
			runtime, err := newEdgeKVMCPServer("9.8.7", []string{"writeEdgeKVItem"}, io.Discard, false, config)
			require.NoError(t, err)
			clientSession := connectEdgeKVMCPClient(t, runtime)

			result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
				Name: "writeEdgeKVItem",
				Arguments: map[string]any{
					"network":     "production",
					"namespaceId": "default",
					"groupId":     "g1",
					"itemId":      "item-1",
					"value":       value,
				},
			})
			require.NoError(t, err)
			require.True(t, result.IsError)
			assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "value must be a string or an object")
		})
	}
}

func TestEdgeKVMCPDeleteItemAbortSkipsClientCreationAndAPI(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"deleteEdgeKVItem"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "deleteEdgeKVItem",
		Arguments: map[string]any{
			"network":     "staging",
			"namespaceId": "default",
			"groupId":     "g1",
			"itemId":      "item-1",
			"confirm":     false,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Aborted: deleteEdgeKVItem requires confirm to be true. Set confirm: true to proceed.", result.Content[0].(*mcp.TextContent).Text)
}

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

func TestEdgeKVMCPCreateTokenUsesExpectedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgekv/v1/tokens", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, "token-a", body["name"])
		assert.Equal(t, true, body["allowOnProduction"])
		assert.Equal(t, true, body["allowOnStaging"])
		assert.Equal(t, "2026-10-10T00:00:00Z", body["expiry"])
		assert.Equal(t, []any{"ewid-1", "ewid-2"}, body["restrictToEdgeWorkerIds"])

		namespacePermissions, ok := body["namespacePermissions"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, []any{"r", "w"}, namespacePermissions["default"])

		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"name":"token-a","uuid":"u-1"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"createEdgeKVToken"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createEdgeKVToken",
		Arguments: map[string]any{
			"name":                 "token-a",
			"allowOnProduction":    true,
			"allowOnStaging":       true,
			"expiry":               "2026-10-10T00:00:00Z",
			"namespacePermissions": map[string]any{"default": []any{"r", "w"}},
			"restrictToEdgeWorkerIds": []any{
				"ewid-1",
				"ewid-2",
			},
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"name\": \"token-a\",\n  \"uuid\": \"u-1\"\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPCreateTokenRejectsOverlengthNameBeforeClientCreation(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"createEdgeKVToken"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createEdgeKVToken",
		Arguments: map[string]any{
			"name":                 "123456789012345678901234567890123",
			"allowOnProduction":    true,
			"allowOnStaging":       true,
			"expiry":               "2026-10-10T00:00:00Z",
			"namespacePermissions": map[string]any{"default": []any{"r"}},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "name must be 32 characters or fewer")
}

func TestEdgeKVMCPCreateTokenRejectsInvalidPermissionBeforeClientCreation(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"createEdgeKVToken"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createEdgeKVToken",
		Arguments: map[string]any{
			"name":                 "token-a",
			"allowOnProduction":    true,
			"allowOnStaging":       true,
			"expiry":               "2026-10-10T00:00:00Z",
			"namespacePermissions": map[string]any{"default": []any{"x"}},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, `invalid permission "x"`)
}

func TestEdgeKVMCPListTokensIncludesIncludeExpiredQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/tokens", request.URL.Path)
		assert.Equal(t, "true", request.URL.Query().Get("includeExpired"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"tokens":[{"name":"token-a"}]}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"listEdgeKVTokens"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "listEdgeKVTokens",
		Arguments: map[string]any{
			"includeExpired": true,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"tokens\": [\n    {\n      \"name\": \"token-a\"\n    }\n  ]\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPGetAndRefreshTokenUseExpectedPaths(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		switch requestCount {
		case 1:
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/edgekv/v1/tokens/token-a", request.URL.Path)
			response.Header().Set("Content-Type", "application/json")
			fmt.Fprint(response, `{"name":"token-a","value":"abc"}`)
		case 2:
			assert.Equal(t, http.MethodPost, request.Method)
			assert.Equal(t, "/edgekv/v1/tokens/token-a/refresh", request.URL.Path)
			response.Header().Set("Content-Type", "application/json")
			fmt.Fprint(response, `{"name":"token-a","value":"def"}`)
		default:
			t.Fatalf("unexpected request %d", requestCount)
		}
	}))
	t.Cleanup(server.Close)

	config := newEdgeKVTestConfig(t, server)
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"getEdgeKVToken", "refreshEdgeKVToken"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "getEdgeKVToken",
		Arguments: map[string]any{"tokenName": "token-a"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"name\": \"token-a\",\n  \"value\": \"abc\"\n}", result.Content[0].(*mcp.TextContent).Text)

	result, err = clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "refreshEdgeKVToken",
		Arguments: map[string]any{"tokenName": "token-a"},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"name\": \"token-a\",\n  \"value\": \"def\"\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeKVMCPDeleteTokenAbortSkipsClientCreationAndAPI(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeKVMCPServer("9.8.7", []string{"deleteEdgeKVToken"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeKVMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "deleteEdgeKVToken",
		Arguments: map[string]any{
			"tokenName": "token-a",
			"confirm":   false,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Aborted: deleteEdgeKVToken requires confirm to be true. Set confirm: true to proceed.", result.Content[0].(*mcp.TextContent).Text)
}

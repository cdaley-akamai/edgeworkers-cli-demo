package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListItemsUsesMaxItemsAndSandboxQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/networks/staging/namespaces/default/groups/g1", request.URL.Path)
		assert.Equal(t, "25", request.URL.Query().Get("maxItems"))
		assert.Equal(t, "sbx-123", request.URL.Query().Get("sandboxId"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `["item-a","item-b"]`)
	}))
	t.Cleanup(server.Close)

	maxItems := 25.0
	items, err := newTestClient(t, server).ListItems(context.Background(), "staging", "default", "g1", &maxItems, "sbx-123")
	require.NoError(t, err)
	assert.Equal(t, []string{"item-a", "item-b"}, items)
}

func TestWriteItemObjectSerializesAsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", request.Header.Get("Accept"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		assert.Equal(t, "v", payload["k"])
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"status":"ok"}`)
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(t, server).WriteItem(context.Background(), "production", "default", "g1", "i1", map[string]any{"k": "v"}, "", "")
	require.NoError(t, err)
}

func TestWriteItemTextPreservesPlainText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "text/plain", request.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", request.Header.Get("Accept"))
		data, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"status":"ok"}`)
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(t, server).WriteItem(context.Background(), "production", "default", "g1", "i1", "hello world", "", "")
	require.NoError(t, err)
}

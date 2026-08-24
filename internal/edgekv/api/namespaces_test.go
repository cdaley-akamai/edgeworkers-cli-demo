package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNamespacesUsesExpectedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/networks/staging/namespaces", request.URL.Path)
		assert.Equal(t, "true", request.URL.Query().Get("details"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"namespaces":[{"namespace":"default","retentionInSeconds":0}]}`)
	}))
	t.Cleanup(server.Close)

	namespaces, err := newTestClient(t, server).ListNamespaces(context.Background(), "staging", true)
	require.NoError(t, err)
	require.Len(t, namespaces, 1)
	assert.Equal(t, "default", namespaces[0].Name)
}

func TestRescheduleNamespaceDeleteUsesExpectedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "/edgekv/v1/networks/production/namespaces/sample/status/scheduled-delete", request.URL.Path)
		var body map[string]string
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, "2026-09-10T12:34:56Z", body["scheduledDeleteTime"])
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"scheduledDeleteTime":"2026-09-10T12:34:56Z"}`)
	}))
	t.Cleanup(server.Close)

	result, err := newTestClient(t, server).RescheduleNamespaceDelete(context.Background(), "production", "sample", "2026-09-10T12:34:56Z")
	require.NoError(t, err)
	assert.Equal(t, "2026-09-10T12:34:56Z", result["scheduledDeleteTime"])
}

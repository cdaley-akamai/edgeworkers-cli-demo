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

func TestListTokensUsesIncludeExpiredQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/tokens", request.URL.Path)
		assert.Equal(t, "true", request.URL.Query().Get("includeExpired"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"tokens":[{"name":"sample"}]}`)
	}))
	t.Cleanup(server.Close)

	includeExpired := true
	result, err := newTestClient(t, server).ListTokens(context.Background(), &includeExpired)
	require.NoError(t, err)
	assert.Equal(t, "sample", result.Tokens[0]["name"])
}

func TestGetRefreshAndDeleteTokenUseExpectedPaths(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		response.Header().Set("Content-Type", "application/json")
		switch requestCount {
		case 1:
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/edgekv/v1/tokens/token-a", request.URL.Path)
			fmt.Fprint(response, `{"name":"token-a","value":"abc"}`)
		case 2:
			assert.Equal(t, http.MethodPost, request.Method)
			assert.Equal(t, "/edgekv/v1/tokens/token-a/refresh", request.URL.Path)
			fmt.Fprint(response, `{"name":"token-a","value":"def"}`)
		case 3:
			assert.Equal(t, http.MethodDelete, request.Method)
			assert.Equal(t, "/edgekv/v1/tokens/token-a", request.URL.Path)
			fmt.Fprint(response, `{"name":"token-a","status":"REVOKED"}`)
		default:
			t.Fatalf("unexpected request %d", requestCount)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	got, err := client.GetToken(context.Background(), "token-a")
	require.NoError(t, err)
	assert.Equal(t, "abc", got["value"])

	refreshed, err := client.RefreshToken(context.Background(), "token-a")
	require.NoError(t, err)
	assert.Equal(t, "def", refreshed["value"])

	revoked, err := client.DeleteToken(context.Background(), "token-a")
	require.NoError(t, err)
	assert.Equal(t, "REVOKED", revoked["status"])
}

func TestCreateTokenBuildsExpectedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgekv/v1/tokens", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, "token-a", body["name"])
		assert.Equal(t, true, body["allowOnStaging"])
		assert.Equal(t, false, body["allowOnProduction"])
		assert.Equal(t, "2026-10-10T00:00:00Z", body["expiry"])
		assert.Equal(t, []any{"123", "456"}, body["restrictToEdgeWorkerIds"])

		namespacePermissions, ok := body["namespacePermissions"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, []any{"r", "w"}, namespacePermissions["default"])
		assert.Equal(t, []any{"d"}, namespacePermissions["archive"])

		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"name":"token-a","uuid":"u-1"}`)
	}))
	t.Cleanup(server.Close)

	created, err := newTestClient(t, server).CreateToken(context.Background(), CreateTokenRequest{
		Name:              "token-a",
		AllowOnStaging:    true,
		AllowOnProduction: false,
		Expiry:            "2026-10-10T00:00:00Z",
		NamespacePermissions: map[string][]string{
			"default": {"r", "w"},
			"archive": {"d"},
		},
		RestrictToEdgeWorkerIDs: []string{"123", "456"},
	})
	require.NoError(t, err)
	assert.Equal(t, "u-1", created["uuid"])
}

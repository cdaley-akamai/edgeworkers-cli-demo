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

func TestListActivationsPreservesOptionalFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123/activations", request.URL.Path)
		assert.Equal(t, "2", request.URL.Query().Get("version"))
		assert.Equal(t, "staging", request.URL.Query().Get("network"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"activations":[]}`)
	}))
	t.Cleanup(server.Close)

	version := "2"
	network := "staging"
	_, err := newTestClient(t, server).ListActivations(context.Background(), 123, &version, &network)
	require.NoError(t, err)
}

func TestActivateEdgeWorkerPreservesNoteAndNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123/activations", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{"network": "staging", "version": "2", "note": "ship it"}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"activationId":456}`)
	}))
	t.Cleanup(server.Close)

	note := "ship it"
	_, err := newTestClient(t, server).ActivateEdgeWorker(context.Background(), 123, DeploymentRequest{
		Network: "staging",
		Version: "2",
		Note:    &note,
	})
	require.NoError(t, err)
}

func TestDeactivateEdgeWorkerPreservesNoteAndNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/ids/123/deactivations", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{"network": "production", "version": "2", "note": "pause"}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"deactivationId":789}`)
	}))
	t.Cleanup(server.Close)

	note := "pause"
	_, err := newTestClient(t, server).DeactivateEdgeWorker(context.Background(), 123, DeploymentRequest{
		Network: "production",
		Version: "2",
		Note:    &note,
	})
	require.NoError(t, err)
}

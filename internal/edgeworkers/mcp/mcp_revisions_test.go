package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeWorkersMCPActivateRevisionPreservesNodeBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/ids/123/revisions/activations", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{"revisionId": "5-1"}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"revisionId":"5-1","status":"PENDING"}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"activateRevision"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "activateRevision",
		Arguments: map[string]any{
			"edgeWorkerId":     123,
			"revisionId":       "5-1",
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeWorkersMCPDownloadRevisionWritesRequestedPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/ids/123/revisions/5-1/content", request.URL.Path)
		fmt.Fprint(response, "bundle-data")
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"downloadRevisionContent"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)
	outputPath := filepath.Join(t.TempDir(), "revision.tgz")

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "downloadRevisionContent",
		Arguments: map[string]any{
			"edgeWorkerId": 123,
			"revisionId":   "5-1",
			"outputPath":   outputPath,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, fmt.Sprintf("{\n  \"message\": \"Revision bundle saved to %s\"\n}", outputPath), result.Content[0].(*mcp.TextContent).Text)
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, []byte("bundle-data"), data)
}

func TestEdgeWorkersMCPRevisionCatalog(t *testing.T) {
	config := Config{}
	patterns := []string{
		"listRevisions", "getRevision", "getRevisionBom", "downloadRevisionContent", "compareRevisions",
		"listRevisionActivations", "activateRevision", "pinRevision", "unpinRevision",
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
	assert.Equal(t,
		"Get the Bill of Materials (BOM) for a revision — lists all sub-worker dependencies and their versions.",
		mustFindEdgeWorkersTool(t, result.Tools, "getRevisionBom").Description,
	)
}

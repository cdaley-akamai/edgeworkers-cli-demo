package mcp

import (
	"context"
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

func TestEdgeWorkersMCPCreateVersionUploadsLocalBundle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123/versions", request.URL.Path)
		assert.Equal(t, "application/gzip", request.Header.Get("Content-Type"))
		data, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("bundle-data"), data)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"version":"1"}`)
	}))
	t.Cleanup(server.Close)

	bundlePath := filepath.Join(t.TempDir(), "bundle.tgz")
	require.NoError(t, os.WriteFile(bundlePath, []byte("bundle-data"), 0o600))
	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"createVersion"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "createVersion",
		Arguments: map[string]any{
			"edgeWorkerId": 123,
			"bundlePath":   bundlePath,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "{\n  \"version\": \"1\"\n}", result.Content[0].(*mcp.TextContent).Text)
}

func TestEdgeWorkersMCPDownloadVersionWritesRequestedPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/ids/123/versions/1/content", request.URL.Path)
		fmt.Fprint(response, "bundle-data")
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"downloadVersionContent"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)
	outputPath := filepath.Join(t.TempDir(), "bundle.tgz")

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "downloadVersionContent",
		Arguments: map[string]any{
			"edgeWorkerId": 123,
			"version":      "1",
			"outputPath":   outputPath,
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, fmt.Sprintf("{\n  \"message\": \"Bundle saved to %s\"\n}", outputPath), result.Content[0].(*mcp.TextContent).Text)
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, []byte("bundle-data"), data)
}

func TestEdgeWorkersMCPVersionCatalog(t *testing.T) {
	config := Config{}
	patterns := []string{"listVersions", "getVersion", "createVersion", "deleteVersion", "downloadVersionContent"}
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
		"Upload a new version of an EdgeWorker code bundle. The bundle must be a GZIP-compressed tarball (.tgz) at the specified local file path.",
		mustFindEdgeWorkersTool(t, result.Tools, "createVersion").Description,
	)
}

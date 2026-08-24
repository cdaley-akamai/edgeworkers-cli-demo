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

func TestEdgeWorkersMCPValidateCodeBundleCatalogAndUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/validations", request.URL.Path)
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		assert.Equal(t, "application/gzip", request.Header.Get("Content-Type"))
		data, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("bundle-data"), data)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"errors":[],"warnings":[]}`)
	}))
	t.Cleanup(server.Close)

	bundlePath := filepath.Join(t.TempDir(), "bundle.tgz")
	require.NoError(t, os.WriteFile(bundlePath, []byte("bundle-data"), 0o600))
	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"validateCodeBundle"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	tools, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	tool := mustFindEdgeWorkersTool(t, tools.Tools, "validateCodeBundle")
	assert.Equal(t, "Validate an EdgeWorker code bundle before uploading it as a new version. Returns a list of errors and warnings. The bundle must be a GZIP-compressed tarball.", tool.Description)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "validateCodeBundle",
		Arguments: map[string]any{
			"bundlePath":       bundlePath,
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

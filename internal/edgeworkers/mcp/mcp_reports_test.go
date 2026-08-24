package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeWorkersMCPGetReportForwardsAllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/reports/5", request.URL.Path)
		query := request.URL.Query()
		assert.Equal(t, "runtimeError", query.Get("status"))
		assert.Equal(t, "onOriginRequest", query.Get("eventHandler"))
		assert.Equal(t, "production", query.Get("network"))
		assert.Equal(t, "3-1", query.Get("revisionId"))
		assert.Equal(t, "A-CCT123", query.Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"reportId":5,"data":[]}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"getReport"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "getReport",
		Arguments: map[string]any{
			"reportId":         5,
			"start":            "2022-10-18T14:10:50Z",
			"end":              "2022-10-19T14:10:50Z",
			"edgeWorker":       "42-1.0",
			"status":           "runtimeError",
			"eventHandler":     "onOriginRequest",
			"network":          "production",
			"revisionId":       "3-1",
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeWorkersMCPReportCatalog(t *testing.T) {
	config := Config{}
	patterns := []string{"listReports", "getReport"}
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
	assert.Equal(t, "List available EdgeWorker reports. Note: reports 2 and 4 are deprecated.", mustFindEdgeWorkersTool(t, result.Tools, "listReports").Description)
}

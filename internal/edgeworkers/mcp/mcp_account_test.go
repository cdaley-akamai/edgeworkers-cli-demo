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

func TestEdgeWorkersMCPListResourceTiersForwardsOptionalContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/resource-tiers", request.URL.Path)
		assert.Equal(t, "ctr_123", request.URL.Query().Get("contractId"))
		assert.Equal(t, "A-CCT123", request.URL.Query().Get("accountSwitchKey"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"resourceTiers":[]}`)
	}))
	t.Cleanup(server.Close)

	config := newEdgeWorkersTestConfig(t, server)
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"listResourceTiers"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "listResourceTiers",
		Arguments: map[string]any{
			"contractId":       "ctr_123",
			"accountSwitchKey": "A-CCT123",
		},
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestEdgeWorkersMCPAccountMetadataCatalog(t *testing.T) {
	config := Config{}
	patterns := []string{"listGroups", "getGroup", "listContracts", "listResourceTiers", "listLimits"}
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

	groups := mustFindEdgeWorkersTool(t, result.Tools, "listGroups")
	assert.Equal(t, "List all permission groups available in the account that can be used with EdgeWorkers.", groups.Description)
}

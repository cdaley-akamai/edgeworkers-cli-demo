package mcp

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeWorkersMCPRegistersCompleteCatalog(t *testing.T) {
	config := Config{}
	runtime, err := newEdgeWorkersMCPServer("9.8.7", []string{"*"}, io.Discard, false, config)
	require.NoError(t, err)
	clientSession := connectEdgeWorkersMCPClient(t, runtime)

	result, err := clientSession.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	assert.ElementsMatch(t, []string{
		"hello_world",
		"listEdgeworkers", "getEdgeworker", "createEdgeworker", "updateEdgeworker", "deleteEdgeworker", "cloneEdgeworker", "getEdgeworkerResourceTier",
		"listVersions", "getVersion", "createVersion", "deleteVersion", "downloadVersionContent",
		"listActivations", "getActivation", "activateEdgeworker", "cancelActivation", "rollbackEdgeworker",
		"listDeactivations", "getDeactivation", "deactivateEdgeworker", "listProperties",
		"listGroups", "getGroup", "listContracts", "listResourceTiers", "listLimits",
		"listReports", "getReport",
		"listLoggingOverrides", "getLoggingOverride", "createLoggingOverride",
		"createSecureToken", "validateCodeBundle",
		"listRevisions", "getRevision", "getRevisionBom", "downloadRevisionContent", "compareRevisions",
		"listRevisionActivations", "activateRevision", "pinRevision", "unpinRevision",
	}, names)
}

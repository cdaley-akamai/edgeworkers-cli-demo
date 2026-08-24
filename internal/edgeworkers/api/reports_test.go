package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListReportsUsesExpectedPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgeworkers/v1/reports", request.URL.Path)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"reports":[]}`)
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(t, server).ListReports(context.Background())
	require.NoError(t, err)
}

func TestGetReportPreservesAllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgeworkers/v1/reports/5", request.URL.Path)
		query := request.URL.Query()
		assert.Equal(t, "2022-10-18T14:10:50Z", query.Get("start"))
		assert.Equal(t, "2022-10-19T14:10:50Z", query.Get("end"))
		assert.Equal(t, "42-1.0", query.Get("edgeWorker"))
		assert.Equal(t, "runtimeError", query.Get("status"))
		assert.Equal(t, "onOriginRequest", query.Get("eventHandler"))
		assert.Equal(t, "production", query.Get("network"))
		assert.Equal(t, "3-1", query.Get("revisionId"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"reportId":5,"data":[]}`)
	}))
	t.Cleanup(server.Close)

	eventHandler := "onOriginRequest"
	network := "production"
	revisionID := "3-1"
	_, err := newTestClient(t, server).GetReport(context.Background(), ReportRequest{
		ReportID:     5,
		Start:        "2022-10-18T14:10:50Z",
		End:          "2022-10-19T14:10:50Z",
		EdgeWorker:   "42-1.0",
		Status:       "runtimeError",
		EventHandler: &eventHandler,
		Network:      &network,
		RevisionID:   &revisionID,
	})
	require.NoError(t, err)
}

func TestListActiveCustomersSupportsAccountAndEdgeWorkerScopes(t *testing.T) {
	expectedPaths := []string{
		"/edgeworkers/v1/ids/active-customers",
		"/edgeworkers/v1/ids/123/active-customers",
	}
	requestIndex := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, expectedPaths[requestIndex], request.URL.Path)
		requestIndex++
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{}`)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.ListActiveCustomers(context.Background(), nil)
	require.NoError(t, err)
	edgeWorkerID := float64(123)
	_, err = client.ListActiveCustomers(context.Background(), &edgeWorkerID)
	require.NoError(t, err)
}

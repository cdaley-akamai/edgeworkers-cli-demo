package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgegrid"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/session"
	"github.com/akamai/edgeworkers-cli/internal/apiclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTransport struct {
	server *httptest.Server
}

func (transport *testTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	cloned := request.Clone(request.Context())
	cloned.URL.Scheme = "http"
	cloned.URL.Host = transport.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(cloned)
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	config := &edgegrid.Config{
		Host:         server.Listener.Addr().String(),
		ClientToken:  "fake-ct",
		ClientSecret: "fake-cs",
		AccessToken:  "fake-at",
	}
	sess, err := session.New(
		session.WithSigner(config),
		session.WithClient(&http.Client{Transport: &testTransport{server: server}}),
	)
	require.NoError(t, err)
	return New(apiclient.FromSession(sess))
}

func TestListEdgeWorkersPreservesFiltersAndResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids", request.URL.Path)
		assert.Equal(t, "12.5", request.URL.Query().Get("groupId"))
		assert.Equal(t, "200.25", request.URL.Query().Get("resourceTierId"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"edgeWorkerIds":[{"edgeWorkerId":123,"description":"kept"}]}`)
	}))
	t.Cleanup(server.Close)

	groupID := 12.5
	resourceTierID := 200.25
	result, err := newTestClient(t, server).ListEdgeWorkers(context.Background(), &groupID, &resourceTierID)
	require.NoError(t, err)

	objects, ok := result.(map[string]any)
	require.True(t, ok)
	edgeWorkers, ok := objects["edgeWorkerIds"].([]any)
	require.True(t, ok)
	assert.Equal(t, "kept", edgeWorkers[0].(map[string]any)["description"])
}

func TestCreateEdgeWorkerSendsOptionalDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"groupId":        12.5,
			"name":           "worker",
			"resourceTierId": 200.25,
			"description":    "kept",
		}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"edgeWorkerId":123}`)
	}))
	t.Cleanup(server.Close)

	description := "kept"
	resourceTierID := 200.25
	_, err := newTestClient(t, server).CreateEdgeWorker(context.Background(), EdgeWorkerMutation{
		GroupID:        12.5,
		Name:           "worker",
		ResourceTierID: &resourceTierID,
		Description:    &description,
	})
	require.NoError(t, err)
}

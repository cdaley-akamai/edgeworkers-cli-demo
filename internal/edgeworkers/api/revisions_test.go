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

func TestRevisionReadsUseExpectedPaths(t *testing.T) {
	expectedPaths := []string{
		"/edgeworkers/v1/ids/123/revisions",
		"/edgeworkers/v1/ids/123/revisions/5-1",
		"/edgeworkers/v1/ids/123/revisions/5-1/bom",
		"/edgeworkers/v1/ids/123/revisions/activations",
	}
	requestIndex := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, expectedPaths[requestIndex], request.URL.Path)
		requestIndex++
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{}`)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.ListRevisions(context.Background(), 123, RevisionListOptions{})
	require.NoError(t, err)
	_, err = client.GetRevision(context.Background(), 123, "5-1")
	require.NoError(t, err)
	_, err = client.GetRevisionBOM(context.Background(), 123, "5-1", RevisionBOMOptions{})
	require.NoError(t, err)
	_, err = client.ListRevisionActivations(context.Background(), 123)
	require.NoError(t, err)
}

func TestRevisionWritesPreserveBodies(t *testing.T) {
	tests := []struct {
		name string
		path string
		body map[string]any
		call func(*Client) error
	}{
		{name: "compare", path: "/edgeworkers/v1/ids/123/revisions/5-1/compare", body: map[string]any{"revisionId": "4-1"}, call: func(client *Client) error {
			_, err := client.CompareRevisions(context.Background(), 123, "5-1", "4-1")
			return err
		}},
		{name: "activate", path: "/edgeworkers/v1/ids/123/revisions/activations", body: map[string]any{"revisionId": "5-1"}, call: func(client *Client) error {
			_, err := client.ActivateRevision(context.Background(), 123, RevisionActivationRequest{RevisionID: "5-1"})
			return err
		}},
		{name: "pin", path: "/edgeworkers/v1/ids/123/revisions/5-1/pin", body: map[string]any{}, call: func(client *Client) error {
			_, err := client.PinRevision(context.Background(), 123, "5-1", RevisionPinRequest{})
			return err
		}},
		{name: "unpin", path: "/edgeworkers/v1/ids/123/revisions/5-1/unpin", body: map[string]any{}, call: func(client *Client) error {
			_, err := client.UnpinRevision(context.Background(), 123, "5-1", RevisionUnpinRequest{})
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodPost, request.Method)
				assert.Equal(t, test.path, request.URL.Path)
				var body map[string]any
				require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
				assert.Equal(t, test.body, body)
				response.Header().Set("Content-Type", "application/json")
				fmt.Fprint(response, `{}`)
			}))
			t.Cleanup(server.Close)
			require.NoError(t, test.call(newTestClient(t, server)))
		})
	}
}

func TestRevisionOperationsPreserveOptions(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		switch requests {
		case 1:
			query := request.URL.Query()
			assert.Equal(t, "2", query.Get("version"))
			assert.Equal(t, "456", query.Get("activationId"))
			assert.Equal(t, "PRODUCTION", query.Get("network"))
			assert.Equal(t, "true", query.Get("pinnedOnly"))
			assert.Equal(t, "true", query.Get("currentlyPinned"))
		case 2:
			assert.Equal(t, "true", request.URL.Query().Get("includeActiveVersions"))
			assert.Equal(t, "true", request.URL.Query().Get("includeCurrentlyPinnedRevisions"))
		case 3:
			var body map[string]any
			require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
			assert.Equal(t, map[string]any{"revisionId": "5-1", "network": "STAGING", "note": "ship"}, body)
		case 4:
			var body map[string]any
			require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
			assert.Equal(t, map[string]any{"unpinNote": "done"}, body)
		}
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{}`)
	}))
	t.Cleanup(server.Close)

	version := "2"
	activationID := "456"
	network := "PRODUCTION"
	trueValue := true
	client := newTestClient(t, server)
	_, err := client.ListRevisions(context.Background(), 123, RevisionListOptions{
		Version:         &version,
		ActivationID:    &activationID,
		Network:         &network,
		PinnedOnly:      &trueValue,
		CurrentlyPinned: &trueValue,
	})
	require.NoError(t, err)
	_, err = client.GetRevisionBOM(context.Background(), 123, "5-1", RevisionBOMOptions{
		IncludeActiveVersions:           &trueValue,
		IncludeCurrentlyPinnedRevisions: &trueValue,
	})
	require.NoError(t, err)
	note := "ship"
	activationNetwork := "STAGING"
	_, err = client.ActivateRevision(context.Background(), 123, RevisionActivationRequest{
		RevisionID: "5-1",
		Network:    &activationNetwork,
		Note:       &note,
	})
	require.NoError(t, err)
	unpinNote := "done"
	_, err = client.UnpinRevision(context.Background(), 123, "5-1", RevisionUnpinRequest{UnpinNote: &unpinNote})
	require.NoError(t, err)
}

func TestDownloadRevisionContentReturnsBundle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/ids/123/revisions/5-1/content", request.URL.Path)
		assert.Equal(t, "application/gzip", request.Header.Get("Accept"))
		fmt.Fprint(response, "bundle-data")
	}))
	t.Cleanup(server.Close)

	data, err := newTestClient(t, server).DownloadRevisionContent(context.Background(), 123, "5-1")
	require.NoError(t, err)
	assert.Equal(t, []byte("bundle-data"), data)
}

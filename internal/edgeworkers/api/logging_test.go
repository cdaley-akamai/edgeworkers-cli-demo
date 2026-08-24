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

func TestLoggingReadsUseExpectedPaths(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		assert.Equal(t, http.MethodGet, request.Method)
		if requests == 1 {
			assert.Equal(t, "/edgeworkers/v1/ids/123/loggings", request.URL.Path)
		} else {
			assert.Equal(t, "/edgeworkers/v1/ids/123/loggings/456", request.URL.Path)
		}
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{}`)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.ListLoggingOverrides(context.Background(), 123)
	require.NoError(t, err)
	_, err = client.GetLoggingOverride(context.Background(), 123, 456)
	require.NoError(t, err)
}

func TestCreateLoggingOverridePreservesPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123/loggings", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"level":   "DEBUG",
			"network": "staging",
			"timeout": "2023-10-23T16:37:32Z",
			"schema":  "v1",
			"ds2Id":   float64(789),
		}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"loggingId":456}`)
	}))
	t.Cleanup(server.Close)

	timeout := "2023-10-23T16:37:32Z"
	schema := "v1"
	ds2ID := float64(789)
	_, err := newTestClient(t, server).CreateLoggingOverride(context.Background(), 123, LoggingOverrideRequest{
		Level:   "DEBUG",
		Network: "staging",
		Timeout: &timeout,
		Schema:  &schema,
		DS2ID:   &ds2ID,
	})
	require.NoError(t, err)
}

func TestCreateLoggingOverrideOmitsUnsetFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{"level": "INFO", "network": "STAGING"}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"loggingId":456}`)
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(t, server).CreateLoggingOverride(context.Background(), 123, LoggingOverrideRequest{
		Level:   "INFO",
		Network: "STAGING",
	})
	require.NoError(t, err)
}

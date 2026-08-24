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

func TestCreateSecureTokenPreservesPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/secure-token", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"propertyId": "prp_123",
			"hostname":   "www.example.com",
			"expiry":     float64(15),
			"hostnames":  []any{"a.example.com", "b.example.com"},
		}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"akamaiEwTrace":"jwt"}`)
	}))
	t.Cleanup(server.Close)

	propertyID := "prp_123"
	hostname := "www.example.com"
	expiry := float64(15)
	hostnames := []string{"a.example.com", "b.example.com"}
	_, err := newTestClient(t, server).CreateSecureToken(context.Background(), SecureTokenRequest{
		PropertyID: &propertyID,
		Hostname:   &hostname,
		Expiry:     &expiry,
		Hostnames:  &hostnames,
	})
	require.NoError(t, err)
}

func TestCreateSecureTokenOmitsAbsentFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.NotContains(t, body, "propertyId")
		assert.NotContains(t, body, "hostname")
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"akamaiEwTrace":"jwt"}`)
	}))
	t.Cleanup(server.Close)

	expiry := float64(300)
	_, err := newTestClient(t, server).CreateSecureToken(context.Background(), SecureTokenRequest{Expiry: &expiry})
	require.NoError(t, err)
}

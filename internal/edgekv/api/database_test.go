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

func TestInitializeDatabaseWrapsDataAccessPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "/edgekv/v1/initialize", request.URL.Path)
		var body map[string]DataAccessPolicy
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, DataAccessPolicy{RestrictDataAccess: true}, body["dataAccessPolicy"])
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"accountStatus":"INITIALIZED","cpcode":"1234"}`)
	}))
	t.Cleanup(server.Close)

	result, err := newTestClient(t, server).InitializeDatabase(context.Background(), &DataAccessPolicy{RestrictDataAccess: true})
	require.NoError(t, err)
	assert.Equal(t, "INITIALIZED", result.AccountStatus)
	assert.Equal(t, "1234", result.CPCode)
}

func TestGetDatabaseStatusUsesInitializePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgekv/v1/initialize", request.URL.Path)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"accountStatus":"INITIALIZED"}`)
	}))
	t.Cleanup(server.Close)

	result, err := newTestClient(t, server).GetDatabaseStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "INITIALIZED", result.AccountStatus)
}

func TestUpdateDatabasePolicyUsesDirectPolicyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPut, request.Method)
		assert.Equal(t, "/edgekv/v1/auth/database", request.URL.Path)
		var body DataAccessPolicy
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, DataAccessPolicy{RestrictDataAccess: true, AllowNamespacePolicyOverride: false}, body)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"accountStatus":"INITIALIZED"}`)
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(t, server).UpdateDatabasePolicy(context.Background(), DataAccessPolicy{RestrictDataAccess: true})
	require.NoError(t, err)
}

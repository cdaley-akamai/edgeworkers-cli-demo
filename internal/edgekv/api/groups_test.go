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

func TestListAndGetGroupsUseExpectedAuthPaths(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestCount++
		switch requestCount {
		case 1:
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/edgekv/v1/auth/groups", request.URL.Path)
			response.Header().Set("Content-Type", "application/json")
			fmt.Fprint(response, `[{"groupId":123,"groupName":"team-a"}]`)
		case 2:
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/edgekv/v1/auth/groups/123", request.URL.Path)
			response.Header().Set("Content-Type", "application/json")
			fmt.Fprint(response, `{"groupId":123,"groupName":"team-a","capabilities":["READ","WRITE"]}`)
		default:
			t.Fatalf("unexpected request %d", requestCount)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	listResult, err := client.ListGroups(context.Background())
	require.NoError(t, err)
	var expectedList any
	require.NoError(t, json.Unmarshal([]byte(`[{"groupId":123,"groupName":"team-a"}]`), &expectedList))
	assert.Equal(t, expectedList, listResult)

	getResult, err := client.GetGroup(context.Background(), "123")
	require.NoError(t, err)
	var expectedGroup any
	require.NoError(t, json.Unmarshal([]byte(`{"groupId":123,"groupName":"team-a","capabilities":["READ","WRITE"]}`), &expectedGroup))
	assert.Equal(t, expectedGroup, getResult)
}

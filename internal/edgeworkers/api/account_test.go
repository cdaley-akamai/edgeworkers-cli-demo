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

func TestAccountMetadataUsesExpectedPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
		call func(*Client) error
	}{
		{name: "list groups", path: "/edgeworkers/v1/groups", call: func(client *Client) error {
			_, err := client.ListGroups(context.Background())
			return err
		}},
		{name: "get group", path: "/edgeworkers/v1/groups/123", call: func(client *Client) error {
			_, err := client.GetGroup(context.Background(), 123)
			return err
		}},
		{name: "list contracts", path: "/edgeworkers/v1/contracts", call: func(client *Client) error {
			_, err := client.ListContracts(context.Background())
			return err
		}},
		{name: "list limits", path: "/edgeworkers/v1/limits", call: func(client *Client) error {
			_, err := client.ListLimits(context.Background())
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				assert.Equal(t, http.MethodGet, request.Method)
				assert.Equal(t, test.path, request.URL.Path)
				response.Header().Set("Content-Type", "application/json")
				fmt.Fprint(response, `{}`)
			}))
			t.Cleanup(server.Close)
			require.NoError(t, test.call(newTestClient(t, server)))
		})
	}
}

func TestListResourceTiersPreservesOptionalContract(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		assert.Equal(t, "/edgeworkers/v1/resource-tiers", request.URL.Path)
		if requests == 1 {
			assert.False(t, request.URL.Query().Has("contractId"))
		} else {
			assert.Equal(t, "ctr_123", request.URL.Query().Get("contractId"))
		}
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{}`)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(t, server)
	_, err := client.ListResourceTiers(context.Background(), nil)
	require.NoError(t, err)
	contractID := "ctr_123"
	_, err = client.ListResourceTiers(context.Background(), &contractID)
	require.NoError(t, err)
}

func TestListPropertiesPreservesExplicitFalseOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/edgeworkers/v1/ids/123/properties", request.URL.Path)
		assert.Equal(t, "false", request.URL.Query().Get("activeOnly"))
		assert.Equal(t, "false", request.URL.Query().Get("details"))
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"properties":[]}`)
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(t, server).ListProperties(context.Background(), 123, false, false)
	require.NoError(t, err)
}

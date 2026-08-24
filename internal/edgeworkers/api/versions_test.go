package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionMetadataUsesExpectedMethodsAndPaths(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		call   func(*Client) error
	}{
		{name: "list", method: http.MethodGet, path: "/edgeworkers/v1/ids/123/versions", call: func(client *Client) error {
			_, err := client.ListVersions(context.Background(), 123)
			return err
		}},
		{name: "get", method: http.MethodGet, path: "/edgeworkers/v1/ids/123/versions/1", call: func(client *Client) error {
			_, err := client.GetVersion(context.Background(), 123, "1")
			return err
		}},
		{name: "delete", method: http.MethodDelete, path: "/edgeworkers/v1/ids/123/versions/1", call: func(client *Client) error {
			return client.DeleteVersion(context.Background(), 123, "1")
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				assert.Equal(t, test.method, request.Method)
				assert.Equal(t, test.path, request.URL.Path)
				if test.method != http.MethodDelete {
					response.Header().Set("Content-Type", "application/json")
					fmt.Fprint(response, `{}`)
				}
			}))
			t.Cleanup(server.Close)
			require.NoError(t, test.call(newTestClient(t, server)))
		})
	}
}

func TestCreateVersionUploadsGzipBundle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123/versions", request.URL.Path)
		assert.Equal(t, "application/gzip", request.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", request.Header.Get("Accept"))
		data, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("bundle-data"), data)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"edgeWorkerId":123,"version":"1"}`)
	}))
	t.Cleanup(server.Close)

	result, err := newTestClient(t, server).CreateVersion(context.Background(), 123, bytes.NewBufferString("bundle-data"))
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"edgeWorkerId": float64(123), "version": "1"}, result)
}

func TestDownloadVersionContentReturnsBundle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, "/edgeworkers/v1/ids/123/versions/1/content", request.URL.Path)
		assert.Equal(t, "application/gzip", request.Header.Get("Accept"))
		fmt.Fprint(response, "bundle-data")
	}))
	t.Cleanup(server.Close)

	data, err := newTestClient(t, server).DownloadVersionContent(context.Background(), 123, "1")
	require.NoError(t, err)
	assert.Equal(t, []byte("bundle-data"), data)
}

func TestValidateCodeBundleUploadsGzip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/edgeworkers/v1/validations", request.URL.Path)
		assert.Equal(t, "application/gzip", request.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", request.Header.Get("Accept"))
		data, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, []byte("bundle-data"), data)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"errors":[],"warnings":[]}`)
	}))
	t.Cleanup(server.Close)

	result, err := newTestClient(t, server).ValidateCodeBundle(context.Background(), bytes.NewBufferString("bundle-data"))
	require.NoError(t, err)
	assert.NotNil(t, result)
}

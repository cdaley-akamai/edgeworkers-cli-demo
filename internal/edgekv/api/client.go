// Package api provides context-aware access to Akamai EdgeKV APIs.
package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/akamai/edgeworkers-cli/internal/apiclient"
)

// Client performs authenticated EdgeKV API operations.
type Client struct {
	api *apiclient.Client
}

// APIError represents a non-2xx response from an Akamai API.
type APIError = apiclient.APIError

// New wraps an authenticated Akamai API client.
func New(client *apiclient.Client) *Client {
	return &Client{api: client}
}

// NewClient builds a Client from .edgerc credentials.
func NewClient(edgercPath, section, accountKey string, timeout time.Duration, debug bool) (*Client, error) {
	client, err := apiclient.New(edgercPath, section, accountKey, timeout, debug)
	if err != nil {
		return nil, err
	}
	return New(client), nil
}

// NewClientWithHTTPClient builds a Client with an optional custom HTTP client.
func NewClientWithHTTPClient(edgercPath, section, accountKey string, timeout time.Duration, debug bool, httpClient *http.Client) (*Client, error) {
	client, err := apiclient.NewWithHTTPClient(edgercPath, section, accountKey, timeout, debug, httpClient)
	if err != nil {
		return nil, err
	}
	return New(client), nil
}

func (client *Client) do(ctx context.Context, method, path string, body, output any) error {
	return client.api.Do(ctx, method, path, body, output)
}

func (client *Client) doRaw(ctx context.Context, method, path string, body io.Reader, contentType, accept string) ([]byte, error) {
	return client.api.DoRaw(ctx, method, path, body, contentType, accept)
}

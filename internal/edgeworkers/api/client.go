// Package api provides context-aware access to Akamai EdgeWorkers APIs.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/akamai/edgeworkers-cli/internal/apiclient"
)

// Client performs authenticated EdgeWorkers API operations.
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

func decodeRawResult(data []byte) any {
	if len(data) == 0 {
		return map[string]any{}
	}
	var result any
	if err := json.Unmarshal(data, &result); err == nil {
		return result
	}
	return string(data)
}

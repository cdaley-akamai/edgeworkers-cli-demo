// Package apiclient creates signed clients shared by the EdgeWorkers and EdgeKV CLIs.
// Its Client combines SDK support with generic helpers for unsupported API endpoints.
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgegrid"
	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/session"
)

// Client holds a signed session and its EdgeWorkers and EdgeKV SDK client.
type Client struct {
	// Session signs requests made outside the SDK.
	Session session.Session
	// Edgeworkers serves EdgeWorkers and EdgeKV SDK operations.
	Edgeworkers sdkew.Edgeworkers
}

// APIError represents a non-2xx Akamai API response.
type APIError struct {
	// StatusCode is the HTTP status returned by the API.
	StatusCode int
	// Title is the API's brief problem summary.
	Title string `json:"title"`
	// Detail contains the API's detailed problem description.
	Detail string `json:"detail"`
	// Instance identifies the API error occurrence when provided.
	Instance string `json:"instance"`
}

// Error returns a concise API error description.
func (e *APIError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Detail)
}

// New builds a signed client from .edgerc credentials.
func New(edgercPath, section, accountKey string, timeout time.Duration, debug bool) (*Client, error) {
	return NewWithHTTPClient(edgercPath, section, accountKey, timeout, debug, nil)
}

// NewWithHTTPClient builds a signed client using an optional custom HTTP client.
func NewWithHTTPClient(edgercPath, section, accountKey string, timeout time.Duration, debug bool, httpClient *http.Client) (*Client, error) {
	if edgercPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot find home directory: %w", err)
		}
		edgercPath = home + "/.edgerc"
	}

	cfg, err := edgegrid.New(
		edgegrid.WithFile(edgercPath),
		edgegrid.WithSection(section),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load credentials from %s [%s]:\n  %w\n  Ensure host, client_token, client_secret, and access_token are set.",
			edgercPath, section, err,
		)
	}
	if accountKey != "" {
		cfg.AccountKey = accountKey
	}
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout == 0 {
		copy := *httpClient
		copy.Timeout = timeout
		httpClient = &copy
	}

	opts := []session.Option{
		session.WithSigner(cfg),
		session.WithClient(httpClient),
	}
	if debug {
		opts = append(opts, session.WithHTTPTracing(true))
	}

	sess, err := session.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return FromSession(sess), nil
}

// FromSession wraps an existing signed session. It supports tests and SDK reuse.
func FromSession(sess session.Session) *Client {
	return &Client{Session: sess, Edgeworkers: sdkew.Client(sess)}
}

// Do signs and executes a JSON request. It accepts any successful HTTP status.
func (c *Client) Do(ctx context.Context, method, path string, body, out interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	data, err := c.do(ctx, method, path, reader, contentType(body != nil, "application/json"), "application/json")
	if err != nil {
		return err
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// DoRaw signs and executes a request with a caller-provided body and content type.
func (c *Client) DoRaw(ctx context.Context, method, path string, body io.Reader, contentType, accept string) ([]byte, error) {
	return c.do(ctx, method, path, body, contentType, accept)
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	req.Header.Set("X-EW-CLIENT", "CLI")

	if err := c.Session.Sign(req); err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}
	resp, err := c.Session.Client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		_ = json.Unmarshal(data, apiErr)
		if apiErr.Detail == "" {
			apiErr.Detail = string(data)
		}
		return nil, apiErr
	}
	return data, nil
}

func contentType(set bool, value string) string {
	if set {
		return value
	}
	return ""
}

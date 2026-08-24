package api

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// ListVersions returns versions for an EdgeWorker ID.
func (client *Client) ListVersions(ctx context.Context, edgeWorkerID float64) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, edgeWorkerIDPath(edgeWorkerID)+"/versions", nil)
}

// GetVersion returns a specific EdgeWorker version.
func (client *Client) GetVersion(ctx context.Context, edgeWorkerID float64, version string) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, versionPath(edgeWorkerID, version), nil)
}

// CreateVersion uploads a GZIP code bundle as a new EdgeWorker version.
func (client *Client) CreateVersion(ctx context.Context, edgeWorkerID float64, bundle io.Reader) (any, error) {
	data, err := client.api.DoRaw(
		ctx,
		http.MethodPost,
		edgeWorkerIDPath(edgeWorkerID)+"/versions",
		bundle,
		"application/gzip",
		"application/json",
	)
	if err != nil {
		return nil, err
	}
	return decodeRawResult(data), nil
}

// ValidateCodeBundle remotely validates a GZIP EdgeWorkers code bundle.
func (client *Client) ValidateCodeBundle(ctx context.Context, bundle io.Reader) (any, error) {
	data, err := client.api.DoRaw(
		ctx,
		http.MethodPost,
		"/edgeworkers/v1/validations",
		bundle,
		"application/gzip",
		"application/json",
	)
	if err != nil {
		return nil, err
	}
	return decodeRawResult(data), nil
}

// DeleteVersion permanently deletes an EdgeWorker version.
func (client *Client) DeleteVersion(ctx context.Context, edgeWorkerID float64, version string) error {
	return client.api.Do(ctx, http.MethodDelete, versionPath(edgeWorkerID, version), nil, nil)
}

// DownloadVersionContent returns the GZIP code bundle for an EdgeWorker version.
func (client *Client) DownloadVersionContent(ctx context.Context, edgeWorkerID float64, version string) ([]byte, error) {
	return client.api.DoRaw(ctx, http.MethodGet, versionPath(edgeWorkerID, version)+"/content", nil, "", "application/gzip")
}

func versionPath(edgeWorkerID float64, version string) string {
	return edgeWorkerIDPath(edgeWorkerID) + "/versions/" + url.PathEscape(version)
}

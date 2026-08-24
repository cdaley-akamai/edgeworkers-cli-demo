package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// EdgeWorkerMutation describes fields accepted when creating, updating, or cloning an EdgeWorker ID.
type EdgeWorkerMutation struct {
	// GroupID identifies the permission group that owns the EdgeWorker ID.
	GroupID float64 `json:"groupId"`
	// Name is the EdgeWorker ID name.
	Name string `json:"name"`
	// ResourceTierID selects the resource tier when provided.
	ResourceTierID *float64 `json:"resourceTierId,omitempty"`
	// Description is optional descriptive text.
	Description *string `json:"description,omitempty"`
}

// ListEdgeWorkers returns EdgeWorker IDs, optionally filtered by group and resource tier.
func (client *Client) ListEdgeWorkers(ctx context.Context, groupID, resourceTierID *float64) (any, error) {
	query := url.Values{}
	if groupID != nil {
		query.Set("groupId", strconv.FormatFloat(*groupID, 'f', -1, 64))
	}
	if resourceTierID != nil {
		query.Set("resourceTierId", strconv.FormatFloat(*resourceTierID, 'f', -1, 64))
	}
	path := "/edgeworkers/v1/ids"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// GetEdgeWorker returns an EdgeWorker ID by identifier.
func (client *Client) GetEdgeWorker(ctx context.Context, edgeWorkerID float64) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, edgeWorkerIDPath(edgeWorkerID), nil)
}

// CreateEdgeWorker registers an EdgeWorker ID.
func (client *Client) CreateEdgeWorker(ctx context.Context, request EdgeWorkerMutation) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, "/edgeworkers/v1/ids", request)
}

// UpdateEdgeWorker updates an EdgeWorker ID.
func (client *Client) UpdateEdgeWorker(ctx context.Context, edgeWorkerID float64, request EdgeWorkerMutation) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPut, edgeWorkerIDPath(edgeWorkerID), request)
}

// DeleteEdgeWorker permanently deletes an EdgeWorker ID.
func (client *Client) DeleteEdgeWorker(ctx context.Context, edgeWorkerID float64) error {
	return client.api.Do(ctx, http.MethodDelete, edgeWorkerIDPath(edgeWorkerID), nil, nil)
}

// CloneEdgeWorker clones an EdgeWorker ID to another resource tier.
func (client *Client) CloneEdgeWorker(ctx context.Context, edgeWorkerID float64, request EdgeWorkerMutation) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, edgeWorkerIDPath(edgeWorkerID)+"/clone", request)
}

// GetEdgeWorkerResourceTier returns the resource tier assigned to an EdgeWorker ID.
func (client *Client) GetEdgeWorkerResourceTier(ctx context.Context, edgeWorkerID float64) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, edgeWorkerIDPath(edgeWorkerID)+"/resource-tier", nil)
}

func (client *Client) edgeWorkerResult(ctx context.Context, method, path string, body any) (any, error) {
	var result any
	if err := client.api.Do(ctx, method, path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func edgeWorkerIDPath(edgeWorkerID float64) string {
	return "/edgeworkers/v1/ids/" + strconv.FormatFloat(edgeWorkerID, 'f', -1, 64)
}

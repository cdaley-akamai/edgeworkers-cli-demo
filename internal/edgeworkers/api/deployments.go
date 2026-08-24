package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// DeploymentRequest describes an activation, rollback, or deactivation request.
type DeploymentRequest struct {
	// Network is the staging or production deployment network.
	Network string `json:"network"`
	// Version is the EdgeWorker version when the operation requires one.
	Version string `json:"version,omitempty"`
	// Note is optional text associated with the operation.
	Note *string `json:"note,omitempty"`
	// AutoPin controls automatic revision pinning when supported.
	AutoPin *bool `json:"autoPin,omitempty"`
}

// ListActivations returns activations for an EdgeWorker ID.
func (client *Client) ListActivations(ctx context.Context, edgeWorkerID float64, version, network *string) (any, error) {
	return client.listDeploymentEvents(ctx, edgeWorkerID, "activations", version, network)
}

// GetActivation returns an activation by identifier.
func (client *Client) GetActivation(ctx context.Context, edgeWorkerID, activationID float64) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, deploymentEventPath(edgeWorkerID, "activations", activationID), nil)
}

// ActivateEdgeWorker deploys an EdgeWorker version to a network.
func (client *Client) ActivateEdgeWorker(ctx context.Context, edgeWorkerID float64, request DeploymentRequest) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, edgeWorkerIDPath(edgeWorkerID)+"/activations", request)
}

// CancelActivation cancels a pending activation.
func (client *Client) CancelActivation(ctx context.Context, edgeWorkerID, activationID float64) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodDelete, deploymentEventPath(edgeWorkerID, "activations", activationID), nil)
}

// RollbackEdgeWorker rolls an EdgeWorker back on a network.
func (client *Client) RollbackEdgeWorker(ctx context.Context, edgeWorkerID float64, request DeploymentRequest) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, edgeWorkerIDPath(edgeWorkerID)+"/activations/rollback", request)
}

// ListDeactivations returns deactivations for an EdgeWorker ID.
func (client *Client) ListDeactivations(ctx context.Context, edgeWorkerID float64, version, network *string) (any, error) {
	return client.listDeploymentEvents(ctx, edgeWorkerID, "deactivations", version, network)
}

// GetDeactivation returns a deactivation by identifier.
func (client *Client) GetDeactivation(ctx context.Context, edgeWorkerID, deactivationID float64) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, deploymentEventPath(edgeWorkerID, "deactivations", deactivationID), nil)
}

// DeactivateEdgeWorker removes an EdgeWorker version from a network.
func (client *Client) DeactivateEdgeWorker(ctx context.Context, edgeWorkerID float64, request DeploymentRequest) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, edgeWorkerIDPath(edgeWorkerID)+"/deactivations", request)
}

func (client *Client) listDeploymentEvents(ctx context.Context, edgeWorkerID float64, resource string, version, network *string) (any, error) {
	query := url.Values{}
	if version != nil {
		query.Set("version", *version)
	}
	if network != nil {
		query.Set("network", *network)
	}
	path := edgeWorkerIDPath(edgeWorkerID) + "/" + resource
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

func deploymentEventPath(edgeWorkerID float64, resource string, eventID float64) string {
	return edgeWorkerIDPath(edgeWorkerID) + "/" + resource + "/" + strconv.FormatFloat(eventID, 'f', -1, 64)
}

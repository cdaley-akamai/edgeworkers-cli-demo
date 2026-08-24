package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ListGroups returns permission groups available to EdgeWorkers.
func (client *Client) ListGroups(ctx context.Context) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, "/edgeworkers/v1/groups", nil)
}

// GetGroup returns a permission group by identifier.
func (client *Client) GetGroup(ctx context.Context, groupID float64) (any, error) {
	path := "/edgeworkers/v1/groups/" + strconv.FormatFloat(groupID, 'f', -1, 64)
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// ListContracts returns contracts that support EdgeWorkers.
func (client *Client) ListContracts(ctx context.Context) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, "/edgeworkers/v1/contracts", nil)
}

// ListResourceTiers returns resource tiers, optionally filtered by contract.
func (client *Client) ListResourceTiers(ctx context.Context, contractID *string) (any, error) {
	path := "/edgeworkers/v1/resource-tiers"
	if contractID != nil {
		path += "?" + url.Values{"contractId": {*contractID}}.Encode()
	}
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// ListLimits returns account-level EdgeWorkers resource limits.
func (client *Client) ListLimits(ctx context.Context) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, "/edgeworkers/v1/limits", nil)
}

// ListProperties returns properties associated with an EdgeWorker ID.
func (client *Client) ListProperties(ctx context.Context, edgeWorkerID float64, activeOnly, details bool) (any, error) {
	query := url.Values{}
	query.Set("activeOnly", strconv.FormatBool(activeOnly))
	query.Set("details", strconv.FormatBool(details))
	path := edgeWorkerIDPath(edgeWorkerID) + "/properties?" + query.Encode()
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

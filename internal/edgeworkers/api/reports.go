package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ReportRequest describes filters for running an EdgeWorkers report.
type ReportRequest struct {
	// ReportID identifies the report type.
	ReportID float64
	// Start is the inclusive report start timestamp.
	Start string
	// End is the inclusive report end timestamp.
	End string
	// EdgeWorker filters by EdgeWorker ID or ID-version.
	EdgeWorker string
	// Status filters by execution status when non-empty.
	Status string
	// EventHandler optionally filters by event handler.
	EventHandler *string
	// Network optionally filters by staging or production network.
	Network *string
	// RevisionID optionally filters by revision identifier.
	RevisionID *string
}

// ListReports returns available EdgeWorkers report types.
func (client *Client) ListReports(ctx context.Context) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, "/edgeworkers/v1/reports", nil)
}

// GetReport runs an EdgeWorkers report with the supplied filters.
func (client *Client) GetReport(ctx context.Context, request ReportRequest) (any, error) {
	query := url.Values{
		"start":      {request.Start},
		"end":        {request.End},
		"edgeWorker": {request.EdgeWorker},
	}
	if request.Status != "" {
		query.Set("status", request.Status)
	}
	if request.EventHandler != nil {
		query.Set("eventHandler", *request.EventHandler)
	}
	if request.Network != nil {
		query.Set("network", *request.Network)
	}
	if request.RevisionID != nil {
		query.Set("revisionId", *request.RevisionID)
	}
	path := "/edgeworkers/v1/reports/" + strconv.FormatFloat(request.ReportID, 'f', -1, 64) + "?" + query.Encode()
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// ListActiveCustomers returns active customers at account or EdgeWorker scope.
func (client *Client) ListActiveCustomers(ctx context.Context, edgeWorkerID *float64) (any, error) {
	path := "/edgeworkers/v1/ids"
	if edgeWorkerID != nil {
		path += "/" + strconv.FormatFloat(*edgeWorkerID, 'f', -1, 64)
	}
	return client.edgeWorkerResult(ctx, http.MethodGet, path+"/active-customers", nil)
}

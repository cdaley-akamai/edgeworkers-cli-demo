package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Revision represents a single immutable revision object.
type Revision struct {
	// RevisionID identifies the immutable revision.
	RevisionID string `json:"revisionId"`
	// Version is the bundle version assigned to the revision.
	Version string `json:"version"`
	// Network is the Akamai network associated with the revision.
	Network string `json:"network"`
	// Status is the revision lifecycle state.
	Status string `json:"status"`
	// CreatedTime is the API-provided revision creation time.
	CreatedTime string `json:"createdTime"`
}

// RevisionList is the response returned when listing revisions.
type RevisionList struct {
	// Revisions contains immutable revision records.
	Revisions []Revision `json:"revisions"`
}

// RevisionActivation holds the activation status for a revision.
type RevisionActivation struct {
	// RevisionID identifies the activated revision.
	RevisionID string `json:"revisionId"`
	// Network is the Akamai network where activation occurs.
	Network string `json:"network"`
	// Status is the activation lifecycle state.
	Status string `json:"status"`
}

// RevisionActivationList is the response returned when listing revision activations.
type RevisionActivationList struct {
	// Activations contains revision activation records.
	Activations []RevisionActivation `json:"activations"`
}

// RevisionComparison holds dependency differences between two revisions.
type RevisionComparison struct {
	// Added contains entries added by the comparison.
	Added []any `json:"added"`
	// Removed contains entries removed by the comparison.
	Removed []any `json:"removed"`
	// Changed contains entries changed by the comparison.
	Changed []any `json:"changed"`
}

// RevisionBOM holds the bill of materials for a revision.
type RevisionBOM struct {
	// RevisionID identifies the revision described by Packages.
	RevisionID string `json:"revisionId"`
	// Packages contains bill-of-materials entries for RevisionID.
	Packages []any `json:"packages"`
}

// PinResult is returned by pin and unpin operations.
type PinResult struct {
	// RevisionID identifies the revision changed by the operation.
	RevisionID string `json:"revisionId"`
	// Pinned reports whether RevisionID is pinned after the operation.
	Pinned bool `json:"pinned"`
}

// RevisionListOptions controls optional revision list filters.
type RevisionListOptions struct {
	// Version optionally filters by bundle version.
	Version *string
	// ActivationID optionally filters by activation identifier.
	ActivationID *string
	// Network optionally filters by network.
	Network *string
	// PinnedOnly optionally filters to pinned revisions.
	PinnedOnly *bool
	// CurrentlyPinned optionally filters to currently pinned revisions.
	CurrentlyPinned *bool
}

// RevisionBOMOptions controls optional bill-of-materials inclusions.
type RevisionBOMOptions struct {
	// IncludeActiveVersions includes active versions when set.
	IncludeActiveVersions *bool
	// IncludeCurrentlyPinnedRevisions includes currently pinned revisions when set.
	IncludeCurrentlyPinnedRevisions *bool
}

// RevisionActivationRequest describes a revision activation.
type RevisionActivationRequest struct {
	// RevisionID identifies the revision to activate.
	RevisionID string `json:"revisionId"`
	// Network optionally selects the activation network.
	Network *string `json:"network,omitempty"`
	// Note is optional activation text.
	Note *string `json:"note,omitempty"`
}

// RevisionPinRequest describes a revision pin operation.
type RevisionPinRequest struct {
	// PinNote is optional context for the pin.
	PinNote *string `json:"pinNote,omitempty"`
}

// RevisionUnpinRequest describes a revision unpin operation.
type RevisionUnpinRequest struct {
	// UnpinNote is optional context for the unpin.
	UnpinNote *string `json:"unpinNote,omitempty"`
}

// ListRevisions returns revisions for an EdgeWorker ID.
func (client *Client) ListRevisions(ctx context.Context, edgeWorkerID float64, options RevisionListOptions) (any, error) {
	query := url.Values{}
	if options.Version != nil {
		query.Set("version", *options.Version)
	}
	if options.ActivationID != nil {
		query.Set("activationId", *options.ActivationID)
	}
	if options.Network != nil {
		query.Set("network", *options.Network)
	}
	if options.PinnedOnly != nil {
		query.Set("pinnedOnly", strconv.FormatBool(*options.PinnedOnly))
	}
	if options.CurrentlyPinned != nil {
		query.Set("currentlyPinned", strconv.FormatBool(*options.CurrentlyPinned))
	}
	path := edgeWorkerIDPath(edgeWorkerID) + "/revisions"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// GetRevision returns a revision by identifier.
func (client *Client) GetRevision(ctx context.Context, edgeWorkerID float64, revisionID string) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, revisionPath(edgeWorkerID, revisionID), nil)
}

// GetRevisionBOM returns the bill of materials for a revision.
func (client *Client) GetRevisionBOM(ctx context.Context, edgeWorkerID float64, revisionID string, options RevisionBOMOptions) (any, error) {
	query := url.Values{}
	if options.IncludeActiveVersions != nil {
		query.Set("includeActiveVersions", strconv.FormatBool(*options.IncludeActiveVersions))
	}
	if options.IncludeCurrentlyPinnedRevisions != nil {
		query.Set("includeCurrentlyPinnedRevisions", strconv.FormatBool(*options.IncludeCurrentlyPinnedRevisions))
	}
	path := revisionPath(edgeWorkerID, revisionID) + "/bom"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// DownloadRevisionContent returns the combined GZIP bundle for a revision.
func (client *Client) DownloadRevisionContent(ctx context.Context, edgeWorkerID float64, revisionID string) ([]byte, error) {
	return client.api.DoRaw(ctx, http.MethodGet, revisionPath(edgeWorkerID, revisionID)+"/content", nil, "", "application/gzip")
}

// CompareRevisions returns dependency differences between two revisions.
func (client *Client) CompareRevisions(ctx context.Context, edgeWorkerID float64, revisionID, compareRevisionID string) (any, error) {
	body := map[string]string{"revisionId": compareRevisionID}
	return client.edgeWorkerResult(ctx, http.MethodPost, revisionPath(edgeWorkerID, revisionID)+"/compare", body)
}

// ListRevisionActivations returns revision activations for an EdgeWorker ID.
func (client *Client) ListRevisionActivations(ctx context.Context, edgeWorkerID float64) (any, error) {
	path := edgeWorkerIDPath(edgeWorkerID) + "/revisions/activations"
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// ActivateRevision activates a revision for an EdgeWorker ID.
func (client *Client) ActivateRevision(ctx context.Context, edgeWorkerID float64, request RevisionActivationRequest) (any, error) {
	path := edgeWorkerIDPath(edgeWorkerID) + "/revisions/activations"
	return client.edgeWorkerResult(ctx, http.MethodPost, path, request)
}

// PinRevision pins a revision.
func (client *Client) PinRevision(ctx context.Context, edgeWorkerID float64, revisionID string, request RevisionPinRequest) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, revisionPath(edgeWorkerID, revisionID)+"/pin", request)
}

// UnpinRevision unpins a revision.
func (client *Client) UnpinRevision(ctx context.Context, edgeWorkerID float64, revisionID string, request RevisionUnpinRequest) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, revisionPath(edgeWorkerID, revisionID)+"/unpin", request)
}

func revisionPath(edgeWorkerID float64, revisionID string) string {
	return edgeWorkerIDPath(edgeWorkerID) + "/revisions/" + url.PathEscape(revisionID)
}

package api

import (
	"context"
	"net/http"
	"strconv"
)

// LoggingOverrideRequest describes a temporary logging override.
type LoggingOverrideRequest struct {
	// Level is the requested logging level.
	Level string `json:"level"`
	// Network is the staging or production network.
	Network string `json:"network"`
	// Timeout is the optional override expiration time.
	Timeout *string `json:"timeout,omitempty"`
	// Schema is the optional logging schema version.
	Schema *string `json:"schema,omitempty"`
	// DS2ID is the optional DataStream 2 stream identifier.
	DS2ID *float64 `json:"ds2Id,omitempty"`
}

// LogLevel represents a logging override for an EdgeWorker ID.
type LogLevel struct {
	// LoggingID identifies this logging override.
	LoggingID string `json:"loggingId"`
	// Level is the selected logging override level.
	Level string `json:"level"`
	// Network is the Akamai network where the logging override applies.
	Network string `json:"network"`
	// ExpiresAt is the API-provided expiration time for this override.
	ExpiresAt string `json:"expiresAt"`
}

// LogLevelList is the response returned when listing logging overrides.
type LogLevelList struct {
	// Loggings contains the logging overrides.
	Loggings []LogLevel `json:"loggings"`
}

// ListLoggingOverrides returns all logging overrides for an EdgeWorker ID.
func (client *Client) ListLoggingOverrides(ctx context.Context, edgeWorkerID float64) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodGet, edgeWorkerIDPath(edgeWorkerID)+"/loggings", nil)
}

// GetLoggingOverride returns a logging override by identifier.
func (client *Client) GetLoggingOverride(ctx context.Context, edgeWorkerID, loggingID float64) (any, error) {
	path := edgeWorkerIDPath(edgeWorkerID) + "/loggings/" + strconv.FormatFloat(loggingID, 'f', -1, 64)
	return client.edgeWorkerResult(ctx, http.MethodGet, path, nil)
}

// CreateLoggingOverride creates a temporary logging override.
func (client *Client) CreateLoggingOverride(ctx context.Context, edgeWorkerID float64, request LoggingOverrideRequest) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, edgeWorkerIDPath(edgeWorkerID)+"/loggings", request)
}

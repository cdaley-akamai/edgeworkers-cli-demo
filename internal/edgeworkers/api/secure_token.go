package api

import (
	"context"
	"net/http"
)

// SecureTokenRequest describes a secure-token request for enhanced debug headers.
type SecureTokenRequest struct {
	// PropertyID optionally identifies the property to secure.
	PropertyID *string `json:"propertyId,omitempty"`
	// Hostname optionally identifies the primary hostname.
	Hostname *string `json:"hostname,omitempty"`
	// Expiry optionally sets the token lifetime in minutes, from 1 to 720.
	Expiry *float64 `json:"expiry,omitempty"`
	// Hostnames optionally lists additional allowed hostnames. Specify "/*" to allow all hosts.
	Hostnames *[]string `json:"hostnames,omitempty"`
}

// AuthToken holds the result of a secure-token request.
type AuthToken struct {
	// Token is the JWT authentication token returned by the API.
	Token string `json:"akamaiEwTrace"`
}

// CreateSecureToken creates a JWT token for secure EdgeWorkers logging.
func (client *Client) CreateSecureToken(ctx context.Context, request SecureTokenRequest) (any, error) {
	return client.edgeWorkerResult(ctx, http.MethodPost, "/edgeworkers/v1/secure-token", request)
}

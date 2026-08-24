package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ListTokensResponse contains EdgeKV access token metadata.
type ListTokensResponse struct {
	// Tokens contains one metadata object per access token.
	Tokens []map[string]any `json:"tokens"`
}

// CreateTokenRequest describes a new EdgeKV access token.
type CreateTokenRequest struct {
	// Name identifies the token.
	Name string
	// AllowOnStaging permits token use on the staging network.
	AllowOnStaging bool
	// AllowOnProduction permits token use on the production network.
	AllowOnProduction bool
	// Expiry is the optional token expiration timestamp.
	Expiry string
	// NamespacePermissions maps namespace names to r, w, and d permission codes.
	NamespacePermissions map[string][]string
	// RestrictToEdgeWorkerIDs limits token use to the listed EdgeWorker IDs.
	RestrictToEdgeWorkerIDs []string
}

// ListTokens returns EdgeKV access token metadata.
func (client *Client) ListTokens(ctx context.Context, includeExpired *bool) (ListTokensResponse, error) {
	path := "/edgekv/v1/tokens"
	if includeExpired != nil {
		query := url.Values{}
		query.Set("includeExpired", strconv.FormatBool(*includeExpired))
		path += "?" + query.Encode()
	}
	var result ListTokensResponse
	if err := client.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return ListTokensResponse{}, err
	}
	return result, nil
}

// GetToken downloads an EdgeKV access token by name.
func (client *Client) GetToken(ctx context.Context, tokenName string) (map[string]any, error) {
	var result map[string]any
	path := "/edgekv/v1/tokens/" + url.PathEscape(tokenName)
	if err := client.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateToken creates an EdgeKV access token.
func (client *Client) CreateToken(ctx context.Context, request CreateTokenRequest) (map[string]any, error) {
	body := map[string]any{
		"name":                 request.Name,
		"allowOnStaging":       request.AllowOnStaging,
		"allowOnProduction":    request.AllowOnProduction,
		"namespacePermissions": request.NamespacePermissions,
	}
	if request.Expiry != "" {
		body["expiry"] = request.Expiry
	}
	if len(request.RestrictToEdgeWorkerIDs) > 0 {
		body["restrictToEdgeWorkerIds"] = request.RestrictToEdgeWorkerIDs
	}
	var result map[string]any
	if err := client.do(ctx, http.MethodPost, "/edgekv/v1/tokens", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteToken revokes an EdgeKV access token.
func (client *Client) DeleteToken(ctx context.Context, tokenName string) (map[string]any, error) {
	var result map[string]any
	path := "/edgekv/v1/tokens/" + url.PathEscape(tokenName)
	if err := client.do(ctx, http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RefreshToken requests an early refresh of an EdgeKV access token.
func (client *Client) RefreshToken(ctx context.Context, tokenName string) (map[string]any, error) {
	var result map[string]any
	path := "/edgekv/v1/tokens/" + url.PathEscape(tokenName) + "/refresh"
	if err := client.do(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

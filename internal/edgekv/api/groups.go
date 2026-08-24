package api

import (
	"context"
	"net/http"
	"net/url"
)

// ListGroups returns permission groups with EdgeKV capabilities.
func (client *Client) ListGroups(ctx context.Context) (any, error) {
	var result any
	if err := client.do(ctx, http.MethodGet, "/edgekv/v1/auth/groups", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetGroup returns an EdgeKV permission group by identifier.
func (client *Client) GetGroup(ctx context.Context, groupID string) (any, error) {
	var result any
	path := "/edgekv/v1/auth/groups/" + url.PathEscape(groupID)
	if err := client.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

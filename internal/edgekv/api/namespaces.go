package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	sdkew "github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgeworkers"
)

// NamespaceDetails describes an EdgeKV namespace and its current state.
type NamespaceDetails struct {
	// Name is the namespace identifier.
	Name string `json:"namespace"`
	// NamespaceStatus is the namespace lifecycle state.
	NamespaceStatus string `json:"namespaceStatus"`
	// Retention is the data retention period in seconds.
	Retention *int `json:"retentionInSeconds"`
	// GroupID is the access-control group identifier.
	GroupID *int `json:"groupId"`
	// GeoLocation is the persistent storage location.
	GeoLocation string `json:"geoLocation"`
}

// ListNamespaces returns namespaces on an EdgeKV network.
func (client *Client) ListNamespaces(ctx context.Context, network string, details bool) ([]sdkew.Namespace, error) {
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces", url.PathEscape(network))
	if details {
		path += "?details=true"
	}
	result := &struct {
		Namespaces []sdkew.Namespace `json:"namespaces"`
	}{}
	if err := client.do(ctx, http.MethodGet, path, nil, result); err != nil {
		return nil, err
	}
	return result.Namespaces, nil
}

// GetNamespace returns a namespace on an EdgeKV network.
func (client *Client) GetNamespace(ctx context.Context, network, namespace string) (*NamespaceDetails, error) {
	result := &NamespaceDetails{}
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s", url.PathEscape(network), url.PathEscape(namespace))
	if err := client.do(ctx, http.MethodGet, path, nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateNamespace creates a namespace on an EdgeKV network.
func (client *Client) CreateNamespace(ctx context.Context, network string, body map[string]any) (*sdkew.Namespace, error) {
	result := &sdkew.Namespace{}
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces", url.PathEscape(network))
	if err := client.do(ctx, http.MethodPost, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateNamespace updates a namespace on an EdgeKV network.
func (client *Client) UpdateNamespace(ctx context.Context, network, namespace string, body map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s", url.PathEscape(network), url.PathEscape(namespace))
	if err := client.do(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteNamespace schedules deletion of a namespace on an EdgeKV network.
func (client *Client) DeleteNamespace(ctx context.Context, network, namespace string) (map[string]any, error) {
	result := make(map[string]any)
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s", url.PathEscape(network), url.PathEscape(namespace))
	if err := client.do(ctx, http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListNamespaceGroups returns data groups in a namespace.
func (client *Client) ListNamespaceGroups(ctx context.Context, network, namespace string) ([]string, error) {
	var result []string
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s/groups", url.PathEscape(network), url.PathEscape(namespace))
	if err := client.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ReauthorizeNamespace changes a namespace's access-control group.
func (client *Client) ReauthorizeNamespace(ctx context.Context, namespace string, groupID float64) (map[string]any, error) {
	result := make(map[string]any)
	path := "/edgekv/v1/auth/namespaces/" + url.PathEscape(namespace)
	if err := client.do(ctx, http.MethodPut, path, map[string]float64{"groupId": groupID}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetScheduledDelete returns a namespace's scheduled deletion state.
func (client *Client) GetScheduledDelete(ctx context.Context, network, namespace string) (map[string]any, error) {
	result := make(map[string]any)
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s/status/scheduled-delete", url.PathEscape(network), url.PathEscape(namespace))
	if err := client.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RescheduleNamespaceDelete changes a namespace's scheduled deletion time.
func (client *Client) RescheduleNamespaceDelete(ctx context.Context, network, namespace, scheduledDeleteTime string) (map[string]any, error) {
	result := make(map[string]any)
	body := map[string]string{"scheduledDeleteTime": scheduledDeleteTime}
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s/status/scheduled-delete", url.PathEscape(network), url.PathEscape(namespace))
	if err := client.do(ctx, http.MethodPut, path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CancelNamespaceDelete cancels a namespace's scheduled deletion.
func (client *Client) CancelNamespaceDelete(ctx context.Context, network, namespace string) (map[string]any, error) {
	result := make(map[string]any)
	path := fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s/status/scheduled-delete", url.PathEscape(network), url.PathEscape(namespace))
	if err := client.do(ctx, http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

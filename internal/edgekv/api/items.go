package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListItems returns item identifiers in an EdgeKV data group.
func (client *Client) ListItems(ctx context.Context, network, namespace, groupID string, maxItems *float64, sandboxID string) ([]string, error) {
	var result []string
	path := itemGroupPath(network, namespace, groupID)
	query := url.Values{}
	if maxItems != nil {
		query.Set("maxItems", strconv.FormatFloat(*maxItems, 'f', -1, 64))
	}
	if sandboxID != "" {
		query.Set("sandboxId", sandboxID)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := client.do(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetItem returns the raw value of an EdgeKV item.
func (client *Client) GetItem(ctx context.Context, network, namespace, groupID, itemID, sandboxID, accept string) ([]byte, error) {
	path := itemPath(network, namespace, groupID, itemID)
	if sandboxID != "" {
		path += "?sandboxId=" + url.QueryEscape(sandboxID)
	}
	if accept == "" {
		accept = "application/json, text/plain"
	}
	return client.doRaw(ctx, http.MethodGet, path, nil, "", accept)
}

// WriteItem creates or replaces an EdgeKV item.
func (client *Client) WriteItem(ctx context.Context, network, namespace, groupID, itemID string, value any, contentType, sandboxID string) (any, error) {
	path := itemPath(network, namespace, groupID, itemID)
	if sandboxID != "" {
		path += "?sandboxId=" + url.QueryEscape(sandboxID)
	}
	body, resolvedContentType, err := itemRequestBody(value, contentType)
	if err != nil {
		return nil, err
	}
	data, err := client.doRaw(ctx, http.MethodPut, path, bytes.NewReader(body), resolvedContentType, "application/json")
	if err != nil {
		return nil, err
	}
	return decodeItemMutationResult(data), nil
}

// DeleteItem deletes an EdgeKV item.
func (client *Client) DeleteItem(ctx context.Context, network, namespace, groupID, itemID, sandboxID string) (any, error) {
	path := itemPath(network, namespace, groupID, itemID)
	if sandboxID != "" {
		path += "?sandboxId=" + url.QueryEscape(sandboxID)
	}
	data, err := client.doRaw(ctx, http.MethodDelete, path, nil, "", "application/json")
	if err != nil {
		return nil, err
	}
	return decodeItemMutationResult(data), nil
}

func itemRequestBody(value any, contentType string) ([]byte, string, error) {
	resolvedContentType := contentType
	if resolvedContentType == "" {
		if _, isString := value.(string); isString {
			resolvedContentType = "text/plain"
		} else {
			resolvedContentType = "application/json"
		}
	}
	if resolvedContentType == "text/plain" {
		text, ok := value.(string)
		if !ok {
			return nil, "", fmt.Errorf("value must be a string when Content-Type is text/plain")
		}
		return []byte(text), resolvedContentType, nil
	}
	if resolvedContentType == "application/json" {
		if text, ok := value.(string); ok {
			return []byte(text), resolvedContentType, nil
		}
		data, err := json.Marshal(value)
		if err != nil {
			return nil, "", fmt.Errorf("marshal JSON item value: %w", err)
		}
		return data, resolvedContentType, nil
	}
	return nil, "", fmt.Errorf("unsupported Content-Type %q", resolvedContentType)
}

func decodeItemMutationResult(data []byte) any {
	if len(data) == 0 {
		return map[string]any{}
	}
	var result any
	if err := json.Unmarshal(data, &result); err == nil {
		return result
	}
	return string(data)
}

func itemGroupPath(network, namespace, groupID string) string {
	return fmt.Sprintf("/edgekv/v1/networks/%s/namespaces/%s/groups/%s", url.PathEscape(network), url.PathEscape(namespace), url.PathEscape(groupID))
}

func itemPath(network, namespace, groupID, itemID string) string {
	return itemGroupPath(network, namespace, groupID) + "/items/" + url.PathEscape(itemID)
}

package devserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHandleRequest_AcceptsFullInvocation(t *testing.T) {
	body := []byte(`{
		"edgeWorkerId": 1234,
		"eventHandler": "onClientResponse",
		"requestId": "1234",
		"resourceTier": 200,
		"settings": {
			"ignoreLimits": false,
			"subworker": { "isSubworker": false, "count": 15 },
			"profile": true
		},
		"request": {
			"method": "GET",
			"url": "https://www.akamai.com/test.html?asdf=1&asdf=2",
			"headers": [ { "name": "name", "value": "value" } ],
			"body": "Thisisabody",
			"features": { "userLocation": {} },
			"userVariables": [ { "name": "PMUSER_TEST", "value": "1234" } ]
		},
		"response": {
			"headers": [ { "name": "name", "value": "value" } ],
			"status": 404
		}
	}`)

	if _, err := handleRequest(body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleRequest_AcceptsMinimalRequiredFields(t *testing.T) {
	body := []byte(`{"edgeWorkerId":1234,"eventHandler":"onClientRequest","resourceTier":200}`)
	if _, err := handleRequest(body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleRequest_RejectsMissingRequiredFields(t *testing.T) {
	if _, err := handleRequest([]byte(`{}`)); err == nil {
		t.Fatal("expected error for missing edgeWorkerId, eventHandler, and resourceTier")
	}
}

func TestHandleRequest_GeneratesRequestIDWhenAbsent(t *testing.T) {
	body := []byte(`{"edgeWorkerId":1234,"eventHandler":"onClientRequest","resourceTier":200}`)
	result, err := handleRequest(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded struct {
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(decoded.RequestID) != 8 {
		t.Fatalf("expected an 8-character generated requestId, got %q", decoded.RequestID)
	}
}

func TestHandleRequest_RejectsInvalidJSON(t *testing.T) {
	if _, err := handleRequest([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleRequest_RejectsUnknownTopLevelProperty(t *testing.T) {
	_, err := handleRequest([]byte(`{"unknown":true}`))
	if err == nil {
		t.Fatal("expected error for unknown property")
	}
	if !strings.Contains(err.Error(), "validate invocation JSON") {
		t.Fatalf("expected validation error, got: %v", err)
	}
}

func TestHandleRequest_RejectsInvalidEventHandler(t *testing.T) {
	if _, err := handleRequest([]byte(`{"eventHandler":"onFooBar"}`)); err == nil {
		t.Fatal("expected error for invalid eventHandler")
	}
}

func TestHandleRequest_RejectsMalformedHeaders(t *testing.T) {
	body := []byte(`{"request":{"headers":[{"name":"X"}]}}`)
	if _, err := handleRequest(body); err == nil {
		t.Fatal("expected error for header missing value")
	}
}

func TestHandleRequest_RejectsResponseStatusOutOfRange(t *testing.T) {
	if _, err := handleRequest([]byte(`{"response":{"status":100}}`)); err == nil {
		t.Fatal("expected error for out-of-range status")
	}
}

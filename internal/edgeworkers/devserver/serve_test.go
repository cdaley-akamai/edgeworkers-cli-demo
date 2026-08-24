package devserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRunRequest_EchoesRequestBody(t *testing.T) {
	logger := newTestLogger()
	body := `{"edgeWorkerId":1234,"eventHandler":"onClientRequest","request":{"method":"GET","url":"/"},"requestId":"abcd1234","resourceTier":200}`
	req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handleRunRequest(rec, req, logger)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != body {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestHandleRunRequest_RejectsSchemaViolation(t *testing.T) {
	logger := newTestLogger()
	req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(`{"a":1}`))
	rec := httptest.NewRecorder()

	handleRunRequest(rec, req, logger)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestHandleRunRequest_RejectsNonPost(t *testing.T) {
	logger := newTestLogger()
	req := httptest.NewRequest(http.MethodGet, "/run", nil)
	rec := httptest.NewRecorder()

	handleRunRequest(rec, req, logger)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestHandleRunRequest_RejectsInvalidJSON(t *testing.T) {
	logger := newTestLogger()
	req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(`not json`))
	rec := httptest.NewRecorder()

	handleRunRequest(rec, req, logger)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestDebugMux_RespondsNotImplemented(t *testing.T) {
	logger := newTestLogger()
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rec := httptest.NewRecorder()

	newDebugMux(logger).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

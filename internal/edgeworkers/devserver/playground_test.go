package devserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlaygroundPage_ServesEmbeddedHTML(t *testing.T) {
	logger := newTestLogger()
	mux := http.NewServeMux()
	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) { handleRunRequest(w, r, logger) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(renderedPlaygroundPage())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if !bytes.Equal(rec.Body.Bytes(), renderedPlaygroundPage()) {
		t.Fatal("response body did not match rendered playground page")
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(invocationSchemaPlaceholder)) {
		t.Fatal("response body still contains the unresolved schema placeholder")
	}
	if !bytes.Contains(rec.Body.Bytes(), invocationSchemaFile) {
		t.Fatal("response body does not contain the embedded invocation schema")
	}
}

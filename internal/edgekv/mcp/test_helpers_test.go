package mcp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type edgeKVTestTransport struct {
	server *httptest.Server
}

func (transport *edgeKVTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	cloned := request.Clone(request.Context())
	cloned.URL.Scheme = "http"
	cloned.URL.Host = transport.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(cloned)
}

func newEdgeKVTestConfig(t *testing.T, server *httptest.Server) Config {
	t.Helper()
	edgercPath := filepath.Join(t.TempDir(), ".edgerc")
	require.NoError(t, os.WriteFile(edgercPath, []byte("[default]\nhost = akamai.test\nclient_token = client-token\nclient_secret = client-secret\naccess_token = access-token\n"), 0o600))
	return Config{
		EdgeRc:     edgercPath,
		Section:    "default",
		Timeout:    time.Second,
		HTTPClient: &http.Client{Transport: &edgeKVTestTransport{server: server}},
	}
}
